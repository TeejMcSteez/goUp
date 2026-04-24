import { useState } from "react";
import type { ServiceConfig } from "../../types";

interface ServiceFormData {
  Name: string;
  URL: string;
  API_URL: string;
  Valid_Responses: string;
  Retry_Requests: string;
}

interface ServiceFormProps {
  initial: ServiceFormData;
  onSubmit: (payload: Partial<ServiceConfig>) => void;
  onCancel: () => void;
  submitLabel: string;
}

export default function ServiceForm({ initial, onSubmit, onCancel, submitLabel }: ServiceFormProps) {
  const [form, setForm] = useState<ServiceFormData>(initial);

  const set = (field: keyof ServiceFormData) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const validResponses = form.Valid_Responses
      ? form.Valid_Responses.split(",").map((s) => s.trim()).filter(Boolean)
      : null;
    onSubmit({
      Name: form.Name,
      URL: form.URL,
      ...(form.API_URL && { API_URL: form.API_URL }),
      ...(validResponses?.length && { Valid_Responses: validResponses.map(Number) }),
      ...(form.Retry_Requests && { Retry_Requests: parseInt(form.Retry_Requests) }),
    });
  };

  const inputClass =
    "px-4 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.15)] placeholder:text-muted placeholder:opacity-50";

  const btnBase =
    "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

  return (
    <form className="flex flex-col gap-4 p-4 bg-elevated border border-border rounded-lg" onSubmit={handleSubmit}>
      <div className="grid grid-cols-[repeat(auto-fit,minmax(220px,1fr))] gap-4">
        <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
          Name
          <input className={inputClass} value={form.Name} onChange={set("Name")} placeholder="My Service" required />
        </label>
        <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
          URL
          <input className={inputClass} value={form.URL} onChange={set("URL")} placeholder="https://example.com" required />
        </label>
        <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
          API URL (optional)
          <input className={inputClass} value={form.API_URL} onChange={set("API_URL")} placeholder="https://api.example.com" />
        </label>
        <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
          Valid Responses (optional)
          <input className={inputClass} value={form.Valid_Responses} onChange={set("Valid_Responses")} placeholder="200, 201, 204" />
        </label>
        <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
          Retries (optional)
          <input className={inputClass} type="number" min="0" value={form.Retry_Requests} onChange={set("Retry_Requests")} placeholder="3" />
        </label>
      </div>
      <div className="flex gap-2 flex-wrap">
        <button type="submit" className={`${btnBase} border-primary text-primary hover:bg-primary/10`}>{submitLabel}</button>
        <button type="button" className={btnBase} onClick={onCancel}>Cancel</button>
      </div>
    </form>
  );
}
