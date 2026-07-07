import { useState } from "react";
import StatusMessage from "./StatusMessage";
import type { StatusMessage as StatusMsg, SlackPanelProps } from "../../types";

const inputClass =
  "px-4 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.15)] placeholder:text-muted placeholder:opacity-50";

const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

interface SlackFormData {
  Slack_Token: string;
  Slack_Channel: string;
  Bot_Username: string;
  Backoff_Period: string;
}

export default function SlackPanel({ slack, onRefresh }: SlackPanelProps) {
  const [form, setForm] = useState<SlackFormData>({
    Slack_Token: slack?.Slack_Token ?? "",
    Slack_Channel: slack?.Slack_Channel ?? "",
    Bot_Username: slack?.Bot_Username ?? "",
    Backoff_Period: slack?.Backoff_Period ?? "",
  });
  const [status, setStatus] = useState<StatusMsg | null>(null);

  const set =
    (field: keyof SlackFormData) => (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    const res = await fetch("/api/config/slack", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        Slack_Token: form.Slack_Token || null,
        Slack_Channel: form.Slack_Channel || null,
        Bot_Username: form.Bot_Username || null,
        Backoff_Period: form.Backoff_Period || null,
      }),
    });
    if (res.ok) {
      setStatus({ text: "Slack config saved.", error: false });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleClear = async () => {
    const res = await fetch("/api/config/slack", { method: "DELETE" });
    if (res.ok) {
      setStatus({ text: "Slack config cleared.", error: false });
      setForm({
        Slack_Token: "",
        Slack_Channel: "",
        Bot_Username: "",
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
              value={form.Slack_Token}
              onChange={set("Slack_Token")}
              placeholder="xoxb-••••••••"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Channel
            <input
              className={inputClass}
              value={form.Slack_Channel}
              onChange={set("Slack_Channel")}
              placeholder="#alerts"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Bot Username
            <input
              className={inputClass}
              value={form.Bot_Username}
              onChange={set("Bot_Username")}
              placeholder="GoUp Bot"
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
