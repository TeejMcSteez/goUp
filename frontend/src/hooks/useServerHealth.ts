import { useQuery } from "@tanstack/react-query";

interface ServerHealth {
  serverDown: boolean;
  networkOnline: boolean;
}

async function fetchHealth(): Promise<ServerHealth> {
  const res = await fetch("/health");
  if (!res.ok) throw new Error("Failed to fetch server health");
  return res.json();
}

export function useServerHealth(intervalMs = 10000): ServerHealth {
  const { data } = useQuery({
    queryKey: ["health"],
    queryFn: fetchHealth,
    refetchInterval: intervalMs
  });

  return {
    serverDown: data?.networkOnline ?? false,
    networkOnline: navigator.onLine
  }
}
