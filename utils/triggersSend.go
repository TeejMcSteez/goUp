package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// buildPayload is buildMessage for postJSON: returns the value to marshal
// instead of an encoded reader, applying any custom message template.
func (w *WebhookTrigger) buildPayload(data []ServiceData) (any, error) {
	if w.Custom_message == nil {
		return data, nil
	}
	var customData map[string]any
	if err := json.Unmarshal([]byte(*w.Custom_message), &customData); err != nil {
		return nil, err
	}
	customData["services"] = data
	return customData, nil
}

// reconnectedText builds the reconnect message, appending m.Body as extra
// detail (e.g. "offline at x, online at y") when the caller provides one.
func reconnectedText(body string) string {
	const msg = "Host went offline and is back online"
	if body == "" {
		return msg
	}
	return msg + "\n" + body
}

func buildMqtt(m Message) any {
	switch m.Kind {
	case KindAlert:
		return m.Services
	case KindClear:
		return "services has recovered"
	case KindOffline:
		return "host is offline"
	case KindReconnected:
		return "host is online"
	case KindInfo:
		return map[string]any{
			"Title": m.Title,
			"Body":  m.Body,
		}
	default:
		return "invalid Kind provided to MQTT message"
	}
}

func (m *MQTTTrigger) Send(ctx context.Context, msg Message) error {
	if shouldBackoff("MQTT", m.lastFired, m.backoffDuration) {
		return fmt.Errorf("MQTT is currently in back off for: %s", m.lastFired.Add(m.backoffDuration))
	}
	m.lastFired = time.Now()

	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		slog.Info("Connected to MQTT broker")
	}

	var lostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		slog.Error("MQTT connection lost", "error", err)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(*m.Mqtt_broker)
	opts.SetClientID("goUp MQTT")
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = lostHandler
	if m.Mqtt_username != nil || m.Mqtt_key != nil {
		opts.SetUsername(*m.Mqtt_username)
		opts.SetPassword(*m.Mqtt_key)
	}

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		slog.Error("Error pushing to MQTT client", "error", token.Error())
		return token.Error()
	}

	if jsonData, err := json.Marshal(buildMqtt(msg)); err != nil {
		slog.Error("Error formatting service data into JSON", "error", err)
		return err
	} else {
		// Keep retain to true to store last known good message
		token := client.Publish("goup_status", 0, true, jsonData)
		token.Done()

		if err := token.Error(); err != nil {
			slog.Error("Error with MQTT token", "error", err)
			return err
		}
		slog.Info("Disconnecting from MQTT broker, sent message complete")
	}
	return nil
}

func (e *SMTPTrigger) Send(ctx context.Context, m Message) error {
	if shouldBackoff("SMTP", e.lastFired, e.backoffDuration) {
		return fmt.Errorf("SMTP is currently in backoff period for: %s", e.backoffDuration.String())
	}
	e.lastFired = time.Now()

	host, _, err := net.SplitHostPort(*e.SMTPServer)
	if err != nil {
		host = *e.SMTPServer
	}

	var header []byte
	header = fmt.Appendf(header, "To: %s\r\nSubject: %s\r\n\r\n", *e.Email, m.Title)
	var msg []byte
	switch m.Kind {
	case KindAlert:
		for _, entry := range m.Services {
			msg = fmt.Appendf(msg, "Service: %s\r\nMessage: %s\r\nAPI Response: %s\r\n",
				entry.ServiceName, entry.ServiceHTTPResponse, entry.ServiceAPIResponse)
		}
	case KindInfo:
		msg = fmt.Appendf(msg, "Body: %s\r\n", m.Body)
	case KindClear:
		msg = fmt.Appendf(msg, "All services recovered\r\n")
	case KindOffline:
		msg = fmt.Append(msg, "Host is offline\r\n")
	case KindReconnected:
		msg = fmt.Append(msg, strings.ReplaceAll(reconnectedText(m.Body), "\n", "\r\n"))
	default:
		msg = fmt.Append(msg, "No Kind in message")
	}

	email := append(header, msg...)
	if err := smtp.SendMail(*e.SMTPServer, smtp.PlainAuth("", *e.Email, *e.App_Password, host), *e.Email, []string{*e.Email}, email); err != nil {
		slog.Error("Error SMTP request", "error", err)
		return err
	}
	slog.Info("SMTP message sent")
	return nil
}

// servicesText formats downed services for plain-text chat triggers.
func servicesText(services []ServiceData) string {
	var text strings.Builder
	for _, entry := range services {
		fmt.Fprintf(&text, "Service: %s\nResponse: %s\nAPI Response: %s\n\n", entry.ServiceName, entry.ServiceHTTPResponse, entry.ServiceAPIResponse)
	}
	return text.String()
}

func (w *WebhookTrigger) Send(ctx context.Context, m Message) error {
	if !w.IsConfigured() {
		return errors.New("webhook trigger is not configured")
	}

	var payload any
	switch m.Kind {
	case KindAlert:
		p, err := w.buildPayload(m.Services)
		if err != nil {
			return fmt.Errorf("building webhook payload: %w", err)
		}
		payload = p
	case KindInfo:
		payload = map[string]any{"message": m.Body}
	case KindClear:
		payload = map[string]any{"message": "All services recovered"}
	case KindOffline:
		payload = map[string]any{"message": "Host is offline"}
	case KindReconnected:
		payload = map[string]any{"message": reconnectedText(m.Body)}
	default:

		return nil
	}

	if shouldBackoff("Webhook", w.lastFired, w.backoffDuration) {
		return fmt.Errorf("webhook is in backoff period for: %s", w.backoffDuration)
	}
	w.lastFired = time.Now()

	var opts []reqOption
	if w.Webhook_key_string != nil {
		opts = append(opts, header("Authorization", *w.Webhook_key_string))
	}
	// No expect option: any 2xx is success, since webhook receivers vary (200, 202, 204...)
	return postJSON(context.Background(), *w.Webhook_url, payload, opts...)
}

func (g *GotifyTrigger) Send(ctx context.Context, m Message) error {
	if !g.IsConfigured() {
		return errors.New("gotify trigger is not configured")
	}

	var message string
	switch m.Kind {
	case KindAlert:
		var body strings.Builder
		for _, entry := range m.Services {
			fmt.Fprintf(&body, "Service: %s\nMessage: %s\nAPI Response: %s\n\n",
				entry.ServiceName, entry.ServiceHTTPResponse, entry.ServiceAPIResponse)
		}
		message = body.String()
	case KindInfo:
		message = m.Body
	case KindClear:
		message = "All services recovered"
	case KindOffline:
		message = "Host is offline"
	case KindReconnected:
		message = reconnectedText(m.Body)
	default:

		return nil
	}

	if shouldBackoff("Gotify", g.lastFired, g.backoffDuration) {
		return fmt.Errorf("gotify is in backoff period for: %s", g.backoffDuration)
	}
	g.lastFired = time.Now()

	title := "GoUp Alert"
	if g.Gotify_Title != nil {
		title = *g.Gotify_Title
	}
	priority := 5
	if g.Gotify_Priority != nil {
		priority = *g.Gotify_Priority
	}

	url := strings.TrimRight(*g.Gotify_Server, "/") + "/message"
	return postJSON(context.Background(), url, map[string]any{
		"title":    title,
		"message":  message,
		"priority": priority,
	}, header("X-Gotify-Key", *g.Gotify_Token), expectStatus(200))
}

func (s *SlackTrigger) Send(ctx context.Context, m Message) error {
	if !s.IsConfigured() {
		return errors.New("slack trigger is not configured")
	}

	var text string
	switch m.Kind {
	case KindAlert:
		text = servicesText(m.Services)
	case KindInfo:
		text = m.Body
	case KindClear:
		text = "All services recovered"
	case KindOffline:
		text = "Host is offline"
	case KindReconnected:
		text = reconnectedText(m.Body)
	default:

		return nil
	}

	if shouldBackoff("Slack", s.lastFired, s.backoffDuration) {
		return fmt.Errorf("slack is in backoff period for: %s", s.backoffDuration)
	}
	s.lastFired = time.Now()

	username := "GoUp Bot"
	if s.Bot_Username != nil {
		username = *s.Bot_Username
	}

	// Slack returns 200 even on failure; the real result is in the "ok" field
	return postJSON(context.Background(), "https://slack.com/api/chat.postMessage", map[string]any{
		"channel":  *s.Slack_Channel,
		"username": username,
		"text":     text,
	}, header("Authorization", "Bearer "+*s.Slack_Token), expectOkField())
}

func (t *TelegramTrigger) Send(ctx context.Context, m Message) error {
	if !t.IsConfigured() {
		return errors.New("telegram trigger is not configured")
	}

	var text string
	switch m.Kind {
	case KindAlert:
		text = servicesText(m.Services)
	case KindInfo:
		text = m.Body
	case KindClear:
		text = "All services recovered"
	case KindOffline:
		text = "Host is offline"
	case KindReconnected:
		text = reconnectedText(m.Body)
	default:

		return nil
	}

	if shouldBackoff("Telegram", t.lastFired, t.backoffDuration) {
		return fmt.Errorf("telegram is in backoff period for: %s", t.backoffDuration)
	}
	t.lastFired = time.Now()

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", *t.Telegram_Token)
	return postJSON(context.Background(), url, map[string]any{
		"chat_id": *t.Telegram_Channel_Id,
		"text":    text,
	}, expectOkField())
}

func (h *HATrigger) Send(ctx context.Context, m Message) error {
	if !h.IsConfigured() {
		return errors.New("home assistant trigger is not configured")
	}

	var details string
	switch m.Kind {
	case KindAlert:
		details = servicesText(m.Services)
	case KindInfo:
		details = m.Body
	case KindClear:
		details = "All services recovered"
	case KindOffline:
		details = "Host is offline"
	case KindReconnected:
		details = reconnectedText(m.Body)
	default:

		return nil
	}

	if shouldBackoff("Home Assistant", h.lastFired, h.backoffDuration) {
		return fmt.Errorf("home assistant is in backoff period for: %s", h.backoffDuration)
	}
	h.lastFired = time.Now()

	url := strings.TrimRight(*h.HA_URL, "/") + "/api/events/goup_alert"
	return postJSON(context.Background(), url, map[string]any{
		"details": details,
	}, header("Authorization", "Bearer "+*h.HA_Token), expectStatus(200, 201))
}

// TODO: Handle chunked messages for discord 2000 char limit on API content
func (d *DiscordTrigger) Send(ctx context.Context, m Message) error {
	if !d.IsConfigured() {
		return errors.New("discord trigger is not configured")
	}

	var content string
	switch m.Kind {
	case KindAlert:
		var text strings.Builder
		for _, entry := range m.Services {
			fmt.Fprintf(&text, "\nService: %s\nResponse: %s\nAPI Response: %s", entry.ServiceName, entry.ServiceHTTPResponse, entry.ServiceAPIResponse)
		}
		content = "**Server Error Detected**\n" + text.String()
	case KindInfo:
		content = m.Body
	case KindClear:
		content = "All services recovered"
	case KindOffline:
		content = "Host is offline"
	case KindReconnected:
		content = reconnectedText(m.Body)
	default:

		return nil
	}

	if shouldBackoff("Discord", d.lastFired, d.backoffDuration) {
		return fmt.Errorf("discord is in backoff period for: %s", d.backoffDuration)
	}
	d.lastFired = time.Now()

	url := fmt.Sprintf("https://discord.com/api/channels/%s/messages", *d.Discord_Channel)
	return postJSON(context.Background(), url, map[string]any{
		"content": content,
	}, header("Authorization", *d.Discord_Auth), expectStatus(200))
}
