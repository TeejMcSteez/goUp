# GoUp

Under-Development!

## Triggers

### MQTT Trigger

Client ID: goUp MQTT
State Topic: goUp status

Will publish to the broker specified under the "goUp status" topic

## Example services.yml

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

