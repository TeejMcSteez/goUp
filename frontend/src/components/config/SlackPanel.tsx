import FormPanel, { type FieldDef } from "./FormPanel";
import type { SlackConfig, SlackPanelProps } from "../../types";

interface SlackFormData {
  [key: string]: string;
  Slack_Token: string;
  Slack_Channel: string;
  Bot_Username: string;
  Backoff_Period: string;
}

const EMPTY: SlackFormData = {
  Slack_Token: "",
  Slack_Channel: "",
  Bot_Username: "",
  Backoff_Period: "",
};

const FIELDS: FieldDef<SlackFormData>[] = [
  { key: "Slack_Token", label: "Bot Token", type: "password", placeholder: "xoxb-••••••••" },
  { key: "Slack_Channel", label: "Channel", placeholder: "#alerts" },
  { key: "Bot_Username", label: "Bot Username", placeholder: "GoUp Bot" },
  { key: "Backoff_Period", label: "Backoff Period", placeholder: "5m" },
];

function toForm(slack?: SlackConfig): SlackFormData {
  return {
    Slack_Token: slack?.Slack_Token ?? "",
    Slack_Channel: slack?.Slack_Channel ?? "",
    Bot_Username: slack?.Bot_Username ?? "",
    Backoff_Period: slack?.Backoff_Period ?? "",
  };
}

export default function SlackPanel({ slack, onRefresh }: SlackPanelProps) {
  return (
    <FormPanel
      endpoint="/api/config/slack"
      entityLabel="Slack"
      fields={FIELDS}
      initial={toForm(slack)}
      empty={EMPTY}
      onRefresh={onRefresh}
    />
  );
}
