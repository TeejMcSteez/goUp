package utils

import (
	"sync"
	"time"
)

// Main yaml config
type Config struct {
	Database_Location *string            `yaml:"db_path"`
	Database_Max_Size *string            `yaml:"db_max_size"`
	Persist_db        *bool              `yaml:"persist_db"`
	Services          map[string]Service `yaml:"services"`
	Triggers          Trigger            `yaml:"triggers"`

	// ConfigPath is the filesystem path this config was loaded from.
	// Used internally for write-back operations; not serialized to YAML.
	ConfigPath string `yaml:"-"`
}

type ConfigData struct {
	Services map[string]Service `json:"services"`
	MQTT     MQTTTrigger        `json:"mqtt"`
	Webhook  WebhookTrigger     `json:"webhook"`
}

// Endpoints and data for triggers
type Trigger struct {
	MQTT           MQTTTrigger    `yaml:"mqtt"`
	Webhook        WebhookTrigger `yaml:"webhook"`
	Backoff_Period *string        `yaml:"backoff"`

	backoffDuration time.Duration
	lastFired       time.Time
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

// Data for service endponts
type Service struct {
	Name            string    `yaml:"-"`
	URL             string    `yaml:"url"`
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
	ServiceName         string    `json:"name"`
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

// Shared paramters used in scheduler
type ScheduleParameters struct {
	Span     *int
	Interval *string
	Mux      *sync.RWMutex
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
