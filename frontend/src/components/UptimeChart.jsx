import { useState, useEffect } from 'react';
import { Bar } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend
);

export default function UptimeChart() {
    const [chartData, setChartData] = useState(null);
    const [error, setError] = useState(null);

    useEffect(() => {
        const getUptimeAverages = async () => {
            try {
                const res = await fetch("/api/uptime");
                if (!res.ok) {
                    throw new Error(`Error fetching data: ${res.statusText}`);
                }
                const json = await res.json();

                if (Array.isArray(json)) {
                    const labels = json.map(item => item.name);
                    const data = json.map(item => item.average);

                    setChartData({
                        labels,
                        datasets: [{
                            label: 'Total Failure Average',
                            data,
                            backgroundColor: [
                                'rgba(255, 99, 132, 0.5)',
                                'rgba(54, 162, 235, 0.5)',
                                'rgba(255, 206, 86, 0.5)',
                                'rgba(75, 192, 192, 0.5)',
                                'rgba(153, 102, 255, 0.5)',
                                'rgba(255, 159, 64, 0.5)',
                            ],
                            borderColor: [
                                'rgba(255, 99, 132, 1)',
                                'rgba(54, 162, 235, 1)',
                                'rgba(255, 206, 86, 1)',
                                'rgba(75, 192, 192, 1)',
                                'rgba(153, 102, 255, 1)',
                                'rgba(255, 159, 64, 1)',
                            ],
                            borderWidth: 1
                        }]
                    });
                }
            } catch (err) {
                setError(err.message);
                console.error(err);
            }
        };

        getUptimeAverages();

        const intervalId = setInterval(getUptimeAverages, 5000);
        return () => clearInterval(intervalId);
    }, []);

    if (error) {
        return (
            <div id="uptimeChart">
                <p>Could not load chart: {error}</p>
            </div>
        );
    }
    
    if (!chartData) {
        return (
            <div id="uptimeChart">
                <p>Loading chart...</p>
            </div>
        );
    }

    return (
        <>
            <h1>Failure Graph</h1>
            <div id="uptimeChart">
                <Bar 
                    data={chartData}
                    options={{
                        scales: { y: { beginAtZero: true } },
                        maintainAspectRatio: false 
                    }}
                />
            </div>
        </>
    );
}
