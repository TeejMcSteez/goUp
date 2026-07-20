import usePolling from "./usePolling";
import type { ErrorItem } from "../types";
import { usePollRate } from "../context/PollRateContext";

export default function useErrorData(limit = 100, sortOrder = "desc") {
  const { pollRate } = usePollRate();

  const fetchErrors = async (): Promise<ErrorItem[]> => {
    const res = await fetch(`/api/errors?limit=${limit}&sort=${sortOrder}`);
    if (!res.ok) {
      throw new Error(`HTTP error! status: ${res.status}`);
    }
    const data: unknown = await res.json();
    return (data as ErrorItem[]) || [];
  };

  return usePolling(["errors", limit, sortOrder], fetchErrors, pollRate);
}
