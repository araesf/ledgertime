package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/araesf/ledgertime/internal/config"
	"github.com/araesf/ledgertime/internal/db"
	gql "github.com/araesf/ledgertime/internal/graphql"
	"github.com/araesf/ledgertime/internal/ledger"
	"github.com/araesf/ledgertime/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
)


// Global services
var (
	database      *db.DB
	ledgerService *ledger.Service
	gqlSchema     graphql.Schema
	logger        *logger.Logger
)

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger = logger.New(cfg.Log.Level)

	// Initialize database
	database, err = db.New(cfg.Database.DSN, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer database.Close()

	// Initialize services
	ledgerService = ledger.NewService(database, logger)
	initGraphQL()

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// GraphQL endpoint
	r.POST("/graphql", graphqlHandler)
	r.GET("/health", healthCheck)

	// Start server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Starting Ledgertime GraphQL API on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}
	logger.Info("Server shutdown complete")
}


func initGraphQL() {
	resolver := gql.NewResolver(database, ledgerService, logger)
	schema, err := resolver.BuildSchema()
	if err != nil {
		logger.Fatal("Failed to build GraphQL schema", "error", err)
	}
	gqlSchema = schema
}


func graphqlHandler(c *gin.Context) {
	var req struct {
		Query         string                 `json:"query"`
		Variables     map[string]interface{} `json:"variables"`
		OperationName string                 `json:"operationName"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result := graphql.Do(graphql.Params{
		Schema:         gqlSchema,
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName:  req.OperationName,
	})

	if result.HasErrors() {
		logger.Error("GraphQL errors", "errors", result.Errors)
	}

	c.JSON(http.StatusOK, result)
}


func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "ledgertime-graphql-api",
		"database":  "connected",
		"graphql":   "enabled",
		"timestamp": time.Now(),
	})
}
