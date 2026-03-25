package main

import (
	"context"
	"flag"
	"goUp/server"
	"goUp/utils"
	"goUp/workers"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configPath := flag.String("config", "services.yml", "path to config file")
	flag.Parse()

	cfg, err := utils.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if err := utils.Setup(cfg); err != nil {
		log.Fatalf("Failed to setup services: %v", err)
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
		if (cfg.Persist_db != nil && !*cfg.Persist_db) || cfg.Persist_db == nil {
			if err := utils.CleanupDbFiles(); err != nil {
				log.Printf("Error Occured cleaning up the database: %v", err)
			}
		}
	}()

	// listening for SIGINT (Ctrl+C) and SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Starts all background workers
	sch := workers.NewScheduler(db, 30, "seconds")
	defer sch.Stop()
	go workers.StartHotReloader(*configPath, ctx)
	go workers.StartMemoryWatcher(ctx, db)
	go func() {
		log.Println("Starting server on port 8080")
		if err := server.NewServer(db, sch).Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
		log.Println("Server started")
	}()

	// Block until a signal is received
	<-ctx.Done()
	log.Printf("Caught stop signal, starting workers graceful shutdown")
}
