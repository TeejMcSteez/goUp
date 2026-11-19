package utils

import (
	"crypto/sha256"
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
	defer os.Remove(tmp.Name()) // no-op after a successful rename

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil { // flush to disk before the rename
		tmp.Close()
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

func (c *Config) Clone() *Config {
	n := c
	n.Services = maps.Clone(c.Services)
	return n
}

func (s *ConfigStore) Get() *Config { return s.cur.Load() }

func (s *ConfigStore) Update(fn func(c *Config) error) error {
	s.wmu.Lock()
	defer s.wmu.Unlock()

	old := s.cur.Load()
	next := old.Clone()
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
