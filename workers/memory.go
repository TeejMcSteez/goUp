package workers

import (
	"context"
	"database/sql"
	"fmt"
	"goUp/utils"
	"log"
	"time"
)

func StartMemoryWatcher(ctx context.Context, db *sql.DB) {
	maxSize, err := GetMaxSize()
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

// Returns max size set in config in bytes or 1 gigabyte as default size
func GetMaxSize() (float64, error) {
	if utils.Current_Config != nil {
		if utils.Current_Config.Database_Max_Size != nil {
			str := *utils.Current_Config.Database_Max_Size
			size, err := utils.GetSizeFromString(str)
			if err != nil {
				return 0, fmt.Errorf("incorrect format, must be <decimal><a-zA-Z>")
			}
			return size, nil
		}
	}
	log.Print("Database max size not set, defaulting to 1GB max size")
	return utils.GB, nil
}
