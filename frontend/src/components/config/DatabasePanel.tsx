import { useState, useEffect, useRef } from "react";
import StatusMessage from "./StatusMessage";
import {
  DatabaseSizePayload,
  type StatusMessage as StatusMsg,
} from "../../types";

const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

export default function DatabasePanel() {
  const [persists, setPersists] = useState<boolean | null>(null);
  const [status, setStatus] = useState<StatusMsg | null>(null);
  const [size, setSize] = useState<DatabaseSizePayload | null>(null);
  const siRef = useRef<HTMLInputElement | null>(null);
  useEffect(() => {
    const controller = new AbortController();

    fetch("/api/db/persist", { signal: controller.signal })
      .then((r) => r.json())
      .then((data: boolean) => setPersists(data))
      .catch((err: Error) => {
        if (err.name !== "AbortError") {
          setStatus({
            text: "Failed to load persistence state.",
            error: true,
          });
        }
      });

    fetch("/api/config/size", { signal: controller.signal })
      .then((r) => r.json())
      .then((data: DatabaseSizePayload) => setSize(data))
      .catch((err: Error) => {
        if (err.name !== "AbortError") {
          setStatus({
            text: "Failed to load database max size",
            error: true,
          });
        }
      });
    return () => controller.abort();
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

  const handleUpdate = async () => {
    if (siRef.current != null) {
      const newSize: DatabaseSizePayload = { db_max_size: siRef.current.value };
      const res = await fetch("/api/config/size", {
        method: "POST",
        body: JSON.stringify(newSize),
      });
      if (res.ok) {
        setStatus({ text: "Updated size.", error: false });
      }
    }
  };

  return (
    <div className="flex flex-col gap-4">
      <StatusMessage message={status?.text} isError={status?.error} />
      <div className="flex flex-col md:flex-row md:justify-between gap-4">
        <div className="flex gap-2 flex-wrap items-center">
          <span className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Persist database:{" "}
            <strong>{persists === null ? "…" : persists ? "On" : "Off"}</strong>
          </span>
          <button
            className={`${btnBase} border-primary text-primary hover:bg-primary/10`}
            onClick={handleToggle}
            disabled={persists === null}
          >
            Toggle
          </button>
        </div>
        <div className="flex flex-col items-center gap-2 md:justify-end">
          <span className="text-[0.85rem] font-medium text-muted">
            Database Max Size
          </span>
          <div className="flex flex-row">
            <input
              ref={siRef}
              className="flex"
              placeholder={size?.db_max_size}
            />
            <button
              className={`${btnBase} border-primary text-primary hover:bg-primary/10`}
              onClick={handleUpdate}
            >
              Update
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
