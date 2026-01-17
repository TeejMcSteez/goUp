package utils

import (
	"bytes"
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"io"
	"log"
	"net/http"
	"time"
)

// Copies Trigger config from configuration to use in t
func SetupTrigger(cfg *Config) *Trigger {
	if cfg.Triggers.MQTT.Mqtt_broker == nil {
		log.Println("No MQTT Broker Found in Configuration")
	}
	if cfg.Triggers.Webhook.Webhook_url == nil {
		log.Println("No Webhook Found in Configuration")
	}

	t := &cfg.Triggers

	if cfg.Backoff_Period != nil && *cfg.Backoff_Period != "" {
		dur, err := time.ParseDuration(*cfg.Backoff_Period)
		if err != nil {
			log.Printf("Invalid Backoff_Period '%s': %v. Disabling backoff.", *cfg.Backoff_Period, err)
			t.backoffDuration = 0
		} else {
			log.Printf("Trigger backoff period set to %s", dur)
			t.backoffDuration = dur
		}
	}

	log.Println("Triggers setup")
	return t
}

// Takes service data and fires all t
func (t *Trigger) Fire(data []ServiceData) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.backoffDuration > 0 && !t.lastFired.IsZero() && time.Since(t.lastFired) < t.backoffDuration {
		log.Printf("Trigger backoff period active, skipping. Last trigger was %s ago.", time.Since(t.lastFired))
		return
	}

	hasMQTT := t.MQTT.Mqtt_broker != nil
	hasWebhook := t.Webhook.Webhook_url != nil

	if !hasMQTT && !hasWebhook {
		return // No triggers configured
	}

	if hasMQTT {
		go t.FireMqtt(data)
	}
	if hasWebhook {
		go t.FireWebhook(data)
	}

	t.lastFired = time.Now()
}

// Takes current bad service data and fires message to configured mqtt broker
func (t *Trigger) FireMqtt(data []ServiceData) {
	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		log.Println("Connected to MQTT Broker")
	}

	var lostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		log.Println("Connection Lost: " + err.Error())
	}

	if t.MQTT.Mqtt_broker != nil {
		opts := mqtt.NewClientOptions()
		opts.AddBroker(*t.MQTT.Mqtt_broker)
		opts.SetClientID("goUp MQTT")
		opts.OnConnect = connectHandler
		opts.OnConnectionLost = lostHandler
		if t.MQTT.Mqtt_username != nil || t.MQTT.Mqtt_key != nil {
			opts.SetUsername(*t.MQTT.Mqtt_username)
			opts.SetPassword(*t.MQTT.Mqtt_key)
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
	} else {
		log.Println("No MQTT broker setup")
	}
}

// Takes in trigger message and checks config for any special message paramters
func (t *Trigger) getWebhookMessage(data []ServiceData) (io.Reader, error) {
	// If their is a custom message defined it will use that as a template to map into
	if t.Webhook.Custom_message != nil {
		var customData map[string]any
		if err := json.Unmarshal([]byte(*t.Webhook.Custom_message), &customData); err != nil {
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

func (t *Trigger) FireWebhook(data []ServiceData) {
	if t.Webhook.Webhook_url != nil {
		jsonMessage, err := t.getWebhookMessage(data)
		if err != nil {
			log.Printf("Failed parsing json service data message: %v\n", err)
		}
		log.Println("Firing webhook")

		req, err := http.NewRequest("POST", *t.Webhook.Webhook_url, jsonMessage)
		if err != nil {
			log.Printf("Error creating webhook request: %v\n", err)
			return
		}
		req.Header.Add("Content-Type", "application/json")
		if t.Webhook.Webhook_key_string != nil {
			req.Header.Add("Authorization", *t.Webhook.Webhook_key_string)
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

	} else {
		log.Println("No webhook setup")
	}
}
