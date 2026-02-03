# GoUp

GoUp is a server monitor featuring an HTTP web interface and API routes to track uptime, response time, status, and more!

The frontend is built with [React](https://react.dev/) and [Vite](https://vite.dev/), utilizing [Chart.js](https://www.chartjs.org/) for uptime visualization.

The entire frontend directory is embedded directly into the Go server using Go's [embed](https://pkg.go.dev/embed) package. 

This application does not include built-in authentication and is designed for use within a LAN or trusted networks:
1. Developing authentication solutions was not a primary focus, as it falls outside the scope of this project's initial intent.
2. The goal is to avoid imposing a specific authentication solution on users.

## Example services.yml

The database path is specified by `db_path`.

`backoff` specifies the delay between trigger activations. Options include `<int><s/m/h>` (e.g., `15s`).

Services are defined by their name, URL, and optional `api_url` and `api_key`.

Only the database path, service name, and service URL are mandatory; `api_url` and `api_key` are optional.

Triggers contain the necessary information for sending various messages:
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
"timestamp" TEXT,
"service_response_time" TEXT,
"error" INTEGER NOT NULL DEFAULT 0
```

## Triggers

### MQTT Trigger

Client ID: goUp MQTT
State Topic: goup_status

Username and password credentials are utilized only if provided in `services.yml`; otherwise, the system will attempt an unauthorized TCP connection.

Messages will be published to the broker under the "goup_status" topic.

### Webhook Trigger

It sends JSON data regarding service failures to the specified webhook URL. If custom messages are defined, they will be appended to the JSON payload.

The provided key will be injected as an `Authorization` header, supporting values like `Basic`, `Bearer`, etc., as string values in the YAML configuration.

## Arch. Support

- Linux ARM64/AMD64
- Windows ARM64/AMD64

This is a pure Go program; it does not require additional libraries and compiles directly to a single executable with the Go compiler.

While Go supports building for Darwin (Apple) ARM64/AMD64, I do not actively build for these platforms due to a lack of Apple hardware for testing. However, I can add Darwin builds to the GitHub Actions workflow upon request.

## Example Use - N8N Message on Error

To integrate, create a POST Webhook trigger. When activated, this trigger will receive error information for any failed servers.

This information can then be utilized as desired. Below is an example HTML message that maps error items to a card, which is subsequently sent via Gmail from N8N.


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

Upon a goUp webhook trigger firing, the user will receive a message containing the error information.