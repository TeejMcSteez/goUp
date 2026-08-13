package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (m *MQTTTrigger) Fire(data []ServiceData) {
	if shouldBackoff("MQTT", m.lastFired, m.backoffDuration) {
		return
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
	}

	if jsonData, err := json.Marshal(data); err != nil {
		slog.Error("Error formatting service data into JSON", "error", err)
	} else {
		// Keep retain to true to store last known good message
		token := client.Publish("goup_status", 0, true, jsonData)
		token.Done()

		if err := token.Error(); err != nil {
			slog.Error("Error with MQTT token", "error", err)
		}
		slog.Info("Disconnecting from MQTT broker, sent message complete")
	}

	client.Disconnect(500)
}

// Fire sends service data to the configured webhook URL.
func (w *WebhookTrigger) Fire(data []ServiceData) {
	if shouldBackoff("Webhook", w.lastFired, w.backoffDuration) {
		return
	}
	w.lastFired = time.Now()

	jsonMessage, err := w.buildMessage(data)
	if err != nil {
		slog.Error("Failed parsing json service data message", "error", err)
	}
	slog.Info("Firing webhook")

	req, err := http.NewRequest("POST", *w.Webhook_url, jsonMessage)
	if err != nil {
		slog.Error("Error creating webhook request", "error", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	if w.Webhook_key_string != nil {
		req.Header.Add("Authorization", *w.Webhook_key_string)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Error sending webhook", "error", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Error closing webhook request", "error", err)
		}
	}()

	slog.Info("Webhook sent", "status", res.Status)
}

// buildMessage returns the JSON body for the webhook, applying any custom message template.
func (w *WebhookTrigger) buildMessage(data []ServiceData) (io.Reader, error) {
	if w.Custom_message != nil {
		var customData map[string]any
		if err := json.Unmarshal([]byte(*w.Custom_message), &customData); err != nil {
			return nil, err
		}
		customData["services"] = data
		finalJson, err := json.Marshal(customData)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(finalJson), nil
	}

	jsonSvcData, err := json.Marshal(data)
	if err != nil {
		slog.Error("Error occured while parsing webhook message", "error", err)
		return nil, err
	}
	return bytes.NewBuffer(jsonSvcData), nil
}

// Fire sends all downed service data as an email
func (e *SMTPTrigger) Fire(data []ServiceData) {
	if shouldBackoff("SMTP", e.lastFired, e.backoffDuration) {
		return
	}
	e.lastFired = time.Now()

	host, _, err := net.SplitHostPort(*e.SMTPServer)
	if err != nil {
		host = *e.SMTPServer
	}

	var header []byte
	header = fmt.Appendf(header, "To: %s\r\nSubject: GoUp Failure Message\r\n\r\n", *e.Email)
	var msg []byte
	for _, entry := range data {
		msg = fmt.Appendf(msg, "Service: %s\r\nMessage: %s\r\nAPI Response: %s\r\n",
			entry.ServiceName, entry.ServiceHTTPResponse, entry.ServiceAPIResponse)
	}
	email := append(header, msg...)
	if err := smtp.SendMail(*e.SMTPServer, smtp.PlainAuth("", *e.Email, *e.App_Password, host), *e.Email, []string{*e.Email}, email); err != nil {
		slog.Error("Error SMTP request", "error", err)
		return
	}
	slog.Info("SMTP message sent")
}

// Fire sends a Gotify push notification for all downed services.
func (g *GotifyTrigger) Fire(data []ServiceData) {
	if shouldBackoff("Gotify", g.lastFired, g.backoffDuration) {
		return
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

	var body []byte
	for _, entry := range data {
		body = fmt.Appendf(body, "Service: %s\nMessage: %s\nAPI Response: %s\n\n",
			entry.ServiceName, entry.ServiceHTTPResponse, entry.ServiceAPIResponse)
	}

	payload, err := json.Marshal(map[string]any{
		"title":    title,
		"message":  string(body),
		"priority": priority,
	})
	if err != nil {
		slog.Error("Error building Gotify payload", "error", err)
		return
	}

	url := strings.TrimRight(*g.Gotify_Server, "/") + "/message"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		slog.Error("Error creating Gotify request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", *g.Gotify_Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Error sending Gotify notification", "error", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Error closing Gotify response", "error", err)
		}
	}()

	slog.Info("Gotify notification sent", "status", res.Status)
}

func (s *SlackTrigger) Fire(data []ServiceData) {
	if shouldBackoff("Slack", s.lastFired, s.backoffDuration) {
		return
	}
	s.lastFired = time.Now()

	var text strings.Builder
	for _, entry := range data {
		fmt.Fprintf(&text, "Service: %s\nResponse: %s\nAPI Response: %s", entry.ServiceName, entry.ServiceHTTPResponse, entry.ServiceAPIResponse)
	}
	getUsername := func(str *string) string {
		if str == nil {
			return "GoUp Bot"
		}
		return *str
	}
	payload, err := json.Marshal(map[string]any{
		"channel":  *s.Slack_Channel,
		"username": getUsername(s.Bot_Username),
		"text":     text.String(),
	})
	if err != nil {
		slog.Error("Failed craft payload for Slack trigger", "error", err)
		return
	}
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(payload))
	if err != nil {
		slog.Error("Failed to create request for slack", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+*s.Slack_Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed to send request to slack", "error", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Error closing slack response body", "error", err)
		}
	}()
	var slackResp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&slackResp); err != nil {
		slog.Error("Failed to decode Slack response", "error", err)
		return
	}
	if !slackResp.OK {
		slog.Error("Slack notification failed", "error", slackResp.Error)
		return
	}
	slog.Info("Slack notification sent")
}

func (t *TelegramTrigger) Fire(data []ServiceData) {
	if shouldBackoff("Telegram", t.lastFired, t.backoffDuration) {
		return
	}
	t.lastFired = time.Now()

	var text strings.Builder
	for _, entry := range data {
		fmt.Fprintf(&text, "Service: %s\nResponse: %s\nAPI Response: %s", entry.ServiceName, entry.ServiceHTTPResponse, entry.ServiceAPIResponse)
	}
	payload, err := json.Marshal(map[string]any{
		"chat_id": *t.Telegram_Channel_Id,
		"text":    text.String(),
	})
	if err != nil {
		slog.Error("failed to craft payload for telegram", "error", err)
		return
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", *t.Telegram_Token)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		slog.Error("failed to create request for telegram", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to send request for telegram", "error", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Error closing Telegram response body", "error", err)
		}
	}()
	var telegramResp struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(res.Body).Decode(&telegramResp); err != nil {
		slog.Error("Failed to decode Telegram response", "error", err)
		return
	}
	if !telegramResp.OK {
		slog.Error("Telegram notification failed", "description", telegramResp.Description)
		return
	}
	slog.Info("Telegram notification sent")
}

func (h *HATrigger) Fire(data []ServiceData) {
	if shouldBackoff("Home Assistant", h.lastFired, h.backoffDuration) {
		return
	}
	h.lastFired = time.Now()

	var extraStateAttribs strings.Builder
	for _, entry := range data {
		fmt.Fprintf(&extraStateAttribs, "Service: %s\nResponse: %s\nAPI Response: %s", entry.ServiceName, entry.ServiceHTTPResponse, entry.ServiceAPIResponse)
	}
	body, err := json.Marshal(map[string]any{
		"details": extraStateAttribs.String(),
	})
	if err != nil {
		slog.Error("failed to create HA json payload", "error", err)
		return
	}
	url := strings.TrimRight(*h.HA_URL, "/") + "/api/events/goup_alert"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		slog.Error("failed to create HA request", "error", err)
		return
	}
	req.Header.Add("Authorization", "Bearer "+*h.HA_Token)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to send request for HA", "error", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("error closing HA response body", "error", err)
		}
	}()
	if res.StatusCode != 200 && res.StatusCode != 201 {
		slog.Error("Failed to send or post sensor state", "status_code", res.StatusCode)
		return
	}
	slog.Info("HA notification sent")
}

// TODO: Handle chunked messages for discord 2000 char limit on API content
func (d *DiscordTrigger) Fire(data []ServiceData) {
	if shouldBackoff("Discord", d.lastFired, d.backoffDuration) {
		return
	}
	d.lastFired = time.Now()

	var text strings.Builder

	for _, entry := range data {
		fmt.Fprintf(&text, "\nService: %s\nResponse: %s\nAPI Response: %s", entry.ServiceName, entry.ServiceHTTPResponse, entry.ServiceAPIResponse)
	}
	payload, err := json.Marshal(map[string]any{
		"content": "**Server Error Detected**\n" + text.String(),
	})
	if err != nil {
		slog.Error("failed to create payload for discord notification request", "error", err)
		return
	}
	api_url := "https://discord.com/api"
	req, err := http.NewRequest("POST", api_url+fmt.Sprintf("/channels/%s/messages", *d.Discord_Channel), bytes.NewBuffer(payload))
	if err != nil {
		slog.Error("failed to create request for discord notification", "error", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", *d.Discord_Auth)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("error sending discord notification", "error", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("error closing HA response body", "error", err)
		}
	}()
	if res.StatusCode != 200 {
		slog.Error("error in discord notification response", "status_code", res.StatusCode)
		return
	}
	slog.Info("discord notification sent")
}
