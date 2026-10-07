import { useMemo, useState } from "react";
import UplotReact from "uplot-react";
import uPlot, { type AlignedData } from "uplot";
import "uplot/dist/uPlot.min.css";
import useUptimeData, { type UptimeRange } from "../hooks/useUptimeData";
import useElementWidth from "../hooks/useElementWidth";
import type { UptimeChartData } from "../types";

const CHART_HEIGHT = 400;

const RANGE_OPTIONS: { label: string; value: UptimeRange }[] = [
  { label: "All Time", value: "" },
  { label: "1 Hour", value: "1hr" },
  { label: "12 Hours", value: "12hr" },
  { label: "Day", value: "day" },
  { label: "Week", value: "week" },
  { label: "Month", value: "month" },
  { label: "Year", value: "year" },
];

const bars = uPlot.paths.bars!({ size: [0.6, 100] });

function buildOptions(chartData: UptimeChartData, width: number): uPlot.Options {
  const { labels, datasets } = chartData;
  // x values are label indices; map them back to service names for display
  const labelAt = (i: number | null) => (i == null ? "" : (labels[i] ?? ""));

  return {
    width,
    height: CHART_HEIGHT,
    scales: {
      x: { time: false, range: () => [-0.5, labels.length - 0.5] },
      y: { range: (_u, _min, max) => [0, max > 0 ? max * 1.1 : 1] },
    },
    axes: [
      {
        stroke: "#9ca3af",
        splits: () => labels.map((_, i) => i),
        values: (_u, splits) => splits.map(labelAt),
        grid: { show: false },
      },
      {
        stroke: "#9ca3af",
        values: (_u, splits) => splits.map((v) => `${v}%`),
      },
    ],
    series: [
      { label: "Service", value: (_u, v) => labelAt(v) },
      ...datasets.map((ds) => ({
        label: ds.label,
        value: (_u: uPlot, v: number | null) => (v == null ? "" : `${v}%`),
        paths: bars,
        points: { show: false },
        stroke: ds.borderColor[0],
        fill: ds.backgroundColor[0],
        width: ds.borderWidth,
      })),
    ],
  };
}

export default function UptimeChart() {
  const [range, setRange] = useState<UptimeRange>("");
  const { data: chartData, loading, error } = useUptimeData(range);

  const [containerRef, width] = useElementWidth<HTMLDivElement>();

  const options = useMemo(
    () => (chartData ? buildOptions(chartData, width) : null),
    [chartData, width],
  );

  const alignedData = useMemo<AlignedData | null>(
    () =>
      chartData
        ? [
            chartData.labels.map((_, i) => i),
            ...chartData.datasets.map((ds) => ds.data),
          ]
        : null,
    [chartData],
  );

  const rangeSelector = (
    <select
      className="border rounded px-2 py-1 text-sm"
      value={range}
      onChange={(e) => setRange(e.target.value as UptimeRange)}
    >
      {RANGE_OPTIONS.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );

  let body;
  if (error) {
    body = <p>Could not load chart: {error}</p>;
  } else if (loading && !chartData) {
    body = <p>Loading chart...</p>;
  } else if (!options || !alignedData) {
    body = <p>No chart data available</p>;
  } else if (width > 0) {
    body = <UplotReact options={options} data={alignedData} />;
  }

  return (
    <div className="w-full flex flex-col items-center justify-center p-4 min-h-100">
      <div className="w-full flex justify-end mb-2">{rangeSelector}</div>
      <div
        ref={containerRef}
        className="w-full flex-1 flex items-center justify-center"
      >
        {body}
      </div>
    </div>
  );
}
