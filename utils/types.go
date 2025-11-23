package utils

import "sync"

type Config struct {
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	URL string `yaml:"url"`
}

type ServiceData struct {
	ServiceName         string `json:"name"`
	ServiceHTTPResponse int    `json:"response"`
}

type SharedData struct {
	mu   sync.RWMutex
	data []ServiceData
}
