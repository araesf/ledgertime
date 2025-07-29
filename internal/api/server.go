package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/araesf/ledgertime/internal/config"
	"github.com/araesf/ledgertime/internal/db"
	"github.com/araesf/ledgertime/internal/ledger"
	"github.com/araesf/ledgertime/internal/models"
	"github.com/araesf/ledgertime/pkg/logger"
)

// Server represents the HTTP server
type Server struct {
	router        *mux.Router
	server        *http.Server
	db            *db.DB
	ledgerService *ledger.Service
	logger        *logger.Logger
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, database *db.DB, log *logger.Logger) *Server {
	ledgerService := ledger.NewService(database, log)
	
	s := &Server{
		router:        mux.NewRouter(),
		db:            database,
		ledgerService: ledgerService,
		logger:        log,
	}

	s.setupRoutes()
	
	s.server = &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      s.router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return s
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")
	
	// User routes
	s.router.HandleFunc("/users", s.createUser).Methods("POST")
	s.router.HandleFunc("/users/{id}", s.getUser).Methods("GET")
	
	// Card routes
	s.router.HandleFunc("/cards", s.createCard).Methods("POST")
	s.router.HandleFunc("/cards/{cardNumber}", s.getCard).Methods("GET")
	
	// Transaction routes
	s.router.HandleFunc("/transactions", s.createTransaction).Methods("POST")
	s.router.HandleFunc("/users/{id}/transactions", s.getUserTransactions).Methods("GET")
	s.router.HandleFunc("/users/{id}/summary", s.getUserSummary).Methods("GET")
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// Health check endpoint
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "ledgertime",
	}
	s.writeJSON(w, http.StatusOK, response)
}

// Create user endpoint
func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Email == "" {
		s.writeError(w, http.StatusBadRequest, "Name and email are required")
		return
	}

	now := time.Now()
	user := &models.User{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.db.CreateUser(user); err != nil {
		s.logger.Error("Failed to create user", "error", err)
		s.writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	s.writeJSON(w, http.StatusCreated, user)
}

// Get user endpoint
func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := s.db.GetUser(userID)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err, "user_id", userID)
		s.writeError(w, http.StatusNotFound, "User not found")
		return
	}

	s.writeJSON(w, http.StatusOK, user)
}

// Create card endpoint
func (s *Server) createCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID     string `json:"user_id"`
		CardNumber string `json:"card_number"`
		CardType   string `json:"card_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" || req.CardNumber == "" || req.CardType == "" {
		s.writeError(w, http.StatusBadRequest, "All fields are required")
		return
	}

	card := &models.Card{
		ID:         uuid.New().String(),
		UserID:     req.UserID,
		CardNumber: req.CardNumber,
		CardType:   req.CardType,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	if err := s.db.CreateCard(card); err != nil {
		s.logger.Error("Failed to create card", "error", err)
		s.writeError(w, http.StatusInternalServerError, "Failed to create card")
		return
	}

	s.writeJSON(w, http.StatusCreated, card)
}

// Get card endpoint
func (s *Server) getCard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cardNumber := vars["cardNumber"]

	card, err := s.db.GetCardByNumber(cardNumber)
	if err != nil {
		s.logger.Error("Failed to get card", "error", err, "card_number", cardNumber)
		s.writeError(w, http.StatusNotFound, "Card not found")
		return
	}

	s.writeJSON(w, http.StatusOK, card)
}

// Create transaction endpoint
func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	var payload models.CardPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	transaction, err := s.ledgerService.ProcessCardPayload(payload)
	if err != nil {
		s.logger.Error("Failed to process transaction", "error", err)
		s.writeError(w, http.StatusInternalServerError, "Failed to process transaction")
		return
	}

	s.writeJSON(w, http.StatusCreated, transaction)
}

// Get user transactions endpoint
func (s *Server) getUserTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	transactions, err := s.ledgerService.GetUserTransactions(userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get user transactions", "error", err, "user_id", userID)
		s.writeError(w, http.StatusInternalServerError, "Failed to get transactions")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": transactions,
		"limit":        limit,
		"offset":       offset,
	})
}

// Get user summary endpoint
func (s *Server) getUserSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	summary, err := s.ledgerService.GetUserSummary(userID)
	if err != nil {
		s.logger.Error("Failed to get user summary", "error", err, "user_id", userID)
		s.writeError(w, http.StatusInternalServerError, "Failed to get summary")
		return
	}

	s.writeJSON(w, http.StatusOK, summary)
}

// Helper methods
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}
