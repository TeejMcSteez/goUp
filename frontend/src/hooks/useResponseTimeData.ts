import { useMemo } from "react";
import type { UptimeChartData } from "../types";
import useResponseTimeSeries, {
  type ResponseTimeSeries,
} from "./useResponseTimeSeries";

// Average response time per service, computed from the cached series so it
// rides the same snapshot + append fetches instead of hitting /api/rt again.
function toChartData(series: ResponseTimeSeries): UptimeChartData | null {
  const [, ...cols] = series.data as unknown as [
    unknown,
    ...(number | null)[][],
  ];

  const labels: string[] = [];
  const data: number[] = [];
  cols.forEach((col, i) => {
    let sum = 0;
    let count = 0;
    for (const ms of col) {
      if (ms == null || ms <= 0) continue;
      sum += ms;
      count++;
    }
    if (count === 0) return;
    labels.push(series.names[i]);
    data.push(Math.round(sum / count));
  });

  if (labels.length === 0) return null;

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
  const { data: series, ...rest } = useResponseTimeSeries();
  const data = useMemo(() => (series ? toChartData(series) : null), [series]);
  return { ...rest, data };
}
