package workers

import (
	"context"
	"database/sql"
	"goUp/utils"
	"log/slog"
	"time"
)

func StartHotReloader(path string, ctx context.Context, db *sql.DB) {
	startHotReloader(path, ctx, db, 5*time.Second)
}

// StartHotReloaderWithInterval is StartHotReloader with a configurable poll
// interval. Intended for tests that need the reloader to fire quickly.
func StartHotReloaderWithInterval(path string, ctx context.Context, db *sql.DB, interval time.Duration) {
	startHotReloader(path, ctx, db, interval)
}

func startHotReloader(path string, ctx context.Context, db *sql.DB, interval time.Duration) {
	initialModTime, err := utils.GetFileTimestamp(path)
	if err != nil {
		slog.Error("Failed to get file information while starting hot reloading service", "error", err)
	}

	notify := utils.ConfigWriteNotify()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t, err := utils.GetFileTimestamp(path)
			if err != nil {
				slog.Error("Failed to get file information while reloading", "error", err)
			}
			if !t.Equal(initialModTime) {
				select {
				case <-notify:
					// Program wrote this change and already called Setup — just advance baseline.
					slog.Info("Config updated by program, skipping redundant hot reload")
					initialModTime = t
				default:
					// External edit — do the full reload.
					slog.Info("File change detected, reloading configuration")
					cfg, err := utils.LoadConfig(path)
					if err != nil {
						slog.Error("Hot reload failed loading config", "error", err)
						return
					}
					if err := utils.Setup(cfg); err != nil {
						slog.Error("Hot reload failed", "error", err)
						return
					}
					if db != nil {
						if err := utils.DbGarbageCollect(db, cfg); err != nil {
							slog.Error("Hot reload GC failed", "error", err)
						}
					} else {
						slog.Error("Hot reload GC failed: database is nil")
					}
					initialModTime = t
				}
			}
		case <-ctx.Done():
			slog.Info("Reloader Worker recieved termination signal")
			return
		}
	}
}
