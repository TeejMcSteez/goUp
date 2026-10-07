import type { QueryClient } from "@tanstack/react-query";
import type { LiveData } from "../types";

// Query keys whose data arrives in the /ws payload itself, so they're
// written straight into the cache on each push without refetching.
const PUSHED_QUERY_KEYS = [["services"], ["status"], ["tls"]];

// Query keys backed by DB aggregates the payload doesn't carry (and some are
// parameterised per view), so a push still invalidates them to refetch.
const REFETCHED_QUERY_KEYS = [["errors"], ["uptime"], ["responseTime"]];

const MIN_RECONNECT_MS = 1000;
const MAX_RECONNECT_MS = 30000;

// "connecting" only covers the first attempt; once a socket has closed the
// status stays "closed" through retries so the down banner doesn't flicker.
export type LiveStatus = "connecting" | "open" | "closed";

// Refetches every live query, used where no payload is at hand (reconnect,
// config changes).
export function refreshLiveQueries(queryClient: QueryClient) {
  for (const queryKey of [...PUSHED_QUERY_KEYS, ...REFETCHED_QUERY_KEYS]) {
    void queryClient.invalidateQueries({ queryKey });
  }
}

function applyLiveData(queryClient: QueryClient, data: LiveData) {
  queryClient.setQueryData(["services"], data.services ?? []);
  queryClient.setQueryData(["status"], data.downed_services ?? []);
  queryClient.setQueryData(["tls"], data.tls_status ?? []);
  for (const queryKey of REFETCHED_QUERY_KEYS) {
    void queryClient.invalidateQueries({ queryKey });
  }
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

  function setStatus(next: LiveStatus) {
    if (status === next) return;
    status = next;
    listeners.forEach((notify) => notify());
  }

  function connect() {
    const ws = new WebSocket(wsUrl());
    socket = ws;

    // Events from a socket we've already replaced (e.g. StrictMode's
    // subscribe/unsubscribe/subscribe) arrive late and must be ignored.
    ws.onopen = () => {
      if (socket !== ws) return;
      delay = MIN_RECONNECT_MS;
      setStatus("open");
      // pick up anything missed while disconnected
      refreshLiveQueries(queryClient);
    };

    ws.onmessage = (event: MessageEvent<string>) => {
      if (socket !== ws) return;
      let data: LiveData;
      try {
        data = JSON.parse(event.data) as LiveData;
      } catch (err) {
        console.error("Malformed /ws payload, refetching instead:", err);
        refreshLiveQueries(queryClient);
        return;
      }
      applyLiveData(queryClient, data);
    };

    // Browsers follow error with close, but some runtimes (Node's undici)
    // only fire error on a refused connection. Whichever lands first
    // schedules the retry; the socket guard makes the second a no-op.
    ws.onerror = ws.onclose = () => {
      if (socket !== ws) return;
      socket = null;
      ws.close();
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
  };
}
