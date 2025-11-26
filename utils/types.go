package utils

import (
	"sync"
)

type Config struct {
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	URL string `yaml:"url"`
	API_URL *string `yaml:"api_url"`
	API_KEY *string `yaml:"api_key"`
}

type ServiceData struct {
	ServiceName         string `json:"name"`
	ServiceHTTPResponse string    `json:"response"`
	ServiceAPIResponse string `json:"data"`
}

type SharedData struct {
	mu   sync.RWMutex
	data []ServiceData
}

type ScheduleParamters struct {
	Span *int
	Interval *string
	Mux *sync.RWMutex
}

type ParamtersData struct {
	Span int
	Interval string
}
