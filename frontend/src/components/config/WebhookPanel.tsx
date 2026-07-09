import FormPanel, { type FieldDef } from "./FormPanel";
import type { WebhookConfig, WebhookPanelProps } from "../../types";

interface WebhookFormData {
  [key: string]: string;
  Webhook_url: string;
  Webhook_key_string: string;
  Custom_message: string;
  Backoff_Period: string;
}

const EMPTY: WebhookFormData = {
  Webhook_url: "",
  Webhook_key_string: "",
  Custom_message: "",
  Backoff_Period: "",
};

const FIELDS: FieldDef<WebhookFormData>[] = [
  { key: "Webhook_url", label: "Webhook URL", placeholder: "https://hooks.example.com/..." },
  { key: "Webhook_key_string", label: "Authorization Header", placeholder: "Bearer <token>" },
  { key: "Custom_message", label: "Custom Message", placeholder: "A service is down!" },
  { key: "Backoff_Period", label: "Backoff Period", placeholder: "5m" },
];

function toForm(webhook?: WebhookConfig): WebhookFormData {
  return {
    Webhook_url: webhook?.Webhook_url ?? "",
    Webhook_key_string: webhook?.Webhook_key_string ?? "",
    Custom_message: webhook?.Custom_message ?? "",
    Backoff_Period: webhook?.Backoff_Period ?? "",
  };
}

export default function WebhookPanel({ webhook, onRefresh }: WebhookPanelProps) {
  return (
    <FormPanel
      endpoint="/api/config/webhook"
      entityLabel="Webhook"
      fields={FIELDS}
      initial={toForm(webhook)}
      empty={EMPTY}
      onRefresh={onRefresh}
    />
  );
}
