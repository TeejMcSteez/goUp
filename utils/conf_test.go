package utils_test

import (
	"goUp/utils"
	"os"
	"testing"
)

func TestLoadConfigSuccess(t *testing.T) {
	content := `
db_path: "./test.db"
services:
  my-service:
    url: "http://example.com"
`
	filePath := "test_config.yml"
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}
	defer func() {
		if err := os.Remove(filePath); err != nil {
			t.Errorf("Failed to remove test database file")
		}
	}()

	cfg, err := utils.LoadConfig(filePath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected config to be loaded, but it was nil")
		return
	}
	if *cfg.Database_Location != "./test.db" {
		t.Errorf("Expected db_path to be './test.db', got '%s'", *cfg.Database_Location)
	}

	if _, ok := cfg.Services["my-service"]; !ok {
		t.Error("Expected 'my-service' to be in the services map")
	}
}

func TestAddServiceToConfig(t *testing.T) {
	ymlContent1 := `db_path: "./test_data.db"
services:
  service2:
    url: "https://example.com"
    retry: 2
  service3:
    url: "https://www.apple.com"
    retry: 2
`
	cleanup := createTestYML(ymlContent1, t)
	defer cleanup()

	newService := utils.Service{Name: "google", URL: "https://google.com/"}

	conf, err := utils.LoadConfig("./services.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	crn_len := len(conf.Services)

	if err := utils.AddConfigService(conf, newService); err != nil {
		t.Fatalf("UpdateConfigService failed: %v", err)
	}

	conf, err = utils.LoadConfig("./services.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	upd_len := len(conf.Services)

	if upd_len != crn_len+1 {
		t.Fatalf("New service not added to file after reload")
	}
}

func TestAddMQTTTrigger(t *testing.T) {
	ymlContent := `db_path: "./test_data.db"
services:
  service2:
    url: "https://example.com"
    retry: 2
`
	cleanup := createTestYML(ymlContent, t)
	defer cleanup()

	broker := "tcp://broker.example.com:1883"
	username := "mqttuser"
	key := "mqttpassword"
	newMQTT := utils.MQTTTrigger{Mqtt_broker: &broker, Mqtt_username: &username, Mqtt_key: &key}

	conf, err := utils.LoadConfig("./services.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if err := utils.AddConfigMQTTTrigger(conf, newMQTT); err != nil {
		t.Fatalf("UpdateConfigMQTTTrigger failed: %v", err)
	}

	conf, err = utils.LoadConfig("./services.yml")
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if conf.Triggers.MQTT.Mqtt_broker == nil {
		t.Fatal("Failed to add MQTT trigger: broker is nil")
	}
	if *conf.Triggers.MQTT.Mqtt_broker != broker {
		t.Errorf("Expected broker %q, got %q", broker, *conf.Triggers.MQTT.Mqtt_broker)
	}
}

func TestAddWebhookTrigger(t *testing.T) {
	ymlContent1 := `db_path: "./test_data.db"
services:
  service2:
    url: "https://example.com"
    retry: 2
  service3:
    url: "https://www.apple.com"
    retry: 2
`
	cleanup := createTestYML(ymlContent1, t)
	defer cleanup()
	url := "google.com"
	key := "something"
	custom_message := "lol"
	new_webhook := utils.WebhookTrigger{Webhook_url: &url, Webhook_key_string: &key, Custom_message: &custom_message}
	conf, err := utils.LoadConfig("./services.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if err := utils.AddConfigWebhookTrigger(conf, new_webhook); err != nil {
		t.Fatalf("UpdateConfigWebhookTrigger failed: %v", err)
	}

	conf, err = utils.LoadConfig("./services.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if conf.Triggers.Webhook.Webhook_url == nil {
		t.Fatal("Failed to add webhook")
	}

}

func TestDeleteService(t *testing.T) {
	ymlContent1 := `db_path: "./test_data.db"
services:
  example:
    url: "https://example.com"
    retry: 2
  apple:
    url: "https://www.apple.com"
    retry: 2
`
	cleanup := createTestYML(ymlContent1, t)
	defer cleanup()

	conf, err := utils.LoadConfig("./services.yml")

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	current_size := len(conf.Services)
	serviceToDelete := conf.Services["example"]

	db, cleanupDb := setupTestDB(t)
	defer cleanupDb()

	if err := utils.DeleteConfigService(conf, serviceToDelete, db); err != nil {
		t.Fatalf("Failed to delete config: %v", err)
	}

	conf, err = utils.LoadConfig("./services.yml")

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	new_size := len(conf.Services)

	if new_size >= current_size {
		t.Fatalf("Failed to delete service from config")
	}
}

func TestDeleteMQTT(t *testing.T) {
	ymlContent1 := `db_path: "./test_data.db"
services:
  "https://example.com":
    url: "https://example.com"
    retry: 2
  "https://www.apple.com":
    url: "https://www.apple.com"
    retry: 2
triggers:
  backoff: "30m"
  mqtt:
    mqtt_broker: "192.168.1.30:1883"
    mqtt_user: "teej"
    mqtt_key: "Intel11900k"
`
	cleanup := createTestYML(ymlContent1, t)
	defer cleanup()

	conf, err := utils.LoadConfig("./services.yml")

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if conf.Triggers.MQTT.Mqtt_broker == nil {
		t.Fatal("Failed to load MQTT config from setup")
	}

	if err := utils.DeleteConfigMQTT(conf); err != nil {
		t.Fatalf("Error occured deleting MQTT: %v", err)
	}

	if conf.Triggers.MQTT.Mqtt_broker != nil {
		t.Fatal("Broker is not nil")
	}

}

func TestDeleteWebhook(t *testing.T) {
	ymlContent1 := `db_path: "./test_data.db"
services:
  "https://example.com":
    url: "https://example.com"
    retry: 2
  "https://www.apple.com":
    url: "https://www.apple.com"
    retry: 2
triggers:
  backoff: "30m"
  webhook:
    webhook_url: "192.168.1.30:1883"
`
	cleanup := createTestYML(ymlContent1, t)
	defer cleanup()

	conf, err := utils.LoadConfig("./services.yml")

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if conf.Triggers.Webhook.Webhook_url == nil {
		t.Fatal("Failed to load Webhook config from setup")
	}

	if err := utils.DeleteConfigTrigger(conf); err != nil {
		t.Fatalf("Error occured deleting Webhook: %v", err)
	}

	if conf.Triggers.Webhook.Webhook_url != nil {
		t.Fatal("URL is not nil")
	}

}

func TestReadService(t *testing.T) {
	ymlContent1 := `db_path: "./test_data.db"
services:
  ex1:
    url: "https://example.com"
    retry: 2
  ex2:
    url: "https://www.apple.com"
    retry: 2
triggers:
  backoff: "30m"
  webhook:
    webhook_url: "192.168.1.30:1883"
`
	cleanup := createTestYML(ymlContent1, t)
	defer cleanup()

	conf, err := utils.LoadConfig("./services.yml")

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	services := utils.ReadConfigServices(conf)

	if _, ok := services["ex1"]; !ok {
		t.Fatal("Failed to find key")
	}

	if _, ok := services["ex2"]; !ok {
		t.Fatal("Failed to find key")
	}

}

func TestReadMQTT(t *testing.T) {
	ymlContent1 := `db_path: "./test_data.db"
services:
  ex_1:
    url: "https://example.com"
    retry: 2
  ex_2:
    url: "https://www.apple.com"
    retry: 2
triggers:
  backoff: "30m"
  mqtt:
    mqtt_broker: "192.168.1.30:1883"
`
	cleanup := createTestYML(ymlContent1, t)
	defer cleanup()

	conf, err := utils.LoadConfig("./services.yml")

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if m := utils.ReadConfigMQTT(conf); m.Mqtt_broker == nil || *m.Mqtt_broker != "192.168.1.30:1883" {
		t.Fatal("MQTT broker is nil or invalid")
	}
}

func TestReadWebhook(t *testing.T) {
	ymlContent1 := `db_path: "./test_data.db"
services:
  ex_1:
    url: "https://example.com"
    retry: 2
  ex_2:
    url: "https://www.apple.com"
    retry: 2
triggers:
  backoff: "30m"
  webhook:
    webhook_url: "192.168.1.30:1883"
`
	cleanup := createTestYML(ymlContent1, t)
	defer cleanup()

	conf, err := utils.LoadConfig("./services.yml")

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if w := utils.ReadConfigWebhook(conf); w.Webhook_url == nil || *w.Webhook_url != "192.168.1.30:1883" {
		t.Fatal("Webhook is nil or invalid")
	}
}
