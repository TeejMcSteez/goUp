# GoUp

Server monitor with HTTP web display and API routes for uptime, response time, status, and more!

![Example Image](.github/image.png)

Frontend is built with [React](https://react.dev/) + [Vite](https://vite.dev/) alongside [Chart.js](https://www.chartjs.org/) for uptime visualization.

Entire frontend directory should be right at 4.0kb including assets which is embedded into the Go server using Go's [embed](https://pkg.go.dev/embed) package. 

This application does not contain **ANY** built in authentication and is intended for LAN or trusted networks:
1. I do not want to spend excessive time programming authentication solutions (it's not interesting to me)
2. I do not want to lock in users to the authentication solution I choose

## Example services.yml

The database path is specified at `db_path`

`backoff` signifies the wait period in between triggers being fired options can be `<int><s/m/h>` (Ex: 15s)

Services give name, url, api_url (if any), and api_key (if any) to the program.

No keys or api urls are required just the database path, service name, and service url

Triggers contain information for firing different messages:
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
db_path: "./database.db"
backoff: "30m"
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
"error" INTEGER NOT NULL DEFAULT 0
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

## Arch. Support

- Linux ARM64/AMD64
- Windows ARM64/AMD64

This a pure Go program it does not require any other libraries and is compiled directly to source with the Go compiler.

Go does build Darwin (Apple) ARM64/AMD64, I just do not use any Apple products so I do not build for it. That being said if anyone does want to me to build for Darwin I can add it to the GitHub Actions workflow.

## Example Use - N8N Message on Error

Create a POST Webhook trigger, when the trigger is activated it will contain the error information for the servers that have failed.

One can then use this information however they want below is an example HTML message mapping error items to a card which is then sent via Gmail message from N8N.


```JavaScript
{{ $json.body.map(item =>
`<tr style="background-color: ${item.error ? '#fcf2f2' : '#f2fcf2'}; border-bottom: 1px solid #e0e0e0;">
<td style="padding: 12px; font-weight: 500; color: #333;">${item.name}</td>
<td style="padding: 12px;">
<span style="background-color: ${item.error ? '#d9534f' : '#5cb85c'}; color: white; font-size: 12px; font-weight: 600; padding: 4px 8px; border-radius: 12px;">
${item.error ? 'Error' : 'OK'}
</span>
</td>
<td style="padding: 12px; font-size: 14px; color: ${item.error ? '#c9302c' : '#333'}; font-family: 'SF Mono', 'Fira Code', 'Fira Mono', 'Roboto Mono', monospace;">
${item.response}
</td>
<td style="padding: 12px; font-size: 14px; color: #555;">
${item.response_time}
</td>
</tr>`).join('') }}
```

Once a webhook trigger is fired from goUp the user will receive a message with the error information!