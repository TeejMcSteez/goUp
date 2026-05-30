import Island from "../layout/Island";
import Update from "../Update";
import DatabaseInfo from "../DatabaseInfo";
import ConfigEditor from "../ConfigEditor";
import DisplaySettings from "../DisplaySettings";
import RefreshSettingsView from "../RefreshSettings";

export default function SettingsView() {
  return (
    <div className="w-full max-w-full overflow-x-hidden">
      <Island title="Update Schedule">
        <Update />
      </Island>
      <Island title="Configuration">
        <ConfigEditor />
      </Island>
      <Island title="Database Management">
        <DatabaseInfo />
      </Island>
      <Island title="Display">
        <DisplaySettings />
        <RefreshSettingsView />
      </Island>
    </div>
  );
}
