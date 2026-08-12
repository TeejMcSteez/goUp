import usePolling from "./usePolling";
import type { UptimeItem, UptimeChartData } from "../types";
import { usePollRate } from "../context/PollRateContext";

export type UptimeRange =
  | ""
  | "1hr"
  | "12hr"
  | "day"
  | "week"
  | "month"
  | "year";

async function fetchUptimeData(
  range: UptimeRange,
): Promise<UptimeChartData | null> {
  const res = await fetch(`/api/uptime?range=${range}`);
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
}

export default function useUptimeData(range: UptimeRange = "") {
  const { pollRate } = usePollRate();
  return usePolling(["uptime", range], () => fetchUptimeData(range), pollRate);
}
