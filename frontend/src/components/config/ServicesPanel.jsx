import { useState } from "react";
import ServiceForm from "./ServiceForm";
import StatusMessage from "./StatusMessage";

const EMPTY_FORM = { Name: "", URL: "", API_URL: "", Valid_Responses: "", Retry_Requests: "" };

function svcToForm(svc) {
  return {
    Name: svc.Name ?? "",
    URL: svc.URL ?? "",
    API_URL: svc.API_URL ?? "",
    Valid_Responses: svc.Valid_Responses?.join(", ") ?? "",
    Retry_Requests: svc.Retry_Requests != null ? String(svc.Retry_Requests) : "",
  };
}

// mode: { type: "list" } | { type: "add" } | { type: "edit", key: string }
export default function ServicesPanel({ services, onRefresh }) {
  const [mode, setMode] = useState({ type: "list" });
  const [status, setStatus] = useState(null);

  const handleAdd = async (payload) => {
    const res = await fetch("/api/config/service", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (res.ok) {
      setStatus({ text: "Service added.", error: false });
      setMode({ type: "list" });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleUpdate = async (oldName, payload) => {
    const res = await fetch("/api/config/service", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ old_name: oldName, service: payload }),
    });
    if (res.ok) {
      setStatus({ text: "Service updated.", error: false });
      setMode({ type: "list" });
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
              {mode.type === "edit" && mode.key === key ? (
                <ServiceForm
                  key={key}
                  initial={svcToForm(svc)}
                  onSubmit={(payload) => handleUpdate(key, payload)}
                  onCancel={() => setMode({ type: "list" })}
                  submitLabel="Save"
                />
              ) : (
                <>
                  <div className="config-service-info">
                    <span className="config-service-name">{svc.Name || key}</span>
                    <span className="config-service-url">{svc.URL}</span>
                    {svc.Valid_Responses?.length > 0 && (
                      <span className="config-service-url">Valid: {svc.Valid_Responses.join(", ")}</span>
                    )}
                  </div>
                  <div className="config-form-actions">
                    <button className="config-btn" onClick={() => setMode({ type: "edit", key })}>Edit</button>
                    <button className="config-btn config-btn--danger" onClick={() => handleDelete(svc)}>Remove</button>
                  </div>
                </>
              )}
            </li>
          ))}
        </ul>
      )}

      {mode.type === "add" ? (
        <ServiceForm
          key="add"
          initial={EMPTY_FORM}
          onSubmit={handleAdd}
          onCancel={() => setMode({ type: "list" })}
          submitLabel="Add"
        />
      ) : mode.type === "list" && (
        <button className="config-btn config-btn--primary config-add-btn" onClick={() => setMode({ type: "add" })}>
          + Add Service
        </button>
      )}
    </div>
  );
}
