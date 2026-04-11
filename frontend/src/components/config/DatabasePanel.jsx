import { useState, useEffect } from "react";
import StatusMessage from "./StatusMessage";

const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

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
    <div className="flex flex-col gap-4">
      <StatusMessage message={status?.text} isError={status?.error} />
      <div className="flex gap-2 flex-wrap items-center">
        <span className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
          Persist database: <strong>{persists === null ? "…" : persists ? "On" : "Off"}</strong>
        </span>
        <button
          className={`${btnBase} border-primary text-primary hover:bg-primary/10`}
          onClick={handleToggle}
          disabled={persists === null}
        >
          Toggle
        </button>
      </div>
    </div>
  );
}
