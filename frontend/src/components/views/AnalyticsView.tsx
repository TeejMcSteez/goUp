import Island from "../layout/Island";
import UptimeChart from "../UptimeChart";
import ErrorLogViewer from "../ErrorLogViewer";

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
