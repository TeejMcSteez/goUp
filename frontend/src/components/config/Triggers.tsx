import { useState } from "react";
import { MQTTPanelProps, WebhookPanelProps } from "../../types";
import MQTTPanel from "./MQTTPanel";
import WebhookPanel from "./WebhookPanel";

interface TriggerProps {
  mqttProps: MQTTPanelProps | undefined;
  webhookProps: WebhookPanelProps | undefined;
}

const tabBtn = (active: boolean) =>
  `px-4 py-1.5 text-[0.85rem] font-medium rounded-md border-none cursor-pointer transition-all duration-200 ${
    active ? "bg-surface text-primary" : "bg-transparent text-muted hover:text-fg"
  }`;

export default function Triggers({ mqttProps, webhookProps }: TriggerProps) {
  const [tab, setTab] = useState<"mqtt" | "webhook">("mqtt");

  return (
    <div className="flex flex-col gap-4">
      <div className="flex gap-1 p-1 bg-elevated rounded-lg w-fit">
        <button className={tabBtn(tab === "mqtt")} onClick={() => setTab("mqtt")}>
          MQTT
        </button>
        <button className={tabBtn(tab === "webhook")} onClick={() => setTab("webhook")}>
          Webhook
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
    </div>
  );
}
