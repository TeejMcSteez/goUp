import useLiveQuery from "./useLiveQuery";
import { LIVE_QUERY_KEY, LIVE_STALE_TIME } from "./liveConnection";
import { liveConnection } from "../queryClient";
import type { LiveData, TlsStatus } from "../types";

const selectTls = (data: LiveData): TlsStatus[] => data.tls_status ?? [];

export default function useTlsData() {
  return useLiveQuery(LIVE_QUERY_KEY, liveConnection.fetchLive, {
    select: selectTls,
    staleTime: LIVE_STALE_TIME,
  });
}
