// Package main is the entry point for the goUp service uptime monitor.
//
// @title goUp API
// @version 1.0
// @description REST API for monitoring service uptime, response times, and managing configuration.
// @host localhost:8101
// @BasePath /
// @schemes http
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
	"strings"
	"syscall"
)

func main() {
	configPath := flag.String("config", "services.yml", "path to config file")
	serveReact := flag.String("ui", "y", "serve frontend React file, otherwise just run API server")
	flag.Parse()
	serve := string([]byte(*serveReact)[0])
	if strings.ToLower(*serveReact) != "y" && strings.ToLower(serve) != "n" {
		log.Fatalf("UI flag must be 'y' or 'n'\nwas: %v", *serveReact)
	}

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
	// listening for SIGINT (Ctrl+C) and SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Starts all background workers
	sch := workers.NewScheduler(db, cfg)
	defer sch.Stop()
	go workers.StartHotReloader(*configPath, ctx)
	go workers.StartMemoryWatcher(ctx, db)
	go func() {
		log.Println("Starting server on port 8080")
		if err := server.NewServer(db, sch, &serve).Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
		log.Println("Server started")
	}()

	// Block until a signal is received
	<-ctx.Done()
	log.Printf("Caught stop signal, starting workers graceful shutdown")
	log.Printf("Loading fresh configuration file before stoppping services")
	cfg, err = utils.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load fresh config on shutdown: %v", err)
	}
	if err := db.Close(); err != nil {
		log.Fatalf("error closing database: %s", err)
	}
	log.Println("Database connection closed")
	if (cfg.Persist_db != nil && !*cfg.Persist_db) || cfg.Persist_db == nil {
		if err := utils.CleanupDbFiles(); err != nil {
			log.Printf("Error occured cleaning up the database: %v", err)
		}
	} else {
		log.Printf("Database persistence is set to true, skipping database cleanup")
	}
}
