import { useSyncExternalStore } from "react";
import { liveConnection } from "../queryClient";
import type { LiveStatus } from "./liveConnection";

// Keeps the /ws connection open while mounted and reports its status, which
// doubles as the server health check.
export default function useLiveUpdates(): LiveStatus {
  return useSyncExternalStore(liveConnection.subscribe, liveConnection.getSnapshot);
}
