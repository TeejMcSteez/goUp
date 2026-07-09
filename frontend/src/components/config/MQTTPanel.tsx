import FormPanel, { type FieldDef } from "./FormPanel";
import type { MQTTConfig, MQTTPanelProps } from "../../types";

interface MQTTFormData {
  [key: string]: string;
  Mqtt_broker: string;
  Mqtt_username: string;
  Mqtt_key: string;
  Backoff_Period: string;
}

const EMPTY: MQTTFormData = {
  Mqtt_broker: "",
  Mqtt_username: "",
  Mqtt_key: "",
  Backoff_Period: "",
};

const FIELDS: FieldDef<MQTTFormData>[] = [
  { key: "Mqtt_broker", label: "Broker URL", placeholder: "mqtt://broker.example.com" },
  { key: "Mqtt_username", label: "Username", placeholder: "username" },
  { key: "Mqtt_key", label: "Key / Password", type: "password", placeholder: "••••••••" },
  { key: "Backoff_Period", label: "Backoff Period", placeholder: "5m" },
];

function toForm(mqtt?: MQTTConfig): MQTTFormData {
  return {
    Mqtt_broker: mqtt?.Mqtt_broker ?? "",
    Mqtt_username: mqtt?.Mqtt_username ?? "",
    Mqtt_key: mqtt?.Mqtt_key ?? "",
    Backoff_Period: mqtt?.Backoff_Period ?? "",
  };
}

export default function MQTTPanel({ mqtt, onRefresh }: MQTTPanelProps) {
  return (
    <FormPanel
      endpoint="/api/config/mqtt"
      entityLabel="MQTT"
      fields={FIELDS}
      initial={toForm(mqtt)}
      empty={EMPTY}
      onRefresh={onRefresh}
    />
  );
}
