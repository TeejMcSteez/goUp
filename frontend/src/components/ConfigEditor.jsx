import { useState, useEffect } from "react";
import { useConfigData } from "../hooks/useConfigData";

function StatusMessage({ message, isError }) {
  if (!message) return null;
  return (
    <p
      className={`config-status ${isError ? "config-status--error" : "config-status--success"}`}
    >
      {message}
    </p>
  );
}

function ServicesPanel({ services, onRefresh }) {
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({
    Name: "",
    URL: "",
    API_URL: "",
    Retry_Requests: "",
  });
  const [status, setStatus] = useState(null);

  const handleAdd = async (e) => {
    e.preventDefault();
    const payload = {
      Name: form.Name,
      URL: form.URL,
      ...(form.API_URL && { API_URL: form.API_URL }),
      ...(form.Retry_Requests && {
        Retry_Requests: parseInt(form.Retry_Requests),
      }),
    };
    const res = await fetch("/api/config/service", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (res.ok) {
      setStatus({ text: "Service added.", error: false });
      setForm({ Name: "", URL: "", API_URL: "", Retry_Requests: "" });
      setShowForm(false);
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleDelete = async (svc) => {
    const res = await fetch("/api/config/service", {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(svc),
    });
    if (res.ok) {
      setStatus({ text: "Service removed.", error: false });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const serviceList = Object.entries(services || {});

  return (
    <div className="config-panel">
      <StatusMessage message={status?.text} isError={status?.error} />
      {serviceList.length === 0 ? (
        <p className="config-empty">No services configured.</p>
      ) : (
        <ul className="config-service-list">
          {serviceList.map(([key, svc]) => (
            <li key={key} className="config-service-item">
              <div className="config-service-info">
                <span className="config-service-name">{svc.Name || key}</span>
                <span className="config-service-url">{svc.URL}</span>
              </div>
              <button
                className="config-btn config-btn--danger"
                onClick={() => handleDelete(svc)}
              >
                Remove
              </button>
            </li>
          ))}
        </ul>
      )}

      {showForm ? (
        <form className="config-form" onSubmit={handleAdd}>
          <div className="config-form-grid">
            <label className="config-label">
              Name
              <input
                className="config-input"
                value={form.Name}
                onChange={(e) =>
                  setForm((f) => ({ ...f, Name: e.target.value }))
                }
                placeholder="My Service"
                required
              />
            </label>
            <label className="config-label">
              URL
              <input
                className="config-input"
                value={form.URL}
                onChange={(e) =>
                  setForm((f) => ({ ...f, URL: e.target.value }))
                }
                placeholder="https://example.com"
                required
              />
            </label>
            <label className="config-label">
              API URL (optional)
              <input
                className="config-input"
                value={form.API_URL}
                onChange={(e) =>
                  setForm((f) => ({ ...f, API_URL: e.target.value }))
                }
                placeholder="https://api.example.com"
              />
            </label>
            <label className="config-label">
              Retries (optional)
              <input
                className="config-input"
                type="number"
                min="0"
                value={form.Retry_Requests}
                onChange={(e) =>
                  setForm((f) => ({ ...f, Retry_Requests: e.target.value }))
                }
                placeholder="3"
              />
            </label>
          </div>
          <div className="config-form-actions">
            <button type="submit" className="config-btn config-btn--primary">
              Add
            </button>
            <button
              type="button"
              className="config-btn"
              onClick={() => setShowForm(false)}
            >
              Cancel
            </button>
          </div>
        </form>
      ) : (
        <button
          className="config-btn config-btn--primary config-add-btn"
          onClick={() => setShowForm(true)}
        >
          + Add Service
        </button>
      )}
    </div>
  );
}

function MQTTPanel({ mqtt, onRefresh }) {
  const [form, setForm] = useState({
    Mqtt_broker: mqtt?.Mqtt_broker ?? "",
    Mqtt_username: mqtt?.Mqtt_username ?? "",
    Mqtt_key: mqtt?.Mqtt_key ?? "",
  });
  const [status, setStatus] = useState(null);

  useEffect(() => {
    setForm({
      Mqtt_broker: mqtt?.Mqtt_broker ?? "",
      Mqtt_username: mqtt?.Mqtt_username ?? "",
      Mqtt_key: mqtt?.Mqtt_key ?? "",
    });
  }, [mqtt]);

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
            <input
              className="config-input"
              value={form.Mqtt_broker}
              onChange={(e) =>
                setForm((f) => ({ ...f, Mqtt_broker: e.target.value }))
              }
              placeholder="mqtt://broker.example.com"
            />
          </label>
          <label className="config-label">
            Username
            <input
              className="config-input"
              value={form.Mqtt_username}
              onChange={(e) =>
                setForm((f) => ({ ...f, Mqtt_username: e.target.value }))
              }
              placeholder="username"
            />
          </label>
          <label className="config-label">
            Key / Password
            <input
              className="config-input"
              type="password"
              value={form.Mqtt_key}
              onChange={(e) =>
                setForm((f) => ({ ...f, Mqtt_key: e.target.value }))
              }
              placeholder="••••••••"
            />
          </label>
        </div>
        <div className="config-form-actions">
          <button type="submit" className="config-btn config-btn--primary">
            Save
          </button>
          <button
            type="button"
            className="config-btn config-btn--danger"
            onClick={handleClear}
          >
            Clear
          </button>
        </div>
      </form>
    </div>
  );
}

function WebhookPanel({ webhook, onRefresh }) {
  const [form, setForm] = useState({
    Webhook_url: webhook?.Webhook_url ?? "",
    Webhook_key_string: webhook?.Webhook_key_string ?? "",
    Custom_message: webhook?.Custom_message ?? "",
  });
  const [status, setStatus] = useState(null);

  useEffect(() => {
    setForm({
      Webhook_url: webhook?.Webhook_url ?? "",
      Webhook_key_string: webhook?.Webhook_key_string ?? "",
      Custom_message: webhook?.Custom_message ?? "",
    });
  }, [webhook]);

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
            <input
              className="config-input"
              value={form.Webhook_url}
              onChange={(e) =>
                setForm((f) => ({ ...f, Webhook_url: e.target.value }))
              }
              placeholder="https://hooks.example.com/..."
            />
          </label>
          <label className="config-label">
            Authorization Header
            <input
              className="config-input"
              value={form.Webhook_key_string}
              onChange={(e) =>
                setForm((f) => ({ ...f, Webhook_key_string: e.target.value }))
              }
              placeholder="Bearer <token>"
            />
          </label>
          <label className="config-label">
            Custom Message
            <input
              className="config-input"
              value={form.Custom_message}
              onChange={(e) =>
                setForm((f) => ({ ...f, Custom_message: e.target.value }))
              }
              placeholder="A service is down!"
            />
          </label>
        </div>
        <div className="config-form-actions">
          <button type="submit" className="config-btn config-btn--primary">
            Save
          </button>
          <button
            type="button"
            className="config-btn config-btn--danger"
            onClick={handleClear}
          >
            Clear
          </button>
        </div>
      </form>
    </div>
  );
}

function DatabasePanel() {
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
          Persist database:{" "}
          <strong>{persists === null ? "…" : persists ? "On" : "Off"}</strong>
        </span>
        <button className="config-btn config-btn--primary" onClick={handleToggle} disabled={persists === null}>
          Toggle
        </button>
      </div>
    </div>
  );
}

const SECTIONS = [
  { id: "services", label: "Services" },
  { id: "mqtt", label: "MQTT" },
  { id: "webhook", label: "Webhook" },
  { id: "database", label: "Database" },
];

export default function ConfigEditor() {
  const { config, loading, error, refresh } = useConfigData();
  const [activeSection, setActiveSection] = useState("services");

  if (loading) return <p className="config-loading">Loading configuration…</p>;
  if (error)
    return <p className="config-status config-status--error">{error}</p>;

  return (
    <div className="config-editor">
      <div className="config-section-tabs">
        {SECTIONS.map((s) => (
          <button
            key={s.id}
            className={`config-section-tab ${activeSection === s.id ? "active" : ""}`}
            onClick={() => setActiveSection(s.id)}
          >
            {s.label}
          </button>
        ))}
      </div>
      <div className="config-section-content">
        {activeSection === "services" && (
          <ServicesPanel services={config?.services} onRefresh={refresh} />
        )}
        {activeSection === "mqtt" && (
          <MQTTPanel mqtt={config?.mqtt} onRefresh={refresh} />
        )}
        {activeSection === "webhook" && (
          <WebhookPanel webhook={config?.webhook} onRefresh={refresh} />
        )}
        {activeSection === "database" && <DatabasePanel />}
      </div>
    </div>
  );
}
