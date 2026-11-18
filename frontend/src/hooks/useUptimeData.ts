import useLiveQuery from "./useLiveQuery";
import { liveConnection } from "../queryClient";
import type { UptimeItem, UptimeChartData } from "../types";

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
  const items = await liveConnection.request<UptimeItem[] | null>({
    type: "uptime",
    uptimeParams: { range },
  });

  // empty means no service had usable data yet
  if (items && items.length > 0) {
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
  return useLiveQuery(["uptime", range], () => fetchUptimeData(range));
}
