# GoUp

GoUp is a server monitor featuring a web interface and API routes to track server(s) information.

The frontend is built with [React](https://react.dev/) and [Vite](https://vite.dev/), utilizing [Chart.js](https://www.chartjs.org/) for uptime visualization.

The entire frontend directory is embedded directly into the Go server using Go's [embed](https://pkg.go.dev/embed) package. 

**READ BELOW**

This application does not include built-in authentication and is designed for use within a LAN or trusted networks:
1. Developing authentication solutions was not a primary focus, as it falls outside the scope of this project's initial intent.
2. The goal is to avoid imposing a specific authentication solution on users.

Development focusing on security may arise but is not planned hence this warning

## API

A full interactive API reference is available at `/swagger/index.html` when the server is running. Run `make docs` to regenerate it after modifying handler annotations.

## AI Disclosure

Portions of this project were generated with the assistance of AI tools.
All generated code was reviewed and modified by the author.

## Example services.yml

The database path is specified by `db_path` as well as the `db_max_size` which is in the format <number><size> (Ex: 1gb, 20mb, 40kb,etc.) the minimum size is rougly 64KB as the database writes in WAL mode so when running it uses a WAL, SHM, and DB file for operation. 

When shrinking the database the system switches off WAL mode so the row deletion (and VACUUM) happens on disk then switches back into WAL mode for concurrent read/writes.

One can also specify whether or not to keep data after stopping the program with `persist_db` being `true` or `false`

Services are defined by their name, URL, and optional fields. Only `db_path`, the service name, and `url` are mandatory; all other fields are optional.

| Service field | Description |
|---|---|
| `url` | URL to monitor |
| `description` | Human-readable label shown in the UI |
| `api_url` | Secondary endpoint to query for API-level health |
| `api_key` | Bearer token sent with `api_url` requests |
| `valid_responses` | List of HTTP status codes considered healthy (default: `["200"]`) |
| `retry` | Number of times to retry a failed request before triggering |
| `active` | Tells the program whether the service is currently being monitored |
| `skip_insecure` | Skips TLS certificate verification for this service (e.g. self-signed or hostname-mismatched certs) |

Triggers fire when a service is detected as down. All trigger types are optional; only configure the ones you need.

- `backoff` — minimum time between trigger activations, e.g. `15s`, `5m`, `1h`. Set at the top level of `triggers` to apply globally. Each trigger type also accepts its own `backoff` field, which overrides the global value for that trigger only; triggers without their own `backoff` fall back to the global setting.
- `mqtt` — publishes service state JSON to a broker
  - `mqtt_broker` — broker URL (e.g. `mqtt://broker.example.com`)
  - `mqtt_user` — username (optional)
  - `mqtt_key` — password (optional)
  - `backoff` — per-trigger override of the global backoff period
- `webhook` — HTTP POST to any URL
  - `webhook_url` — target URL
  - `webhook_key` — `Authorization` header value (e.g. `Bearer <token>`)
  - `custom_message` — JSON object merged with the service payload
  - `backoff` — per-trigger override of the global backoff period
- `smtp` — sends an email via SMTP plain auth
  - `smtp_server` — server and port (e.g. `smtp.gmail.com:587`)
  - `email` — sender and recipient address
  - `app_password` — app password or SMTP credential
  - `backoff` — per-trigger override of the global backoff period
- `gotify` — push notification via a Gotify server
  - `gotify_server` — base URL of the Gotify instance
  - `gotify_app_token` — application token
  - `gotify_application` — application name (optional)
  - `gotify_title` — notification title (default: `GoUp Alert`)
  - `gotify_priority` — priority level 0–10 (default: `5`)
  - `backoff` — per-trigger override of the global backoff period
- `slack` — posts to a Slack channel via bot token
  - `slack_token` — bot OAuth token (`xoxb-...`)
  - `slack_channel` — channel ID used in the POST request (e.g. `C1234567890`)
  - `username` — display name for the bot (default: `GoUp Bot`)
  - `backoff` — per-trigger override of the global backoff period
- `telegram` — sends a message via Telegram Bot API
  - `telegram_token` — bot token from [@BotFather](https://t.me/BotFather)
  - `telegram_channel_id` — chat or channel ID (e.g. `@channelname` or `-100123456789`)
  - `backoff` — per-trigger override of the global backoff period
- `home_assistant` — fires a `goup_alert` event on the HA event bus; use an automation with `trigger: event_type: goup_alert` to act on it. Event data contains a `details` field with downed service info.
  - `ha_url` — base URL of your HA instance (e.g. `http://homeassistant.local:8123`)
  - `ha_token` — long-lived access token
  - `backoff` — per-trigger override of the global backoff period
- `discord` — posts a message to a Discord channel via the bot API
  - `discord_auth` — authorization header value (`Bot <token>` or `Bearer <token>`)
  - `discord_channel_id` — target channel ID (e.g. `123456789012345678`)
  - `backoff` — per-trigger override of the global backoff period

### Example

```yaml
db_path: "./database.db"
db_max_size: "1gb"
persist_db: true

services:
  home_assistant:
    url: "https://ha.example.com/"
    description: "Home automation hub"
    api_url: "https://ha.example.com/api/"
    api_key: "<token>"
    valid_responses: ["200", "201"]
    retry: 2
    active: null # Can be true or null for active monitoring

  truenas:
    url: "http://nas.example.com/"
    description: "NAS storage"
    active: false # Skips fetching data until changed to true or null

  streamer:
    url: "https://streamer.example.com:8080/"
    description: "Internal service with a self-signed cert"
    skip_insecure: true # Skips TLS certificate verification for this service only

triggers:
  backoff: "30m"
  mqtt:
    mqtt_broker: "mqtt://broker.example.com"
    mqtt_user: "<user>"
    mqtt_key: "<password>"
  webhook:
    webhook_url: "https://hooks.example.com/alert"
    webhook_key: "Bearer <token>"
    custom_message: '{ "source": "goUp" }'
    backoff: "5m" # overrides the global 30m backoff for this trigger only
  smtp:
    smtp_server: "smtp.gmail.com:587"
    email: "you@gmail.com"
    app_password: "<app-password>"
  gotify:
    gotify_server: "https://gotify.example.com"
    gotify_app_token: "<token>"
    gotify_title: "GoUp Alert"
    gotify_priority: 5
  slack:
    slack_token: "xoxb-<token>"
    slack_channel: "C1234567890"
    username: "GoUp Bot"
  telegram:
    telegram_token: "<bot-token>"
    telegram_channel_id: "@channelname"
  home_assistant:
    ha_url: "http://homeassistant.local:8123"
    ha_token: "<long-lived-access-token>"
  discord:
    discord_auth: "Bot <token>"
    discord_channel_id: "123456789012345678"
```
## Notifications

These are **subject to change** 

- [x] Webhook
- [x] MQTT
- [x] SMTP (via Gmail App Password currently)
- [x] Slack (via bot token)
- [x] Gotify
- [x] Telegram
- [x] Home Assistant (via event bus)
- [x] Discord

## Building

Requires:
- Node package manager (npm, pnpm, yarn, or bun) 
- [swag](https://github.com/swaggo/swag) CLI (`go install github.com/swaggo/swag/cmd/swag@latest`)
- [sqlc](https://sqlc.dev/) CLI, v1.31.1 (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1`)
- [goose]() CLI, v3.27.2 (`go install github.com/pressly/goose/v3/cmd/goose@v3.27.2`)
### Development:
- [golangci-lint](https://golangci-lint.run/)
- [pprof](https://github.com/google/pprof) - for performance profiling

The Makefile auto-detects the package manager by matching the existing lock file, then falls back to whichever is installed.

| Target | Description |
|--------|-------------|
| `make` / `make all` | Generate API docs, build frontend, then compile the Go binary |
| `make docs` | Regenerate Swagger docs from handler annotations (served at `/swagger/index.html`) |
| `make build` | Build the frontend, SQL codegen, then compile the Go binary (`./goUp`) |
| `make fmt` | Format Go source files via `golangci-lint fmt` |
| `make lint` | Lint Go source files via `golangci-lint run` |
| `make test` | Run all Go tests with coverage (`go test ./... -cover`) |
| `make prof` | Run benchmarks and write CPU + memory profiles to `perf/` use [pprof](https://github.com/google/pprof) for viewing (example: `pprof -http=:8080 perf/workers.mem.out`)|
| `make clean` | Remove the compiled binary and `perf/` directory |

### SQL codegen

Database access code is generated by [sqlc](https://sqlc.dev/) from `db/schema.sql` and `db/query.sql` (config: `db/sqlc.yaml`), producing Go code in `internal/db`. After editing the schema or a query, regenerate with:

```sh
cd db && sqlc generate
```

`make build` and `make dev` run this automatically, and the Docker build regenerates it in the builder stage as well, so the files in `internal/db` should always match `db/schema.sql` and `db/query.sql`.

## Storage

Uses SQLite with Go [database/sql](https://pkg.go.dev/database/sql) package and [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) driver as a dependency

Schema: 

```sql
"id" INTEGER PRIMARY KEY AUTOINCREMENT,
"service_url" TEXT,
"service_name" TEXT,
"service_description" TEXT,
"service_HTTP_response" TEXT,
"service_API_response" TEXT,
"service_response_time" INTEGER,
"timestamp" TEXT,
"error" INTEGER NOT NULL DEFAULT 0
"active" INTEGER NOT NULL DEFAULT 1
```

`service_response_time` is stored as nanoseconds (INTEGER) for efficient range queries and indexing. The column is indexed together with `service_name` via `idx_service_data_lookup`. Existing databases with legacy TEXT response times (e.g. `"1.234ms"`) are migrated automatically on startup. (now using [goose](https://pkg.go.dev/github.com/pressly/goose/v3))

```sql
"service_name" TEXT PRIMARY KEY,
"fingerprint" TEXT NOT NULL, -- sha256 of the soonest certs DER
"not_after" INTEGER NOT NULL, -- soonest expiry in chain
"subject" TEXT,
"issuer" TEXT,
"is_expired" INTEGER NOT NULL DEFAULT 0,
"chain" TEXT,
"first_seen" TEXT NOT NULL,
"last_checked" TEXT NOT NULL
```

## Arch. Support

- Linux ARM64/AMD64
- Windows ARM64/AMD64
- Darwin ARM64/AMD64

### Tested Arch.

- Debian x86_64 (AMD64)

## Current Test Coverage (Dev)

After listening to some talks from the creator of SQLite I was pushed to want more testing and overall test coverage in my code.

With this in mind below are the current coverages of the code found by running `go test ./... -cover`

- goUp coverage: 0.0% of statements
- goUp/server coverage: 25.7% of statements
- goUp/utils coverage: 63.7% of statements
- goUp/workers coverage: 51.4% of statements
