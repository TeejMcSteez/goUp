package utils

import (
	"log"
	"os"
	"time"
)

func GetDatabaseSize() (int64, error) {
	if Current_Config.Database_Location == nil {
		return 0, &NoConfigError{"error getting size", "configuration not found in memory"}
	}
	loc := *Current_Config.Database_Location
	total := int64(0)
	for _, path := range []string{loc, loc + "-wal", loc + "-shm"} {
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return 0, err
		}
		total += info.Size()
	}
	return total, nil
}

func GetFileTimestamp(path string) (time.Time, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to open file while getting file timestamp: %v", err)
		return time.Now(), err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Failed to get file information while getting file timestamp: %v", err)
		return time.Now(), err
	}
	return fileInfo.ModTime(), nil
}
