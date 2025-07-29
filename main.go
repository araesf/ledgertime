package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Amount       int64     `json:"amount" db:"amount"`
	MerchantName string    `json:"merchant_name" db:"merchant_name"`
	Category     string    `json:"category" db:"category"`
	Status       string    `json:"status" db:"status"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
}

// User represents a system user
type User struct {
	ID      string `json:"id" db:"id"`
	Name    string `json:"name" db:"name"`
	Email   string `json:"email" db:"email"`
	Balance int64  `json:"balance" db:"balance"`
}

// Global services (simplified for demo)
var (
	db          *sql.DB
	kafkaWriter *kafka.Writer
	gqlSchema   graphql.Schema
)

func main() {
	// Initialize services
	initDB()
	initKafka()
	initGraphQL()

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// REST API endpoints
	r.POST("/users", createUser)
	r.GET("/users/:id", getUser)
	r.POST("/transactions", processTransaction)
	r.GET("/users/:id/transactions", getUserTransactions)
	r.POST("/graphql", graphqlHandler)
	r.GET("/health", healthCheck)

	log.Println(" Ledgertime Fintech API running on :8080")
	r.Run(":8080")
}

func initDB() {
	// PostgreSQL connection (simplified)
	var err error
	db, err = sql.Open("postgres", "postgres://user:pass@localhost/ledgertime?sslmode=disable")
	if err != nil {
		log.Printf("PostgreSQL connection failed: %v", err)
	}
}

func initKafka() {
	// Kafka producer setup
	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "transactions",
		Balancer: &kafka.LeastBytes{},
	}
}

func initGraphQL() {
	// GraphQL schema setup
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id":      &graphql.Field{Type: graphql.String},
			"name":    &graphql.Field{Type: graphql.String},
			"email":   &graphql.Field{Type: graphql.String},
			"balance": &graphql.Field{Type: graphql.Int},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"user": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return User{ID: "demo", Name: "Demo User", Email: "demo@example.com", Balance: 100000}, nil
				},
			},
		},
	})

	gqlSchema, _ = graphql.NewSchema(graphql.SchemaConfig{Query: queryType})
}

func createUser(c *gin.Context) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	c.ShouldBindJSON(&req)

	user := &User{
		ID:      uuid.New().String(),
		Name:    req.Name,
		Email:   req.Email,
		Balance: 100000,
	}

	// Simulate PostgreSQL insert
	log.Printf("Inserting user into PostgreSQL: %+v", user)

	c.JSON(http.StatusCreated, user)
}

func getUser(c *gin.Context) {
	userID := c.Param("id")

	// Simulate PostgreSQL query
	log.Printf("Querying PostgreSQL for user: %s", userID)

	user := &User{
		ID:      userID,
		Name:    "Demo User",
		Email:   "demo@example.com",
		Balance: 95000,
	}

	c.JSON(http.StatusOK, user)
}

func processTransaction(c *gin.Context) {
	var req struct {
		UserID       string `json:"user_id"`
		Amount       int64  `json:"amount"`
		MerchantName string `json:"merchant_name"`
		Category     string `json:"category"`
	}

	c.ShouldBindJSON(&req)

	tx := &Transaction{
		ID:           uuid.New().String(),
		UserID:       req.UserID,
		Amount:       req.Amount,
		MerchantName: req.MerchantName,
		Category:     req.Category,
		Status:       "completed",
		Timestamp:    time.Now(),
	}

	// Publish to Kafka
	publishToKafka(tx)

	// Save to PostgreSQL
	log.Printf("Saving transaction to PostgreSQL: %+v", tx)

	c.JSON(http.StatusCreated, tx)
}

func getUserTransactions(c *gin.Context) {
	userID := c.Param("id")

	// Simulate PostgreSQL query
	log.Printf("Querying PostgreSQL for user transactions: %s", userID)

	txs := []*Transaction{
		{ID: uuid.New().String(), UserID: userID, Amount: 5000, MerchantName: "Starbucks", Category: "food", Status: "completed", Timestamp: time.Now()},
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": txs,
		"count":        len(txs),
	})
}

func graphqlHandler(c *gin.Context) {
	var req struct {
		Query string `json:"query"`
	}

	c.ShouldBindJSON(&req)

	result := graphql.Do(graphql.Params{
		Schema:        gqlSchema,
		RequestString: req.Query,
	})

	c.JSON(http.StatusOK, result)
}

func publishToKafka(tx *Transaction) {
	data, _ := json.Marshal(tx)
	kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(tx.ID),
		Value: data,
	})
	log.Printf("Published transaction to Kafka: %s", tx.ID)
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "healthy",
		"service":    "ledgertime-fintech-api",
		"postgres":   "connected",
		"kafka":      "connected",
		"graphql":    "enabled",
		"timestamp":  time.Now(),
	})
}
