package scheduler

import (
	"goUp/utils"
	"log"
	"os"
	"time"
	"context"
)

func StartHotReloader(path string, ctx context.Context) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to open file while starting hot reloading service: %v", err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Failed to get file information while starting hot reloading service: %v", err)
	}
	initialModTime := fileInfo.ModTime()
	// Checks for file mods every 5 seconds
	ticker := time.NewTicker(5 * time.Second)

	defer ticker.Stop()
	// Runs forever on a ticker and closes with the program
	// Similar to the scheduler
	for {
		select {
		case <-ticker.C:
			file, err := os.Open(path)
			if err != nil {
				log.Printf("Failed to open file while reloading: %v", err)
			}
			fileInfo, err := file.Stat()
			if err != nil {
				log.Printf("Failed to get file information while reloading: %v", err)
			}
			if !fileInfo.ModTime().Equal(initialModTime) {
				log.Println("Reloading configuration")
				if err := utils.Setup(); err != nil {
					log.Printf("Hot reload failed: %v", err)
					return
				}
				initialModTime = fileInfo.ModTime()
			}
		case <-ctx.Done():
			log.Println("Reloader Worker recieved termination signal")
			return
		}
	}

}