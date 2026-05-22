import Island from "../layout/Island";
import UptimeChart from "../UptimeChart";
import ErrorLogViewer from "../ErrorLogViewer";
import ResponseTimeChart from "../ResponseTimeChart";

export default function AnalyticsView() {
  return (
    <div className="w-full max-w-full overflow-x-hidden">
      <Island title="Error Log">
        <ErrorLogViewer />
      </Island>
      <Island title="Failure Graph">
        <UptimeChart />
      </Island>
      <Island title="Response Times">
        <ResponseTimeChart />
      </Island>
    </div>
  );
}
