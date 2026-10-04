import Island from "../layout/Island";
import Update from "../Update";
import ConfigEditor from "../ConfigEditor";
import DisplaySettings from "../DisplaySettings";

export default function SettingsView() {
  return (
    <div className="w-full max-w-full overflow-x-hidden">
      <Island title="Server Refresh Schedule">
        <Update />
      </Island>
      <Island title="Configuration">
        <ConfigEditor />
      </Island>
      <Island title="Display">
        <DisplaySettings />
      </Island>
    </div>
  );
}
