import useLiveQuery from "./useLiveQuery";
import { liveConnection } from "../queryClient";
import type { ErrorItem } from "../types";

export default function useErrorData(limit = 100, sortOrder = "desc") {
  const fetchErrors = async (): Promise<ErrorItem[]> => {
    const data = await liveConnection.request<ErrorItem[] | null>({
      type: "errors",
      errParams: { limit, sortOrder },
    });
    return data ?? [];
  };

  return useLiveQuery(["errors", limit, sortOrder], fetchErrors);
}
