package utils

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var triggers *Trigger

func SetupTrigger(cfg *Config) {

	triggers = &cfg.Triggers

	fmt.Println("Triggers setup")
}

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
		if err := token.Error(); err != nil {
			panic(err)
		}
		fmt.Println("Disconnecting from MQTT broker, sent message complete")
		client.Disconnect(250)
	} else {
		fmt.Println("No MQTT broker setup")
	}
}
