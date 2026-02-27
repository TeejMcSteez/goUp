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
	got := workers.GetMaxSize()
	if got != oneGB {
		t.Errorf("Expected 1GB default for nil config, got %v", got)
	}
}

func TestGetMaxSizeNilMaxSize(t *testing.T) {
	utils.Current_Config = &utils.Config{}
	defer func() { utils.Current_Config = nil }()

	got := workers.GetMaxSize()
	if got != oneGB {
		t.Errorf("Expected 1GB default when Database_Max_Size is nil, got %v", got)
	}
}

func TestGetMaxSizeKB(t *testing.T) {
	defer setMaxSize("100kb")()

	got := workers.GetMaxSize()
	want := 100.0 * 1000
	if got != want {
		t.Errorf("Expected %v bytes for '100kb', got %v", want, got)
	}
}

func TestGetMaxSizeMB(t *testing.T) {
	defer setMaxSize("500mb")()

	got := workers.GetMaxSize()
	want := 500.0 * 1e6
	if got != want {
		t.Errorf("Expected %v bytes for '500mb', got %v", want, got)
	}
}

func TestGetMaxSizeGB(t *testing.T) {
	defer setMaxSize("2gb")()

	got := workers.GetMaxSize()
	want := 2.0 * 1e9
	if got != want {
		t.Errorf("Expected %v bytes for '2gb', got %v", want, got)
	}
}

// Values below 4KB are clamped to the 4KB minimum
func TestGetMaxSizeBelowMinimumKB(t *testing.T) {
	defer setMaxSize("3kb")()

	got := workers.GetMaxSize()
	want := 4.0 * 1000
	if got != want {
		t.Errorf("Expected minimum %v bytes for '3kb', got %v", want, got)
	}
}

func TestGetMaxSizeInvalidUnit(t *testing.T) {
	defer setMaxSize("100tb")()

	got := workers.GetMaxSize()
	if got != oneGB {
		t.Errorf("Expected 1GB default for unknown unit 'tb', got %v", got)
	}
}

func TestGetMaxSizeInvalidFormat(t *testing.T) {
	defer setMaxSize("notasize")()

	got := workers.GetMaxSize()
	if got != oneGB {
		t.Errorf("Expected 1GB default for invalid format, got %v", got)
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
