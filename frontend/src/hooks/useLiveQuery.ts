import { useQuery, type QueryKey } from "@tanstack/react-query";

interface LiveQueryResult<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

interface LiveQueryOptions<TQueryFnData, TData> {
  // Pass a module-level function so react-query can memoise it.
  select?: (data: TQueryFnData) => TData;
  staleTime?: number;
}

// Wraps useQuery for data delivered over the /ws connection: either pushed
// into the cache by liveConnection, or re-requested over the socket when
// liveConnection invalidates the key after a scheduler broadcast.
export default function useLiveQuery<TQueryFnData, TData = TQueryFnData>(
  queryKey: QueryKey,
  fetchFunction: () => Promise<TQueryFnData>,
  options: LiveQueryOptions<TQueryFnData, TData> = {},
): LiveQueryResult<TData> {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey,
    queryFn: fetchFunction,
    ...options,
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
