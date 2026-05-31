package utils_test

import (
	"fmt"
	"goUp/utils"
	"os"
	"sync"
	"testing"
	"time"
)

// setupIntegrationDB creates an isolated DB + config for a single integration
// test. Using a unique path per test avoids conflicts when tests run in parallel.
func setupIntegrationDB(t *testing.T, services map[string]utils.Service) (*utils.Config, func()) {
	t.Helper()
	dbPath := fmt.Sprintf("./integration_%s.db", t.Name())

	cfg := &utils.Config{
		Database_Location: &dbPath,
		Services:          services,
	}
	utils.Current_Config = cfg

	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}

	return cfg, func() {
		db.Close()
		os.Remove(dbPath)
		utils.Current_Config = nil
	}
}

// TestGetRecentDataFiltersStaleRowsAfterRename verifies that rows for a service
// removed from config are excluded from GetRecentData even before GC runs.
func TestGetRecentDataFiltersStaleRowsAfterRename(t *testing.T) {
	cfg, cleanup := setupIntegrationDB(t, map[string]utils.Service{
		"Website": {Name: "Website", URL: "https://example.com"},
	})
	defer cleanup()

	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	// Insert a stale row for the old name "Web" and a live row for "Website".
	if err := utils.InsertData(db, utils.ServiceData{ServiceName: "Web", ServiceHTTPResponse: "200"}); err != nil {
		t.Fatalf("InsertData Web: %v", err)
	}
	if err := utils.InsertData(db, utils.ServiceData{ServiceName: "Website", ServiceHTTPResponse: "200"}); err != nil {
		t.Fatalf("InsertData Website: %v", err)
	}

	// Config only knows "Website" — stale "Web" row must be filtered.
	got, err := utils.GetRecentData(db)
	if err != nil {
		t.Fatalf("GetRecentData: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("Expected 1 result, got %d: %+v", len(got), got)
	}
	if got[0].ServiceName != "Website" {
		t.Errorf("Expected 'Website', got %q", got[0].ServiceName)
	}
	_ = cfg
}

// TestGetRecentDataNilConfigReturnsAll verifies that when no config is loaded
// the filter is skipped and all rows are returned.
func TestGetRecentDataNilConfigReturnsAll(t *testing.T) {
	_, cleanup := setupIntegrationDB(t, nil)
	defer cleanup()

	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	for _, name := range []string{"alpha", "beta", "gamma"} {
		if err := utils.InsertData(db, utils.ServiceData{ServiceName: name, ServiceHTTPResponse: "200"}); err != nil {
			t.Fatalf("InsertData %s: %v", name, err)
		}
	}

	utils.Current_Config = nil

	got, err := utils.GetRecentData(db)
	if err != nil {
		t.Fatalf("GetRecentData: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("Expected 3 rows with nil config, got %d", len(got))
	}
}

// TestConcurrentRenameAndGetRecentData simulates the real-world race where the
// scheduler keeps inserting the old service name while a rename is in progress.
// GetRecentData must never return the stale name once config is updated.
func TestConcurrentRenameAndGetRecentData(t *testing.T) {
	cfg, cleanup := setupIntegrationDB(t, map[string]utils.Service{
		"OldName": {Name: "OldName", URL: "https://example.com"},
	})
	defer cleanup()

	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	// Seed some initial "OldName" rows.
	for range 5 {
		_ = utils.InsertData(db, utils.ServiceData{ServiceName: "OldName", ServiceHTTPResponse: "200"})
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Goroutine simulating the scheduler inserting under the old name.
	wg.Go(func() {
		for {
			select {
			case <-stop:
				return
			default:
				_ = utils.InsertData(db, utils.ServiceData{ServiceName: "OldName", ServiceHTTPResponse: "200"})
				time.Sleep(2 * time.Millisecond)
			}
		}
	})

	// Simulate rename: update config to "NewName".
	time.Sleep(20 * time.Millisecond)
	cfg.Services = map[string]utils.Service{
		"NewName": {Name: "NewName", URL: "https://example.com"},
	}
	utils.Current_Config = cfg
	_ = utils.InsertData(db, utils.ServiceData{ServiceName: "NewName", ServiceHTTPResponse: "200"})

	// After rename, GetRecentData must not return "OldName".
	for range 10 {
		got, err := utils.GetRecentData(db)
		if err != nil {
			t.Fatalf("GetRecentData: %v", err)
		}
		for _, s := range got {
			if s.ServiceName == "OldName" {
				t.Errorf("GetRecentData returned stale name 'OldName' after rename")
			}
		}
		time.Sleep(5 * time.Millisecond)
	}

	close(stop)
	wg.Wait()
}

// TestConcurrentGCAndInserts verifies DbGarbageCollect does not panic or
// deadlock when concurrent inserts are happening at the same time.
func TestConcurrentGCAndInserts(t *testing.T) {
	cfg, cleanup := setupIntegrationDB(t, map[string]utils.Service{
		"live": {Name: "live", URL: "https://example.com"},
	})
	defer cleanup()

	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Writer goroutine inserts rows for both live and stale services.
	wg.Go(func() {
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
				name := "live"
				if i%3 == 0 {
					name = "stale"
				}
				_ = utils.InsertData(db, utils.ServiceData{ServiceName: name, ServiceHTTPResponse: "200"})
				time.Sleep(1 * time.Millisecond)
			}
		}
	})

	// Run GC several times concurrently with the inserts.
	for range 5 {
		if err := utils.DbGarbageCollect(db, cfg); err != nil {
			t.Errorf("DbGarbageCollect: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	close(stop)
	wg.Wait()

	// After stopping, one final GC and verify only "live" rows remain.
	if err := utils.DbGarbageCollect(db, cfg); err != nil {
		t.Fatalf("Final GC: %v", err)
	}

	all, err := utils.GetData(db)
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	for _, s := range all {
		if s.ServiceName != "live" {
			t.Errorf("Found stale row after GC: %q", s.ServiceName)
		}
	}
}

// TestGCRemovesOrphanedRows verifies that rows for services no longer in
// config are deleted and rows for current services are preserved.
func TestGCRemovesOrphanedRows(t *testing.T) {
	cfg, cleanup := setupIntegrationDB(t, map[string]utils.Service{
		"kept": {Name: "kept", URL: "https://example.com"},
	})
	defer cleanup()

	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	for _, name := range []string{"kept", "orphan1", "orphan2"} {
		if err := utils.InsertData(db, utils.ServiceData{ServiceName: name, ServiceHTTPResponse: "200"}); err != nil {
			t.Fatalf("InsertData %s: %v", name, err)
		}
	}

	if err := utils.DbGarbageCollect(db, cfg); err != nil {
		t.Fatalf("DbGarbageCollect: %v", err)
	}

	all, err := utils.GetData(db)
	if err != nil {
		t.Fatalf("GetData: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("Expected 1 row after GC, got %d", len(all))
	}
	if all[0].ServiceName != "kept" {
		t.Errorf("Expected 'kept', got %q", all[0].ServiceName)
	}
}

// TestGetRecentDataReturnsLatestPerService verifies that only the most recent
// row per service name is returned regardless of total row count.
func TestGetRecentDataReturnsLatestPerService(t *testing.T) {
	_, cleanup := setupIntegrationDB(t, map[string]utils.Service{
		"svcA": {Name: "svcA", URL: "https://a.example.com"},
		"svcB": {Name: "svcB", URL: "https://b.example.com"},
	})
	defer cleanup()

	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer db.Close()

	responses := []struct {
		name string
		code string
	}{
		{"svcA", "200"},
		{"svcB", "200"},
		{"svcA", "500"},
		{"svcB", "503"},
		{"svcA", "200"},
	}
	for _, r := range responses {
		if err := utils.InsertData(db, utils.ServiceData{ServiceName: r.name, ServiceHTTPResponse: r.code}); err != nil {
			t.Fatalf("InsertData: %v", err)
		}
	}

	got, err := utils.GetRecentData(db)
	if err != nil {
		t.Fatalf("GetRecentData: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(got))
	}

	index := map[string]string{}
	for _, s := range got {
		index[s.ServiceName] = s.ServiceHTTPResponse
	}
	if index["svcA"] != "200" {
		t.Errorf("svcA: expected latest response '200', got %q", index["svcA"])
	}
	if index["svcB"] != "503" {
		t.Errorf("svcB: expected latest response '503', got %q", index["svcB"])
	}
}
