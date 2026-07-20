import type { AppConfig } from "../types";
import { useQuery } from "@tanstack/react-query";

interface ConfigDataHook {
  config: AppConfig | null;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
}

async function fetchConfig(): Promise<AppConfig> {
  const res = await fetch("/api/config");
  if (!res.ok) throw new Error("Failed to fetch config");
  return res.json();
}

export function useConfigData(): ConfigDataHook {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["config"],
    queryFn: fetchConfig,
  });

  return {
    config: data ?? null,
    loading: isLoading,
    error: error ? (error as Error).message : null,
    refresh: async () => {
      await refetch();
    },
  };
}
