export interface Service {
  name: string;
  url: string;
  response?: string;
  response_time?: string;
  data?: string;
  error?: string;
}

export interface ErrorItem {
  name: string;
  timestamp: string;
  response: string;
  data?: string;
}

export interface UptimeItem {
  name: string;
  average: number;
}

export interface ChartDataset {
  label: string;
  data: number[];
  backgroundColor: string[];
  borderColor: string[];
  borderWidth: number;
}

export interface UptimeChartData {
  labels: string[];
  datasets: ChartDataset[];
}

export interface ServiceConfig {
  Name: string;
  URL: string;
  API_URL?: string;
  Valid_Responses?: number[];
  Retry_Requests?: number;
}

export interface MQTTConfig {
  Mqtt_broker?: string;
  Mqtt_username?: string;
  Mqtt_key?: string;
}

export interface WebhookConfig {
  Webhook_url?: string;
  Webhook_key_string?: string;
  Custom_message?: string;
}

export interface AppConfig {
  services?: Record<string, ServiceConfig>;
  mqtt?: MQTTConfig;
  webhook?: WebhookConfig;
  database?: boolean;
}

export interface Schedule {
  timespan: number;
  interval: string;
}

export interface ResponseTimeEntry {
  service_data: {
    name: string;
    url: string;
    response: string;
    response_time: string;
    error: boolean;
  };
  response_time: string;
}

export interface DbSizeResponse {
  size?: number;
  size_string?: string;
}

export interface StatusMessage {
  text: string;
  error: boolean;
}
