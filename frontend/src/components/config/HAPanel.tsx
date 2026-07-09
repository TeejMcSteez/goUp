import FormPanel, { type FieldDef } from "./FormPanel";
import type { HAConfig, HAPanelProps } from "../../types";

interface HAFormData {
  [key: string]: string;
  HA_URL: string;
  HA_Token: string;
  Backoff_Period: string;
}

const EMPTY: HAFormData = { HA_URL: "", HA_Token: "", Backoff_Period: "" };

const FIELDS: FieldDef<HAFormData>[] = [
  { key: "HA_URL", label: "Instance URL", placeholder: "http://homeassistant.local:8123" },
  { key: "HA_Token", label: "Long-Lived Access Token", type: "password", placeholder: "••••••••" },
  { key: "Backoff_Period", label: "Backoff Period", placeholder: "5m" },
];

const DESCRIPTION = (
  <p className="text-[0.8rem] text-muted m-0">
    Fires a <code className="text-fg">goup_alert</code> event on the HA
    event bus when a service goes down. Use an automation with{" "}
    <code className="text-fg">trigger: event_type: goup_alert</code> to
    act on it.
  </p>
);

function toForm(ha?: HAConfig): HAFormData {
  return {
    HA_URL: ha?.HA_URL ?? "",
    HA_Token: ha?.HA_Token ?? "",
    Backoff_Period: ha?.Backoff_Period ?? "",
  };
}

export default function HAPanel({ ha, onRefresh }: HAPanelProps) {
  return (
    <FormPanel
      endpoint="/api/config/ha"
      entityLabel="Home Assistant"
      fields={FIELDS}
      initial={toForm(ha)}
      empty={EMPTY}
      description={DESCRIPTION}
      onRefresh={onRefresh}
    />
  );
}
