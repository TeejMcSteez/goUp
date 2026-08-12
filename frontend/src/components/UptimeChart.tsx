import { useState } from "react";
import { Bar } from "react-chartjs-2";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from "chart.js";
import useUptimeData, { type UptimeRange } from "../hooks/useUptimeData";

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
);

const RANGE_OPTIONS: { label: string; value: UptimeRange }[] = [
  { label: "All Time", value: "" },
  { label: "1 Hour", value: "1hr" },
  { label: "12 Hours", value: "12hr" },
  { label: "Day", value: "day" },
  { label: "Week", value: "week" },
  { label: "Month", value: "month" },
  { label: "Year", value: "year" },
];

export default function UptimeChart() {
  const [range, setRange] = useState<UptimeRange>("");
  const { data: chartData, loading, error } = useUptimeData(range);

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
  } else if (!chartData) {
    body = <p>No chart data available</p>;
  } else {
    body = (
      <Bar
        data={chartData}
        options={{
          scales: {
            y: {
              beginAtZero: true,
              ticks: {
                callback: function (value) {
                  return value + "%";
                },
              },
            },
          },
          plugins: {
            tooltip: {
              callbacks: {
                label: function (context) {
                  let label = context.dataset.label ?? "";
                  if (label) {
                    label += ": ";
                  }
                  if (context.parsed.y !== null) {
                    label += context.parsed.y + "%";
                  }
                  return label;
                },
              },
            },
          },
          maintainAspectRatio: false,
        }}
      />
    );
  }

  return (
    <div className="w-full flex flex-col items-center justify-center p-4 min-h-100">
      <div className="w-full flex justify-end mb-2">{rangeSelector}</div>
      <div className="w-full flex-1 flex items-center justify-center">
        {body}
      </div>
    </div>
  );
}
