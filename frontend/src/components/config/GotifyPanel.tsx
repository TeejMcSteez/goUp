import FormPanel, { type FieldDef } from "./FormPanel";
import type { GotifyConfig, GotifyPanelProps } from "../../types";

interface GotifyFormData {
  [key: string]: string;
  Gotify_Server: string;
  Gotify_Token: string;
  Gotify_Application: string;
  Gotify_Title: string;
  Gotify_Priority: string;
  Backoff_Period: string;
}

const EMPTY: GotifyFormData = {
  Gotify_Server: "",
  Gotify_Token: "",
  Gotify_Application: "",
  Gotify_Title: "",
  Gotify_Priority: "",
  Backoff_Period: "",
};

const FIELDS: FieldDef<GotifyFormData>[] = [
  { key: "Gotify_Server", label: "Server URL", placeholder: "https://gotify.example.com" },
  { key: "Gotify_Token", label: "App Token", type: "password", placeholder: "••••••••" },
  { key: "Gotify_Application", label: "Application", placeholder: "GoUp" },
  { key: "Gotify_Title", label: "Title", placeholder: "GoUp Alert" },
  { key: "Gotify_Priority", label: "Priority", type: "number", min: "0", max: "10", placeholder: "5" },
  { key: "Backoff_Period", label: "Backoff Period", placeholder: "5m" },
];

function toForm(gotify?: GotifyConfig): GotifyFormData {
  return {
    Gotify_Server: gotify?.Gotify_Server ?? "",
    Gotify_Token: gotify?.Gotify_Token ?? "",
    Gotify_Application: gotify?.Gotify_Application ?? "",
    Gotify_Title: gotify?.Gotify_Title ?? "",
    Gotify_Priority: gotify?.Gotify_Priority?.toString() ?? "",
    Backoff_Period: gotify?.Backoff_Period ?? "",
  };
}

function serialize(form: GotifyFormData) {
  return {
    Gotify_Server: form.Gotify_Server || null,
    Gotify_Token: form.Gotify_Token || null,
    Gotify_Application: form.Gotify_Application || null,
    Gotify_Title: form.Gotify_Title || null,
    Gotify_Priority: form.Gotify_Priority ? parseInt(form.Gotify_Priority, 10) : null,
    Backoff_Period: form.Backoff_Period || null,
  };
}

export default function GotifyPanel({ gotify, onRefresh }: GotifyPanelProps) {
  return (
    <FormPanel
      endpoint="/api/config/gotify"
      entityLabel="Gotify"
      fields={FIELDS}
      initial={toForm(gotify)}
      empty={EMPTY}
      serialize={serialize}
      onRefresh={onRefresh}
    />
  );
}
