import { useQuery } from "@tanstack/react-query";

interface ServerHealth {
  serverDown: boolean;
  networkOnline: boolean;
}

async function fetchHealth(): Promise<{ ok: boolean }> {
  const res = await fetch("/health");
  if (!res.ok) throw new Error("Failed to fetch server health");
  return res.json();
}

export function useServerHealth(intervalMs = 10000): ServerHealth {
  const { isError } = useQuery({
    queryKey: ["health"],
    queryFn: fetchHealth,
    refetchInterval: intervalMs,
    retry: false,
  });

  return {
    serverDown: isError,
    networkOnline: navigator.onLine,
  };
}
