package workers

import (
	"context"
	"database/sql"
	"goUp/utils"
	"log"
	"time"
)

func StartMemoryWatcher(ctx context.Context, db *sql.DB) {
	maxSize, err := utils.GetMaxSize()
	if err != nil {
		log.Fatalf("Failed to get database max size: %v", err)
	}

	ticker := time.NewTicker(1 * time.Minute)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t, err := utils.GetDatabaseSize()
			if err != nil {
				log.Printf("Failed to get file size while watching database memory: %v", err)
			}
			if t > int64(maxSize) {
				log.Printf("File size is %v, clearing database memory", t)
				if err := utils.ClearDatabase(db); err != nil {
					log.Printf("error occured clearing database: %v", err)
				}

			}
		case <-ctx.Done():
			log.Println("Memory watcher received termination signal")
			return
		}
	}
}
