import { useCallback } from "react";
import usePolling from "./usePolling";
import type { ErrorItem } from "../types";
import { POLL_RATE } from "../constants";

export default function useErrorData(limit = 100, sortOrder = "desc") {
  const fetchErrors = useCallback(async (): Promise<ErrorItem[]> => {
    const res = await fetch(`/api/errors?limit=${limit}&sort=${sortOrder}`);
    if (!res.ok) {
      throw new Error(`HTTP error! status: ${res.status}`);
    }
    const data: unknown = await res.json();
    return (data as ErrorItem[]) || [];
  }, [limit, sortOrder]);

  return usePolling(fetchErrors, POLL_RATE);
}
