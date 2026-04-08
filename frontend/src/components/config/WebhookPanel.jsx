import { useState, useEffect } from "react";
import StatusMessage from "./StatusMessage";

export default function WebhookPanel({ webhook, onRefresh }) {
  const [form, setForm] = useState({ Webhook_url: "", Webhook_key_string: "", Custom_message: "" });
  const [status, setStatus] = useState(null);

  useEffect(() => {
    setForm({
      Webhook_url: webhook?.Webhook_url ?? "",
      Webhook_key_string: webhook?.Webhook_key_string ?? "",
      Custom_message: webhook?.Custom_message ?? "",
    });
  }, [webhook]);

  const set = (field) => (e) => setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSave = async (e) => {
    e.preventDefault();
    const res = await fetch("/api/config/webhook", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        Webhook_url: form.Webhook_url || null,
        Webhook_key_string: form.Webhook_key_string || null,
        Custom_message: form.Custom_message || null,
      }),
    });
    if (res.ok) {
      setStatus({ text: "Webhook config saved.", error: false });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleClear = async () => {
    const res = await fetch("/api/config/webhook", { method: "DELETE" });
    if (res.ok) {
      setStatus({ text: "Webhook config cleared.", error: false });
      setForm({ Webhook_url: "", Webhook_key_string: "", Custom_message: "" });
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
            Webhook URL
            <input className="config-input" value={form.Webhook_url} onChange={set("Webhook_url")} placeholder="https://hooks.example.com/..." />
          </label>
          <label className="config-label">
            Authorization Header
            <input className="config-input" value={form.Webhook_key_string} onChange={set("Webhook_key_string")} placeholder="Bearer <token>" />
          </label>
          <label className="config-label">
            Custom Message
            <input className="config-input" value={form.Custom_message} onChange={set("Custom_message")} placeholder="A service is down!" />
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
