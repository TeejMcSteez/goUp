import { useState } from "react";
import StatusMessage from "./StatusMessage";
import type {
  StatusMessage as StatusMsg,
  DiscordPanelProps,
} from "../../types";

const inputClass =
  "px-4 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] transition-[border-color,box-shadow] duration-200 focus:outline-none focus:border-focus focus:shadow-[0_0_0_3px_rgba(56,189,248,0.15)] placeholder:text-muted placeholder:opacity-50";

const btnBase =
  "px-6 py-2 rounded-lg border border-border bg-surface text-fg text-[0.9rem] cursor-pointer transition-all duration-200 hover:bg-elevated hover:-translate-y-px disabled:cursor-not-allowed disabled:opacity-50";

interface DiscordFormData {
  Discord_Auth: string;
  Discord_Channel: string;
}

export default function DiscordPanel({
  discord,
  onRefresh,
}: DiscordPanelProps) {
  const [form, setForm] = useState<DiscordFormData>({
    Discord_Auth: discord?.Discord_Auth ?? "",
    Discord_Channel: discord?.Discord_Channel ?? "",
  });
  const [status, setStatus] = useState<StatusMsg | null>(null);

  const set =
    (field: keyof DiscordFormData) =>
    (e: React.ChangeEvent<HTMLInputElement>) =>
      setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSave = async (e: React.SubmitEvent) => {
    e.preventDefault();
    const res = await fetch("/api/config/discord", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        Discord_Auth: form.Discord_Auth || null,
        Discord_Channel: form.Discord_Channel || null,
      }),
    });
    if (res.ok) {
      setStatus({ text: "Discord config saved.", error: false });
      onRefresh();
    } else {
      setStatus({ text: await res.text(), error: true });
    }
  };

  const handleClear = async () => {
    const res = await fetch("/api/config/discord", { method: "DELETE" });
    if (res.ok) {
      setStatus({ text: "Discord config cleared.", error: false });
      setForm({ Discord_Auth: "", Discord_Channel: "" });
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
        <p className="text-[0.8rem] text-muted m-0">
          Authorization header value — use{" "}
          <code className="text-fg">Bot &lt;token&gt;</code> for a bot token or{" "}
          <code className="text-fg">Bearer &lt;token&gt;</code> for a user OAuth
          token.
        </p>
        <div className="grid grid-cols-[repeat(auto-fit,minmax(220px,1fr))] gap-4">
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Authorization
            <input
              className={inputClass}
              type="password"
              value={form.Discord_Auth}
              onChange={set("Discord_Auth")}
              placeholder="Bot ••••••••"
            />
          </label>
          <label className="flex flex-col gap-1 text-[0.85rem] font-medium text-muted">
            Channel ID
            <input
              className={inputClass}
              value={form.Discord_Channel}
              onChange={set("Discord_Channel")}
              placeholder="123456789012345678"
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
