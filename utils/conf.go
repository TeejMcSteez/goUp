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
	if err != nil {
		return nil, err
	}

	var cfg = new(Config)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func AddConfigService(conf *Config, newEndpoint Service) error {
	if conf == nil {
		return fmt.Errorf("config is nil")
	}
	org_len := len(conf.Services)
	conf.Services[newEndpoint.URL] = newEndpoint
	new_len := len(conf.Services)
	if new_len == org_len {
		return fmt.Errorf("no new endpoint added")
	}
	if err := os.Remove("./services.yml"); err != nil {
		return err
	}
	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	// TODO: Change from full r/w/x to something more sensible
	if err := os.WriteFile("./services.yml", data, 0777); err != nil {
		return err
	}
	return nil
}

func AddConfigMQTTTrigger(config *Config, newMQTT MQTTTrigger) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	if newMQTT == config.Triggers.MQTT {
		return fmt.Errorf("MQTT trigger is the same")
	}
	config.Triggers.MQTT = newMQTT
	if err := os.Remove("./services.yml"); err != nil {
		return err
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.WriteFile("./services.yml", data, 0777); err != nil {
		return err
	}
	return nil
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
	if err := os.Remove("./services.yml"); err != nil {
		return err
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.WriteFile("./services.yml", data, 0777); err != nil {
		return err
	}
	return nil
}

func DeleteConfigService(config *Config, serviceToDelete Service) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	org_size := len(config.Services)
	delete(config.Services, serviceToDelete.URL)
	new_size := len(config.Services)

	if org_size == new_size {
		return fmt.Errorf("failed to remove the element")
	}

	if err := os.Remove("./services.yml"); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := os.WriteFile("./services.yml", data, 0777); err != nil {
		return err
	}

	return nil
}

func DeleteConfigMQTT(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	// Clears MQTT trigger in config
	config.Triggers.MQTT = MQTTTrigger{}

	if err := os.Remove("./services.yml"); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := os.WriteFile("./services.yml", data, 0777); err != nil {
		return err
	}
	return nil
}

func DeleteConfigTrigger(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	config.Triggers.Webhook = WebhookTrigger{}

	if err := os.Remove("./services.yml"); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := os.WriteFile("./services.yml", data, 0777); err != nil {
		return err
	}

	return nil
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
