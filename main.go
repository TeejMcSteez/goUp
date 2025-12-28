package main

import (
	"goUp/workers"
	"goUp/server"
	"goUp/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
	"context"
)

func main() {
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
	}()

	svcData, err := utils.GetServiceData()
	if err != nil {
		log.Fatalf("Error on initial fetch of service data, panicking out: %v\n", err)
	}
	for data := range svcData.AllServices {
		utils.InsertData(db, svcData.AllServices[data])
	}

	sch := scheduler.NewScheduler(db, 30, "seconds")
	defer sch.Stop()

	// listening for SIGINT (Ctrl+C) and SIGTERM
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	
	ctx, cancel := context.WithCancel(context.Background())

	go scheduler.StartHotReloader("services.yml", ctx)

	go func() {
		log.Println("starting server on port 8080")
		if err := server.NewServer(db, sch).Start(); err != nil {
			log.Fatalf("server failed to start: %v", err)
		}
		log.Println("Server started")
	}()

	// Block until a signal is received
	sig := <-shutdown
	log.Printf("caught signal: %v, starting graceful shutdown", sig)
	cancel()
}
