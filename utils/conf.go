package utils

import (
	"os"

	"gopkg.in/yaml.v3"
)

var Current_Config *Config

/*
Parses yml file for service information
*/
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg = new(Config)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
