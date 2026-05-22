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
import useResponseTimeData from "../hooks/useResponseTimeData";

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
);

const OPTIONS = {
  indexAxis: "y" as const,
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    x: {
      beginAtZero: true,
      ticks: {
        callback: (value: number | string) => `${value}ms`,
      },
      title: {
        display: true,
        text: "Average Response Time (ms)",
        color: "#9ca3af",
      },
    },
    y: {
      ticks: { color: "#e5e7eb" },
    },
  },
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (ctx: { parsed: { x: number | null } }) =>
          ctx.parsed.x != null ? ` ${ctx.parsed.x}ms` : "",
      },
    },
  },
} as const;

export default function ResponseTimeChart() {
  const { data: chartData, loading, error } = useResponseTimeData();

  if (error) {
    return (
      <div className="w-full flex items-center justify-center p-4 min-h-75">
        <p className="text-error text-sm">Could not load chart: {error}</p>
      </div>
    );
  }

  if (loading && !chartData) {
    return (
      <div className="w-full flex items-center justify-center p-4 min-h-75">
        <p className="text-muted text-sm">Loading chart...</p>
      </div>
    );
  }

  if (!chartData) {
    return (
      <div className="w-full flex items-center justify-center p-4 min-h-75">
        <p className="text-muted text-sm">No response time data yet</p>
      </div>
    );
  }

  const height = Math.max(300, chartData.labels.length * 48);

  return (
    <div className="w-full p-4" style={{ height }}>
      <Bar data={chartData} options={OPTIONS} />
    </div>
  );
}
