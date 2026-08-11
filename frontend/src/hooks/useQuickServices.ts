import usePolling from "./usePolling";
import type { Service } from "../types";
import { usePollRate } from "../context/PollRateContext";

async function fetchQuickServices(): Promise<Service[]> {
  const res = await fetch("/api/status");
  if (!res.ok) {
    throw new Error(`Server error: ${res.status}`);
  }
  const data: unknown = await res.json();
  if (data === null) {
    return [];
  }
  if (Array.isArray(data)) {
    return data as Service[];
  }
  console.error(`Expected array from /api/status, got: ${JSON.stringify(data)}`);
  return [];
}

export default function useQuickServices() {
  const { pollRate } = usePollRate();
  return usePolling(["status"], fetchQuickServices, pollRate);
}
