import { useState, useEffect } from "react";

export function useConfigData() {
  const [config, setConfig] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchConfig = async () => {
    try {
      setLoading(true);
      const res = await fetch("/api/config");
      if (!res.ok) throw new Error("Failed to fetch config");
      const data = await res.json();
      setConfig(data);
      setError(null);
    } catch (err) {
      setError(err.message);
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
