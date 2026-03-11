package workers

import (
	"context"
	"database/sql"
	"goUp/utils"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func StartMemoryWatcher(ctx context.Context, db *sql.DB) {
	maxSize := GetMaxSize()

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
					log.Printf("Failed to clear the database: %v", err)
				}

			}
		case <-ctx.Done():
			log.Println("Memory watcher received termination signal")
			return
		}
	}
}

// Returns max size set in config in bytes or 1 gigabyte as default size
//
// Might be slow to use float64 as this is a lot of headroom
// 32 bit machines will have to spend more compute to handle these float64 operations
// However, it will be up to debate as I want it to be accurate and don't use any 32 bit machines
func GetMaxSize() float64 {
	const GB = 1 * 1e9
	if utils.Current_Config != nil {
		if utils.Current_Config.Database_Max_Size != nil {
			str := *utils.Current_Config.Database_Max_Size

			re := regexp.MustCompile(`^(\d+)([a-zA-Z]+)$`)
			matches := re.FindStringSubmatch(str)

			if len(matches) != 3 {
				log.Printf("Invalid Database_Max_Size format: %s. Defaulting to 1GB.", str)
				return GB
			}

			number, err := strconv.ParseFloat(matches[1], 64)
			if err != nil {
				log.Printf("Failed to parse number from Database_Max_Size: %v. Defaulting to 1GB.", err)
				return GB
			}
			sizeUnit := strings.ToLower(matches[2])

			switch sizeUnit {
			case "kb":
				if number < 4 {
					log.Print("Database size must be at least 4KB, returning 4KB.\n")
					return 4 * 1000
				}
				log.Printf("Set max database size to: %v%v", number, sizeUnit)
				return number * 1000
			case "mb":
				log.Printf("Set max database size to: %v%v", number, sizeUnit)
				return number * 1e6
			case "gb":
				log.Printf("Set max database size to: %v%v", number, sizeUnit)
				return number * 1e9
			default:
				log.Printf("Invalid size unit: %s. Defaulting to 1GB.", sizeUnit)
				return GB
			}
		}
	}
	log.Print("Database max size not set, defaulting to 1GB max size")
	return GB
}
