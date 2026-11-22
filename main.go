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
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configPath := flag.String("config", "services.yml", "path to config file")
	serveReact := flag.Bool("ui", true, "serve frontend React file, otherwise just run API server")
	getVersion := flag.Bool("version", false, "Used to display version")
	cors := flag.Bool("cors", false, "Allow CORS aside from localhost\nMainly for Swagger API")
	origin := flag.String("origin", "*", "Sets allowed origin\nDefault is \"*\"")
	flag.Parse()
	if *getVersion {
		log.Print("Build Version: " + utils.Version)
		return
	}
	cfgStore := utils.NewConfigStore(*configPath)
	err := cfgStore.LoadFromFile()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	db, err := utils.InitDB(cfgStore.Get())
	if err != nil {
		log.Print("Error initializing database")
		panic(err)
	}
	// listening for SIGINT (Ctrl+C) and SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Starts all background workers
	hub := server.NewHub()
	go hub.Run()
	// NewScheduler subscribes itself to schedule changes. The scanner reads
	// endpoints from each cycle's config snapshot, so service changes need
	// no subscriber.
	sch := workers.NewScheduler(ctx, db, cfgStore, hub, utils.NewScanner())
	// notify websocket client configuration has been updated
	cfgStore.Subscribe(func(_, new *utils.Config) {
		hub.Broadcast([]byte(`{"message": "configuration updated"}`))
	})
	// collect removed services from database on configuration change
	cfgStore.Subscribe(func(_, new *utils.Config) {
		if err := utils.DbGarbageCollect(db, new); err != nil {
			slog.Error("failed to garbage collect database on store update", "ERROR", err)
		}
	})
	defer sch.Stop()
	if garabgeCollectErr := utils.DbGarbageCollect(db, cfgStore.Get()); garabgeCollectErr != nil {
		log.Printf("Startup GC failed: %v", garabgeCollectErr)
	}
	go workers.StartHotReloader(ctx, cfgStore)
	go workers.StartMemoryWatcher(ctx, db, cfgStore)
	go func() {
		// Next is set in NewServer to use the DefaultServeMux
		handler := server.GoupHandler{
			Cors:   *cors,
			Origin: origin,
		}
		if serverCreateErr := server.NewServer(db, sch, *serveReact, handler, hub, cfgStore).Start(); serverCreateErr != nil {
			log.Fatalf("Server failed to start: %v", serverCreateErr)
		}
		log.Println("Server started")
	}()

	// Block until a signal is received
	<-ctx.Done()
	log.Printf("Caught stop signal, starting workers graceful shutdown")
	log.Printf("Loading fresh configuration file before stoppping services")
	err = cfgStore.LoadFromFile()
	if err != nil {
		log.Fatalf("Failed to load fresh config on shutdown: %v", err)
	}
	if err := db.Close(); err != nil {
		log.Fatalf("error closing database: %s", err)
	} else {
		log.Println("Database connection closed")
	}
	cfg := cfgStore.Get()
	if (cfg.Persist_db != nil && !*cfg.Persist_db) || cfg.Persist_db == nil {
		if err := utils.CleanupDbFiles(cfg); err != nil {
			log.Printf("Error occured cleaning up the database: %v", err)
		}
	} else {
		log.Printf("Database persistence is set to true, skipping database cleanup")
	}
}
