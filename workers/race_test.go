package workers_test

// TestConcurrentProgrammaticAndExternalConfigWrites stress-tests the hot
// reload coordination under concurrent writes.
//
// Two types of writer run in parallel each round:
//
//  1. Programmatic goroutines — add a unique service through the shared
//     ConfigStore.Update (clone, persist via temp file + rename, publish).
//
//  2. External goroutines — write valid YAML directly to the file via
//     os.WriteFile, mimicking a user editing the file in an editor while
//     the server is running.
//
// The hot reloader runs with a 50 ms interval so it fires several times per
// round, calling store.Reload concurrently with the Updates. The test verifies:
//
//   - No panic from concurrent file I/O or store operations.
//   - The hot reloader does not crash or return early.
//   - The config file is still parseable valid YAML after all writes finish.
//
// Run with -race to surface any data races on the store's shared state.

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"goUp/utils"
	"goUp/workers"
)

func TestConcurrentProgrammaticAndExternalConfigWrites(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent-write race test in short mode")
	}

	// Use a temp dir so this test is fully isolated from others.
	dir := t.TempDir()
	path := dir + "/race_config.yml"

	const initialYAML = `db_path: ./race_test.db
services:
  seed:
    url: https://seed.example.com
`
	if err := os.WriteFile(path, []byte(initialYAML), 0644); err != nil {
		t.Fatalf("write initial config: %v", err)
	}

	store := utils.NewConfigStore(path)
	if err := store.LoadFromFile(); err != nil {
		t.Fatalf("load initial config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Hot reloader at 50 ms fires ~3 times per 150 ms inter-round sleep.
	reloaderDone := make(chan struct{})
	go func() {
		defer close(reloaderDone)
		workers.StartHotReloaderWithInterval(ctx, store, 50*time.Millisecond)
	}()

	const (
		rounds        = 5
		programmaticN = 3 // goroutines per round doing AddConfigService
		externalN     = 2 // goroutines per round doing raw os.WriteFile
	)

	for round := range rounds {
		var wg sync.WaitGroup

		// Programmatic writers: all share one store; Update serializes them.
		for i := range programmaticN {
			wg.Add(1)
			go func(round, id int) {
				defer wg.Done()
				svc := utils.Service{
					Name: fmt.Sprintf("prog-r%d-g%d", round, id),
					URL:  fmt.Sprintf("https://prog-r%d-g%d.example.com", round, id),
				}
				if err := store.Update(func(c *utils.Config) error {
					_, err := utils.AddConfigService(c, svc)
					return err
				}); err != nil {
					t.Logf("round %d prog %d: AddConfigService: %v", round, id, err)
				}
			}(round, i)
		}

		// External writers: simulate an editor writing the file directly,
		// bypassing the store.
		for i := range externalN {
			wg.Add(1)
			go func(round, id int) {
				defer wg.Done()
				yaml := fmt.Sprintf(`db_path: ./race_test.db
services:
  ext-r%d-g%d:
    url: https://ext-r%d-g%d.example.com
`, round, id, round, id)
				if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
					t.Logf("round %d ext %d: WriteFile: %v", round, id, err)
				}
			}(round, i)
		}

		wg.Wait()

		// Let the hot reloader fire several times before the next round.
		time.Sleep(150 * time.Millisecond)
	}

	cancel()

	select {
	case <-reloaderDone:
	case <-time.After(2 * time.Second):
		t.Error("hot reloader did not stop within 2 s after context cancellation")
	}

	// Final sanity check: whatever is on disk must be valid parseable YAML.
	final := utils.NewConfigStore(path)
	if err := final.LoadFromFile(); err != nil {
		t.Fatalf("config file corrupted after concurrent writes: %v", err)
	}
	if final.Get() == nil {
		t.Fatal("LoadFromFile published nil after concurrent writes")
	}
}
