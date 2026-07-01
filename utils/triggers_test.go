package utils_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"goUp/utils"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// interceptTLS creates a TLS test server and redirects http.DefaultTransport so all
// outbound connections (including HTTPS to hardcoded external hosts) go to it instead.
// The returned cleanup function restores the original transport and closes the server.
func interceptTLS(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	old := http.DefaultTransport
	http.DefaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("tcp", server.Listener.Addr().String())
		},
	}
	return server, func() {
		http.DefaultTransport = old
		server.Close()
	}
}

func TestSetupTrigger(t *testing.T) {
	broker := "tcp://localhost:1883"
	webhookURL := "http://localhost/webhook"
	cfg := &utils.Config{
		Triggers: utils.Trigger{
			MQTT: utils.MQTTTrigger{
				Mqtt_broker: &broker,
			},
			Webhook: utils.WebhookTrigger{
				Webhook_url: &webhookURL,
			},
		},
	}

	triggers := utils.SetupTrigger(cfg)

	if triggers.MQTT.Mqtt_broker == nil || *triggers.MQTT.Mqtt_broker != broker {
		t.Errorf("Expected MQTT broker %s, got %v", broker, triggers.MQTT.Mqtt_broker)
	}
	if triggers.Webhook.Webhook_url == nil || *triggers.Webhook.Webhook_url != webhookURL {
		t.Errorf("Expected Webhook URL %s, got %v", webhookURL, triggers.Webhook.Webhook_url)
	}
}

func TestFireWebhook(t *testing.T) {
	var receivedData []utils.ServiceData

	// Mock server to receive the webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header 'Bearer test-key', got '%s'", r.Header.Get("Authorization"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Error reading request body: %v", err)
		}
		defer func() {
			if err := r.Body.Close(); err != nil {
				t.Errorf("Failed to close response body: %v", err)
			}
		}()

		err = json.Unmarshal(body, &receivedData)
		if err != nil {
			t.Fatalf("Error unmarshaling received data: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	key := "Bearer test-key"
	trigger := utils.Trigger{
		Webhook: utils.WebhookTrigger{
			Webhook_url:        &server.URL,
			Webhook_key_string: &key,
		},
	}

	testData := []utils.ServiceData{
		{ServiceName: "down_service", ServiceHTTPResponse: "503", Error: true},
	}

	trigger.Webhook.Fire(testData)

	if len(receivedData) != 1 {
		t.Fatalf("Expected to receive 1 service data object, got %d", len(receivedData))
	}
	if receivedData[0].ServiceName != "down_service" {
		t.Errorf("Expected service name 'down_service', got '%s'", receivedData[0].ServiceName)
	}
}

func TestWebhookCustomMessage(t *testing.T) {
	customMessage := `{"message": "Services are down!", "services": []}`
	trigger := utils.Trigger{
		Webhook: utils.WebhookTrigger{
			Custom_message: &customMessage,
		},
	}

	testData := []utils.ServiceData{
		{ServiceName: "down_service", ServiceHTTPResponse: "503", Error: true},
	}

	var receivedPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		defer func() {
			if err := r.Body.Close(); err != nil {
				t.Errorf("Failed to close response body: %v", err)
			}
		}()
		if err := json.Unmarshal(body, &receivedPayload); err != nil {
			t.Errorf("Failed to unmarshal request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	trigger.Webhook.Webhook_url = &server.URL
	trigger.Webhook.Fire(testData)

	if msg, ok := receivedPayload["message"]; !ok || msg != "Services are down!" {
		t.Errorf("Expected custom message 'Services are down!', got '%v'", msg)
	}

	services, ok := receivedPayload["services"].([]interface{})
	if !ok {
		t.Fatalf("Expected 'services' to be a slice, got %T", receivedPayload["services"])
	}
	if len(services) != 1 {
		t.Errorf("Expected 1 service in payload, got %d", len(services))
	}

	serviceMap, ok := services[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected service to be a map, got %T", services[0])
	}

	if serviceMap["name"] != "down_service" {
		t.Errorf("Expected service name 'down_service', got '%s'", serviceMap["name"])
	}
}

func TestMQTTIsConfigured(t *testing.T) {
	m := utils.MQTTTrigger{}
	if m.IsConfigured() {
		t.Error("expected false when broker is nil")
	}
	broker := "tcp://localhost:1883"
	m.Mqtt_broker = &broker
	if !m.IsConfigured() {
		t.Error("expected true when broker is set")
	}
}

func TestWebhookIsConfigured(t *testing.T) {
	w := utils.WebhookTrigger{}
	if w.IsConfigured() {
		t.Error("expected false when URL is nil")
	}
	url := "http://localhost/hook"
	w.Webhook_url = &url
	if !w.IsConfigured() {
		t.Error("expected true when URL is set")
	}
}

func TestSMTPIsConfigured(t *testing.T) {
	e := utils.SMTPTrigger{}
	if e.IsConfigured() {
		t.Error("expected false when unconfigured")
	}
	server := "smtp.example.com:587"
	email := "user@example.com"
	pass := "secret"
	e.SMTPServer = &server
	if e.IsConfigured() {
		t.Error("expected false with only server set")
	}
	e.Email = &email
	if e.IsConfigured() {
		t.Error("expected false with server and email but no password")
	}
	e.App_Password = &pass
	if !e.IsConfigured() {
		t.Error("expected true when all three fields are set")
	}
}

func TestGotifyIsConfigured(t *testing.T) {
	g := utils.GotifyTrigger{}
	if g.IsConfigured() {
		t.Error("expected false when unconfigured")
	}
	srv := "http://gotify.example.com"
	g.Gotify_Server = &srv
	if g.IsConfigured() {
		t.Error("expected false with only server set")
	}
	tok := "token123"
	g.Gotify_Token = &tok
	if !g.IsConfigured() {
		t.Error("expected true when server and token are set")
	}
}

func TestFireGotify(t *testing.T) {
	var gotPath, gotKey string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("X-Gotify-Key")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	title := "Test Alert"
	priority := 7
	token := "gotify-token"
	g := utils.GotifyTrigger{
		Gotify_Server:   &server.URL,
		Gotify_Token:    &token,
		Gotify_Title:    &title,
		Gotify_Priority: &priority,
	}

	data := []utils.ServiceData{
		{ServiceName: "api", ServiceHTTPResponse: "500", ServiceAPIResponse: "error"},
	}
	g.Fire(data)

	if gotPath != "/message" {
		t.Errorf("expected path /message, got %s", gotPath)
	}
	if gotKey != token {
		t.Errorf("expected X-Gotify-Key %q, got %q", token, gotKey)
	}

	var payload map[string]any
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload["title"] != title {
		t.Errorf("expected title %q, got %v", title, payload["title"])
	}
	if payload["priority"].(float64) != float64(priority) {
		t.Errorf("expected priority %d, got %v", priority, payload["priority"])
	}
	if payload["message"] == "" {
		t.Error("expected non-empty message body")
	}
}

func TestFireGotifyDefaults(t *testing.T) {
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	token := "tok"
	g := utils.GotifyTrigger{
		Gotify_Server: &server.URL,
		Gotify_Token:  &token,
	}
	g.Fire([]utils.ServiceData{{ServiceName: "svc"}})

	var payload map[string]any
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload["title"] != "GoUp Alert" {
		t.Errorf("expected default title 'GoUp Alert', got %v", payload["title"])
	}
	if payload["priority"].(float64) != 5 {
		t.Errorf("expected default priority 5, got %v", payload["priority"])
	}
}

func TestSlackIsConfigured(t *testing.T) {
	s := utils.SlackTrigger{}
	if s.IsConfigured() {
		t.Error("expected false when unconfigured")
	}
	tok := "xoxb-token"
	s.Slack_Token = &tok
	if s.IsConfigured() {
		t.Error("expected false with only token set")
	}
	ch := "#alerts"
	s.Slack_Channel = &ch
	if !s.IsConfigured() {
		t.Error("expected true when token and channel are set")
	}
}

func TestFireSlack(t *testing.T) {
	var gotAuth, gotBody string

	_, cleanup := interceptTLS(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	defer cleanup()

	tok := "xoxb-slack-token"
	ch := "#ops"
	s := utils.SlackTrigger{
		Slack_Token:   &tok,
		Slack_Channel: &ch,
	}

	data := []utils.ServiceData{
		{ServiceName: "db", ServiceHTTPResponse: "503"},
	}
	s.Fire(data)

	if gotAuth != "Bearer "+tok {
		t.Errorf("expected Authorization 'Bearer %s', got %q", tok, gotAuth)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
		t.Fatalf("failed to unmarshal Slack payload: %v", err)
	}
	if payload["channel"] != ch {
		t.Errorf("expected channel %q, got %v", ch, payload["channel"])
	}
	if payload["text"] == "" {
		t.Error("expected non-empty text in Slack payload")
	}
	if payload["username"] != "GoUp Bot" {
		t.Errorf("expected default username 'GoUp Bot', got %v", payload["username"])
	}
}

func TestFireSlackCustomUsername(t *testing.T) {
	var gotBody string

	_, cleanup := interceptTLS(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	defer cleanup()

	tok := "tok"
	ch := "#ch"
	name := "Watcher"
	s := utils.SlackTrigger{
		Slack_Token:   &tok,
		Slack_Channel: &ch,
		Bot_Username:  &name,
	}
	s.Fire([]utils.ServiceData{{ServiceName: "svc"}})

	var payload map[string]any
	if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
		print("failed to marshal json")
	}
	if payload["username"] != name {
		t.Errorf("expected username %q, got %v", name, payload["username"])
	}
}

func TestTelegramIsConfigured(t *testing.T) {
	tg := utils.TelegramTrigger{}
	if tg.IsConfigured() {
		t.Error("expected false when unconfigured")
	}
	tok := "bot-token"
	tg.Telegram_Token = &tok
	if tg.IsConfigured() {
		t.Error("expected false with only token set")
	}
	ch := "-1001234567890"
	tg.Telegram_Channel_Id = &ch
	if !tg.IsConfigured() {
		t.Error("expected true when token and channel ID are set")
	}
}

func TestFireTelegram(t *testing.T) {
	var gotBody []byte

	_, cleanup := interceptTLS(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	defer cleanup()

	tok := "123456:ABC-telegram-token"
	ch := "-1001234567890"
	tg := utils.TelegramTrigger{
		Telegram_Token:      &tok,
		Telegram_Channel_Id: &ch,
	}

	data := []utils.ServiceData{
		{ServiceName: "cache", ServiceHTTPResponse: "500"},
	}
	tg.Fire(data)

	var payload map[string]any
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("failed to unmarshal Telegram payload: %v", err)
	}
	if payload["chat_id"] != ch {
		t.Errorf("expected chat_id %q, got %v", ch, payload["chat_id"])
	}
	if payload["text"] == "" {
		t.Error("expected non-empty text in Telegram payload")
	}
}

func TestHAIsConfigured(t *testing.T) {
	h := utils.HATrigger{}
	if h.IsConfigured() {
		t.Error("expected false when unconfigured")
	}
	url := "http://homeassistant.local:8123"
	h.HA_URL = &url
	if h.IsConfigured() {
		t.Error("expected false with only URL set")
	}
	tok := "ha-long-lived-token"
	h.HA_Token = &tok
	if !h.IsConfigured() {
		t.Error("expected true when URL and token are set")
	}
}

func TestFireHA(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tok := "ha-token"
	h := utils.HATrigger{
		HA_URL:   &server.URL,
		HA_Token: &tok,
	}

	data := []utils.ServiceData{
		{ServiceName: "lights", ServiceHTTPResponse: "503"},
	}
	h.Fire(data)

	if gotPath != "/api/events/goup_alert" {
		t.Errorf("expected path /api/events/goup_alert, got %s", gotPath)
	}
	if gotAuth != "Bearer "+tok {
		t.Errorf("expected Authorization 'Bearer %s', got %q", tok, gotAuth)
	}

	var payload map[string]any
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("failed to unmarshal HA payload: %v", err)
	}
	if payload["details"] == "" {
		t.Error("expected non-empty details in HA payload")
	}
}

func TestDiscordIsConfigured(t *testing.T) {
	d := utils.DiscordTrigger{}
	if d.IsConfigured() {
		t.Error("expected false when unconfigured")
	}
	auth := "Bot my-discord-token"
	d.Discord_Auth = &auth
	if d.IsConfigured() {
		t.Error("expected false with only auth set")
	}
	ch := "123456789012345678"
	d.Discord_Channel = &ch
	if !d.IsConfigured() {
		t.Error("expected true when auth and channel are set")
	}
}

func TestFireDiscord(t *testing.T) {
	var gotAuth string
	var gotBody []byte

	_, cleanup := interceptTLS(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	defer cleanup()

	auth := "Bot discord-bot-token"
	ch := "987654321098765432"
	d := utils.DiscordTrigger{
		Discord_Auth:    &auth,
		Discord_Channel: &ch,
	}

	data := []utils.ServiceData{
		{ServiceName: "api", ServiceHTTPResponse: "500"},
	}
	d.Fire(data)

	if gotAuth != auth {
		t.Errorf("expected Authorization %q, got %q", auth, gotAuth)
	}

	var payload map[string]any
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("failed to unmarshal Discord payload: %v", err)
	}
	content, ok := payload["content"].(string)
	if !ok || content == "" {
		t.Errorf("expected non-empty content in Discord payload, got %v", payload["content"])
	}
}

func TestFireWithBackoff(t *testing.T) {
	hits := make(chan struct{}, 10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	backoff := "50ms"
	cfg := &utils.Config{
		Triggers: utils.Trigger{
			Backoff_Period: &backoff,
			Webhook: utils.WebhookTrigger{
				Webhook_url: &server.URL,
			},
		},
	}
	trigger := utils.SetupTrigger(cfg)
	data := []utils.ServiceData{{ServiceName: "svc", ServiceHTTPResponse: "200"}}

	// First fire must reach the server.
	trigger.Fire(data)
	select {
	case <-hits:
	case <-time.After(2 * time.Second):
		t.Fatal("first Fire() never reached webhook")
	}

	// Immediate second fire must be suppressed by backoff.
	trigger.Fire(data)
	select {
	case <-hits:
		t.Error("second Fire() should have been blocked by backoff but reached webhook")
	case <-time.After(100 * time.Millisecond):
		// correct — nothing received within backoff window
	}

	// After the 100 ms wait above the 50 ms backoff has expired; fire again.
	trigger.Fire(data)
	select {
	case <-hits:
	case <-time.After(2 * time.Second):
		t.Fatal("third Fire() never reached webhook after backoff expired")
	}
}
