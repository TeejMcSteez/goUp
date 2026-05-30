import { useCallback } from "react";
import usePolling from "./usePolling";
import type { UptimeItem, UptimeChartData } from "../types";
import { usePollRate } from "../context/PollRateContext";

export default function useUptimeData() {
  const { pollRate } = usePollRate();
  const fetchUptimeData =
    useCallback(async (): Promise<UptimeChartData | null> => {
      const res = await fetch("/api/uptime");
      if (!res.ok) {
        throw new Error(`Error fetching data: ${res.statusText}`);
      }
      const json: unknown = await res.json();

      if (Array.isArray(json)) {
        const items = json as UptimeItem[];
        const labels = items.map((item) => item.name);
        const data = items.map((item) => item.average * 100);

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

  return usePolling(fetchUptimeData, pollRate);
}
