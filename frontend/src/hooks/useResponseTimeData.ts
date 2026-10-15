import useLiveQuery from "./useLiveQuery";
import type { ResponseTimeEntry, UptimeChartData } from "../types";
import { parseDurationMs } from "../lib/formatters";

async function fetchData(): Promise<UptimeChartData | null> {
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
}

export default function useResponseTimeData() {
  return useLiveQuery(["responseTime"], fetchData);
}
