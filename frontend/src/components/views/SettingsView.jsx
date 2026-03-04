import Island from "../layout/Island.jsx";
import Update from "../Update.jsx";
import DatabaseInfo from "../DatabaseInfo.jsx";
import ConfigEditor from "../ConfigEditor.jsx";
import DisplaySettings from "../DisplaySettings.jsx";

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
