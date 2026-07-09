import FormPanel, { type FieldDef } from "./FormPanel";
import type { DiscordConfig, DiscordPanelProps } from "../../types";

interface DiscordFormData {
  [key: string]: string;
  Discord_Auth: string;
  Discord_Channel: string;
  Backoff_Period: string;
}

const EMPTY: DiscordFormData = {
  Discord_Auth: "",
  Discord_Channel: "",
  Backoff_Period: "",
};

const FIELDS: FieldDef<DiscordFormData>[] = [
  { key: "Discord_Auth", label: "Authorization", type: "password", placeholder: "Bot ••••••••" },
  { key: "Discord_Channel", label: "Channel ID", placeholder: "123456789012345678" },
  { key: "Backoff_Period", label: "Backoff Period", placeholder: "5m" },
];

const DESCRIPTION = (
  <p className="text-[0.8rem] text-muted m-0">
    Authorization header value — use{" "}
    <code className="text-fg">Bot &lt;token&gt;</code> for a bot token or{" "}
    <code className="text-fg">Bearer &lt;token&gt;</code> for a user OAuth
    token.
  </p>
);

function toForm(discord?: DiscordConfig): DiscordFormData {
  return {
    Discord_Auth: discord?.Discord_Auth ?? "",
    Discord_Channel: discord?.Discord_Channel ?? "",
    Backoff_Period: discord?.Backoff_Period ?? "",
  };
}

export default function DiscordPanel({ discord, onRefresh }: DiscordPanelProps) {
  return (
    <FormPanel
      endpoint="/api/config/discord"
      entityLabel="Discord"
      fields={FIELDS}
      initial={toForm(discord)}
      empty={EMPTY}
      description={DESCRIPTION}
      onRefresh={onRefresh}
    />
  );
}
