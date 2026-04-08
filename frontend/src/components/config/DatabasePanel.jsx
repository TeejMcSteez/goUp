import { useState, useEffect } from "react";
import StatusMessage from "./StatusMessage";

export default function DatabasePanel() {
  const [persists, setPersists] = useState(null);
  const [status, setStatus] = useState(null);

  useEffect(() => {
    fetch("/api/db/persist")
      .then((r) => r.json())
      .then(setPersists)
      .catch(() => setStatus({ text: "Failed to load persistence state.", error: true }));
  }, []);

  const handleToggle = async () => {
    const res = await fetch("/api/db/persist", { method: "POST" });
    if (res.ok) {
      setPersists((prev) => !prev);
      setStatus({ text: "Database persistence updated.", error: false });
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  return (
    <div className="config-panel">
      <StatusMessage message={status?.text} isError={status?.error} />
      <div className="config-form-actions">
        <span className="config-label">
          Persist database: <strong>{persists === null ? "…" : persists ? "On" : "Off"}</strong>
        </span>
        <button className="config-btn config-btn--primary" onClick={handleToggle} disabled={persists === null}>
          Toggle
        </button>
      </div>
    </div>
  );
}
