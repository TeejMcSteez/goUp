package workers

import (
	"context"
	"database/sql"
	"goUp/utils"
	"log"
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
		log.Printf("Failed to get file information while starting hot reloading service: %v", err)
	}

	notify := utils.ConfigWriteNotify()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t, err := utils.GetFileTimestamp(path)
			if err != nil {
				log.Printf("Failed to get file information while reloading: %v", err)
			}
			if !t.Equal(initialModTime) {
				select {
				case <-notify:
					// Program wrote this change and already called Setup — just advance baseline.
					log.Println("Config updated by program, skipping redundant hot reload")
					initialModTime = t
				default:
					// External edit — do the full reload.
					log.Println("File change detected, reloading configuration")
					cfg, err := utils.LoadConfig(path)
					if err != nil {
						log.Printf("Hot reload failed loading config: %v", err)
						return
					}
					if err := utils.Setup(cfg); err != nil {
						log.Printf("Hot reload failed: %v", err)
						return
					}
					if err := utils.DbGarbageCollect(db, cfg); err != nil {
						log.Printf("Hot reload GC failed: %v", err)
					}
					initialModTime = t
				}
			}
		case <-ctx.Done():
			log.Println("Reloader Worker recieved termination signal")
			return
		}
	}
}
