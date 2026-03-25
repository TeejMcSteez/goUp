package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SetupTrigger copies Trigger config from cfg and registers configured handlers.
func SetupTrigger(cfg *Config) *Trigger {
	t := &cfg.Triggers

	if cfg.Triggers.Backoff_Period != nil && *cfg.Triggers.Backoff_Period != "" {
		dur, err := time.ParseDuration(*cfg.Triggers.Backoff_Period)
		if err != nil {
			log.Printf("Invalid Backoff_Period '%s': %v. Disabling backoff.", *cfg.Triggers.Backoff_Period, err)
			t.backoffDuration = 0
		} else {
			log.Printf("Trigger backoff period set to %s", dur)
			t.backoffDuration = dur
		}
	} else {
		log.Println("No backup period setup!")
	}

	if t.MQTT.IsConfigured() {
		t.handlers = append(t.handlers, &t.MQTT)
	}
	if t.Webhook.IsConfigured() {
		t.handlers = append(t.handlers, &t.Webhook)
	}

	if len(t.handlers) == 0 {
		log.Println("No MQTT broker or Webhook URL setup, exiting trigger setup")
		return t
	}

	log.Println("Triggers setup")
	return t
}

// Fire dispatches service data to all configured trigger handlers.
func (t *Trigger) Fire(data []ServiceData) {
	if len(t.handlers) == 0 {
		return
	}

	if t.backoffDuration > 0 && !t.lastFired.IsZero() && time.Since(t.lastFired) < t.backoffDuration {
		log.Printf("Trigger backoff period active, skipping. Last trigger was %s ago.", time.Since(t.lastFired))
		return
	}

	for _, h := range t.handlers {
		go h.Fire(data)
	}

	t.lastFired = time.Now()
}

func (m *MQTTTrigger) IsConfigured() bool {
	return m.Mqtt_broker != nil
}

func (m *MQTTTrigger) Fire(data []ServiceData) {
	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		log.Println("Connected to MQTT Broker")
	}

	var lostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		log.Println("Connection Lost: " + err.Error())
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(*m.Mqtt_broker)
	opts.SetClientID("goUp MQTT")
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = lostHandler
	if m.Mqtt_username != nil || m.Mqtt_key != nil {
		opts.SetUsername(*m.Mqtt_username)
		opts.SetPassword(*m.Mqtt_key)
	}

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Println("Error pushing to MQTT client: " + token.Error().Error())
	}

	if jsonData, err := json.Marshal(data); err != nil {
		log.Printf("Error formatting service data into JSON: %v\n", err)
	} else {
		// Keep retain to true to store last known good message
		token := client.Publish("goup_status", 0, true, jsonData)
		token.Done()

		if err := token.Error(); err != nil {
			log.Printf("Error with MQTT token: %v\n", err)
		}
		log.Println("Disconnecting from MQTT broker, sent message complete")
	}

	client.Disconnect(500)
}

// IsConfigured reports whether the webhook trigger has a URL set.
func (w *WebhookTrigger) IsConfigured() bool {
	return w.Webhook_url != nil
}

// Fire sends service data to the configured webhook URL.
func (w *WebhookTrigger) Fire(data []ServiceData) {
	jsonMessage, err := w.buildMessage(data)
	if err != nil {
		log.Printf("Failed parsing json service data message: %v\n", err)
	}
	log.Println("Firing webhook")

	req, err := http.NewRequest("POST", *w.Webhook_url, jsonMessage)
	if err != nil {
		log.Printf("Error creating webhook request: %v\n", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	if w.Webhook_key_string != nil {
		req.Header.Add("Authorization", *w.Webhook_key_string)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error sending webhook: %v\n", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Error closing webhook request: %v", err)
		}
	}()

	log.Printf("Webhook sent, status: %s\n", res.Status)
}

// buildMessage returns the JSON body for the webhook, applying any custom message template.
func (w *WebhookTrigger) buildMessage(data []ServiceData) (io.Reader, error) {
	if w.Custom_message != nil {
		var customData map[string]any
		if err := json.Unmarshal([]byte(*w.Custom_message), &customData); err != nil {
			return nil, err
		}
		customData["services"] = data
		finalJson, err := json.Marshal(customData)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(finalJson), nil
	}

	jsonSvcData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error occured while parsing webhook message: %v\n", err)
		return nil, err
	}
	return bytes.NewBuffer(jsonSvcData), nil
}
