import { useState } from "react";
import StatusMessage from "./StatusMessage";
import type { StatusMessage as StatusMsg, TelegramPanelProps } from "../../types";

const inputClass =
  "px-4 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.15)] placeholder:text-muted placeholder:opacity-50";

const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

interface TelegramFormData {
  Telegram_Token: string;
  Telegram_Channel_Id: string;
  Backoff_Period: string;
}

export default function TelegramPanel({ telegram, onRefresh }: TelegramPanelProps) {
  const [form, setForm] = useState<TelegramFormData>({
    Telegram_Token: telegram?.Telegram_Token ?? "",
    Telegram_Channel_Id: telegram?.Telegram_Channel_Id ?? "",
    Backoff_Period: telegram?.Backoff_Period ?? "",
  });
  const [status, setStatus] = useState<StatusMsg | null>(null);

  const set =
    (field: keyof TelegramFormData) => (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    const res = await fetch("/api/config/telegram", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        Telegram_Token: form.Telegram_Token || null,
        Telegram_Channel_Id: form.Telegram_Channel_Id || null,
        Backoff_Period: form.Backoff_Period || null,
      }),
    });
    if (res.ok) {
      setStatus({ text: "Telegram config saved.", error: false });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleClear = async () => {
    const res = await fetch("/api/config/telegram", { method: "DELETE" });
    if (res.ok) {
      setStatus({ text: "Telegram config cleared.", error: false });
      setForm({
        Telegram_Token: "",
        Telegram_Channel_Id: "",
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
            Bot Token
            <input
              className={inputClass}
              type="password"
              value={form.Telegram_Token}
              onChange={set("Telegram_Token")}
              placeholder="••••••••"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Chat / Channel ID
            <input
              className={inputClass}
              value={form.Telegram_Channel_Id}
              onChange={set("Telegram_Channel_Id")}
              placeholder="@channelname or -100123456789"
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
