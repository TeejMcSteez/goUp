package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var Current_Config *Config

/*
Parses yml file for service information
*/
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		defaultDB := "goup.db"
		cfg := &Config{
			ConfigPath:        path,
			Services:          make(map[string]Service),
			Database_Location: &defaultDB,
		}
		if err := writeConfig(cfg); err != nil {
			return nil, fmt.Errorf("could not create default config: %w", err)
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg = new(Config)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	cfg.ConfigPath = path

	for key, svc := range cfg.Services {
		svc.Name = key
		cfg.Services[key] = svc
	}

	return cfg, nil
}

// writeConfig marshals the config back to the file it was loaded from.
func writeConfig(conf *Config) error {
	if conf.ConfigPath == "" {
		return fmt.Errorf("config path is not set")
	}
	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	if err := os.WriteFile(conf.ConfigPath, data, 0644); err != nil {
		return err
	}
	return nil
}

func AddConfigService(conf *Config, newEndpoint Service) error {
	if conf == nil {
		return fmt.Errorf("config is nil")
	}
	org_len := len(conf.Services)
	conf.Services[newEndpoint.Name] = newEndpoint
	new_len := len(conf.Services)
	if new_len == org_len {
		return fmt.Errorf("no new endpoint added")
	}
	return writeConfig(conf)
}

func AddConfigMQTTTrigger(config *Config, newMQTT MQTTTrigger) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	if newMQTT == config.Triggers.MQTT {
		return fmt.Errorf("MQTT trigger is the same")
	}
	config.Triggers.MQTT = newMQTT
	return writeConfig(config)
}

func AddConfigWebhookTrigger(config *Config, newWebhook WebhookTrigger) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	org_webhook := config.Triggers.Webhook
	if newWebhook == org_webhook {
		return fmt.Errorf("webhook is the same")
	}
	config.Triggers.Webhook = newWebhook
	return writeConfig(config)
}

func DeleteConfigService(config *Config, serviceToDelete Service) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	org_size := len(config.Services)
	delete(config.Services, serviceToDelete.Name)
	new_size := len(config.Services)

	if org_size == new_size {
		return fmt.Errorf("failed to remove the element")
	}

	return writeConfig(config)
}

func DeleteConfigMQTT(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	// Clears MQTT trigger in config
	config.Triggers.MQTT = MQTTTrigger{}
	return writeConfig(config)
}

func DeleteConfigTrigger(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	config.Triggers.Webhook = WebhookTrigger{}
	return writeConfig(config)
}

func ReadConfigServices(config *Config) map[string]Service {
	return config.Services
}

func ReadConfigMQTT(config *Config) MQTTTrigger {
	return config.Triggers.MQTT
}

func ReadConfigWebhook(config *Config) WebhookTrigger {
	return config.Triggers.Webhook
}

func ReadConfigDatabasePersistence(config *Config) bool {
	if config.Persist_db == nil {
		return false
	}
	return *config.Persist_db
}
