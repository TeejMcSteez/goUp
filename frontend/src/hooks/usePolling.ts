import { useQuery, type QueryKey } from "@tanstack/react-query";

interface PollingResult<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export default function usePolling<T>(
  queryKey: QueryKey,
  fetchFunction: () => Promise<T>,
  interval = 5000,
): PollingResult<T> {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey,
    queryFn: fetchFunction,
    refetchInterval: interval,
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
