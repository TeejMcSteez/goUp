# GoUp

***Under-Development!***

**[Todo](TODO.md)**

Server monitor with HTTP web display and API routes for uptime, response time, status, and more! 

![Example Image](.github/image.png)

## Example services.yml

Services give url, api_url (if any), and api_key (if any) to the program nothing but the name and url are required.

Triggers contain information for firing
  - mqtt
    - mqtt_broker - URL of the MQTT broker
    - mqtt_user - Username of the MQTT account if any
    - mqtt_key - Key or password for the MQTT account
  - webhook
    - webhook_url - URL to send the webhook message to
    - webhook_key - Authorization type (Bearer, Basic, etc.) and the key/connection string
    - custom_message - Extra fields to add to the pre-defined JSON message

### Example

```yaml
services:
  home_assistant:
    url: "https://ex.com/"
    api_url: "http://ex.com/api/"
    api_key: "<key>"

  truenas_scale:
    url: "http://ex-scale.com/"

triggers:
  mqtt:
    mqtt_broker: "<url>"
    mqtt_user: "<user>"
    mqtt_key: "<key>"
  webhook:
    webhook_url: "https://ex.com/"
    webhook_key: "<auth_type> <key>"
    custom_message: '{ "example_tag": "data" }'

```

## Storage

Uses SQLite with Go [database/sql](https://pkg.go.dev/database/sql) package and [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) driver as a dependency

Schema: 

```sql
"id" INTEGER PRIMARY KEY AUTOINCREMENT,
"service_name" TEXT,
"service_HTTP_response" TEXT,
"service_API_response" TEXT,
"service_response_time" TEXT
```

## Triggers

### MQTT Trigger

Client ID: goUp MQTT
State Topic: goup_status

Username and password credentials will only be used if they are provided in services.yml otherwise it will attempt a basic un-authorized tcp connection.

Will publish to the broker specified under the "goup_status" topic

Currently only MQTT is setup for my Home Assistant instance but if viable would wish to add more such as Email, Telegram, etc. but that is more on the integration than logic side.

### Webhook Trigger

Will send bad service data as JSON to the specified webhook URL, if the user has defined any custom messages it will add that custom message onto the JSON payload.

The key provided will be injected as a `Authorization` header. Meaning it accepts Basic, Bearer, etc. as a string value in the YAML.