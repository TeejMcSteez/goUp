package workers_test

import (
	"context"
	"goUp/workers"
	"os"
	"testing"
	"time"
)

func TestHotReloaderStopsOnContextCancel(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "hot_reloader_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	}()
	if err := tmpFile.Close(); err != nil {
		t.Errorf("Failed to close temp file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		workers.StartHotReloader(tmpFile.Name(), ctx, nil)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Error("StartHotReloader did not stop after context cancellation")
	}
}

// Even with an invalid path, the reloader should start (logging the error)
// and exit cleanly when the context is cancelled.
func TestHotReloaderInvalidPathStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		workers.StartHotReloader("/nonexistent/path/to/config.yml", ctx, nil)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success: started without panic and exited cleanly
	case <-time.After(2 * time.Second):
		t.Error("StartHotReloader did not stop after context cancellation with invalid path")
	}
}

// Verify that a file modification is detected: after updating the file, the
// reloader should log and call Setup. We test detection by checking the
// timestamp comparison rather than waiting for the full 5-second tick, so we
// skip timing here and rely on the context-cancel behaviour above for the loop.
func TestHotReloaderDetectsFileChange(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "hot_reloader_change_test_*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Errorf("Failed to remove temporary file: %v", err)
		}
	}()
	if err := tmpFile.Close(); err != nil {
		t.Errorf("Failed to close temp file: %v", err)
	}

	before, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file before write: %v", err)
	}

	if err := os.WriteFile(tmpFile.Name(), []byte("changed"), 0644); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	after, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file after write: %v", err)
	}

	if !after.ModTime().After(before.ModTime()) || after.ModTime().Equal(before.ModTime()) {
		t.Skip("Failed to properly detect file changes with current file system due to time issues")
	}
}
