import { useState, useEffect } from "react";
import type { AppConfig } from "../types";

interface ConfigDataHook {
  config: AppConfig | null;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
}

export function useConfigData(): ConfigDataHook {
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchConfig = async () => {
    try {
      setLoading(true);
      const res = await fetch("/api/config");
      if (!res.ok) throw new Error("Failed to fetch config");
      const data: AppConfig = await res.json();
      setConfig(data);
      setError(null);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const id = setTimeout(fetchConfig, 0);
    return () => clearTimeout(id);
  }, []);

  return { config, loading, error, refresh: fetchConfig };
}
