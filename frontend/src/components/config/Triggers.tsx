import { useState } from "react";
import { MQTTPanelProps, WebhookPanelProps, SMTPPanelProps, GotifyPanelProps } from "../../types";
import MQTTPanel from "./MQTTPanel";
import WebhookPanel from "./WebhookPanel";
import SMTPPanel from "./SMTPPanel";
import GotifyPanel from "./GotifyPanel";

interface TriggerProps {
  mqttProps: MQTTPanelProps | undefined;
  webhookProps: WebhookPanelProps | undefined;
  smtpProps: SMTPPanelProps | undefined;
  gotifyProps: GotifyPanelProps | undefined;
}

type Tab = "mqtt" | "webhook" | "smtp" | "gotify";

const tabBtn = (active: boolean) =>
  `px-4 py-1.5 text-[0.85rem] font-medium rounded-md border-none cursor-pointer transition-all duration-200 ${
    active ? "bg-surface text-primary" : "bg-transparent text-muted hover:text-fg"
  }`;

export default function Triggers({ mqttProps, webhookProps, smtpProps, gotifyProps }: TriggerProps) {
  const [tab, setTab] = useState<Tab>("mqtt");

  return (
    <div className="flex flex-col gap-4">
      <div className="flex gap-1 p-1 bg-elevated rounded-lg w-fit">
        <button className={tabBtn(tab === "mqtt")} onClick={() => setTab("mqtt")}>
          MQTT
        </button>
        <button className={tabBtn(tab === "webhook")} onClick={() => setTab("webhook")}>
          Webhook
        </button>
        <button className={tabBtn(tab === "smtp")} onClick={() => setTab("smtp")}>
          SMTP
        </button>
        <button className={tabBtn(tab === "gotify")} onClick={() => setTab("gotify")}>
          Gotify
        </button>
      </div>
      {tab === "mqtt" && (
        <MQTTPanel
          onRefresh={mqttProps?.onRefresh ?? function () {}}
          mqtt={mqttProps?.mqtt}
        />
      )}
      {tab === "webhook" && (
        <WebhookPanel
          onRefresh={webhookProps?.onRefresh ?? function () {}}
          webhook={webhookProps?.webhook}
        />
      )}
      {tab === "smtp" && (
        <SMTPPanel
          onRefresh={smtpProps?.onRefresh ?? function () {}}
          smtp={smtpProps?.smtp}
        />
      )}
      {tab === "gotify" && (
        <GotifyPanel
          onRefresh={gotifyProps?.onRefresh ?? function () {}}
          gotify={gotifyProps?.gotify}
        />
      )}
    </div>
  );
}
