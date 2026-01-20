import Island from '../layout/Island.jsx';
import UptimeChart from '../UptimeChart.jsx';
import ErrorLogViewer from '../ErrorLogViewer.jsx';

export default function AnalyticsView() {
    return (
        <div className="analytics-view">
            <Island title="Failure Graph">
                <UptimeChart />
            </Island>
            <Island title="Error Log">
                <ErrorLogViewer />
            </Island>
        </div>
    );
}
