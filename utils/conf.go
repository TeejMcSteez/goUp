package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var Current_Config *Config

// programmaticWrite is signalled by writeConfig so the hot reloader can
// distinguish program-initiated writes from external file edits.
var programmaticWrite = make(chan struct{}, 1)

// ConfigWriteNotify returns the channel the hot reloader should listen on.
func ConfigWriteNotify() <-chan struct{} {
	return programmaticWrite
}

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
	select {
	case programmaticWrite <- struct{}{}:
	default:
	}
	return nil
}

func UpdateDatabaseSize(conf *Config, newSize string) error {
	_, err := GetSizeFromString(newSize)
	if err != nil {
		return err
	}
	if conf.Database_Max_Size != nil && newSize == *conf.Database_Max_Size {
		return fmt.Errorf("new size is equal to current size no change needed")
	}
	conf.Database_Max_Size = &newSize
	return writeConfig(conf)
}

func GetSizeFromString(str string) (float64, error) {

	re, err := regexp.Compile(`^(\d+)([a-zA-Z]+)$`)
	if err != nil {
		return 0, err
	}
	matches := re.FindStringSubmatch(str)

	if len(matches) != 3 {
		log.Printf("Invalid Database_Max_Size format: %s. Defaulting to 1GB.", str)
		return GB, fmt.Errorf("invalid database_max_size format: %s", str)
	}

	number, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		log.Printf("Failed to parse number from Database_Max_Size: %v. Defaulting to 1GB.", err)
		return GB, fmt.Errorf("failed to parse number from database_max_size: %w", err)
	}
	sizeUnit := strings.ToLower(matches[2])

	switch sizeUnit {
	case "kb":
		if number < 4 {
			log.Print("Database size must be at least 4KB, returning 4KB.\n")
			return 4 * 1000, fmt.Errorf("database size must be at least 4KB")
		}
		log.Printf("Set max database size to: %v%v", number, sizeUnit)
		return number * 1000, nil
	case "mb":
		log.Printf("Set max database size to: %v%v", number, sizeUnit)
		return number * 1e6, nil
	case "gb":
		log.Printf("Set max database size to: %v%v", number, sizeUnit)
		return number * 1e9, nil
	default:
		log.Printf("Invalid size unit: %s. Defaulting to 1GB.", sizeUnit)
		return GB, fmt.Errorf("invalid size unit \"%s\": Defaulting to 1GB", sizeUnit)
	}
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

func AddConfigSMTPTrigger(config *Config, newSmtp SMTPTrigger) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	if newSmtp == config.Triggers.SMTP {
		return fmt.Errorf("SMTP config is the same")
	}
	config.Triggers.SMTP = newSmtp
	return writeConfig(config)
}

func AddConfigGotifyTrigger(config *Config, newGotify GotifyTrigger) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	if newGotify == config.Triggers.Gotify {
		return fmt.Errorf("gotify config is the same")
	}
	config.Triggers.Gotify = newGotify
	return writeConfig(config)
}

func DeleteConfigSMTPTrigger(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	if config.Triggers.SMTP.SMTPServer == nil || config.Triggers.SMTP.App_Password == nil || config.Triggers.SMTP.Email == nil {
		return fmt.Errorf("no SMTP config to delete")
	}
	config.Triggers.SMTP = SMTPTrigger{}
	return writeConfig(config)
}

func DeleteConfigGotifyTrigger(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	if config.Triggers.Gotify.Gotify_Server == nil {
		return fmt.Errorf("no Gotify config to delete")
	}
	config.Triggers.Gotify = GotifyTrigger{}
	return writeConfig(config)
}

func UpdateConfigService(conf *Config, oldName string, updated Service, db *sql.DB) error {
	if conf == nil {
		return fmt.Errorf("config is nil")
	}
	if _, exists := conf.Services[oldName]; !exists {
		return fmt.Errorf("service %q not found", oldName)
	}
	if oldName != updated.Name {
		if err := DbServiceRename(db, oldName, updated.Name); err != nil {
			return fmt.Errorf("failed to rename service in database: %w", err)
		}
		delete(conf.Services, oldName)
	}
	conf.Services[updated.Name] = updated
	return writeConfig(conf)
}

func DeleteConfigService(config *Config, serviceToDelete Service, db *sql.DB) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	err := DbServiceDelete(db, serviceToDelete)
	if err != nil {
		return fmt.Errorf("failed to remove data from database: %v", err)
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
	if config.Triggers.MQTT.Mqtt_broker == nil {
		return fmt.Errorf("no MQTT broker to delete")
	}
	// Clears MQTT trigger in config
	config.Triggers.MQTT = MQTTTrigger{}
	return writeConfig(config)
}

func DeleteConfigTrigger(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}
	if config.Triggers.Webhook.Webhook_url == nil {
		return fmt.Errorf("no Webhook to delete")
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

func ReadConfigSMTP(config *Config) SMTPTrigger {
	return config.Triggers.SMTP
}

func ReadConfigGotify(config *Config) GotifyTrigger {
	return config.Triggers.Gotify
}

func ReadConfigDatabasePersistence(config *Config) bool {
	if config.Persist_db == nil {
		return false
	}
	return *config.Persist_db
}

func ReadDatabaseSize(config *Config) (string, error) {
	if config.Persist_db == nil {
		return "", fmt.Errorf("current database size is not set")
	}
	return *config.Database_Max_Size, nil
}

func UpdateConfigSchedule(conf *Config, state ScheduleState) error {
	if conf == nil {
		return fmt.Errorf("config is nil")
	}
	conf.Schedule = &ScheduleState{Span: state.Span, Interval: state.Interval}
	return writeConfig(conf)
}

func UpdateConfigDatabasePersistence(conf *Config) error {
	if conf == nil {
		return fmt.Errorf("Config is nil")
	}

	org_state := conf.Persist_db
	if org_state == nil {
		b := true
		conf.Persist_db = &b
	} else {
		*conf.Persist_db = !*org_state
	}

	return writeConfig(conf)
}
