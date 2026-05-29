import { useCallback } from "react";
import usePolling from "./usePolling";
import type { ResponseTimeEntry, UptimeChartData } from "../types";
import { POLL_RATE } from "../constants";

// Parses Go duration strings (e.g. "12ms", "1.5s", "1m2.3s") to milliseconds.
function parseDurationMs(d: string): number {
  if (!d) return 0;
  let total = 0;
  const re = /(\d+(?:\.\d+)?)(ns|µs|μs|ms|s|m|h)/g;
  let match;
  while ((match = re.exec(d)) !== null) {
    const val = parseFloat(match[1]);
    switch (match[2]) {
      case "ns":
        total += val / 1_000_000;
        break;
      case "µs":
      case "μs":
        total += val / 1_000;
        break;
      case "ms":
        total += val;
        break;
      case "s":
        total += val * 1_000;
        break;
      case "m":
        total += val * 60_000;
        break;
      case "h":
        total += val * 3_600_000;
        break;
    }
  }
  return total;
}

export default function useResponseTimeData() {
  const fetchData = useCallback(async (): Promise<UptimeChartData | null> => {
    const res = await fetch("/api/rt");
    if (!res.ok)
      throw new Error(`Error fetching response times: ${res.statusText}`);

    const json: unknown = await res.json();
    if (!Array.isArray(json) || json.length === 0) return null;

    const items = json as ResponseTimeEntry[];

    // Average response time per service name.
    const totals = new Map<string, { sum: number; count: number }>();
    for (const item of items) {
      const ms = parseDurationMs(item.response_time);
      if (ms <= 0) continue;
      const name = item.service_data.name;
      const existing = totals.get(name) ?? { sum: 0, count: 0 };
      totals.set(name, { sum: existing.sum + ms, count: existing.count + 1 });
    }

    const labels = [...totals.keys()];
    const data = labels.map((name) => {
      const { sum, count } = totals.get(name)!;
      return Math.round(sum / count);
    });

    return {
      labels,
      datasets: [
        {
          label: "Avg Response Time (ms)",
          data,
          backgroundColor: ["rgba(167, 139, 250, 0.5)"],
          borderColor: ["rgba(167, 139, 250, 1)"],
          borderWidth: 1,
        },
      ],
    };
  }, []);

  return usePolling(fetchData, POLL_RATE);
}
