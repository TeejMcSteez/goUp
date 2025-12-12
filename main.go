package main

import (
	"fmt"
	"goUp/scheduler"
	"goUp/server"
	"goUp/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Create blank data store
	db := utils.InitDB()
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("error closing db: %s", err)
		}
		fmt.Println("Database connection closed")
	}()

	// Get current service data before full launch
	svcData := utils.GetServiceData()
	// Adds recently fetched data to the database
	for data := range svcData {
		utils.InsertData(db, svcData[data])
	}

	// Starts scheduler
	sch := scheduler.NewScheduler(db, 30, "seconds")
	defer sch.Stop()

	// Channel to listen for OS signals
	// listening for SIGINT (Ctrl+C) and SIGTERM
	// A buffered channel is used to avoid missing signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Starts http server in a go routine
	go func() {
		fmt.Println("starting server on port 8080")
		if err := server.Start(db, sch); err != nil {
			log.Fatalf("server failed to start: %v", err)
		}
		fmt.Println("Server started")
	}()

	// Block until a signal is received
	sig := <-shutdown
	log.Printf("caught signal: %v, starting graceful shutdown", sig)
}
