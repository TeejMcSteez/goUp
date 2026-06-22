import { useState } from "react";
import { MQTTPanelProps, WebhookPanelProps, SMTPPanelProps, GotifyPanelProps, SlackPanelProps, TelegramPanelProps } from "../../types";
import MQTTPanel from "./MQTTPanel";
import WebhookPanel from "./WebhookPanel";
import SMTPPanel from "./SMTPPanel";
import GotifyPanel from "./GotifyPanel";
import SlackPanel from "./SlackPanel";
import TelegramPanel from "./TelegramPanel";

interface TriggerProps {
  mqttProps: MQTTPanelProps | undefined;
  webhookProps: WebhookPanelProps | undefined;
  smtpProps: SMTPPanelProps | undefined;
  gotifyProps: GotifyPanelProps | undefined;
  slackProps: SlackPanelProps | undefined;
  telegramProps: TelegramPanelProps | undefined;
}

type Tab = "mqtt" | "webhook" | "smtp" | "gotify" | "slack" | "telegram";

const tabBtn = (active: boolean) =>
  `px-4 py-1.5 text-[0.85rem] font-medium rounded-md border-none cursor-pointer transition-all duration-200 ${
    active ? "bg-surface text-primary" : "bg-transparent text-muted hover:text-fg"
  }`;

export default function Triggers({ mqttProps, webhookProps, smtpProps, gotifyProps, slackProps, telegramProps }: TriggerProps) {
  const [tab, setTab] = useState<Tab>("mqtt");

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap gap-1 p-1 bg-elevated rounded-lg w-fit">
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
        <button className={tabBtn(tab === "slack")} onClick={() => setTab("slack")}>
          Slack
        </button>
        <button className={tabBtn(tab === "telegram")} onClick={() => setTab("telegram")}>
          Telegram
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
      {tab === "slack" && (
        <SlackPanel
          onRefresh={slackProps?.onRefresh ?? function () {}}
          slack={slackProps?.slack}
        />
      )}
      {tab === "telegram" && (
        <TelegramPanel
          onRefresh={telegramProps?.onRefresh ?? function () {}}
          telegram={telegramProps?.telegram}
        />
      )}
    </div>
  );
}
