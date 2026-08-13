package utils

import (
	"log/slog"
	"time"
)

// resolveBackoff parses a trigger-specific backoff period, falling back to the
// global trigger backoff when the trigger doesn't define its own.
func resolveBackoff(name string, period *string, fallback time.Duration) time.Duration {
	if period == nil || *period == "" {
		return fallback
	}
	dur, err := time.ParseDuration(*period)
	if err != nil {
		slog.Warn("Invalid trigger backoff period, falling back to default", "period", *period, "trigger", name, "error", err)
		return fallback
	}
	slog.Info("Backoff period set", "trigger", name, "duration", dur)
	return dur
}

// shouldBackoff reports whether a trigger fired within its backoff window.
func shouldBackoff(name string, lastFired time.Time, backoffDuration time.Duration) bool {
	if backoffDuration > 0 && !lastFired.IsZero() && time.Since(lastFired) < backoffDuration {
		slog.Info("Trigger backoff period active, skipping", "trigger", name, "last_fired", time.Since(lastFired))
		return true
	}
	return false
}

// SetupTrigger copies Trigger config from cfg and registers configured handlers.
func SetupTrigger(cfg *Config) *Trigger {
	t := &cfg.Triggers
	t.handlers = nil

	if cfg.Triggers.Backoff_Period != nil && *cfg.Triggers.Backoff_Period != "" {
		dur, err := time.ParseDuration(*cfg.Triggers.Backoff_Period)
		if err != nil {
			slog.Warn("Invalid trigger backoff period, disabling backoff", "period", *cfg.Triggers.Backoff_Period, "error", err)
			t.backoffDuration = 0
		} else {
			slog.Info("Trigger backoff period set", "duration", dur)
			t.backoffDuration = dur
		}
	} else {
		slog.Info("No backoff period setup")
	}

	t.MQTT.backoffDuration = resolveBackoff("MQTT", t.MQTT.Backoff_Period, t.backoffDuration)
	t.Webhook.backoffDuration = resolveBackoff("Webhook", t.Webhook.Backoff_Period, t.backoffDuration)
	t.SMTP.backoffDuration = resolveBackoff("SMTP", t.SMTP.Backoff_Period, t.backoffDuration)
	t.Gotify.backoffDuration = resolveBackoff("Gotify", t.Gotify.Backoff_Period, t.backoffDuration)
	t.Slack.backoffDuration = resolveBackoff("Slack", t.Slack.Backoff_Period, t.backoffDuration)
	t.Telegram.backoffDuration = resolveBackoff("Telegram", t.Telegram.Backoff_Period, t.backoffDuration)
	t.HA.backoffDuration = resolveBackoff("Home Assistant", t.HA.Backoff_Period, t.backoffDuration)
	t.Discord.backoffDuration = resolveBackoff("Discord", t.Discord.Backoff_Period, t.backoffDuration)

	if t.MQTT.IsConfigured() {
		t.handlers = append(t.handlers, &t.MQTT)
	}
	if t.Webhook.IsConfigured() {
		t.handlers = append(t.handlers, &t.Webhook)
	}
	if t.SMTP.IsConfigured() {
		t.handlers = append(t.handlers, &t.SMTP)
	}
	if t.Gotify.IsConfigured() {
		t.handlers = append(t.handlers, &t.Gotify)
	}
	if t.Slack.IsConfigured() {
		t.handlers = append(t.handlers, &t.Slack)
	}
	if t.Telegram.IsConfigured() {
		t.handlers = append(t.handlers, &t.Telegram)
	}
	if t.HA.IsConfigured() {
		t.handlers = append(t.handlers, &t.HA)
	}
	if t.Discord.IsConfigured() {
		t.handlers = append(t.handlers, &t.Discord)
	}
	if len(t.handlers) == 0 {
		slog.Info("No MQTT broker, Webhook URL, or SMTP server setup, exiting trigger setup")
		return t
	}

	slog.Info("Triggers setup")
	return t
}

// Fire dispatches service data to all configured trigger handlers.
func (t *Trigger) Fire(data []ServiceData) {
	if len(t.handlers) == 0 {
		return
	}

	if shouldBackoff("Trigger", t.lastFired, t.backoffDuration) {
		return
	}

	for _, h := range t.handlers {
		go h.Fire(data)
	}

	t.lastFired = time.Now()
}

func (t *Trigger) Clear() {
	if len(t.handlers) == 0 {
		return
	}

	for _, h := range t.handlers {
		go h.Clear()
	}
}
