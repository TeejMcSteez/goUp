import CollapsibleIsland from "../layout/CollapsibleIsland";
import UptimeChart from "../UptimeChart";
import ErrorLogViewer from "../ErrorLogViewer";
import ResponseTimeChart from "../ResponseTimeChart";

export default function AnalyticsView() {
  return (
    <div className="w-full max-w-full overflow-x-hidden">
      <CollapsibleIsland title="Error Log">
        <ErrorLogViewer />
      </CollapsibleIsland>
      <CollapsibleIsland title="Failure Graph">
        <UptimeChart />
      </CollapsibleIsland>
      <CollapsibleIsland title="Response Times">
        <ResponseTimeChart />
      </CollapsibleIsland>
    </div>
  );
}
