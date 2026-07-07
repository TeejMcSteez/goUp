import { useState } from "react";
import StatusMessage from "./StatusMessage";
import type {
  StatusMessage as StatusMsg,
  GlobalBackoffPanelProps,
} from "../../types";

const inputClass =
  "px-4 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.15)] placeholder:text-muted placeholder:opacity-50";

const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

export default function GlobalBackoffPanel({
  backoffPeriod,
  onRefresh,
}: GlobalBackoffPanelProps) {
  const [period, setPeriod] = useState(backoffPeriod ?? "");
  const [status, setStatus] = useState<StatusMsg | null>(null);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    const res = await fetch("/api/config/backoff", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ backoff_period: period }),
    });
    if (res.ok) {
      setStatus({ text: "Global backoff saved.", error: false });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleClear = async () => {
    const res = await fetch("/api/config/backoff", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ backoff_period: "" }),
    });
    if (res.ok) {
      setStatus({ text: "Global backoff cleared.", error: false });
      setPeriod("");
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  return (
    <div className="flex flex-col gap-4">
      <StatusMessage message={status?.text} isError={status?.error} />
      <form
        className="flex flex-col gap-4 p-4 bg-elevated border border-border rounded-lg"
        onSubmit={handleSave}
      >
        <p className="text-[0.8rem] text-muted m-0">
          Default backoff applied to any notification trigger that doesn't
          set its own backoff period. Leave blank to disable backoff by
          default.
        </p>
        <div className="grid grid-cols-[repeat(auto-fit,minmax(220px,1fr))] gap-4">
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Backoff Period
            <input
              className={inputClass}
              value={period}
              onChange={(e) => setPeriod(e.target.value)}
              placeholder="5m"
            />
          </label>
        </div>
        <div className="flex gap-2 flex-wrap">
          <button
            type="submit"
            className={`${btnBase} border-primary text-primary hover:bg-primary/10`}
          >
            Save
          </button>
          <button
            type="button"
            className={`${btnBase} border-error text-error hover:bg-error/10`}
            onClick={handleClear}
          >
            Clear
          </button>
        </div>
      </form>
    </div>
  );
}
