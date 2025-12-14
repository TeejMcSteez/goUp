# GoUp

***Under-Development!***

**[Todo](TODO.md)**

Basic server monitor with HTTP web display and API routes, future implementation will also monitor uptime, response time, etc. 

Currently built to learn Go as well as some other protocols (MQTT currently) and technologies and it's something I will genuinely use for my homelab services. 

![Example Image](.github/image.png)

## Example services.yml

Services provide the name and url of the service to monitor currently just makes a basic HTTP request and make sure it responds with 200.

Triggers will provide information and credentials for managing triggers fired when a scrape happens to tell the user about a failed server(s). 

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
  webhook_url: "https://yourwebhook.com/webhook-id"
  webhook_key: "<Bearer Key (optional)>"
```

## Storage

Uses SQLite with Go [database/sql](https://pkg.go.dev/database/sql) package and https://github.com/mattn/go-sqlite3 driver as a dependency

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

Will send bad service data as JSON to the specified webhook URL

It will use the  `Authorization: Bearer <key>` header if a webhook key is valid for basic authentication for webhook request's 