# GoUp

GoUp is a server monitor featuring an HTTP web interface and API routes to track uptime, response time, status, and more!

The frontend is built with [React](https://react.dev/) and [Vite](https://vite.dev/), utilizing [Chart.js](https://www.chartjs.org/) for uptime visualization.

The entire frontend directory is embedded directly into the Go server using Go's [embed](https://pkg.go.dev/embed) package. 

This application does not include built-in authentication and is designed for use within a LAN or trusted networks:
1. Developing authentication solutions was not a primary focus, as it falls outside the scope of this project's initial intent.
2. The goal is to avoid imposing a specific authentication solution on users.

## AI Disclosure

Portions of this project were generated with the assistance of AI tools.
All generated code was reviewed and modified by the author.

## Note About TLS

This project is soely focused on HTTP, routes, and the logic. A reverse proxy is recommended to handle TLS errors for servers that are HTTPS that one wants to test.

This seperates the concern for this codebase and reduces binary size as well as offloads the work of handling HTTPS to projects that are much more qualified.

## Example services.yml

The database path is specified by `db_path` as well as the `db_max_size` which is in the format <number><size> (Ex: 1gb, 20mb, 40kb,etc.)

Services are defined by their name, URL, and optional `api_url` and `api_key`.

Only the database path, service name, and service URL are mandatory; `api_url` and `api_key` are optional.

Triggers contain the necessary information for sending various messages:
   - `backoff` specifies the delay between trigger activations. Options include `<int><s/m/h>` (e.g., `15s`).
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
db_max_size: "1gb"
services:
  home_assistant:
    url: "https://ex.com/"
    api_url: "http://ex.com/api/"
    api_key: "<key>"

  truenas_scale:
    url: "http://ex-scale.com/"

triggers:
  backoff: "30m"
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

Username and password credentials are utilized only if provided in `services.yml`; otherwise, the system will attempt an unauthorized connection.

Messages will be published to the broker under the "goup_status" topic.

### Webhook Trigger

It sends JSON data regarding service failures to the specified webhook URL. If custom messages are defined, they will be appended to the JSON payload.

The provided key will be injected as an `Authorization` header, supporting values like `Basic`, `Bearer`, etc., as string values in the YAML configuration.

## Arch. Support

- Linux ARM64/AMD64
- Windows ARM64/AMD64

This is a pure Go program; it does not require additional libraries and compiles directly to a single executable with the Go compiler.

While Go supports building for Darwin (Apple) ARM64/AMD64, I do not actively build for these platforms due to a lack of Apple hardware for testing. However, I can add Darwin builds to the GitHub Actions workflow upon issue request but testing will be up to the user's.

### Tested Arch.

- Debian x86_64 (AMD64)
- Linux raspberrypi(4) (Debian) aarch64 (ARM64)

## Current Test Coverage (Dev)

After listening to some talks from the creator of SQLite I was pushed to want more testing and overall test coverage in my code.

With this in mind below are the current coverages of the code found by running `go test ./... -cover`

- utils = 62.6%
- workers = 68.7%
- server = 0%
- main = 0%
