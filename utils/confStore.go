package utils

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"

	"gopkg.in/yaml.v3"
)

type ConfigStore struct {
	cur         atomic.Pointer[Config] // readers: lock-free, always a consistent snapshot
	wmu         sync.Mutex             // serialized writers only
	path        string
	subs        []func(old, new *Config) // Setup, GC, scheduler reschedule, ws broadcast
	lastWritten [32]byte
}

func NewConfigStore(path string) *ConfigStore {
	return &ConfigStore{
		cur:         atomic.Pointer[Config]{},
		wmu:         sync.Mutex{},
		path:        path,
		subs:        []func(old, new *Config){},
		lastWritten: [32]byte{},
	}
}

// Path returns the config file this store reads from and writes to.
func (s *ConfigStore) Path() string { return s.path }

func defaultConfig(path string) *Config {
	defaultDB := "goup.db"
	defaultSize := "5mb"
	defaultPersist := true
	return &Config{
		ConfigPath:        path,
		Services:          make(map[string]Service),
		Database_Location: &defaultDB,
		Database_Max_Size: &defaultSize,
		Schedule:          &ScheduleState{Span: 30, Interval: "seconds"},
		Persist_db:        &defaultPersist,
	}
}

// parseConfig unmarshals YAML into a ready-to-publish Config: path set,
// service names filled from their map keys, and triggers derived.
func parseConfig(path string, data []byte) (*Config, error) {
	cfg := new(Config)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	cfg.ConfigPath = path
	for key, svc := range cfg.Services {
		svc.Name = key
		cfg.Services[key] = svc
	}
	SetupTrigger(cfg)
	return cfg, nil
}

// LoadFromFile reads the config file (writing a default one if missing),
// normalises it back to disk and publishes it. Subscribers are NOT run;
// use Reload for that.
func (s *ConfigStore) LoadFromFile() error {
	s.wmu.Lock()
	defer s.wmu.Unlock()

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		cfg := defaultConfig(s.path)
		SetupTrigger(cfg)
		if err := s.persist(cfg); err != nil {
			return fmt.Errorf("error creating default config: %w", err)
		}
		s.cur.Store(cfg)
		return nil
	}
	if err != nil {
		return err
	}

	cfg, err := parseConfig(s.path, data)
	if err != nil {
		return err
	}
	if err := s.persist(cfg); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	s.cur.Store(cfg)
	return nil
}

// Reload re-reads the file after an external edit, publishes it and runs
// subscribers. Returns false without doing anything when the file content
// is exactly what this store last wrote (its own echo).
func (s *ConfigStore) Reload() (bool, error) {
	s.wmu.Lock()
	defer s.wmu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		return false, err
	}
	sum := sha256.Sum256(data)
	if sum == s.lastWritten {
		return false, nil
	}
	next, err := parseConfig(s.path, data)
	if err != nil {
		return false, err
	}
	s.lastWritten = sum
	old := s.cur.Load()
	s.cur.Store(next)
	if old == nil {
		return true, nil
	}
	for _, sub := range s.subs {
		sub(old, next)
	}
	return true, nil
}

// Observers subscribe handling there own changes from old->new
// Subscribers MUST not call Update as this will deadlock
// Subs are only readers
// TODO: consider moving subs out of the lock to avoid blocking
func (s *ConfigStore) Subscribe(fn func(old, new *Config)) {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	s.subs = append(s.subs, fn)
}

// handles storing on disk and checksum gen
// ensuring future updates aren't duplicate
func (s *ConfigStore) persist(c *Config) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// Write to a temp file in the same dir, then rename over the real one.
	// Rename is atomic on the same filesystem, so a reader sees either the
	// old file or the new one, never a half-written one.
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".config-*.yml")
	if err != nil {
		return err
	}
	defer func(tmp *os.File) {
		if err := os.Remove(tmp.Name()); err != nil && !os.IsNotExist(err) {
			slog.Error("failed to remove temp name persisting config", "ERROR", err)
		}
	}(tmp) // no-op after a successful rename

	if _, err := tmp.Write(data); err != nil {
		if err := tmp.Close(); err != nil {
			slog.Error("failed to close temp file persisting config", "ERROR", err)
		}
		return err
	}
	if err := tmp.Sync(); err != nil { // flush to disk before the rename
		if err := tmp.Close(); err != nil {
			slog.Error("failed to close temp file persisting config", "ERROR", err)
		}
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return err
	}

	// Remember what we wrote so the hot reloader can skip its own echo.
	sum := sha256.Sum256(data)
	s.lastWritten = sum
	return nil
}

// clonePtr returns a pointer to a fresh copy of *p (nil-safe).
func clonePtr[T any](p *T) *T {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

// cloneSlicePtr copies both the slice header AND its backing array.
func cloneSlicePtr[T any](p *[]T) *[]T {
	if p == nil {
		return nil
	}
	v := slices.Clone(*p)
	return &v
}

// Value receivers: `s` is already a shallow copy (including unexported
// fields like backoffDuration), so only pointer fields need replacing.

func (s Service) Clone() Service {
	s.Description = clonePtr(s.Description)
	s.API_URL = clonePtr(s.API_URL)
	s.API_KEY = clonePtr(s.API_KEY)
	s.Valid_Responses = cloneSlicePtr(s.Valid_Responses)
	s.Retry_Requests = clonePtr(s.Retry_Requests)
	s.Active = clonePtr(s.Active)
	s.SkipInsecure = clonePtr(s.SkipInsecure)
	return s
}

func (m MQTTTrigger) clone() MQTTTrigger {
	m.Mqtt_broker = clonePtr(m.Mqtt_broker)
	m.Mqtt_username = clonePtr(m.Mqtt_username)
	m.Mqtt_key = clonePtr(m.Mqtt_key)
	m.Backoff_Period = clonePtr(m.Backoff_Period)
	return m
}

func (w WebhookTrigger) clone() WebhookTrigger {
	w.Webhook_url = clonePtr(w.Webhook_url)
	w.Webhook_key_string = clonePtr(w.Webhook_key_string)
	w.Custom_message = clonePtr(w.Custom_message)
	w.Backoff_Period = clonePtr(w.Backoff_Period)
	return w
}

func (e SMTPTrigger) clone() SMTPTrigger {
	e.Email = clonePtr(e.Email)
	e.App_Password = clonePtr(e.App_Password)
	e.SMTPServer = clonePtr(e.SMTPServer)
	e.Backoff_Period = clonePtr(e.Backoff_Period)
	return e
}

func (g GotifyTrigger) clone() GotifyTrigger {
	g.Gotify_Server = clonePtr(g.Gotify_Server)
	g.Gotify_Token = clonePtr(g.Gotify_Token)
	g.Gotify_Application = clonePtr(g.Gotify_Application)
	g.Gotify_Title = clonePtr(g.Gotify_Title)
	g.Gotify_Priority = clonePtr(g.Gotify_Priority)
	g.Backoff_Period = clonePtr(g.Backoff_Period)
	return g
}

func (s SlackTrigger) clone() SlackTrigger {
	s.Slack_Token = clonePtr(s.Slack_Token)
	s.Slack_Channel = clonePtr(s.Slack_Channel)
	s.Bot_Username = clonePtr(s.Bot_Username)
	s.Backoff_Period = clonePtr(s.Backoff_Period)
	return s
}

func (t TelegramTrigger) clone() TelegramTrigger {
	t.Telegram_Token = clonePtr(t.Telegram_Token)
	t.Telegram_Channel_Id = clonePtr(t.Telegram_Channel_Id)
	t.Backoff_Period = clonePtr(t.Backoff_Period)
	return t
}

func (h HATrigger) clone() HATrigger {
	h.HA_Token = clonePtr(h.HA_Token)
	h.HA_URL = clonePtr(h.HA_URL)
	h.Backoff_Period = clonePtr(h.Backoff_Period)
	return h
}

func (d DiscordTrigger) clone() DiscordTrigger {
	d.Discord_Auth = clonePtr(d.Discord_Auth)
	d.Discord_Channel = clonePtr(d.Discord_Channel)
	d.Backoff_Period = clonePtr(d.Backoff_Period)
	return d
}

// clone deep-copies every sub-trigger. It does NOT bind notifiers: the
// returned value is copied again on assignment, so &n.MQTT here would point
// at a temporary. Call bindNotifiers() on the final location instead.
func (t Trigger) clone() Trigger {
	t.MQTT = t.MQTT.clone()
	t.Webhook = t.Webhook.clone()
	t.SMTP = t.SMTP.clone()
	t.Gotify = t.Gotify.clone()
	t.Slack = t.Slack.clone()
	t.Telegram = t.Telegram.clone()
	t.HA = t.HA.clone()
	t.Discord = t.Discord.clone()
	t.Backoff_Period = clonePtr(t.Backoff_Period)
	t.notifiers = nil
	return t
}

// Deep-copy current config returned as a pointer to config.
// Mutating the clone must never be visible through the original.
func (s *Config) Clone() *Config {
	n := *s
	n.Database_Location = clonePtr(s.Database_Location)
	n.Database_Max_Size = clonePtr(s.Database_Max_Size)
	n.Persist_db = clonePtr(s.Persist_db)
	n.Schedule = clonePtr(s.Schedule)

	if s.Services != nil {
		n.Services = make(map[string]Service, len(s.Services))
		for k, svc := range s.Services {
			n.Services[k] = svc.Clone()
		}
	}

	n.Triggers = s.Triggers.clone()
	n.Triggers.bindNotifiers() // point notifiers at n's fields, not s's
	return &n
}

// return current config (atomic operation)
func (s *ConfigStore) Get() *Config { return s.cur.Load() }

func (s *ConfigStore) Update(fn func(c *Config) error) error {
	s.wmu.Lock()
	defer s.wmu.Unlock()

	old := s.cur.Load()
	if old == nil {
		return fmt.Errorf("config store not loaded")
	}
	next := old.Clone()
	if err := fn(next); err != nil {
		return err
	}
	// re-derive trigger runtime state (backoff durations, notifiers) on the
	// private clone so every writer gets it, not just HTTP handlers
	SetupTrigger(next)

	if err := s.persist(next); err != nil {
		return err
	}
	s.cur.Store(next)
	for _, sub := range s.subs {
		sub(old, next)
	}
	return nil
}
