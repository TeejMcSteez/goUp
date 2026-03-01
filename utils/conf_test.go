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
	defer os.Remove(filePath)

	cfg, err := utils.LoadConfig(filePath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected config to be loaded, but it was nil")
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

	if err := utils.UpdateConfigService(conf, newService); err != nil {
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

	if err := utils.UpdateConfigWebhookTrigger(conf, new_webhook); err != nil {
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
