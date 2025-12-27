package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var triggers *Trigger

// Copies Trigger config from configuration to use in triggers
func SetupTrigger(cfg *Config) {
	triggers = &cfg.Triggers
	if triggers.MQTT.Mqtt_broker == nil {
		log.Println("No MQTT Broker Found in Configuration")
	}
	if triggers.Webhook.Webhook_url == nil {
		log.Println("No Webhook Found in Configuration")
	}

	log.Println("Triggers setup")
}

// Takes service data and fires all triggers
func Fire(data []ServiceData) {
	if triggers != nil {
		if triggers.MQTT.Mqtt_broker != nil {
			go FireMqtt(data)
		}
		if triggers.Webhook.Webhook_url != nil {
			go FireWebhook(data)
		}
	}
}

// Takes current bad service data and fires message to configured mqtt broker
func FireMqtt(data []ServiceData) {
	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		log.Println("Connected to MQTT Broker")
	}

	var lostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		log.Println("Connection Lost: " + err.Error())
	}

	if triggers.MQTT.Mqtt_broker != nil {
		opts := mqtt.NewClientOptions()
		opts.AddBroker(*triggers.MQTT.Mqtt_broker)
		opts.SetClientID("goUp MQTT")
		opts.OnConnect = connectHandler
		opts.OnConnectionLost = lostHandler
		if triggers.MQTT.Mqtt_username != nil || triggers.MQTT.Mqtt_key != nil {
			opts.SetUsername(*triggers.MQTT.Mqtt_username)
			opts.SetPassword(*triggers.MQTT.Mqtt_key)
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
			log.Printf("Published Message: %v\n", data)

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

func FireWebhook(data []ServiceData) {
	if triggers.Webhook.Webhook_url != nil {
		jsonSvcData, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Println("Firing webhook")

		req, err := http.NewRequest("POST", *triggers.Webhook.Webhook_url, bytes.NewBuffer(jsonSvcData))
		if err != nil {
			log.Printf("Error creating webhook request: %v\n", err)
			return
		}
		req.Header.Add("Content-Type", "application/json")
		if triggers.Webhook.Webhook_key_string != nil {
			req.Header.Add("Authorization", *triggers.Webhook.Webhook_key_string)
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
