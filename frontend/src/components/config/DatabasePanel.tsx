import { useState, useEffect, useRef } from "react";
import StatusMessage from "./StatusMessage";
import {
  DatabaseSizePayload,
  DbSizeResponse,
  type StatusMessage as StatusMsg,
} from "../../types";
import ConfirmModal from "./ConfirmModal";

const btnPrimary =
  "px-4 py-2 rounded-lg border border-primary bg-surface text-primary text-sm cursor-pointer transition-all duration-200 hover:bg-primary/10 hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
}

export default function DatabasePanel() {
  const [persists, setPersists] = useState<boolean | null>(null);
  const [status, setStatus] = useState<StatusMsg | null>(null);
  const [size, setSize] = useState<DatabaseSizePayload | null>(null);
  const [dbSize, setDbSize] = useState<string | null>(null);
  const [confirmClear, setConfirmClear] = useState(false);
  const siRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    fetch("/api/db/persist", { signal: controller.signal })
      .then((r) => r.json())
      .then((data: boolean) => setPersists(data))
      .catch((err: Error) => {
        if (err.name !== "AbortError")
          setStatus({ text: "Failed to load persistence state.", error: true });
      });

    fetch("/api/config/size", { signal: controller.signal })
      .then((r) => r.json())
      .then((data: DatabaseSizePayload) => setSize(data))
      .catch((err: Error) => {
        if (err.name !== "AbortError")
          setStatus({ text: "Failed to load database max size", error: true });
      });

    async function loadDbSize() {
      try {
        const res = await fetch("/api/db/size", { signal: controller.signal });
        if (!res.ok) throw new Error(`Server error: ${res.status}`);
        const data: DbSizeResponse = await res.json();
        const sizeValue = data.size ?? data.size_string;
        let sizeInBytes: number | null = null;
        if (typeof sizeValue === "number") {
          sizeInBytes = sizeValue;
        } else if (typeof sizeValue === "string") {
          const parsed = parseInt(sizeValue, 10);
          if (!isNaN(parsed)) sizeInBytes = parsed;
        }
        setDbSize(sizeInBytes !== null ? formatBytes(sizeInBytes) : "N/A");
      } catch (err: unknown) {
        if (err instanceof Error && err.name !== "AbortError")
          setStatus({
            text: `Failed to fetch database size: ${err.message}`,
            error: true,
          });
      }
    }
    void loadDbSize();

    return () => controller.abort();
  }, []);

  const handleClearDatabase = async () => {
    try {
      const res = await fetch("/api/db/clear");
      if (!res.ok) throw new Error(`Server error: ${res.status}`);
      setStatus({ text: "Database cleared.", error: false });
      // Re-fetch database size on clear
      const res_size = await fetch("/api/db/size");
      if (!res_size.ok) throw new Error(`Server error: ${res_size.status}`);
      const data: DbSizeResponse = await res_size.json();
      const sizeValue = data.size ?? data.size_string;
      let sizeInBytes: number | null = null;
      if (typeof sizeValue === "number") {
        sizeInBytes = sizeValue;
      } else if (typeof sizeValue === "string") {
        const parsed = parseInt(sizeValue, 10);
        if (!isNaN(parsed)) sizeInBytes = parsed;
      }
      setDbSize(sizeInBytes !== null ? formatBytes(sizeInBytes) : "N/A");
    } catch (err) {
      setStatus({
        text: `Failed to clear database: ${(err as Error).message}`,
        error: true,
      });
    }
  };

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
    if (siRef.current === null) return;
    const newSize: DatabaseSizePayload = { db_max_size: siRef.current.value };
    const res = await fetch("/api/config/size", {
      method: "POST",
      body: JSON.stringify(newSize),
    });
    if (res.ok) setStatus({ text: "Updated size.", error: false });
  };

  return (
    <div className="flex flex-col gap-4">
      <StatusMessage message={status?.text} isError={status?.error} />
      <div className="flex flex-col md:flex-row md:justify-between gap-4">
        <div className="flex items-center gap-3">
          <span className="text-sm text-muted">
            Persist:{" "}
            <strong className="text-fg">
              {persists === null ? "…" : persists ? "On" : "Off"}
            </strong>
          </span>
          <button
            className={btnPrimary}
            onClick={handleToggle}
            disabled={persists === null}
          >
            Toggle
          </button>
        </div>

        <div className="flex flex-col items-center gap-3">
          <p className="text-sm text-muted m-0">
            {dbSize === null ? "Loading…" : `On disk: ${dbSize}`}
          </p>
          <button
            onClick={() => setConfirmClear(true)}
            disabled={dbSize === null}
            className="hover:border-secondary w-full sm:w-auto"
          >
            Clear Database
          </button>
          {confirmClear && (
            <ConfirmModal
              message="Clear all database memory? This action is irreversible."
              onConfirm={() => {
                void handleClearDatabase();
                setConfirmClear(false);
              }}
              onCancel={() => setConfirmClear(false)}
            />
          )}
        </div>

        <div className="flex items-center gap-2 md:justify-end">
          <span className="text-sm text-muted whitespace-nowrap">Max size</span>
          <input ref={siRef} placeholder={size?.db_max_size} />
          <button className={btnPrimary} onClick={handleUpdate}>
            Update
          </button>
        </div>
      </div>
    </div>
  );
}
