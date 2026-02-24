import { useCallback } from "react";
import usePolling from "./usePolling.js";

export default function useErrorData(limit = 100, sortOrder = "desc") {
  const fetchErrors = useCallback(async () => {
    const res = await fetch(`/api/errors?limit=${limit}&sort=${sortOrder}`);
    if (!res.ok) {
      throw new Error(`HTTP error! status: ${res.status}`);
    }
    const data = await res.json();
    return data || [];
  }, [limit, sortOrder]);

  return usePolling(fetchErrors, 5000);
}
