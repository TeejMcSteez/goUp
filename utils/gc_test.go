package utils_test

import (
	"database/sql"
	"goUp/utils"
	"os"
	"path/filepath"
	"testing"
)

// These tests guard the contract main.go relies on: every config change,
// whether it comes from a handler (store.Update) or the hot reloader
// (store.Reload), runs the store's subscribers, so a GC subscriber alone
// keeps the database clean without mutators calling DbGarbageCollect.

// setupGCStore builds a store and DB in a temp dir with services Alpha and
// Beta, seeds one row for each, and registers a GC subscriber identical to
// the one in main.go. The returned counter records how often it ran.
func setupGCStore(t *testing.T) (*utils.ConfigStore, *sql.DB, *int) {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "services.yml")
	dbPath := filepath.Join(dir, "gc.db")

	writeYML(t, cfgPath, "db_path: "+dbPath+`
services:
  Alpha:
    url: http://alpha.invalid
  Beta:
    url: http://beta.invalid
`)

	store := utils.NewConfigStore(cfgPath)
	if err := store.LoadFromFile(); err != nil {
		t.Fatalf("LoadFromFile: %v", err)
	}
	db, err := utils.InitDB(store.Get())
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close: %v", err)
		}
	})

	for _, name := range []string{"Alpha", "Beta"} {
		if err := utils.InsertData(db, utils.ServiceData{ServiceName: name, ServiceHTTPResponse: "200"}); err != nil {
			t.Fatalf("InsertData %s: %v", name, err)
		}
	}

	runs := 0
	// mirrors the GC subscriber registered in main.go
	store.Subscribe(func(_, new *utils.Config) {
		runs++
		if err := utils.DbGarbageCollect(db, new); err != nil {
			t.Errorf("DbGarbageCollect in subscriber: %v", err)
		}
	})
	return store, db, &runs
}

func writeYML(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func rowCount(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	rows, err := utils.GetDataForService(db, name)
	if err != nil {
		t.Fatalf("GetDataForService %s: %v", name, err)
	}
	return len(rows)
}

// Handlers mutate config through store.Update; removing a service there must
// drop its rows via the subscriber.
func TestStoreUpdateTriggersGC(t *testing.T) {
	store, db, runs := setupGCStore(t)

	if err := store.Update(func(c *utils.Config) error {
		delete(c.Services, "Beta")
		return nil
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if *runs != 1 {
		t.Fatalf("GC subscriber ran %d times, want 1", *runs)
	}
	if n := rowCount(t, db, "Beta"); n != 0 {
		t.Errorf("Beta rows after removal = %d, want 0", n)
	}
	if n := rowCount(t, db, "Alpha"); n != 1 {
		t.Errorf("Alpha rows = %d, want 1 (kept service must not be collected)", n)
	}
}

// The hot reloader calls store.Reload after an external edit; that path must
// run the GC subscriber too.
func TestStoreReloadTriggersGC(t *testing.T) {
	store, db, runs := setupGCStore(t)

	writeYML(t, store.Path(), "db_path: "+*store.Get().Database_Location+`
services:
  Beta:
    url: http://beta.invalid
`)
	changed, err := store.Reload()
	if err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if !changed {
		t.Fatal("Reload reported no change after external edit")
	}

	if *runs != 1 {
		t.Fatalf("GC subscriber ran %d times, want 1", *runs)
	}
	if n := rowCount(t, db, "Alpha"); n != 0 {
		t.Errorf("Alpha rows after external removal = %d, want 0", n)
	}
	if n := rowCount(t, db, "Beta"); n != 1 {
		t.Errorf("Beta rows = %d, want 1", n)
	}
}

// Reload skips the store's own writes, so it must not run GC a second time
// for a change Update already handled.
func TestStoreReloadSkipsOwnWriteNoDoubleGC(t *testing.T) {
	store, _, runs := setupGCStore(t)

	if err := store.Update(func(c *utils.Config) error {
		delete(c.Services, "Beta")
		return nil
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	changed, err := store.Reload()
	if err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if changed {
		t.Error("Reload treated the store's own write as an external change")
	}
	if *runs != 1 {
		t.Errorf("GC subscriber ran %d times, want 1", *runs)
	}
}

// A rename (configServicePut) moves rows inside the Update closure, before
// subscribers run, so GC must keep the renamed rows rather than delete them.
func TestStoreRenameSurvivesGC(t *testing.T) {
	store, db, _ := setupGCStore(t)

	if err := store.Update(func(c *utils.Config) error {
		renamed := c.Services["Beta"]
		renamed.Name = "Gamma"
		_, err := utils.UpdateConfigService(c, "Beta", renamed, db)
		return err
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if n := rowCount(t, db, "Gamma"); n != 1 {
		t.Errorf("Gamma rows after rename = %d, want 1 (GC must not drop renamed rows)", n)
	}
	if n := rowCount(t, db, "Beta"); n != 0 {
		t.Errorf("Beta rows after rename = %d, want 0", n)
	}
}
