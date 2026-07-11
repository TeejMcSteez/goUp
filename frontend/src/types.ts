export interface Service {
  name: string;
  url: string;
  description?: string;
  response?: string;
  response_time?: string;
  data?: string;
  error?: string;
  active: boolean;
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

export interface DatabaseSizePayload {
  db_max_size: string;
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
  Description?: string;
  API_URL?: string;
  Valid_Responses?: number[];
  Retry_Requests?: number;
  Active?: boolean;
}

export interface MQTTConfig {
  Mqtt_broker?: string;
  Mqtt_username?: string;
  Mqtt_key?: string;
  Backoff_Period?: string;
}

export interface MQTTPanelProps {
  mqtt?: MQTTConfig;
  onRefresh: () => void;
}

export interface WebhookConfig {
  Webhook_url?: string;
  Webhook_key_string?: string;
  Custom_message?: string;
  Backoff_Period?: string;
}

export interface WebhookPanelProps {
  webhook?: WebhookConfig;
  onRefresh: () => void;
}

export interface SMTPConfig {
  Email?: string;
  App_Password?: string;
  SMTPServer?: string;
  Backoff_Period?: string;
}

export interface SMTPPanelProps {
  smtp?: SMTPConfig;
  onRefresh: () => void;
}

export interface GotifyConfig {
  Gotify_Server?: string;
  Gotify_Token?: string;
  Gotify_Application?: string;
  Gotify_Title?: string;
  Gotify_Priority?: number;
  Backoff_Period?: string;
}

export interface GotifyPanelProps {
  gotify?: GotifyConfig;
  onRefresh: () => void;
}

export interface SlackConfig {
  Slack_Token?: string;
  Slack_Channel?: string;
  Bot_Username?: string;
  Backoff_Period?: string;
}

export interface SlackPanelProps {
  slack?: SlackConfig;
  onRefresh: () => void;
}

export interface TelegramConfig {
  Telegram_Token?: string;
  Telegram_Channel_Id?: string;
  Backoff_Period?: string;
}

export interface TelegramPanelProps {
  telegram?: TelegramConfig;
  onRefresh: () => void;
}

export interface HAConfig {
  HA_URL?: string;
  HA_Token?: string;
  Backoff_Period?: string;
}

export interface HAPanelProps {
  ha?: HAConfig;
  onRefresh: () => void;
}

export interface DiscordConfig {
  Discord_Auth?: string;
  Discord_Channel?: string;
  Backoff_Period?: string;
}

export interface DiscordPanelProps {
  discord?: DiscordConfig;
  onRefresh: () => void;
}

export interface GlobalBackoffPanelProps {
  backoffPeriod?: string;
  onRefresh: () => void;
}

export interface AppConfig {
  services?: Record<string, ServiceConfig>;
  mqtt?: MQTTConfig;
  webhook?: WebhookConfig;
  smtp?: SMTPConfig;
  gotify?: GotifyConfig;
  slack?: SlackConfig;
  telegram?: TelegramConfig;
  ha?: HAConfig;
  discord?: DiscordConfig;
  database?: boolean;
  backoff_period?: string;
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

export interface ServiceCardProps {
  service: Service;
}
