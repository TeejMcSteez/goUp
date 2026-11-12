package utils

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
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

type resCheck func(res *http.Response, body []byte) error

type reqConfig struct {
	headers map[string]string
	checks  []resCheck
}

type reqOption func(*reqConfig)

func header(k, v string) reqOption {
	return func(c *reqConfig) { c.headers[k] = v }
}

func expectStatus(codes ...int) reqOption {
	return func(c *reqConfig) {
		c.checks = append(c.checks, func(res *http.Response, _ []byte) error {
			if slices.Contains(codes, res.StatusCode) {
				return nil
			}
			return fmt.Errorf("unexpected status: %d", res.StatusCode)
		})
	}
}

func expect2xx(res *http.Response, _ []byte) error {
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("unexpected status: %d", res.StatusCode)
	}
	return nil
}

func expectOkField() reqOption {
	return func(c *reqConfig) {
		c.checks = append(c.checks, func(_ *http.Response, body []byte) error {
			var r struct {
				OK          bool   `json:"ok"`
				Error       string `json:"error"`
				Description string `json:"description"`
			}
			if err := json.Unmarshal(body, &r); err != nil {
				return fmt.Errorf("decoding ok-field response: %w", err)
			}
			if !r.OK {
				return fmt.Errorf("api error: %s", cmp.Or(r.Error, r.Description, "ok=false"))
			}
			return nil
		})
	}
}

// handles posting all data via JSON to triggers
func postJSON(ctx context.Context, url string, payload any, opts ...reqOption) error {
	cfg := reqConfig{headers: map[string]string{"Content-Type": "application/json"}}

	for _, o := range opts {
		o(&cfg)
	}

	if len(cfg.checks) == 0 {
		cfg.checks = []resCheck{expect2xx}
	}

	b, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	for k, v := range cfg.headers {
		req.Header.Set(k, v)
	}

	res, err := httpClient.Do(req) // a client with a Timeout, not DefaultClient
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	for _, check := range cfg.checks {
		if err := check(res, body); err != nil {
			return err
		}
	}
	return nil
}

// SetupTrigger copies Trigger config from cfg and registers configured notifiers.
func SetupTrigger(cfg *Config) *Trigger {
	t := &cfg.Triggers
	t.notifiers = nil

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
		t.notifiers = append(t.notifiers, &t.MQTT)
	}
	if t.Webhook.IsConfigured() {
		t.notifiers = append(t.notifiers, &t.Webhook)
	}
	if t.SMTP.IsConfigured() {
		t.notifiers = append(t.notifiers, &t.SMTP)
	}
	if t.Gotify.IsConfigured() {
		t.notifiers = append(t.notifiers, &t.Gotify)
	}
	if t.Slack.IsConfigured() {
		t.notifiers = append(t.notifiers, &t.Slack)
	}
	if t.Telegram.IsConfigured() {
		t.notifiers = append(t.notifiers, &t.Telegram)
	}
	if t.HA.IsConfigured() {
		t.notifiers = append(t.notifiers, &t.HA)
	}
	if t.Discord.IsConfigured() {
		t.notifiers = append(t.notifiers, &t.Discord)
	}
	if len(t.notifiers) == 0 {
		slog.Info("No MQTT broker, Webhook URL, or SMTP server setup, exiting trigger setup")
		return t
	}

	slog.Info("Triggers setup")
	return t
}

func (t *Trigger) Send(ctx context.Context, m Message) error {
	if len(t.notifiers) == 0 {
		return fmt.Errorf("there are no triggers")
	}
	var ers []error
	// Backoff is enforced per notifier in each Send implementation.
	for _, n := range t.notifiers {
		nctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		err := n.Send(nctx, m)
		cancel()
		if err != nil {
			ers = append(ers, err)
		}

	}
	return errors.Join(ers...)
}
