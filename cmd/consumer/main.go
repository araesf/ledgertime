package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/araesf/ledgertime/internal/config"
	"github.com/araesf/ledgertime/internal/db"
	"github.com/araesf/ledgertime/internal/kafka"
	"github.com/araesf/ledgertime/internal/ledger"
	"github.com/araesf/ledgertime/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.NewLogger()
	log.Info("Starting Ledgertime Kafka consumer")

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

	// Initialize ledger service
	ledgerService := ledger.NewService(database, log)

	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.Kafka, ledgerService, log)
	if err != nil {
		log.Fatal("Failed to create Kafka consumer", "error", err)
	}

	// Start consuming in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Error("Consumer error", "error", err)
			cancel()
		}
	}()

	log.Info("Consumer started successfully")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down consumer...")
	cancel()

	if err := consumer.Close(); err != nil {
		log.Error("Error closing consumer", "error", err)
	}

	log.Info("Consumer exited properly")
}
