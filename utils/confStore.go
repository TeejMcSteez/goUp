package utils

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
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

func (s *ConfigStore) LoadFromFile() error {
	data, err := os.ReadFile(s.path)
	// handle no config -> write default
	if os.IsNotExist(err) {
		defaultDB := "goup.db"
		default_size := "5mb"
		st := ScheduleState{
			30,
			"seconds",
		}
		default_persist := true
		cfg := &Config{
			ConfigPath:        s.path,
			Services:          make(map[string]Service),
			Database_Location: &defaultDB,
			Database_Max_Size: &default_size,
			Schedule:          &st,
			Persist_db:        &default_persist,
		}
		if err := s.persist(cfg); err != nil {
			return fmt.Errorf("error creating default config: %w", err)
		}
		s.cur.Store(cfg)

		return nil
	}
	// non-exist error return
	if err != nil {
		return err
	}

	var cfg = new(Config)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return err
	}

	cfg.ConfigPath = s.path

	for key, svc := range cfg.Services {
		svc.Name = key
		cfg.Services[key] = svc
	}

	if err := s.persist(cfg); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	s.cur.Store(cfg)

	return nil
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

// Deep-copy current config returned as a pointer to config
func (s *Config) Clone() *Config {
	n := *s
	n.Services = maps.Clone(s.Services)
	if s.Schedule != nil {
		s := *s.Schedule
		n.Schedule = &s
	}
	if s.Database_Location != nil {
		s := *s.Database_Location
		n.Database_Location = &s
	}
	// TODO: handle rest of pointers with deep copy
	return &n
}

// return current config (atomic operation)
func (s *ConfigStore) Get() *Config { return s.cur.Load() }

func (s *ConfigStore) Update(fn func(c *Config) error) error {
	s.wmu.Lock()
	defer s.wmu.Unlock()

	old := s.cur.Load()
	next := old.Clone()
	// runs all observer functions
	if err := fn(next); err != nil {
		return err
	}

	if err := s.persist(next); err != nil {
		return err
	}
	s.cur.Store(next)
	for _, sub := range s.subs {
		sub(old, next)
	}
	return nil
}
