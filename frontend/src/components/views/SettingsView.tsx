import Island from "../layout/Island";
import Update from "../Update";
import DatabaseInfo from "../DatabaseInfo";
import ConfigEditor from "../ConfigEditor";
import DisplaySettings from "../DisplaySettings";

export default function SettingsView() {
  return (
    <div className="settings-view">
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
      </Island>
    </div>
  );
}
