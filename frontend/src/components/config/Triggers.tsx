import { useState } from "react";
import type { AppConfig } from "../../types";
import MQTTPanel from "./MQTTPanel";
import WebhookPanel from "./WebhookPanel";
import SMTPPanel from "./SMTPPanel";
import GotifyPanel from "./GotifyPanel";
import SlackPanel from "./SlackPanel";
import TelegramPanel from "./TelegramPanel";
import HAPanel from "./HAPanel";
import DiscordPanel from "./DiscordPanel";
import GlobalBackoffPanel from "./GlobalBackoffPanel";

interface TriggersProps {
  config: AppConfig | null;
  onRefresh: () => void;
}

const TABS = [
  { id: "global", label: "Global" },
  { id: "mqtt", label: "MQTT" },
  { id: "webhook", label: "Webhook" },
  { id: "smtp", label: "SMTP" },
  { id: "gotify", label: "Gotify" },
  { id: "slack", label: "Slack" },
  { id: "telegram", label: "Telegram" },
  { id: "ha", label: "Home Assistant" },
  { id: "discord", label: "Discord" },
] as const;

type Tab = (typeof TABS)[number]["id"];

const tabBtn = (active: boolean) =>
  `px-4 py-1.5 text-[0.85rem] font-medium rounded-md border-none cursor-pointer transition-all duration-200 ${
    active ? "bg-surface text-primary" : "bg-transparent text-muted hover:text-fg"
  }`;

export default function Triggers({ config, onRefresh }: TriggersProps) {
  const [tab, setTab] = useState<Tab>("global");

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap gap-1 p-1 bg-elevated rounded-lg w-fit">
        {TABS.map((t) => (
          <button key={t.id} className={tabBtn(tab === t.id)} onClick={() => setTab(t.id)}>
            {t.label}
          </button>
        ))}
      </div>
      {tab === "global" && (
        <GlobalBackoffPanel backoffPeriod={config?.backoff_period} onRefresh={onRefresh} />
      )}
      {tab === "mqtt" && <MQTTPanel mqtt={config?.mqtt} onRefresh={onRefresh} />}
      {tab === "webhook" && <WebhookPanel webhook={config?.webhook} onRefresh={onRefresh} />}
      {tab === "smtp" && <SMTPPanel smtp={config?.smtp} onRefresh={onRefresh} />}
      {tab === "gotify" && <GotifyPanel gotify={config?.gotify} onRefresh={onRefresh} />}
      {tab === "slack" && <SlackPanel slack={config?.slack} onRefresh={onRefresh} />}
      {tab === "telegram" && <TelegramPanel telegram={config?.telegram} onRefresh={onRefresh} />}
      {tab === "ha" && <HAPanel ha={config?.ha} onRefresh={onRefresh} />}
      {tab === "discord" && <DiscordPanel discord={config?.discord} onRefresh={onRefresh} />}
    </div>
  );
}
