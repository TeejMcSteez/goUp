package main

import (
	"context"
	"goUp/server"
	"goUp/utils"
	"goUp/workers"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	svcData, err := utils.GetServiceData()
	if err != nil {
		log.Fatalf("Error on initial fetch of service data, panicking out: %v\n", err)
	}

	db, err := utils.InitDB()
	if err != nil {
		log.Print("Error initializing database")
		panic(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("error closing db: %s", err)
		}
		log.Println("Database connection closed")
		if err := utils.CleanupDb(); err != nil {
			log.Printf("Error Occured cleaning up the database: %v", err)
		}
	}()

	for data := range svcData.AllServices {
		if err := utils.InsertData(db, svcData.AllServices[data]); err != nil {
			log.Printf("Failed inserting data on initial fetch: %v\n", err)
		}
	}

	sch := scheduler.NewScheduler(db, 30, "seconds")
	defer sch.Stop()

	// listening for SIGINT (Ctrl+C) and SIGTERM
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go scheduler.StartHotReloader("services.yml", ctx)

	go func() {
		log.Println("Starting server on port 8080")
		if err := server.NewServer(db, sch).Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
		log.Println("Server started")
	}()

	// Block until a signal is received
	sig := <-shutdown
	log.Printf("Caught signal: %v, starting graceful shutdown", sig)
	cancel()
}
