import { useQuery, type QueryKey } from "@tanstack/react-query";

interface LiveQueryResult<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

// Fetches once on mount, then refetches whenever liveConnection invalidates
// the key after a websocket push from the scheduler.
export default function useLiveQuery<T>(
  queryKey: QueryKey,
  fetchFunction: () => Promise<T>,
): LiveQueryResult<T> {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey,
    queryFn: fetchFunction,
  });

  return {
    data: data ?? null,
    loading: isLoading,
    error: error ? (error as Error).message : null,
    refetch: async () => {
      await refetch();
    },
  };
}
