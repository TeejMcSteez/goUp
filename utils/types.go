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
	Services map[string]Service `json:"services"`
	MQTT     MQTTTrigger        `json:"mqtt"`
	Webhook  WebhookTrigger     `json:"webhook"`
	SMTP     SMTPTrigger        `json:"smtp"`
	Gotify   GotifyTrigger      `json:"gotify"`
}

// Endpoints and data for triggers
type Trigger struct {
	MQTT           MQTTTrigger    `yaml:"mqtt"`
	Webhook        WebhookTrigger `yaml:"webhook"`
	SMTP           SMTPTrigger    `yaml:"smtp"`
	Gotify         GotifyTrigger  `yaml:"gotify"`
	Slack          SlackTrigger   `yaml:"slack"`
	Backoff_Period *string        `yaml:"backoff"`

	backoffDuration time.Duration
	lastFired       time.Time
	handlers        []TriggerHandler
}

type TriggerHandler interface {
	Fire(data []ServiceData)
	IsConfigured() bool
}

type MQTTTrigger struct {
	Mqtt_broker   *string `yaml:"mqtt_broker"`
	Mqtt_username *string `yaml:"mqtt_user"`
	Mqtt_key      *string `yaml:"mqtt_key"`
}

type WebhookTrigger struct {
	Webhook_url *string `yaml:"webhook_url"`
	// Can use any authorization header string
	// Ex: Basic <creds>/Digest username=<username>.../Bearer <token_string>
	Webhook_key_string *string `yaml:"webhook_key"`
	Custom_message     *string `yaml:"custom_message"`
}

type SMTPTrigger struct {
	Email        *string `yaml:"email"`
	App_Password *string `yaml:"app_password"`
	SMTPServer   *string `yaml:"smtp_server"`
}

type GotifyTrigger struct {
	Gotify_Server      *string `yaml:"gotify_server"`
	Gotify_Token       *string `yaml:"gotify_app_token"`
	Gotify_Application *string `yaml:"gotify_application"`
	Gotify_Title       *string `yaml:"gotify_title"`
	Gotify_Priority    *int    `yaml:"gotify_priority"`
}

type SlackTrigger struct {
	Slack_Token   *string `yaml:"slack_token"`
	Slack_Channel *string `yaml:"slack_channel"`
	Bot_Username  *string `yaml:"username"`
}

// Data for service endponts
type Service struct {
	Name            string    `yaml:"-"`
	URL             string    `yaml:"url"`
	Description     *string   `yaml:"description"`
	API_URL         *string   `yaml:"api_url"`
	API_KEY         *string   `yaml:"api_key"`
	Valid_Responses *[]string `yaml:"valid_responses"`
	Retry_Requests  *int      `yaml:"retry"`
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
