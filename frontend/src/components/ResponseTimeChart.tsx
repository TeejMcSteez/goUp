import { useMemo, useState } from "react";
import UplotReact from "uplot-react";
import uPlot from "uplot";
import "uplot/dist/uPlot.min.css";
import useResponseTimeSeries, {
  type ResponseTimeRange,
} from "../hooks/useResponseTimeSeries";
import useElementWidth from "../hooks/useElementWidth";

const CHART_HEIGHT = 300;

const RANGE_OPTIONS: { label: string; value: ResponseTimeRange }[] = [
  { label: "All Time", value: "" },
  { label: "1 Hour", value: "1hr" },
  { label: "12 Hours", value: "12hr" },
  { label: "Day", value: "day" },
  { label: "Week", value: "week" },
  { label: "Month", value: "month" },
  { label: "Year", value: "year" },
];

const PALETTE = [
  "rgb(167, 139, 250)",
  "rgb(96, 165, 250)",
  "rgb(52, 211, 153)",
  "rgb(251, 191, 36)",
  "rgb(248, 113, 113)",
  "rgb(244, 114, 182)",
  "rgb(45, 212, 191)",
  "rgb(163, 230, 53)",
];

// Server sends nanosecond precision; show sub-ms detail only when it matters.
// Switch to seconds past 1000ms so wide labels like "15000ms" don't overflow
// the fixed-width y axis and get their leading digits clipped.
const formatMs = (v: number) => {
  if (v >= 1000) {
    const s = v / 1000;
    return `${s < 10 ? +s.toFixed(2) : s < 100 ? +s.toFixed(1) : Math.round(s)}s`;
  }
  return `${v < 10 ? v.toFixed(2) : v < 100 ? v.toFixed(1) : Math.round(v)}ms`;
};

function buildOptions(names: string[], width: number): uPlot.Options {
  return {
    width,
    height: CHART_HEIGHT,
    scales: { x: { time: true } },
    axes: [
      { stroke: "#9ca3af" },
      {
        stroke: "#9ca3af",
        size: 60,
        values: (_u, splits) => splits.map(formatMs),
      },
    ],
    series: [
      { label: "Time", value: "{YYYY}-{MM}-{DD} {h}:{mm}:{ss}{aa}" },
      ...names.map((name, i) => ({
        label: name,
        stroke: PALETTE[i % PALETTE.length],
        width: 1.5,
        spanGaps: true,
        value: (_u: uPlot, v: number | null) => (v == null ? "" : formatMs(v)),
      })),
    ],
  };
}

export default function ResponseTimeChart() {
  const [range, setRange] = useState<ResponseTimeRange>("");
  const { data: series, loading, error } = useResponseTimeSeries(range);
  const [containerRef, width] = useElementWidth<HTMLDivElement>();

  // Keyed on names (kept by reference across appends) so new rows only
  // update the data instead of rebuilding the whole chart.
  const names = series?.names;
  const options = useMemo(
    () => (names ? buildOptions(names, width) : null),
    [names, width],
  );

  const rangeSelector = (
    <select
      className="border rounded px-2 py-1 text-sm"
      value={range}
      onChange={(e) => setRange(e.target.value as ResponseTimeRange)}
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
    body = <p className="text-error text-sm">Could not load chart: {error}</p>;
  } else if (loading && !series) {
    body = <p className="text-muted text-sm">Loading chart...</p>;
  } else if (!series || !options) {
    body = (
      <p className="text-muted text-sm">
        {range ? "No response time data in this range" : "No response time data yet"}
      </p>
    );
  } else if (width > 0) {
    body = <UplotReact options={options} data={series.data} />;
  }

  return (
    <div className="w-full flex flex-col items-center justify-center p-4 min-h-75">
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
