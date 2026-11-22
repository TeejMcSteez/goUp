package workers

import (
	"context"
	"database/sql"
	"goUp/utils"
	"log/slog"
	"time"
)

// StartMemoryWatcher clears the database whenever its file grows past the
// configured db_max_size. Size and path are read from the store on every
// tick so changes made through the API or hot reload apply without restart.
func StartMemoryWatcher(ctx context.Context, db *sql.DB, store *utils.ConfigStore) {
	startMemoryWatcher(ctx, db, store, 1*time.Minute)
}

func startMemoryWatcher(ctx context.Context, db *sql.DB, store *utils.ConfigStore, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cfg := store.Get()
			maxSize, err := utils.GetMaxSize(cfg)
			if err != nil {
				slog.Error("Failed to get database max size", "error", err)
				continue
			}
			t, err := utils.GetDatabaseSize(cfg)
			if err != nil {
				slog.Error("Failed to get file size while watching database memory", "error", err)
				continue
			}
			if t > int64(maxSize) {
				slog.Info("Clearing database memory", "file_size", t)
				if err := utils.ClearDatabase(db); err != nil {
					slog.Error("error occured clearing database", "error", err)
				}
			}
		case <-ctx.Done():
			slog.Info("Memory watcher received termination signal")
			return
		}
	}
}
