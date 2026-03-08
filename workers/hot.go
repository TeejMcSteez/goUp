package workers

import (
	"context"
	"goUp/utils"
	"log"
	"time"
)

func StartHotReloader(path string, ctx context.Context) {
	initialModTime, err := utils.GetFileTimestamp(path)
	if err != nil {
		log.Printf("Failed to get file information while starting hot reloading service: %v", err)
	}
	// Checks for file mods every 5 seconds
	ticker := time.NewTicker(5 * time.Second)

	defer ticker.Stop()
	// Runs forever on a ticker and closes with the program
	// Similar to the scheduler
	for {
		select {
		case <-ticker.C:
			t, err := utils.GetFileTimestamp(path)
			if err != nil {
				log.Printf("Failed to get file information while reloading: %v", err)
			}
			if !t.Equal(initialModTime) {
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
				initialModTime = t
			}
		case <-ctx.Done():
			log.Println("Reloader Worker recieved termination signal")
			return
		}
	}

}
