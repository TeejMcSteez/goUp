import FormPanel, { type FieldDef } from "./FormPanel";
import type { TelegramConfig, TelegramPanelProps } from "../../types";

interface TelegramFormData {
  [key: string]: string;
  Telegram_Token: string;
  Telegram_Channel_Id: string;
  Backoff_Period: string;
}

const EMPTY: TelegramFormData = {
  Telegram_Token: "",
  Telegram_Channel_Id: "",
  Backoff_Period: "",
};

const FIELDS: FieldDef<TelegramFormData>[] = [
  { key: "Telegram_Token", label: "Bot Token", type: "password", placeholder: "••••••••" },
  { key: "Telegram_Channel_Id", label: "Chat / Channel ID", placeholder: "@channelname or -100123456789" },
  { key: "Backoff_Period", label: "Backoff Period", placeholder: "5m" },
];

function toForm(telegram?: TelegramConfig): TelegramFormData {
  return {
    Telegram_Token: telegram?.Telegram_Token ?? "",
    Telegram_Channel_Id: telegram?.Telegram_Channel_Id ?? "",
    Backoff_Period: telegram?.Backoff_Period ?? "",
  };
}

export default function TelegramPanel({ telegram, onRefresh }: TelegramPanelProps) {
  return (
    <FormPanel
      endpoint="/api/config/telegram"
      entityLabel="Telegram"
      fields={FIELDS}
      initial={toForm(telegram)}
      empty={EMPTY}
      onRefresh={onRefresh}
    />
  );
}
