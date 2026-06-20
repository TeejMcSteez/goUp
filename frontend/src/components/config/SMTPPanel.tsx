import { useState } from "react";
import StatusMessage from "./StatusMessage";
import type { StatusMessage as StatusMsg, SMTPPanelProps } from "../../types";

const inputClass =
  "px-4 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.15)] placeholder:text-muted placeholder:opacity-50";

const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

interface SMTPFormData {
  Email: string;
  App_Password: string;
  SMTPServer: string;
}

export default function SMTPPanel({ smtp, onRefresh }: SMTPPanelProps) {
  const [form, setForm] = useState<SMTPFormData>({
    Email: smtp?.Email ?? "",
    App_Password: smtp?.App_Password ?? "",
    SMTPServer: smtp?.SMTPServer ?? "",
  });
  const [status, setStatus] = useState<StatusMsg | null>(null);

  const set =
    (field: keyof SMTPFormData) => (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSave = async (e: React.SubmitEvent) => {
    e.preventDefault();
    const res = await fetch("/api/config/smtp", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        Email: form.Email || null,
        App_Password: form.App_Password || null,
        SMTPServer: form.SMTPServer || null,
      }),
    });
    if (res.ok) {
      setStatus({ text: "SMTP config saved.", error: false });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleClear = async () => {
    const res = await fetch("/api/config/smtp", { method: "DELETE" });
    if (res.ok) {
      setStatus({ text: "SMTP config cleared.", error: false });
      setForm({ Email: "", App_Password: "", SMTPServer: "" });
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
            SMTP Server
            <input
              className={inputClass}
              value={form.SMTPServer}
              onChange={set("SMTPServer")}
              placeholder="smtp.example.com:587"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Email
            <input
              className={inputClass}
              type="email"
              value={form.Email}
              onChange={set("Email")}
              placeholder="you@example.com"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            App Password
            <input
              className={inputClass}
              type="password"
              value={form.App_Password}
              onChange={set("App_Password")}
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
