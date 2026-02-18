import { useCallback } from "react";
import usePolling from "./usePolling.js";

export default function useUptimeData() {
  const fetchUptimeData = useCallback(async () => {
    const res = await fetch("/api/uptime");
    if (!res.ok) {
      throw new Error(`Error fetching data: ${res.statusText}`);
    }
    const json = await res.json();

    if (Array.isArray(json)) {
      const labels = json.map((item) => item.name);
      const data = json.map((item) => item.average * 100);

      return {
        labels,
        datasets: [
          {
            label: "Total Failure Average",
            data,
            backgroundColor: ["rgba(255, 99, 132, 0.5)"],
            borderColor: ["rgba(255, 99, 132, 1)"],
            borderWidth: 1,
          },
        ],
      };
    }
    return null;
  }, []);

  return usePolling(fetchUptimeData, 5000);
}
