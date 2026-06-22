package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SetupTrigger copies Trigger config from cfg and registers configured handlers.
func SetupTrigger(cfg *Config) *Trigger {
	t := &cfg.Triggers
	t.handlers = nil

	if cfg.Triggers.Backoff_Period != nil && *cfg.Triggers.Backoff_Period != "" {
		dur, err := time.ParseDuration(*cfg.Triggers.Backoff_Period)
		if err != nil {
			log.Printf("Invalid Backoff_Period '%s': %v. Disabling backoff.", *cfg.Triggers.Backoff_Period, err)
			t.backoffDuration = 0
		} else {
			log.Printf("Trigger backoff period set to %s", dur)
			t.backoffDuration = dur
		}
	} else {
		log.Println("No backup period setup!")
	}

	if t.MQTT.IsConfigured() {
		t.handlers = append(t.handlers, &t.MQTT)
	}
	if t.Webhook.IsConfigured() {
		t.handlers = append(t.handlers, &t.Webhook)
	}
	if t.SMTP.IsConfigured() {
		t.handlers = append(t.handlers, &t.SMTP)
	}
	if t.Gotify.IsConfigured() {
		t.handlers = append(t.handlers, &t.Gotify)
	}
	if t.Slack.IsConfigured() {
		t.handlers = append(t.handlers, &t.Slack)
	}
	if t.Telegram.IsConfigured() {
		t.handlers = append(t.handlers, &t.Telegram)
	}
	if t.HA.IsConfigured() {
		t.handlers = append(t.handlers, &t.HA)
	}
	if len(t.handlers) == 0 {
		log.Println("No MQTT broker, Webhook URL, or SMTP server setup, exiting trigger setup")
		return t
	}

	log.Println("Triggers setup")
	return t
}

// Fire dispatches service data to all configured trigger handlers.
func (t *Trigger) Fire(data []ServiceData) {
	if len(t.handlers) == 0 {
		return
	}

	if t.backoffDuration > 0 && !t.lastFired.IsZero() && time.Since(t.lastFired) < t.backoffDuration {
		log.Printf("Trigger backoff period active, skipping. Last trigger was %s ago.", time.Since(t.lastFired))
		return
	}

	for _, h := range t.handlers {
		go h.Fire(data)
	}

	t.lastFired = time.Now()
}

func (m *MQTTTrigger) IsConfigured() bool {
	return m.Mqtt_broker != nil
}

func (m *MQTTTrigger) Fire(data []ServiceData) {
	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		log.Println("Connected to MQTT Broker")
	}

	var lostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		log.Println("Connection Lost: " + err.Error())
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
		log.Println("Error pushing to MQTT client: " + token.Error().Error())
	}

	if jsonData, err := json.Marshal(data); err != nil {
		log.Printf("Error formatting service data into JSON: %v\n", err)
	} else {
		// Keep retain to true to store last known good message
		token := client.Publish("goup_status", 0, true, jsonData)
		token.Done()

		if err := token.Error(); err != nil {
			log.Printf("Error with MQTT token: %v\n", err)
		}
		log.Println("Disconnecting from MQTT broker, sent message complete")
	}

	client.Disconnect(500)
}

// IsConfigured reports whether the webhook trigger has a URL set.
func (w *WebhookTrigger) IsConfigured() bool {
	return w.Webhook_url != nil
}

// Fire sends service data to the configured webhook URL.
func (w *WebhookTrigger) Fire(data []ServiceData) {
	jsonMessage, err := w.buildMessage(data)
	if err != nil {
		log.Printf("Failed parsing json service data message: %v\n", err)
	}
	log.Println("Firing webhook")

	req, err := http.NewRequest("POST", *w.Webhook_url, jsonMessage)
	if err != nil {
		log.Printf("Error creating webhook request: %v\n", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	if w.Webhook_key_string != nil {
		req.Header.Add("Authorization", *w.Webhook_key_string)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error sending webhook: %v\n", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Error closing webhook request: %v", err)
		}
	}()

	log.Printf("Webhook sent, status: %s\n", res.Status)
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
		log.Printf("Error occured while parsing webhook message: %v\n", err)
		return nil, err
	}
	return bytes.NewBuffer(jsonSvcData), nil
}

// Fire sends all downed service data as an email
func (e *SMTPTrigger) Fire(data []ServiceData) {
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
		log.Printf("Error SMTP request: %v", err)
		return
	}
	log.Print("SMTP message sent")
}

// Checks if all parameters for the SMTP trigger is is configured
func (e *SMTPTrigger) IsConfigured() bool {
	return e.SMTPServer != nil && e.Email != nil && e.App_Password != nil
}

// Fire sends a Gotify push notification for all downed services.
func (g *GotifyTrigger) Fire(data []ServiceData) {
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
		log.Printf("Error building Gotify payload: %v", err)
		return
	}

	url := strings.TrimRight(*g.Gotify_Server, "/") + "/message"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error creating Gotify request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", *g.Gotify_Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error sending Gotify notification: %v", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Error closing Gotify response: %v", err)
		}
	}()

	log.Printf("Gotify notification sent, status: %s\n", res.Status)
}

func (g *GotifyTrigger) IsConfigured() bool {
	return g.Gotify_Server != nil && g.Gotify_Token != nil
}

func (s *SlackTrigger) Fire(data []ServiceData) {
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
		log.Printf("Failed craft payload for Slack trigger: %v", err)
		return
	}
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Failed to create request for slack: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+*s.Slack_Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to send request to slack: %v", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Error closing slack response body: %v", err)
		}
	}()
	var slackResp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&slackResp); err != nil {
		log.Printf("Failed to decode Slack response: %v", err)
		return
	}
	if !slackResp.OK {
		log.Printf("Slack notification failed: %s", slackResp.Error)
		return
	}
	log.Println("Slack notification sent")
}

func (s *SlackTrigger) IsConfigured() bool {
	return s.Slack_Token != nil && s.Slack_Channel != nil
}

func (t *TelegramTrigger) Fire(data []ServiceData) {
	var text strings.Builder
	for _, entry := range data {
		fmt.Fprintf(&text, "Service: %s\nResponse: %s\nAPI Response: %s", entry.ServiceName, entry.ServiceHTTPResponse, entry.ServiceAPIResponse)
	}
	payload, err := json.Marshal(map[string]any{
		"chat_id": *t.Telegram_Channel_Id,
		"text":    text.String(),
	})
	if err != nil {
		log.Printf("failed to craft payload for telegram: %v", err)
		return
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", *t.Telegram_Token)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("failed to create request for telegram: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("failed to send request for telegram: %v", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Error closing Telegram response body: %v", err)
		}
	}()
	var telegramResp struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(res.Body).Decode(&telegramResp); err != nil {
		log.Printf("Failed to decode Telegram response: %v", err)
		return
	}
	if !telegramResp.OK {
		log.Printf("Telegram notification failed: %s", telegramResp.Description)
		return
	}
	log.Println("Telegram notification sent")
}

func (t *TelegramTrigger) IsConfigured() bool {
	return t.Telegram_Channel_Id != nil && t.Telegram_Token != nil
}

func (h *HATrigger) Fire(data []ServiceData) {
	var extraStateAttribs strings.Builder
	for _, entry := range data {
		fmt.Fprintf(&extraStateAttribs, "Service: %s\nResponse: %s\nAPI Response: %s", entry.ServiceName, entry.ServiceHTTPResponse, entry.ServiceAPIResponse)
	}
	body, err := json.Marshal(map[string]any{
		"details": extraStateAttribs.String(),
	})
	if err != nil {
		log.Printf("failed to create HA json payload: %v", err)
		return
	}
	url := strings.TrimRight(*h.HA_URL, "/") + "/api/events/goup_alert"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("failed to create HA request: %v", err)
		return
	}
	req.Header.Add("Authorization", "Bearer "+*h.HA_Token)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("failed to send request for HA: %v", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("error closing HA response body: %v", err)
		}
	}()
	if res.StatusCode != 200 && res.StatusCode != 201 {
		log.Printf("Failed to send or post sensor state\n Status Code: %d", res.StatusCode)
		return
	}
	log.Println("HA notification sent")
}

func (h *HATrigger) IsConfigured() bool {
	return h.HA_URL != nil && h.HA_Token != nil
}
