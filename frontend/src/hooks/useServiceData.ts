import useLiveQuery from "./useLiveQuery";
import { LIVE_QUERY_KEY, LIVE_STALE_TIME } from "./liveConnection";
import { liveConnection } from "../queryClient";
import type { LiveData, Service } from "../types";

const selectServices = (data: LiveData): Service[] => data.services ?? [];

export default function useServiceData() {
  return useLiveQuery(LIVE_QUERY_KEY, liveConnection.fetchLive, {
    select: selectServices,
    staleTime: LIVE_STALE_TIME,
  });
}
