package utils

import (
	"sync"
)

// Main yaml config
type Config struct {
	Services map[string]Service `yaml:"services"`
	Triggers Trigger            `yaml:"triggers"`
}

// Endpoints and data for triggers
type Trigger struct {
	Mqtt_broker   *string `yaml:"mqtt_broker"`
	Mqtt_username *string `yaml:"mqtt_user"`
	Mqtt_key      *string `yaml:"mqtt_key"`
	Webhook_url *string `yaml:"webhook_url"`
	Webhook_key *string `yaml:"webhook_key"`
}

// Data for service endponts
type Service struct {
	URL     string  `yaml:"url"`
	API_URL *string `yaml:"api_url"`
	API_KEY *string `yaml:"api_key"`
}

// Shared service endpoint struct
type ServiceEndpoints struct {
	Mux             *sync.RWMutex
	ServiceEndpoint []Service
}

// Service data from GetServiceData
type ServiceData struct {
	ServiceName         string `json:"name"`
	ServiceHTTPResponse string `json:"response"`
	ServiceAPIResponse  string `json:"data"`
	ServiceResponseTime string `json:"response_time"`
}

// Shared paramters used in scheduler
type ScheduleParameters struct {
	Span     *int
	Interval *string
	Mux      *sync.RWMutex
}

// Base parameter struct used in shared parameters
type ParamtersData struct {
	Span     int    `json:"timespan"`
	Interval string `json:"timeInterval"`
}

type AverageData struct {
	Name string `json:"name"`
	Average float64 `json:"average"`
}