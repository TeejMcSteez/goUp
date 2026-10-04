import useLiveQuery from "./useLiveQuery";
import type { ErrorItem } from "../types";

export default function useErrorData(limit = 100, sortOrder = "desc") {
  const fetchErrors = async (): Promise<ErrorItem[]> => {
    const res = await fetch(`/api/errors?limit=${limit}&sort=${sortOrder}`);
    if (!res.ok) {
      throw new Error(`HTTP error! status: ${res.status}`);
    }
    const data: unknown = await res.json();
    return (data as ErrorItem[]) || [];
  };

  return useLiveQuery(["errors", limit, sortOrder], fetchErrors);
}
