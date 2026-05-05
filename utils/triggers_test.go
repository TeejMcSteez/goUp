package utils_test

import (
	"encoding/json"
	"goUp/utils"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
