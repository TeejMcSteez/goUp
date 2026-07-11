package utils

import (
	"sync"
	"time"
)

const GB = 1 * 1e9

// Main yaml config
type Config struct {
	Database_Location *string            `yaml:"db_path"`
	Database_Max_Size *string            `yaml:"db_max_size"`
	Persist_db        *bool              `yaml:"persist_db"`
	Services          map[string]Service `yaml:"services"`
	Triggers          Trigger            `yaml:"triggers"`

	Schedule *ScheduleState `yaml:"schedule"`

	// ConfigPath is the filesystem path this config was loaded from.
	// Used internally for write-back operations; not serialized to YAML.
	ConfigPath string `yaml:"-"`
}

type ConfigData struct {
	Services       map[string]Service `json:"services"`
	MQTT           MQTTTrigger        `json:"mqtt"`
	Webhook        WebhookTrigger     `json:"webhook"`
	SMTP           SMTPTrigger        `json:"smtp"`
	Gotify         GotifyTrigger      `json:"gotify"`
	Slack          SlackTrigger       `json:"slack"`
	Telegram       TelegramTrigger    `json:"telegram"`
	HA             HATrigger          `json:"ha"`
	Discord        DiscordTrigger     `json:"discord"`
	Backoff_Period string             `json:"backoff_period"`
}

// Endpoints and data for triggers
type Trigger struct {
	MQTT           MQTTTrigger     `yaml:"mqtt"`
	Webhook        WebhookTrigger  `yaml:"webhook"`
	SMTP           SMTPTrigger     `yaml:"smtp"`
	Gotify         GotifyTrigger   `yaml:"gotify"`
	Slack          SlackTrigger    `yaml:"slack"`
	Telegram       TelegramTrigger `yaml:"telegram"`
	HA             HATrigger       `yaml:"home_assistant"`
	Discord        DiscordTrigger  `yaml:"discord"`
	Backoff_Period *string         `yaml:"backoff"`

	backoffDuration time.Duration
	lastFired       time.Time
	handlers        []TriggerHandler
}

type TriggerHandler interface {
	Fire(data []ServiceData)
	IsConfigured() bool
}

type MQTTTrigger struct {
	Mqtt_broker     *string `yaml:"mqtt_broker"`
	Mqtt_username   *string `yaml:"mqtt_user"`
	Mqtt_key        *string `yaml:"mqtt_key"`
	Backoff_Period  *string `yaml:"backoff"`
	backoffDuration time.Duration
	lastFired       time.Time
}

type WebhookTrigger struct {
	Webhook_url *string `yaml:"webhook_url"`
	// Can use any authorization header string
	// Ex: Basic <creds>/Digest username=<username>.../Bearer <token_string>
	Webhook_key_string *string `yaml:"webhook_key"`
	Custom_message     *string `yaml:"custom_message"`
	Backoff_Period     *string `yaml:"backoff"`
	backoffDuration    time.Duration
	lastFired          time.Time
}

type SMTPTrigger struct {
	Email           *string `yaml:"email"`
	App_Password    *string `yaml:"app_password"`
	SMTPServer      *string `yaml:"smtp_server"`
	Backoff_Period  *string `yaml:"backoff"`
	backoffDuration time.Duration
	lastFired       time.Time
}

type GotifyTrigger struct {
	Gotify_Server      *string `yaml:"gotify_server"`
	Gotify_Token       *string `yaml:"gotify_app_token"`
	Gotify_Application *string `yaml:"gotify_application"`
	Gotify_Title       *string `yaml:"gotify_title"`
	Gotify_Priority    *int    `yaml:"gotify_priority"`
	Backoff_Period     *string `yaml:"backoff"`
	backoffDuration    time.Duration
	lastFired          time.Time
}

type SlackTrigger struct {
	Slack_Token     *string `yaml:"slack_token"`
	Slack_Channel   *string `yaml:"slack_channel"`
	Bot_Username    *string `yaml:"username"`
	Backoff_Period  *string `yaml:"backoff"`
	backoffDuration time.Duration
	lastFired       time.Time
}

type TelegramTrigger struct {
	Telegram_Token      *string `yaml:"telegram_token"`
	Telegram_Channel_Id *string `yaml:"telegram_channel_id"`
	Backoff_Period      *string `yaml:"backoff"`
	backoffDuration     time.Duration
	lastFired           time.Time
}

type HATrigger struct {
	HA_Token        *string `yaml:"ha_token"`
	HA_URL          *string `yaml:"ha_url"`
	Backoff_Period  *string `yaml:"backoff"`
	backoffDuration time.Duration
	lastFired       time.Time
}

type DiscordTrigger struct {
	// Can be Bot <token> or Bearer <token>
	Discord_Auth *string `yaml:"discord_auth"`
	// For now only one channel is supported
	// However, if needed many channels would be optional
	// Just need to use a string array over string pointer to send message to all channels
	Discord_Channel *string `yaml:"discord_channel_id"`
	Backoff_Period  *string `yaml:"backoff"`
	backoffDuration time.Duration
	lastFired       time.Time
}

// Data for service endponts
//
// JSON tags match the Go field names verbatim (the frontend has always
// read these fields as e.g. svc.Name / svc.URL, relying on encoding/json's
// untagged default of using the field name as-is) — don't switch these to
// snake_case without updating every frontend consumer.
type Service struct {
	Name            string    `yaml:"-" json:"Name"`
	URL             string    `yaml:"url" json:"URL"`
	Description     *string   `yaml:"description" json:"Description,omitempty"`
	API_URL         *string   `yaml:"api_url" json:"API_URL,omitempty"`
	API_KEY         *string   `yaml:"api_key" json:"API_KEY,omitempty"`
	Valid_Responses *[]string `yaml:"valid_responses" json:"Valid_Responses,omitempty"`
	Retry_Requests  *int      `yaml:"retry" json:"Retry_Requests,omitempty"`
	// Active is a *bool so an unset value (nil) can default to true,
	// distinct from an explicit `active: false`.
	Active *bool `yaml:"active" json:"Active,omitempty"`
}

// IsActive reports whether the service should be actively monitored.
// A service defaults to active unless explicitly disabled in config.
func (s Service) IsActive() bool {
	return s.Active == nil || *s.Active
}

// Shared service endpoint struct
type ServiceEndpoints struct {
	Mux             *sync.RWMutex
	ServiceEndpoint []Service
}

// Service data from GetServiceData
type ServiceData struct {
	ServiceURL          string    `json:"url"`
	ServiceName         string    `json:"name"`
	ServiceDescription  string    `json:"description"`
	ServiceHTTPResponse string    `json:"response"`
	ServiceAPIResponse  string    `json:"data"`
	ServiceResponseTime string    `json:"response_time"`
	Timestamp           time.Time `json:"timestamp"`
	Error               bool      `json:"error"`
	Active              bool      `json:"active"`
}

type ServiceResponse struct {
	AllServices  []ServiceData `json:"services"`
	DownServices []ServiceData `json:"downed_services"`
}

type ServiceResponseTime struct {
	Svc          ServiceData `json:"service_data"`
	ResponseTime string      `json:"response_time"`
}

// Shared paramters used in scheduler
type ScheduleParameters struct {
	Span     *int
	Interval *string
	Mux      *sync.RWMutex
}

// Schedule state persisted to config
type ScheduleState struct {
	Span     int    `yaml:"span" json:"timespan"`
	Interval string `yaml:"interval" json:"interval"`
}

// Average data for frontend graph
type AverageData struct {
	Name    string  `json:"name"`
	Average float64 `json:"average"`
}

// Types come from: https://sqlite.org/dbstat.html
type DatabaseStatistic struct {
	Name       string
	Path       string
	Pageno     int
	Pagetype   string
	Ncell      int
	Payload    int
	Unused     int
	Mx_payload int
	Pgoffset   int
	Pgsize     int
}

type DatabaseSizePayload struct {
	Size int64 `json:"size_string"`
}
