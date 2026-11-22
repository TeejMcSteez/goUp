package workers_test

import (
	"context"
	"goUp/utils"
	"goUp/workers"
	"testing"
	"time"
)

const oneGB = 1 * 1e9

func maxSizeConfig(s string) *utils.Config {
	return &utils.Config{Database_Max_Size: &s}
}

func TestGetMaxSizeNilConfig(t *testing.T) {
	got, err := utils.GetMaxSize(nil)
	if err != nil {
		t.Fatalf("Unexpected error for nil config: %v", err)
	}
	if got != oneGB {
		t.Errorf("Expected 1GB default for nil config, got %v", got)
	}
}

func TestGetMaxSizeNilMaxSize(t *testing.T) {
	got, err := utils.GetMaxSize(&utils.Config{})
	if err != nil {
		t.Fatalf("Unexpected error when Database_Max_Size is nil: %v", err)
	}
	if got != oneGB {
		t.Errorf("Expected 1GB default when Database_Max_Size is nil, got %v", got)
	}
}

func TestGetMaxSizeKB(t *testing.T) {
	got, err := utils.GetMaxSize(maxSizeConfig("100kb"))
	if err != nil {
		t.Fatalf("Unexpected error for '100kb': %v", err)
	}
	want := 100.0 * 1000
	if got != want {
		t.Errorf("Expected %v bytes for '100kb', got %v", want, got)
	}
}

func TestGetMaxSizeMB(t *testing.T) {
	got, err := utils.GetMaxSize(maxSizeConfig("500mb"))
	if err != nil {
		t.Fatalf("Unexpected error for '500mb': %v", err)
	}
	want := 500.0 * 1e6
	if got != want {
		t.Errorf("Expected %v bytes for '500mb', got %v", want, got)
	}
}

func TestGetMaxSizeGB(t *testing.T) {
	got, err := utils.GetMaxSize(maxSizeConfig("2gb"))
	if err != nil {
		t.Fatalf("Unexpected error for '2gb': %v", err)
	}
	want := 2.0 * 1e9
	if got != want {
		t.Errorf("Expected %v bytes for '2gb', got %v", want, got)
	}
}

func TestGetMaxSizeBelowMinimumKB(t *testing.T) {
	_, err := utils.GetMaxSize(maxSizeConfig("3kb"))
	if err == nil {
		t.Error("Expected error for size below 4KB minimum, got nil")
	}
}

func TestGetMaxSizeInvalidUnit(t *testing.T) {
	_, err := utils.GetMaxSize(maxSizeConfig("100tb"))
	if err == nil {
		t.Error("Expected error for unknown unit 'tb', got nil")
	}
}

func TestGetMaxSizeInvalidFormat(t *testing.T) {
	_, err := utils.GetMaxSize(maxSizeConfig("notasize"))
	if err == nil {
		t.Error("Expected error for invalid format 'notasize', got nil")
	}
}

func TestMemoryWatcherStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		workers.StartMemoryWatcher(ctx, nil, utils.NewConfigStore("unused.yml"))
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Error("StartMemoryWatcher did not stop after context cancellation")
	}
}
