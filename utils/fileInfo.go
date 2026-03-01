package utils

import (
	"log"
	"os"
	"time"
)

func GetDatabaseSize() (int64, error) {
	if Current_Config.Database_Location != nil {
		file, err := os.Open(*Current_Config.Database_Location)
		if err != nil {
			return 0, err
		}
		defer file.Close()
		fileStats, err := file.Stat()
		if err != nil {
			return 0, err
		}
		return fileStats.Size(), nil

	}
	return 0, &NoConfigError{"error getting size", "configuration not found in memory"}
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
