package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/araesf/ledgertime/internal/api"
	"github.com/araesf/ledgertime/internal/config"
	"github.com/araesf/ledgertime/internal/db"
	"github.com/araesf/ledgertime/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.NewLogger()
	log.Info("Starting Ledgertime API server")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Connect to database
	database, err := db.Connect(cfg.Database, log)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer database.Close()

	// Initialize API server
	server := api.NewServer(cfg, database, log)

	// Start server in a goroutine
	go func() {
		log.Info("Starting HTTP server", "port", cfg.Server.Port)
		if err := server.Start(); err != nil {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}

	log.Info("Server exited properly")
}
