import { useEffect } from "react";
import { useQueryClient, type QueryClient } from "@tanstack/react-query";
import type { LiveData } from "../types";

// Query keys whose data arrives in the /ws payload itself, so they're
// written straight into the cache on each push without refetching.
const PUSHED_QUERY_KEYS = [["services"], ["status"], ["tls"]];

// Query keys backed by DB aggregates the payload doesn't carry (and some are
// parameterised per view), so a push still invalidates them to refetch.
const REFETCHED_QUERY_KEYS = [["errors"], ["uptime"], ["responseTime"]];

const MIN_RECONNECT_MS = 1000;
const MAX_RECONNECT_MS = 30000;

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

export default function useLiveUpdates() {
  const queryClient = useQueryClient();

  useEffect(() => {
    let socket: WebSocket | null = null;
    let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
    let delay = MIN_RECONNECT_MS;
    let closed = false;

    function connect() {
      socket = new WebSocket(wsUrl());

      socket.onopen = () => {
        delay = MIN_RECONNECT_MS;
        // pick up anything missed while disconnected
        refreshLiveQueries(queryClient);
      };

      socket.onmessage = (event: MessageEvent<string>) => {
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

      socket.onclose = () => {
        if (closed) return;
        reconnectTimer = setTimeout(connect, delay);
        delay = Math.min(delay * 2, MAX_RECONNECT_MS);
      };
    }

    connect();

    return () => {
      closed = true;
      clearTimeout(reconnectTimer);
      socket?.close();
    };
  }, [queryClient]);
}
