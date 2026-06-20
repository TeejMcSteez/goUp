import { useState } from "react";
import StatusMessage from "./StatusMessage";
import type { StatusMessage as StatusMsg, MQTTPanelProps } from "../../types";

const inputClass =
  "px-4 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.15)] placeholder:text-muted placeholder:opacity-50";

const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

interface MQTTFormData {
  Mqtt_broker: string;
  Mqtt_username: string;
  Mqtt_key: string;
}

export default function MQTTPanel({ mqtt, onRefresh }: MQTTPanelProps) {
  const [form, setForm] = useState<MQTTFormData>({
    Mqtt_broker: mqtt?.Mqtt_broker ?? "",
    Mqtt_username: mqtt?.Mqtt_username ?? "",
    Mqtt_key: mqtt?.Mqtt_key ?? "",
  });
  const [status, setStatus] = useState<StatusMsg | null>(null);

  const set =
    (field: keyof MQTTFormData) => (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    const res = await fetch("/api/config/mqtt", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        Mqtt_broker: form.Mqtt_broker || null,
        Mqtt_username: form.Mqtt_username || null,
        Mqtt_key: form.Mqtt_key || null,
      }),
    });
    if (res.ok) {
      setStatus({ text: "MQTT config saved.", error: false });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleClear = async () => {
    const res = await fetch("/api/config/mqtt", { method: "DELETE" });
    if (res.ok) {
      setStatus({ text: "MQTT config cleared.", error: false });
      setForm({ Mqtt_broker: "", Mqtt_username: "", Mqtt_key: "" });
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
        <div className="grid grid-cols-[repeat(auto-fit,minmax(220px,1fr))] gap-4">
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Broker URL
            <input
              className={inputClass}
              value={form.Mqtt_broker}
              onChange={set("Mqtt_broker")}
              placeholder="mqtt://broker.example.com"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Username
            <input
              className={inputClass}
              value={form.Mqtt_username}
              onChange={set("Mqtt_username")}
              placeholder="username"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Key / Password
            <input
              className={inputClass}
              type="password"
              value={form.Mqtt_key}
              onChange={set("Mqtt_key")}
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
