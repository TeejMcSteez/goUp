package workers

import (
	"context"
	"goUp/utils"
	"log/slog"
	"time"
)

// StartHotReloader polls the store's config file and republishes it through
// the store when it is edited externally. The store's subscribers (reschedule,
// websocket notice, database GC) run as part of Reload.
func StartHotReloader(ctx context.Context, store *utils.ConfigStore) {
	startHotReloader(ctx, store, 5*time.Second)
}

// StartHotReloaderWithInterval is StartHotReloader with a configurable poll
// interval. Intended for tests that need the reloader to fire quickly.
func StartHotReloaderWithInterval(ctx context.Context, store *utils.ConfigStore, interval time.Duration) {
	startHotReloader(ctx, store, interval)
}

func startHotReloader(ctx context.Context, store *utils.ConfigStore, interval time.Duration) {
	path := store.Path()
	lastModTime, err := utils.GetFileTimestamp(path)
	if err != nil {
		slog.Error("Failed to get file information while starting hot reloading service", "error", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t, err := utils.GetFileTimestamp(path)
			if err != nil {
				slog.Error("Failed to get file information while reloading", "error", err)
				continue
			}
			if t.Equal(lastModTime) {
				continue
			}
			lastModTime = t
			// Reload skips content the store wrote itself (hash match), so
			// program writes don't trigger a redundant reload.
			changed, err := store.Reload()
			if err != nil {
				// keep serving the last good config; a later save may fix it
				slog.Error("Hot reload failed loading config", "error", err)
				continue
			}
			if !changed {
				slog.Info("Config updated by program, skipping redundant hot reload")
				continue
			}
			slog.Info("File change detected, configuration reloaded")
		case <-ctx.Done():
			slog.Info("Reloader Worker recieved termination signal")
			return
		}
	}
}
