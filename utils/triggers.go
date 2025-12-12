package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var triggers *Trigger

// Copies Trigger config from configuration to use in triggers
func SetupTrigger(cfg *Config) {
	triggers = &cfg.Triggers
	if triggers.Mqtt_broker == nil {
		fmt.Println("No MQTT Broker Found in Configuration")
	}
	if triggers.Webhook_url == nil {
		fmt.Println("No Webhook Found in Configuration")
	}

	fmt.Println("Triggers setup")
}

// Takes service data and fires all triggers
func Fire(data []ServiceData) {
	if triggers != nil {
		if triggers.Mqtt_broker != nil {
			FireMqtt(data)
		}
		if triggers.Webhook_url != nil {
			FireWebhook(data)
		}
	}
}

// Takes current bad service data and fires message to configured mqtt broker
func FireMqtt(data []ServiceData) {
	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		fmt.Println("Connected to MQTT Broker")
	}

	var lostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		fmt.Println("Connection Lost: " + err.Error())
	}

	if triggers.Mqtt_broker != nil {
		opts := mqtt.NewClientOptions()
		opts.AddBroker(*triggers.Mqtt_broker)
		opts.SetClientID("goUp MQTT")
		opts.OnConnect = connectHandler
		opts.OnConnectionLost = lostHandler
		if triggers.Mqtt_username != nil || triggers.Mqtt_key != nil {
			opts.SetUsername(*triggers.Mqtt_username)
			opts.SetPassword(*triggers.Mqtt_key)
		}

		client := mqtt.NewClient(opts)

		if token := client.Connect(); token.Wait() && token.Error() != nil {
			fmt.Println("Error pushing to MQTT client: " + token.Error().Error())
		}

		var message string

		for _, c := range data {
			message += "Name: " + c.ServiceName + "\nHTTPResponse: " + c.ServiceHTTPResponse + "\nAPIResponse: " + c.ServiceAPIResponse
		}

		token := client.Publish("goUp status", 0, false, message)
		fmt.Println("Published Message: " + message)

		token.Done()
		// TODO: If connection fails will panic out
		if err := token.Error(); err != nil {
			panic(err)
		}
		fmt.Println("Disconnecting from MQTT broker, sent message complete")
		client.Disconnect(250)
	} else {
		fmt.Println("No MQTT broker setup")
	}
}

func FireWebhook(data []ServiceData) {
	if triggers.Webhook_url != nil {
		jsonSvcData, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println("Firing webhook")

		req, err := http.NewRequest("POST", *triggers.Webhook_url, bytes.NewBuffer(jsonSvcData))
		if err != nil {
			log.Printf("Error creating webhook request: %v\n", err)
			return
		}
		req.Header.Add("Content-Type", "application/json")
		if triggers.Webhook_key != nil {
			req.Header.Add("Authorization", "Basic "+*triggers.Webhook_key)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error sending webhook: %v\n", err)
			return
		}
		defer func() {
			if err := res.Body.Close(); err != nil {
				fmt.Printf("Error closing webhook request: %v", err)
			}
		}()

		fmt.Printf("Webhook sent, status: %s\n", res.Status)

	} else {
		fmt.Println("No webhook setup")
	}
}
