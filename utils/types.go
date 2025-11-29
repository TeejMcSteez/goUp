package utils

import (
	"sync"
)

type Config struct {
	Services map[string]Service `yaml:"services"`
	Triggers Trigger            `yaml:"triggers"`
}

type Trigger struct {
	Mqtt_broker   *string `yaml:"mqtt_broker"`
	Mqtt_username *string `yaml:"mqtt_user"`
	Mqtt_key      *string `yaml:"mqtt_key"`
}

type Service struct {
	URL     string  `yaml:"url"`
	API_URL *string `yaml:"api_url"`
	API_KEY *string `yaml:"api_key"`
}

type ServiceEndpoints struct {
	Mux             *sync.RWMutex
	ServiceEndpoint []Service
}

type ServiceData struct {
	ServiceName         string `json:"name"`
	ServiceHTTPResponse string `json:"response"`
	ServiceAPIResponse  string `json:"data"`
}

type SharedData struct {
	mu   sync.RWMutex
	data []ServiceData
}

type ScheduleParameters struct {
	Span     *int
	Interval *string
	Mux      *sync.RWMutex
}

type ParamtersData struct {
	Span     int    `json:"timespan"`
	Interval string `json:"timeInterval"`
}
