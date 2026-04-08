import { useState } from "react";

// Pass a `key` prop from the parent to reset form state when switching services.
export default function ServiceForm({ initial, onSubmit, onCancel, submitLabel }) {
  const [form, setForm] = useState(initial);

  const set = (field) => (e) => setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSubmit = (e) => {
    e.preventDefault();
    const validResponses = form.Valid_Responses
      ? form.Valid_Responses.split(",").map((s) => s.trim()).filter(Boolean)
      : null;
    onSubmit({
      Name: form.Name,
      URL: form.URL,
      ...(form.API_URL && { API_URL: form.API_URL }),
      ...(validResponses?.length && { Valid_Responses: validResponses }),
      ...(form.Retry_Requests && { Retry_Requests: parseInt(form.Retry_Requests) }),
    });
  };

  return (
    <form className="config-form" onSubmit={handleSubmit}>
      <div className="config-form-grid">
        <label className="config-label">
          Name
          <input className="config-input" value={form.Name} onChange={set("Name")} placeholder="My Service" required />
        </label>
        <label className="config-label">
          URL
          <input className="config-input" value={form.URL} onChange={set("URL")} placeholder="https://example.com" required />
        </label>
        <label className="config-label">
          API URL (optional)
          <input className="config-input" value={form.API_URL} onChange={set("API_URL")} placeholder="https://api.example.com" />
        </label>
        <label className="config-label">
          Valid Responses (optional)
          <input className="config-input" value={form.Valid_Responses} onChange={set("Valid_Responses")} placeholder="200, 201, 204" />
        </label>
        <label className="config-label">
          Retries (optional)
          <input className="config-input" type="number" min="0" value={form.Retry_Requests} onChange={set("Retry_Requests")} placeholder="3" />
        </label>
      </div>
      <div className="config-form-actions">
        <button type="submit" className="config-btn config-btn--primary">{submitLabel}</button>
        <button type="button" className="config-btn" onClick={onCancel}>Cancel</button>
      </div>
    </form>
  );
}
