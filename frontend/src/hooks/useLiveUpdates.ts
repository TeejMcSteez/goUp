import { useEffect } from "react";
import { useQueryClient, type QueryClient } from "@tanstack/react-query";

// Query keys backed by data the scheduler refreshes each fetch cycle.
// The server pushes over /ws after every cycle, so these refetch once per
// push instead of polling on a timer.
const LIVE_QUERY_KEYS = [
  ["services"],
  ["status"],
  ["tls"],
  ["errors"],
  ["uptime"],
  ["responseTime"],
];

const MIN_RECONNECT_MS = 1000;
const MAX_RECONNECT_MS = 30000;

export function refreshLiveQueries(queryClient: QueryClient) {
  for (const queryKey of LIVE_QUERY_KEYS) {
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

      socket.onmessage = () => {
        refreshLiveQueries(queryClient);
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
