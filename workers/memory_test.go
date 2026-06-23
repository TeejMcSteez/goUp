package workers_test

import (
	"context"
	"goUp/utils"
	"goUp/workers"
	"testing"
	"time"
)

const oneGB = 1 * 1e9

func setMaxSize(s string) func() {
	utils.Current_Config = &utils.Config{Database_Max_Size: &s}
	return func() { utils.Current_Config = nil }
}

func TestGetMaxSizeNilConfig(t *testing.T) {
	utils.Current_Config = nil
	got, err := utils.GetMaxSize()
	if err != nil {
		t.Fatalf("Unexpected error for nil config: %v", err)
	}
	if got != oneGB {
		t.Errorf("Expected 1GB default for nil config, got %v", got)
	}
}

func TestGetMaxSizeNilMaxSize(t *testing.T) {
	utils.Current_Config = &utils.Config{}
	defer func() { utils.Current_Config = nil }()

	got, err := utils.GetMaxSize()
	if err != nil {
		t.Fatalf("Unexpected error when Database_Max_Size is nil: %v", err)
	}
	if got != oneGB {
		t.Errorf("Expected 1GB default when Database_Max_Size is nil, got %v", got)
	}
}

func TestGetMaxSizeKB(t *testing.T) {
	defer setMaxSize("100kb")()

	got, err := utils.GetMaxSize()
	if err != nil {
		t.Fatalf("Unexpected error for '100kb': %v", err)
	}
	want := 100.0 * 1000
	if got != want {
		t.Errorf("Expected %v bytes for '100kb', got %v", want, got)
	}
}

func TestGetMaxSizeMB(t *testing.T) {
	defer setMaxSize("500mb")()

	got, err := utils.GetMaxSize()
	if err != nil {
		t.Fatalf("Unexpected error for '500mb': %v", err)
	}
	want := 500.0 * 1e6
	if got != want {
		t.Errorf("Expected %v bytes for '500mb', got %v", want, got)
	}
}

func TestGetMaxSizeGB(t *testing.T) {
	defer setMaxSize("2gb")()

	got, err := utils.GetMaxSize()
	if err != nil {
		t.Fatalf("Unexpected error for '2gb': %v", err)
	}
	want := 2.0 * 1e9
	if got != want {
		t.Errorf("Expected %v bytes for '2gb', got %v", want, got)
	}
}

func TestGetMaxSizeBelowMinimumKB(t *testing.T) {
	defer setMaxSize("3kb")()

	_, err := utils.GetMaxSize()
	if err == nil {
		t.Error("Expected error for size below 4KB minimum, got nil")
	}
}

func TestGetMaxSizeInvalidUnit(t *testing.T) {
	defer setMaxSize("100tb")()

	_, err := utils.GetMaxSize()
	if err == nil {
		t.Error("Expected error for unknown unit 'tb', got nil")
	}
}

func TestGetMaxSizeInvalidFormat(t *testing.T) {
	defer setMaxSize("notasize")()

	_, err := utils.GetMaxSize()
	if err == nil {
		t.Error("Expected error for invalid format 'notasize', got nil")
	}
}

func TestMemoryWatcherStopsOnContextCancel(t *testing.T) {
	utils.Current_Config = nil // triggers default 1GB path in GetMaxSize
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		workers.StartMemoryWatcher(ctx, nil)
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
