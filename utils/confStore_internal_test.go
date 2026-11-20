package utils

import "testing"

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
