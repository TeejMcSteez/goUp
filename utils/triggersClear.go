package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/smtp"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (m *MQTTTrigger) Clear() {
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

	if jsonData, err := json.Marshal("Services all clear"); err != nil {
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

// Clear sends an all-clear notification to the configured webhook URL.
func (w *WebhookTrigger) Clear() {
	payload, err := json.Marshal(map[string]any{
		"status":  "clear",
		"message": "All services recovered",
	})
	if err != nil {
		slog.Error("Failed to build webhook clear payload", "error", err)
		return
	}

	req, err := http.NewRequest("POST", *w.Webhook_url, bytes.NewBuffer(payload))
	if err != nil {
		slog.Error("Error creating webhook clear request", "error", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	if w.Webhook_key_string != nil {
		req.Header.Add("Authorization", *w.Webhook_key_string)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Error sending webhook clear notification", "error", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Error closing webhook clear response", "error", err)
		}
	}()

	slog.Info("Webhook clear notification sent", "status", res.Status)
}

// Clear sends an all-clear email once previously downed services recover.
func (e *SMTPTrigger) Clear() {
	host, _, err := net.SplitHostPort(*e.SMTPServer)
	if err != nil {
		host = *e.SMTPServer
	}

	var header []byte
	header = fmt.Appendf(header, "To: %s\r\nSubject: GoUp All Clear\r\n\r\n", *e.Email)
	msg := []byte("All services have recovered.\r\n")
	email := append(header, msg...)
	if err := smtp.SendMail(*e.SMTPServer, smtp.PlainAuth("", *e.Email, *e.App_Password, host), *e.Email, []string{*e.Email}, email); err != nil {
		slog.Error("Error SMTP clear request", "error", err)
		return
	}
	slog.Info("SMTP clear message sent")
}

// Clear sends a Gotify push notification once previously downed services recover.
func (g *GotifyTrigger) Clear() {
	title := "GoUp Alert"
	if g.Gotify_Title != nil {
		title = *g.Gotify_Title
	}

	payload, err := json.Marshal(map[string]any{
		"title":    title,
		"message":  "All services have recovered",
		"priority": 1,
	})
	if err != nil {
		slog.Error("Error building Gotify clear payload", "error", err)
		return
	}

	url := strings.TrimRight(*g.Gotify_Server, "/") + "/message"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		slog.Error("Error creating Gotify clear request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", *g.Gotify_Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Error sending Gotify clear notification", "error", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Error closing Gotify clear response", "error", err)
		}
	}()

	slog.Info("Gotify clear notification sent", "status", res.Status)
}

// Clear sends an all-clear message to Slack once previously downed services recover.
func (s *SlackTrigger) Clear() {
	getUsername := func(str *string) string {
		if str == nil {
			return "GoUp Bot"
		}
		return *str
	}
	payload, err := json.Marshal(map[string]any{
		"channel":  *s.Slack_Channel,
		"username": getUsername(s.Bot_Username),
		"text":     "All services have recovered",
	})
	if err != nil {
		slog.Error("Failed craft clear payload for Slack trigger", "error", err)
		return
	}
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(payload))
	if err != nil {
		slog.Error("Failed to create clear request for slack", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+*s.Slack_Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed to send clear request to slack", "error", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Error closing slack clear response body", "error", err)
		}
	}()
	var slackResp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&slackResp); err != nil {
		slog.Error("Failed to decode Slack clear response", "error", err)
		return
	}
	if !slackResp.OK {
		slog.Error("Slack clear notification failed", "error", slackResp.Error)
		return
	}
	slog.Info("Slack clear notification sent")
}

// Clear sends an all-clear message to Telegram once previously downed services recover.
func (t *TelegramTrigger) Clear() {
	payload, err := json.Marshal(map[string]any{
		"chat_id": *t.Telegram_Channel_Id,
		"text":    "All services have recovered",
	})
	if err != nil {
		slog.Error("failed to craft clear payload for telegram", "error", err)
		return
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", *t.Telegram_Token)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		slog.Error("failed to create clear request for telegram", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to send clear request for telegram", "error", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Error closing Telegram clear response body", "error", err)
		}
	}()
	var telegramResp struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(res.Body).Decode(&telegramResp); err != nil {
		slog.Error("Failed to decode Telegram clear response", "error", err)
		return
	}
	if !telegramResp.OK {
		slog.Error("Telegram clear notification failed", "description", telegramResp.Description)
		return
	}
	slog.Info("Telegram clear notification sent")
}

// Clear sends an all-clear event to Home Assistant once previously downed services recover.
func (h *HATrigger) Clear() {
	body, err := json.Marshal(map[string]any{
		"details": "All services have recovered",
	})
	if err != nil {
		slog.Error("failed to create HA clear json payload", "error", err)
		return
	}
	url := strings.TrimRight(*h.HA_URL, "/") + "/api/events/goup_alert"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		slog.Error("failed to create HA clear request", "error", err)
		return
	}
	req.Header.Add("Authorization", "Bearer "+*h.HA_Token)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to send clear request for HA", "error", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("error closing HA clear response body", "error", err)
		}
	}()
	if res.StatusCode != 200 && res.StatusCode != 201 {
		slog.Error("Failed to send clear sensor state", "status_code", res.StatusCode)
		return
	}
	slog.Info("HA clear notification sent")
}

// Clear sends an all-clear message to Discord once previously downed services recover.
func (d *DiscordTrigger) Clear() {
	payload, err := json.Marshal(map[string]any{
		"content": "**All Clear**\nAll services have recovered",
	})
	if err != nil {
		slog.Error("failed to create clear payload for discord notification request", "error", err)
		return
	}
	api_url := "https://discord.com/api"
	req, err := http.NewRequest("POST", api_url+fmt.Sprintf("/channels/%s/messages", *d.Discord_Channel), bytes.NewBuffer(payload))
	if err != nil {
		slog.Error("failed to create clear request for discord notification", "error", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", *d.Discord_Auth)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("error sending discord clear notification", "error", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("error closing discord clear response body", "error", err)
		}
	}()
	if res.StatusCode != 200 {
		slog.Error("error in discord clear notification response", "status_code", res.StatusCode)
		return
	}
	slog.Info("discord clear notification sent")
}
