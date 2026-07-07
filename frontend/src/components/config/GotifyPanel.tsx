import { useState } from "react";
import StatusMessage from "./StatusMessage";
import type { StatusMessage as StatusMsg, GotifyPanelProps } from "../../types";

const inputClass =
  "px-4 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.15)] placeholder:text-muted placeholder:opacity-50";

const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

interface GotifyFormData {
  Gotify_Server: string;
  Gotify_Token: string;
  Gotify_Application: string;
  Gotify_Title: string;
  Gotify_Priority: string;
  Backoff_Period: string;
}

export default function GotifyPanel({ gotify, onRefresh }: GotifyPanelProps) {
  const [form, setForm] = useState<GotifyFormData>({
    Gotify_Server: gotify?.Gotify_Server ?? "",
    Gotify_Token: gotify?.Gotify_Token ?? "",
    Gotify_Application: gotify?.Gotify_Application ?? "",
    Gotify_Title: gotify?.Gotify_Title ?? "",
    Gotify_Priority: gotify?.Gotify_Priority?.toString() ?? "",
    Backoff_Period: gotify?.Backoff_Period ?? "",
  });
  const [status, setStatus] = useState<StatusMsg | null>(null);

  const set =
    (field: keyof GotifyFormData) => (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSave = async (e: React.SubmitEvent) => {
    e.preventDefault();
    const priority = form.Gotify_Priority
      ? parseInt(form.Gotify_Priority, 10)
      : null;
    const res = await fetch("/api/config/gotify", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        Gotify_Server: form.Gotify_Server || null,
        Gotify_Token: form.Gotify_Token || null,
        Gotify_Application: form.Gotify_Application || null,
        Gotify_Title: form.Gotify_Title || null,
        Gotify_Priority: priority,
        Backoff_Period: form.Backoff_Period || null,
      }),
    });
    if (res.ok) {
      setStatus({ text: "Gotify config saved.", error: false });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleClear = async () => {
    const res = await fetch("/api/config/gotify", { method: "DELETE" });
    if (res.ok) {
      setStatus({ text: "Gotify config cleared.", error: false });
      setForm({
        Gotify_Server: "",
        Gotify_Token: "",
        Gotify_Application: "",
        Gotify_Title: "",
        Gotify_Priority: "",
        Backoff_Period: "",
      });
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
            Server URL
            <input
              className={inputClass}
              value={form.Gotify_Server}
              onChange={set("Gotify_Server")}
              placeholder="https://gotify.example.com"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            App Token
            <input
              className={inputClass}
              type="password"
              value={form.Gotify_Token}
              onChange={set("Gotify_Token")}
              placeholder="••••••••"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Application
            <input
              className={inputClass}
              value={form.Gotify_Application}
              onChange={set("Gotify_Application")}
              placeholder="GoUp"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Title
            <input
              className={inputClass}
              value={form.Gotify_Title}
              onChange={set("Gotify_Title")}
              placeholder="GoUp Alert"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Priority
            <input
              className={inputClass}
              type="number"
              min="0"
              max="10"
              value={form.Gotify_Priority}
              onChange={set("Gotify_Priority")}
              placeholder="5"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Backoff Period
            <input
              className={inputClass}
              value={form.Backoff_Period}
              onChange={set("Backoff_Period")}
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
