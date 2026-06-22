import { useState } from "react";
import StatusMessage from "./StatusMessage";
import type { StatusMessage as StatusMsg, HAPanelProps } from "../../types";

const inputClass =
  "px-4 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.15)] placeholder:text-muted placeholder:opacity-50";

const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

interface HAFormData {
  HA_URL: string;
  HA_Token: string;
}

export default function HAPanel({ ha, onRefresh }: HAPanelProps) {
  const [form, setForm] = useState<HAFormData>({
    HA_URL: ha?.HA_URL ?? "",
    HA_Token: ha?.HA_Token ?? "",
  });
  const [status, setStatus] = useState<StatusMsg | null>(null);

  const set =
    (field: keyof HAFormData) => (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSave = async (e: React.SubmitEvent) => {
    e.preventDefault();
    const res = await fetch("/api/config/ha", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        HA_URL: form.HA_URL || null,
        HA_Token: form.HA_Token || null,
      }),
    });
    if (res.ok) {
      setStatus({ text: "Home Assistant config saved.", error: false });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleClear = async () => {
    const res = await fetch("/api/config/ha", { method: "DELETE" });
    if (res.ok) {
      setStatus({ text: "Home Assistant config cleared.", error: false });
      setForm({ HA_URL: "", HA_Token: "" });
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
          Fires a <code className="text-fg">goup_alert</code> event on the HA
          event bus when a service goes down. Use an automation with{" "}
          <code className="text-fg">trigger: event_type: goup_alert</code> to
          act on it.
        </p>
        <div className="grid grid-cols-[repeat(auto-fit,minmax(220px,1fr))] gap-4">
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Instance URL
            <input
              className={inputClass}
              value={form.HA_URL}
              onChange={set("HA_URL")}
              placeholder="http://homeassistant.local:8123"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Long-Lived Access Token
            <input
              className={inputClass}
              type="password"
              value={form.HA_Token}
              onChange={set("HA_Token")}
              placeholder="••••••••"
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
