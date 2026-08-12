import usePolling from "./usePolling";
import type { TlsStatus } from "../types";
import { usePollRate } from "../context/PollRateContext";

async function fetchTlsData(): Promise<TlsStatus[]> {
  const res = await fetch("/api/tls");
  if (!res.ok) {
    throw new Error(`Server error: ${res.status}`);
  }
  const data: unknown = await res.json();
  return (data as TlsStatus[]) || [];
}

export default function useTlsData() {
  const { pollRate } = usePollRate();
  return usePolling(["tls"], fetchTlsData, pollRate);
}
