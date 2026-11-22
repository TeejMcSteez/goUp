package utils

import (
	"os"
	"testing"
)

func TestConfigCloneIsDeep(t *testing.T) {
	broker := "mqtt://old"
	desc := "old desc"
	codes := []string{"200"}
	orig := &Config{
		Services: map[string]Service{
			"a": {Name: "a", Description: &desc, Valid_Responses: &codes},
		},
		Triggers: Trigger{MQTT: MQTTTrigger{Mqtt_broker: &broker}},
	}
	orig.Triggers.bindNotifiers()

	c := orig.Clone()

	// mutate clone through every layer of pointers
	*c.Triggers.MQTT.Mqtt_broker = "mqtt://new"
	*c.Services["a"].Description = "new desc"
	(*c.Services["a"].Valid_Responses)[0] = "500"

	if *orig.Triggers.MQTT.Mqtt_broker != "mqtt://old" {
		t.Errorf("trigger pointer shared: %s", *orig.Triggers.MQTT.Mqtt_broker)
	}
	if *orig.Services["a"].Description != "old desc" {
		t.Errorf("service pointer shared")
	}
	if codes[0] != "200" {
		t.Errorf("valid_responses backing array shared")
	}

	// notifiers must reference the clone's own fields, not the original's
	if len(c.Triggers.notifiers) != 1 {
		t.Fatalf("want 1 notifier, got %d", len(c.Triggers.notifiers))
	}
	if c.Triggers.notifiers[0] != &c.Triggers.MQTT {
		t.Errorf("clone notifier points outside clone")
	}
	if c.Triggers.notifiers[0] == orig.Triggers.notifiers[0] {
		t.Errorf("clone notifier aliases original")
	}
}

func TestReloadSkipsOwnWritesAndAppliesExternalEdits(t *testing.T) {
	path := t.TempDir() + "/cfg.yml"
	if err := os.WriteFile(path, []byte("db_path: a.db\nservices:\n  one:\n    url: https://one.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := NewConfigStore(path)
	if err := s.LoadFromFile(); err != nil {
		t.Fatal(err)
	}
	var fired int
	s.Subscribe(func(_, _ *Config) { fired++ })

	// the file now holds exactly what LoadFromFile wrote -> own echo
	if changed, err := s.Reload(); err != nil || changed {
		t.Fatalf("own write: changed=%v err=%v, want false/nil", changed, err)
	}

	// program write through Update is also an echo for the reloader
	if err := s.Update(func(c *Config) error {
		c.Services["two"] = Service{Name: "two", URL: "https://two.example.com"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	fired = 0
	if changed, err := s.Reload(); err != nil || changed || fired != 0 {
		t.Fatalf("after Update: changed=%v err=%v fired=%d, want false/nil/0", changed, err, fired)
	}

	// external edit is published, names filled in, subscribers run
	if err := os.WriteFile(path, []byte("db_path: b.db\nservices:\n  ext:\n    url: https://ext.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	changed, err := s.Reload()
	if err != nil || !changed {
		t.Fatalf("external edit: changed=%v err=%v, want true/nil", changed, err)
	}
	cfg := s.Get()
	if *cfg.Database_Location != "b.db" || cfg.Services["ext"].Name != "ext" || len(cfg.Services) != 1 {
		t.Errorf("reloaded config not applied: %+v", cfg)
	}
	if fired != 1 {
		t.Errorf("subscribers fired %d times, want 1", fired)
	}

	// broken YAML keeps the last good config
	if err := os.WriteFile(path, []byte("services: [::"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Reload(); err == nil {
		t.Error("want parse error for invalid YAML")
	}
	if s.Get() != cfg {
		t.Error("invalid YAML replaced the published config")
	}
}
