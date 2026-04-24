import { useState, useEffect, useCallback } from "react";
import ConfirmModal from "./config/ConfirmModal";
import type { DbSizeResponse } from "../types";

function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return "0 Bytes";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
}

export default function DatabaseInfo() {
  const [dbSize, setDbSize] = useState<string | null>(null);
  const [isPersistent, setIsPersistent] = useState<boolean | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [confirmClear, setConfirmClear] = useState(false);

  const fetchDbSize = useCallback(async () => {
    try {
      const res = await fetch("/api/db/size");
      if (!res.ok) {
        throw new Error(`Server error: ${res.status}`);
      }
      const data: DbSizeResponse = await res.json();

      let sizeInBytes: number | null = null;
      const sizeValue = data.size ?? data.size_string;

      if (typeof sizeValue === "number") {
        sizeInBytes = sizeValue;
      } else if (typeof sizeValue === "string") {
        const parsed = parseInt(sizeValue, 10);
        if (!isNaN(parsed)) {
          sizeInBytes = parsed;
        }
      }

      if (sizeInBytes !== null) {
        setDbSize(formatBytes(sizeInBytes));
      } else {
        console.error("Could not determine database size from API response:", data);
        setDbSize("N/A");
      }
    } catch (err) {
      setError((err as Error).message);
      console.error("Error fetching database size:", err);
    }

    try {
      const res = await fetch("/api/db/persist");
      if (!res.ok) throw new Error(`Server Error: ${res.status}`);
      const data: unknown = await res.json();
      setIsPersistent(!!data);
    } catch (e) {
      console.error("Error fetching persistence data:", e);
    }
  }, []);

  useEffect(() => {
    const initialId = setTimeout(fetchDbSize, 0);
    const intervalId = setInterval(fetchDbSize, 5000);
    return () => {
      clearTimeout(initialId);
      clearInterval(intervalId);
    };
  }, [fetchDbSize]);

  const handleClearDatabase = async () => {
    try {
      const res = await fetch("/api/db/clear");
      if (!res.ok) {
        throw new Error(`Server error: ${res.status}`);
      }
      fetchDbSize();
    } catch (err) {
      setError((err as Error).message);
      console.error("Error clearing database:", err);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center p-4 my-4 gap-4">
      {error ? (
        <p>Error loading database size: {error}</p>
      ) : dbSize === null ? (
        <p>Loading database size...</p>
      ) : (
        <p>
          Current database size: {dbSize}
          <br />
          {isPersistent === null
            ? "Checking persistence..."
            : isPersistent
              ? "Database file is persistent"
              : "Database file is not persistent"}
        </p>
      )}
      <button
        onClick={() => setConfirmClear(true)}
        disabled={dbSize === null}
        className="hover:border-secondary"
      >
        Clear Database Memory
      </button>

      {confirmClear && (
        <ConfirmModal
          message="Clear all database memory? This action is irreversible."
          onConfirm={() => { handleClearDatabase(); setConfirmClear(false); }}
          onCancel={() => setConfirmClear(false)}
        />
      )}
    </div>
  );
}
