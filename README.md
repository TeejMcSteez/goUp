# GoUp

***Under-Development!***

Basic server monitor with HTTP web display and API routes, future implementation will also monitor uptime, response time, etc. 

Currently built to learn Go as well as some other protocols (MQTT currently) and technologies and it's something I will genuinely use for my homelab services. 

![Example Image](.github/image.png)

## Triggers

### MQTT Trigger

Client ID: goUp MQTT
State Topic: goUp status

Username and password credentials will only be used if they are provided in services.yml otherwise it will attempt a basic un-authorized tcp connection.

Will publish to the broker specified under the "goUp status" topic

## Example services.yml

Services provide the name and url of the service to monitor currently just makes a basic HTTP request and make sure it responds with 200.

Triggers will provide information and credentials for managing triggers fired when a scrape happens to tell the user about a failed server(s). 

### Triggers

Currently only MQTT is setup for my Home Assistant instance but if viable would wish to add more such as Email, Telegram, etc. but that is more on the integration than logic side.

### Example

```yaml
services:
  home_assistant:
    url: "https://example.com/"
    api_url: "https://example.com/api"
    api_key: "YOUR_API_KEY"

  truenas_scale:
    url: "http://another-example.com/"
triggers:
  mqtt_broker: "URL"
  mqtt_user: "USER"
  mqtt_key: "PASSWORD"
```

