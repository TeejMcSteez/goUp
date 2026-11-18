import useLiveQuery from "./useLiveQuery";
import { LIVE_QUERY_KEY, LIVE_STALE_TIME } from "./liveConnection";
import { liveConnection } from "../queryClient";
import type { LiveData, Service } from "../types";

const selectDowned = (data: LiveData): Service[] => data.downed_services ?? [];

export default function useQuickServices() {
  return useLiveQuery(LIVE_QUERY_KEY, liveConnection.fetchLive, {
    select: selectDowned,
    staleTime: LIVE_STALE_TIME,
  });
}
