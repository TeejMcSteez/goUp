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