import { useState, useEffect, useCallback } from "react";

interface PollingResult<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}

export default function usePolling<T>(
  fetchFunction: () => Promise<T>,
  interval = 5000,
): PollingResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const result = await fetchFunction();
      setData(result);
      setError(null);
    } catch (err) {
      setError((err as Error).message);
      console.error("Error fetching data:", err);
    } finally {
      setLoading(false);
    }
  }, [fetchFunction]);

  useEffect(() => {
    const initialId = setTimeout(fetchData, 0);
    const intervalId = setInterval(fetchData, interval);
    return () => {
      clearTimeout(initialId);
      clearInterval(intervalId);
    };
  }, [fetchData, interval]);

  return { data, loading, error, refetch: fetchData };
}
