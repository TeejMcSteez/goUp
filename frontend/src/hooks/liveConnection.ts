import type { QueryClient } from "@tanstack/react-query";
import type { LiveData } from "../types";

// Services/status/TLS arrive in the /ws payload itself (the server sends
// it on connect and on every fetch cycle), so it's written straight into
// this one cache entry; useServiceData, useQuickServices and useTlsData
// select the field each view needs.
export const LIVE_QUERY_KEY = ["live"] as const;

// Pushes keep the entry current, so it never goes stale on its own;
// refreshLiveQueries still forces a fresh copy.
export const LIVE_STALE_TIME = Infinity;

// Query keys backed by DB aggregates the payload doesn't carry (and some are
// parameterised per view), so a push invalidates them and they re-request
// their data over the same socket (see request()).
const REFETCHED_QUERY_KEYS = [["errors"], ["uptime"], ["responseTime"]];

const MIN_RECONNECT_MS = 1000;
const MAX_RECONNECT_MS = 30000;
// A request with no reply by then is rejected so its query errors out (and
// react-query retries) instead of hanging forever.
const REQUEST_TIMEOUT_MS = 15000;

// Requests the client can send over /ws, mirroring readMessage in
// server/handlers_websocket.go. "id" is added by request().
export type WsRequest =
  | { type: "live" }
  | { type: "errors"; errParams: { limit: number; sortOrder: string } }
  | { type: "uptime"; uptimeParams: { range: string } }
  // rtParams.type 0: full snapshot, 1: snapshot of [from, to] (unix seconds
  // as strings, `to` defaults to the server's now)
  | { type: "rt"; rtParams: { type: 0 } | { type: 1; from: string; to?: string } };

// JSON reply to a request; hub broadcasts (LiveData) carry no "id".
interface ReplyEnvelope {
  id: number;
  type: string;
  data: unknown;
  error?: string;
}

interface PendingRequest {
  payload: string;
  // false until it has gone out on an open socket
  sent: boolean;
  resolve: (value: unknown) => void;
  reject: (err: Error) => void;
  timer: ReturnType<typeof setTimeout>;
}

function isReply(msg: unknown): msg is ReplyEnvelope {
  return (
    typeof msg === "object" &&
    msg !== null &&
    typeof (msg as { id?: unknown }).id === "number"
  );
}

// "connecting" only covers the first attempt; once a socket has closed the
// status stays "closed" through retries so the down banner doesn't flicker.
export type LiveStatus = "connecting" | "open" | "closed";

function refreshAggregates(queryClient: QueryClient) {
  for (const queryKey of REFETCHED_QUERY_KEYS) {
    void queryClient.invalidateQueries({ queryKey });
  }
}

// Refetches every live query, used where no payload is at hand (config
// changes, malformed pushes).
export function refreshLiveQueries(queryClient: QueryClient) {
  void queryClient.invalidateQueries({ queryKey: LIVE_QUERY_KEY });
  refreshAggregates(queryClient);
}

function wsUrl(): string {
  const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${proto}//${window.location.host}/ws`;
}

// A /ws connection living outside React, shaped for useSyncExternalStore:
// the socket opens when the first subscriber arrives and closes when the
// last one leaves, so every component shares a single connection.
export function createLiveConnection(queryClient: QueryClient) {
  const listeners = new Set<() => void>();
  let socket: WebSocket | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
  let delay = MIN_RECONNECT_MS;
  let status: LiveStatus = "connecting";
  // u32 on the wire; wraps long before any id could still be pending
  let nextId = 1;
  const pending = new Map<number, PendingRequest>();
  // fetchLive calls made while disconnected, resolved by the next payload
  let liveWaiters: Array<(data: LiveData) => void> = [];

  function applyLiveData(data: LiveData) {
    queryClient.setQueryData(LIVE_QUERY_KEY, data);
    const waiters = liveWaiters;
    liveWaiters = [];
    waiters.forEach((resolve) => resolve(data));
    refreshAggregates(queryClient);
  }

  function setStatus(next: LiveStatus) {
    if (status === next) return;
    status = next;
    listeners.forEach((notify) => notify());
  }

  function settle(id: number, err: Error | null, value?: unknown) {
    const req = pending.get(id);
    if (!req) return; // timed out, or a reply to a dropped request
    pending.delete(id);
    clearTimeout(req.timer);
    if (err) req.reject(err);
    else req.resolve(value);
  }

  // Requests already sent on a socket that's going away will never get a
  // reply; unsent ones stay queued for the next connection.
  function failSent(reason: string) {
    for (const [id, req] of pending) {
      if (req.sent) settle(id, new Error(reason));
    }
  }

  function flush() {
    if (socket?.readyState !== WebSocket.OPEN) return;
    for (const req of pending.values()) {
      if (req.sent) continue;
      socket.send(req.payload);
      req.sent = true;
    }
  }

  function handleBinary(buf: ArrayBuffer) {
    // response time frame prefixed with the u32 LE request id
    if (buf.byteLength < 4) {
      console.error("Binary /ws message too short for a request id");
      return;
    }
    const id = new DataView(buf).getUint32(0, true);
    settle(id, null, buf.slice(4));
  }

  function handleText(text: string) {
    let msg: unknown;
    try {
      msg = JSON.parse(text);
    } catch (err) {
      console.error("Malformed /ws payload, refetching instead:", err);
      refreshLiveQueries(queryClient);
      return;
    }
    if (isReply(msg)) {
      settle(msg.id, msg.error ? new Error(msg.error) : null, msg.data);
      return;
    }
    applyLiveData(msg as LiveData);
  }

  function connect() {
    const ws = new WebSocket(wsUrl());
    ws.binaryType = "arraybuffer";
    socket = ws;

    // Events from a socket we've already replaced (e.g. StrictMode's
    // subscribe/unsubscribe/subscribe) arrive late and must be ignored.
    ws.onopen = () => {
      if (socket !== ws) return;
      delay = MIN_RECONNECT_MS;
      setStatus("open");
      flush();
      // pick up anything missed while disconnected; services/status/TLS
      // need no request since the server seeds them on connect
      refreshAggregates(queryClient);
    };

    ws.onmessage = (event: MessageEvent<string | ArrayBuffer>) => {
      if (socket !== ws) return;
      if (event.data instanceof ArrayBuffer) handleBinary(event.data);
      else handleText(event.data);
    };

    // Browsers follow error with close, but some runtimes (Node's undici)
    // only fire error on a refused connection. Whichever lands first
    // schedules the retry; the socket guard makes the second a no-op.
    ws.onerror = ws.onclose = () => {
      if (socket !== ws) return;
      socket = null;
      ws.close();
      failSent("Live connection closed");
      setStatus("closed");
      reconnectTimer = setTimeout(connect, delay);
      delay = Math.min(delay * 2, MAX_RECONNECT_MS);
    };
  }

  function disconnect() {
    clearTimeout(reconnectTimer);
    const ws = socket;
    socket = null;
    ws?.close();
    failSent("Live connection closed");
  }

  // Sends a request over the shared socket and resolves with its reply:
  // the envelope's data for JSON replies, the frame bytes for "rt". Queued
  // until the socket is open if it isn't yet.
  function request<T>(req: WsRequest): Promise<T> {
    const id = nextId;
    nextId = nextId >= 0xffffffff ? 1 : nextId + 1;
    return new Promise<T>((resolve, reject) => {
      pending.set(id, {
        payload: JSON.stringify({ ...req, id }),
        sent: false,
        resolve: resolve as (value: unknown) => void,
        reject,
        timer: setTimeout(
          () => settle(id, new Error(`Live ${req.type} request timed out`)),
          REQUEST_TIMEOUT_MS,
        ),
      });
      flush();
    });
  }

  // queryFn for LIVE_QUERY_KEY. While connecting it waits for the seed the
  // server sends on connect rather than requesting a duplicate; once open
  // (an explicit refetch) it asks for a fresh copy.
  function fetchLive(): Promise<LiveData> {
    if (socket?.readyState === WebSocket.OPEN)
      return request<LiveData>({ type: "live" });
    return new Promise<LiveData>((resolve, reject) => {
      const timer = setTimeout(() => {
        liveWaiters = liveWaiters.filter((w) => w !== waiter);
        reject(new Error("Timed out waiting for live data"));
      }, REQUEST_TIMEOUT_MS);
      const waiter = (data: LiveData) => {
        clearTimeout(timer);
        resolve(data);
      };
      liveWaiters.push(waiter);
    });
  }

  return {
    subscribe: (listener: () => void) => {
      listeners.add(listener);
      if (listeners.size === 1) connect();
      return () => {
        listeners.delete(listener);
        if (listeners.size === 0) disconnect();
      };
    },
    getSnapshot: (): LiveStatus => status,
    request,
    fetchLive,
  };
}
