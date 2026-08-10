package workers

import (
	"context"
	"database/sql"
	"goUp/utils"
	"log/slog"
	"os"
	"time"
)

func StartMemoryWatcher(ctx context.Context, db *sql.DB) {
	maxSize, err := utils.GetMaxSize()
	if err != nil {
		slog.Error("Failed to get database max size", "error", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(1 * time.Minute)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t, err := utils.GetDatabaseSize()
			if err != nil {
				slog.Error("Failed to get file size while watching database memory", "error", err)
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
