package utils

import (
	"database/sql"
	"fmt"
	"log/slog"
	"maps"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func UpdateDatabaseSize(conf *Config, newSize string) (*Config, error) {
	_, err := GetSizeFromString(newSize)
	if err != nil {
		return nil, err
	}
	if conf.Database_Max_Size != nil && newSize == *conf.Database_Max_Size {
		return nil, fmt.Errorf("new size is equal to current size no change needed")
	}
	conf.Database_Max_Size = &newSize
	return conf, nil
}

// Returns max size set in config in bytes or 1 gigabyte as default size
func GetMaxSize(cfg *Config) (float64, error) {
	if cfg != nil {
		if cfg.Database_Max_Size != nil {
			str := *cfg.Database_Max_Size
			size, err := GetSizeFromString(str)
			if err != nil {
				return 0, fmt.Errorf("incorrect format, must be <decimal><a-zA-Z>")
			}
			return size, nil
		}
	}
	slog.Warn("Database max size not set, defaulting to 1GB max size")
	return GB, nil
}

func GetSizeFromString(str string) (float64, error) {
	re, err := regexp.Compile(`^(\d+)([a-zA-Z]+)$`)
	if err != nil {
		return 0, err
	}
	matches := re.FindStringSubmatch(str)

	if len(matches) != 3 {
		slog.Warn("Invalid Database_Max_Size format, defaulting to 1GB", "value", str)
		return GB, fmt.Errorf("invalid database_max_size format: %s", str)
	}

	number, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		slog.Warn("Failed to parse number from Database_Max_Size, defaulting to 1GB", "error", err)
		return GB, fmt.Errorf("failed to parse number from database_max_size: %w", err)
	}
	sizeUnit := strings.ToLower(matches[2])

	switch sizeUnit {
	case "kb":
		if number < 4 {
			slog.Warn("Database size must be at least 4KB, returning 4KB")
			return 4 * 1000, fmt.Errorf("database size must be at least 4KB")
		}
		slog.Info("Set max database size", "size", number, "unit", sizeUnit)
		return number * 1000, nil
	case "mb":
		slog.Info("Set max database size", "size", number, "unit", sizeUnit)
		return number * 1e6, nil
	case "gb":
		slog.Info("Set max database size", "size", number, "unit", sizeUnit)
		return number * 1e9, nil
	default:
		slog.Warn("Invalid size unit, defaulting to 1GB", "unit", sizeUnit)
		return GB, fmt.Errorf("invalid size unit \"%s\": Defaulting to 1GB", sizeUnit)
	}
}

func AddConfigService(conf *Config, newEndpoint Service) (*Config, error) {
	if conf == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if conf.Services == nil {
		conf.Services = make(map[string]Service)
	}
	org_len := len(conf.Services)
	conf.Services[newEndpoint.Name] = newEndpoint
	new_len := len(conf.Services)
	if new_len == org_len {
		return nil, fmt.Errorf("no new endpoint added")
	}
	return conf, nil
}

// ServiceURLExists reports whether a service with the given URL is already
// configured. Callers must not range over conf.Services directly since the
// map can be mutated concurrently by the config HTTP handlers.
func ServiceURLExists(conf *Config, url string) bool {
	for _, svc := range conf.Services {
		if strings.TrimSuffix(svc.URL, "/") == strings.TrimSuffix(url, "/") {
			return true
		}
	}
	return false
}

func setTrigger[T comparable](conf *Config, field *T, newVal T, name string) (*Config, error) {
	if conf == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if *field == newVal {
		return nil, fmt.Errorf("%s config is the same", name)
	}
	*field = newVal
	return conf, nil
}

func AddConfigMQTTTrigger(config *Config, newMQTT MQTTTrigger) (*Config, error) {
	return setTrigger(config, &config.Triggers.MQTT, newMQTT, "MQTT")
}

func AddConfigWebhookTrigger(config *Config, newWebhook WebhookTrigger) (*Config, error) {
	return setTrigger(config, &config.Triggers.Webhook, newWebhook, "webhook")
}

func AddConfigSMTPTrigger(config *Config, newSmtp SMTPTrigger) (*Config, error) {
	return setTrigger(config, &config.Triggers.SMTP, newSmtp, "SMTP")
}

func AddConfigGotifyTrigger(config *Config, newGotify GotifyTrigger) (*Config, error) {
	return setTrigger(config, &config.Triggers.Gotify, newGotify, "gotify")
}

func AddConfigSlackTrigger(config *Config, newSlack SlackTrigger) (*Config, error) {
	return setTrigger(config, &config.Triggers.Slack, newSlack, "slack")
}

func AddConfigTelegramTrigger(config *Config, newTelegram TelegramTrigger) (*Config, error) {
	return setTrigger(config, &config.Triggers.Telegram, newTelegram, "telegram")
}

func AddConfigHATrigger(config *Config, newHA HATrigger) (*Config, error) {
	return setTrigger(config, &config.Triggers.HA, newHA, "ha")
}

func AddConfigDiscordTrigger(config *Config, newDiscord DiscordTrigger) (*Config, error) {
	return setTrigger(config, &config.Triggers.Discord, newDiscord, "discord")
}

func UpdateConfigService(conf *Config, oldName string, updated Service, db *sql.DB) (*Config, error) {
	if conf == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if _, exists := conf.Services[oldName]; !exists {
		return nil, fmt.Errorf("service %q not found", oldName)
	}
	if oldName != updated.Name {
		if err := DbServiceRename(db, oldName, updated.Name); err != nil {
			return nil, fmt.Errorf("failed to rename service in database: %w", err)
		}
	}
	delete(conf.Services, oldName)

	conf.Services[updated.Name] = updated
	return conf, nil
}

func DeleteConfigService(config *Config, serviceToDelete Service, db *sql.DB) (*Config, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	org_size := len(config.Services)
	delete(config.Services, serviceToDelete.Name)
	new_size := len(config.Services)

	if org_size == new_size {
		return nil, fmt.Errorf("failed to remove the element")
	}

	return config, nil
}

func DeleteConfigMQTT(config *Config) (*Config, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.Triggers.MQTT.Mqtt_broker == nil {
		return nil, fmt.Errorf("no MQTT broker to delete")
	}
	// Clears MQTT trigger in config
	config.Triggers.MQTT = MQTTTrigger{}
	return config, nil
}

func DeleteConfigWebhook(config *Config) (*Config, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.Triggers.Webhook.Webhook_url == nil {
		return nil, fmt.Errorf("no Webhook to delete")
	}

	config.Triggers.Webhook = WebhookTrigger{}
	return config, nil
}

func DeleteConfigSMTPTrigger(config *Config) (*Config, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.Triggers.SMTP.SMTPServer == nil || config.Triggers.SMTP.App_Password == nil || config.Triggers.SMTP.Email == nil {
		return nil, fmt.Errorf("no SMTP config to delete")
	}
	config.Triggers.SMTP = SMTPTrigger{}
	return config, nil
}

func DeleteConfigGotifyTrigger(config *Config) (*Config, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.Triggers.Gotify.Gotify_Server == nil {
		return nil, fmt.Errorf("no Gotify config to delete")
	}
	config.Triggers.Gotify = GotifyTrigger{}
	return config, nil
}

func DeleteConfigSlackTrigger(config *Config) (*Config, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.Triggers.Slack.Slack_Channel == nil && config.Triggers.Slack.Slack_Token == nil {
		return nil, fmt.Errorf("no slack config to delete")
	}
	config.Triggers.Slack = SlackTrigger{}
	return config, nil
}

func DeleteConfigTelegramTrigger(config *Config) (*Config, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.Triggers.Telegram.Telegram_Token == nil && config.Triggers.Telegram.Telegram_Channel_Id == nil {
		return nil, fmt.Errorf("no telegram config to delete")
	}
	config.Triggers.Telegram = TelegramTrigger{}
	return config, nil
}

func DeleteConfigHATrigger(config *Config) (*Config, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.Triggers.HA.HA_URL == nil && config.Triggers.HA.HA_Token == nil {
		return nil, fmt.Errorf("no HA config to delete")
	}
	config.Triggers.HA = HATrigger{}
	return config, nil
}

func DeleteConfigDiscordTrigger(config *Config) (*Config, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.Triggers.Discord.Discord_Auth == nil && config.Triggers.Discord.Discord_Channel == nil {
		return nil, fmt.Errorf("no discord config to delete")
	}
	config.Triggers.Discord = DiscordTrigger{}
	return config, nil
}

// ReadConfigServices returns a snapshot copy of the configured services.
// It copies rather than returning the live map so callers can range over
// the result without racing the config HTTP handlers' writes.
func ReadConfigServices(config *Config) map[string]Service {
	out := make(map[string]Service, len(config.Services))
	maps.Copy(out, config.Services)
	return out
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

func ReadConfigSlack(config *Config) SlackTrigger {
	return config.Triggers.Slack
}

func ReadConfigTelegram(config *Config) TelegramTrigger {
	return config.Triggers.Telegram
}

func ReadConfigHA(config *Config) HATrigger {
	return config.Triggers.HA
}

func ReadConfigDiscord(config *Config) DiscordTrigger {
	return config.Triggers.Discord
}

func ReadConfigDatabasePersistence(config *Config) bool {
	if config.Persist_db == nil {
		return false
	}
	return *config.Persist_db
}

func ReadDatabaseSize(config *Config) (string, error) {
	if config.Database_Max_Size == nil {
		return "", fmt.Errorf("current database size is not set")
	}
	return *config.Database_Max_Size, nil
}

func UpdateConfigSchedule(conf *Config, state ScheduleState) (*Config, error) {
	if conf == nil {
		return nil, fmt.Errorf("config is nil")
	}
	conf.Schedule = &ScheduleState{Span: state.Span, Interval: state.Interval}
	return conf, nil
}

func UpdateConfigDatabasePersistence(conf *Config) (*Config, error) {
	if conf == nil {
		return nil, fmt.Errorf("Config is nil")
	}

	org_state := conf.Persist_db
	if org_state == nil {
		b := true
		conf.Persist_db = &b
	} else {
		*conf.Persist_db = !*org_state
	}

	return conf, nil
}

// UpdateBackoff sets the backoff period for a single trigger, or the global
// fallback when trigger is "" or "global". A blank newTime clears the period
// so the trigger falls back to the global default (or disables backoff, for
// the global period itself).
func UpdateBackoff(conf *Config, newTime string, trigger string) (*Config, error) {
	if conf == nil {
		return nil, fmt.Errorf("config is nil")
	}

	var period *string
	var dur time.Duration
	if newTime != "" {
		parsed, err := time.ParseDuration(newTime)
		if err != nil {
			return nil, err
		}
		period = &newTime
		dur = parsed
	}

	switch trigger {
	case "mqtt":
		conf.Triggers.MQTT.Backoff_Period = period
		conf.Triggers.MQTT.backoffDuration = dur
	case "webhook":
		conf.Triggers.Webhook.Backoff_Period = period
		conf.Triggers.Webhook.backoffDuration = dur
	case "smtp":
		conf.Triggers.SMTP.Backoff_Period = period
		conf.Triggers.SMTP.backoffDuration = dur
	case "gotify":
		conf.Triggers.Gotify.Backoff_Period = period
		conf.Triggers.Gotify.backoffDuration = dur
	case "slack":
		conf.Triggers.Slack.Backoff_Period = period
		conf.Triggers.Slack.backoffDuration = dur
	case "telegram":
		conf.Triggers.Telegram.Backoff_Period = period
		conf.Triggers.Telegram.backoffDuration = dur
	case "ha":
		conf.Triggers.HA.Backoff_Period = period
		conf.Triggers.HA.backoffDuration = dur
	case "discord":
		conf.Triggers.Discord.Backoff_Period = period
		conf.Triggers.Discord.backoffDuration = dur
	case "", "global":
		conf.Triggers.Backoff_Period = period
		conf.Triggers.backoffDuration = dur
	default:
		return nil, fmt.Errorf("unknown trigger %q", trigger)
	}

	return conf, nil
}

// ReadConfigBackoff returns the global trigger backoff period, or "" if unset.
func ReadConfigBackoff(conf *Config) string {
	if conf.Triggers.Backoff_Period == nil {
		return ""
	}
	return *conf.Triggers.Backoff_Period
}

func ReadConfigService(conf *Config, service Service) (Service, error) {
	if conf == nil {
		return Service{}, fmt.Errorf("config is nil")
	}

	s, ok := conf.Services[service.Name]
	if !ok {
		slog.Error("invalid service", "service", service.Name)
		return Service{}, fmt.Errorf("invalid service %s", service.Name)
	}
	return s, nil
}

func UpdateConfigServiceActive(conf *Config, service Service) (*Config, error) {
	if conf == nil {
		return nil, fmt.Errorf("config is nil")
	}

	s, ok := conf.Services[service.Name]
	if !ok {
		slog.Error("invalid service", "service", service.Name)
		return nil, fmt.Errorf("invalid service %s", service.Name)
	}

	if s.Active != nil {
		active := !s.IsActive()
		s.Active = &active
		conf.Services[service.Name] = s
	} else {
		// defaults to true so on nil create and assigns to false
		a := false
		s.Active = &a
		conf.Services[service.Name] = s
	}

	return conf, nil
}
