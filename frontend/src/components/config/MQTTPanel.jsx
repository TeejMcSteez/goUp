import { useState, useEffect } from "react";
import StatusMessage from "./StatusMessage";

export default function MQTTPanel({ mqtt, onRefresh }) {
  const [form, setForm] = useState({ Mqtt_broker: "", Mqtt_username: "", Mqtt_key: "" });
  const [status, setStatus] = useState(null);

  useEffect(() => {
    setForm({
      Mqtt_broker: mqtt?.Mqtt_broker ?? "",
      Mqtt_username: mqtt?.Mqtt_username ?? "",
      Mqtt_key: mqtt?.Mqtt_key ?? "",
    });
  }, [mqtt]);

  const set = (field) => (e) => setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSave = async (e) => {
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
    <div className="config-panel">
      <StatusMessage message={status?.text} isError={status?.error} />
      <form className="config-form" onSubmit={handleSave}>
        <div className="config-form-grid">
          <label className="config-label">
            Broker URL
            <input className="config-input" value={form.Mqtt_broker} onChange={set("Mqtt_broker")} placeholder="mqtt://broker.example.com" />
          </label>
          <label className="config-label">
            Username
            <input className="config-input" value={form.Mqtt_username} onChange={set("Mqtt_username")} placeholder="username" />
          </label>
          <label className="config-label">
            Key / Password
            <input className="config-input" type="password" value={form.Mqtt_key} onChange={set("Mqtt_key")} placeholder="••••••••" />
          </label>
        </div>
        <div className="config-form-actions">
          <button type="submit" className="config-btn config-btn--primary">Save</button>
          <button type="button" className="config-btn config-btn--danger" onClick={handleClear}>Clear</button>
        </div>
      </form>
    </div>
  );
}
