import { useState, type ReactNode } from "react";
import StatusMessage from "./StatusMessage";
import type { StatusMessage as StatusMsg } from "../../types";

export const inputClass =
  "px-4 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.15)] placeholder:text-muted placeholder:opacity-50";

export const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

export interface FieldDef<T> {
  key: keyof T & string;
  label: string;
  type?: "text" | "password" | "email" | "number";
  placeholder?: string;
  min?: string;
  max?: string;
}

interface FormPanelProps<T extends Record<string, string>> {
  endpoint: string;
  entityLabel: string;
  fields: FieldDef<T>[];
  initial: T;
  empty: T;
  description?: ReactNode;
  onRefresh: () => void;
  serialize?: (form: T) => Record<string, unknown>;
  clear?: () => Promise<Response>;
}

export default function FormPanel<T extends Record<string, string>>({
  endpoint,
  entityLabel,
  fields,
  initial,
  empty,
  description,
  onRefresh,
  serialize,
  clear,
}: FormPanelProps<T>) {
  const [form, setForm] = useState<T>(initial);
  const [status, setStatus] = useState<StatusMsg | null>(null);

  const set =
    (field: keyof T) => (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [field]: e.target.value }));

  const buildBody = (f: T): Record<string, unknown> =>
    serialize
      ? serialize(f)
      : Object.fromEntries(fields.map((fd) => [fd.key, f[fd.key] || null]));

  const handleSave = async (e: React.SubmitEvent) => {
    e.preventDefault();
    const res = await fetch(endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(buildBody(form)),
    });
    if (res.ok) {
      setStatus({ text: `${entityLabel} config saved.`, error: false });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleClear = async () => {
    const res = clear ? await clear() : await fetch(endpoint, { method: "DELETE" });
    if (res.ok) {
      setStatus({ text: `${entityLabel} config cleared.`, error: false });
      setForm(empty);
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
        {description}
        <div className="grid grid-cols-[repeat(auto-fit,minmax(220px,1fr))] gap-4">
          {fields.map((f) => (
            <label
              key={f.key}
              className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted"
            >
              {f.label}
              <input
                className={inputClass}
                type={f.type ?? "text"}
                min={f.min}
                max={f.max}
                value={form[f.key]}
                onChange={set(f.key)}
                placeholder={f.placeholder}
              />
            </label>
          ))}
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
