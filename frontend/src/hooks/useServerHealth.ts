import { useState, useEffect } from "react";

interface ServerHealth {
  serverDown: boolean;
  networkOnline: boolean;
}

export function useServerHealth(intervalMs = 10000): ServerHealth {
  const [serverDown, setServerDown] = useState(false);
  const [networkOnline, setNetworkOnline] = useState(navigator.onLine);

  useEffect(() => {
    const check = async () => {
      setNetworkOnline(navigator.onLine);
      try {
        const res = await fetch("/api", { signal: AbortSignal.timeout(4000) });
        setServerDown(!res.ok);
      } catch {
        setServerDown(true);
      }
    };

    check();
    const id = setInterval(check, intervalMs);
    return () => clearInterval(id);
  }, [intervalMs]);

  return { serverDown, networkOnline };
}
