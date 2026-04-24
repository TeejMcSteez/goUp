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
import useUptimeData from "../hooks/useUptimeData";

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend);

export default function UptimeChart() {
  const { data: chartData, loading, error } = useUptimeData();

  if (error) {
    return (
      <div className="w-full flex items-center justify-center p-4 min-h-[400px]">
        <p>Could not load chart: {error}</p>
      </div>
    );
  }

  if (loading && !chartData) {
    return (
      <div className="w-full flex items-center justify-center p-4 min-h-[400px]">
        <p>Loading chart...</p>
      </div>
    );
  }

  if (!chartData) {
    return (
      <div className="w-full flex items-center justify-center p-4 min-h-[400px]">
        <p>No chart data available</p>
      </div>
    );
  }

  return (
    <div className="w-full flex items-center justify-center p-4 min-h-[400px]">
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
    </div>
  );
}
