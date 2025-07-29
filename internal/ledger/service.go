package ledger

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/araesf/ledgertime/internal/db"
	"github.com/araesf/ledgertime/internal/models"
	"github.com/araesf/ledgertime/pkg/logger"
)

// Service handles ledger operations
type Service struct {
	db     *db.DB
	logger *logger.Logger
}

// NewService creates a new ledger service
func NewService(database *db.DB, log *logger.Logger) *Service {
	return &Service{
		db:     database,
		logger: log,
	}
}

// ProcessCardPayload converts a card payment into a transaction
func (s *Service) ProcessCardPayload(payload models.CardPayload) (*models.Transaction, error) {
	s.logger.Info("Processing card payload", "card_number", payload.CardNumber, "amount", payload.Amount)

	// Find the card and user
	card, err := s.db.GetCardByNumber(payload.CardNumber)
	if err != nil {
		s.logger.Error("Card not found", "error", err, "card_number", payload.CardNumber)
		return nil, fmt.Errorf("card not found: %w", err)
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, payload.Timestamp)
	if err != nil {
		s.logger.Error("Invalid timestamp", "error", err, "timestamp", payload.Timestamp)
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	// Create transaction
	now := time.Now()
	transaction := &models.Transaction{
		ID:           uuid.New().String(),
		UserID:       card.UserID,
		CardID:       card.ID,
		Amount:       payload.Amount,
		MerchantName: payload.MerchantName,
		Category:     payload.Category,
		Description:  fmt.Sprintf("Card payment at %s", payload.MerchantName),
		Status:       models.TransactionStatusPending,
		Timestamp:    timestamp,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Validate transaction
	if err := s.ValidateTransaction(transaction); err != nil {
		s.logger.Error("Transaction validation failed", "error", err, "transaction_id", transaction.ID)
		return nil, fmt.Errorf("transaction validation failed: %w", err)
	}

	// Save to database
	if err := s.db.CreateTransaction(transaction); err != nil {
		s.logger.Error("Failed to save transaction", "error", err, "transaction_id", transaction.ID)
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	// Process the transaction (simulate processing)
	if err := s.processTransaction(transaction); err != nil {
		s.logger.Error("Transaction processing failed", "error", err, "transaction_id", transaction.ID)
		transaction.Status = models.TransactionStatusFailed
	} else {
		transaction.Status = models.TransactionStatusCompleted
	}

	s.logger.Info("Transaction processed", "transaction_id", transaction.ID, "status", transaction.Status)
	return transaction, nil
}

// ValidateTransaction ensures the transaction has valid data
func (s *Service) ValidateTransaction(tx *models.Transaction) error {
	if tx.Amount <= 0 {
		return fmt.Errorf("amount must be positive, got: %d", tx.Amount)
	}

	if tx.UserID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	if tx.CardID == "" {
		return fmt.Errorf("card ID cannot be empty")
	}

	if tx.MerchantName == "" {
		return fmt.Errorf("merchant name cannot be empty")
	}

	if tx.Category == "" {
		return fmt.Errorf("category cannot be empty")
	}

	return nil
}

// GetUserTransactions retrieves transactions for a user
func (s *Service) GetUserTransactions(userID string, limit, offset int) ([]*models.Transaction, error) {
	s.logger.Info("Getting user transactions", "user_id", userID, "limit", limit, "offset", offset)

	transactions, err := s.db.GetTransactionsByUser(userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get user transactions", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get user transactions: %w", err)
	}

	return transactions, nil
}

// GetUserSummary retrieves transaction summary for a user
func (s *Service) GetUserSummary(userID string) (*models.TransactionSummary, error) {
	s.logger.Info("Getting user summary", "user_id", userID)

	summary, err := s.db.GetTransactionSummary(userID)
	if err != nil {
		s.logger.Error("Failed to get user summary", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get user summary: %w", err)
	}

	return summary, nil
}

// processTransaction simulates transaction processing logic
func (s *Service) processTransaction(tx *models.Transaction) error {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Simulate fraud detection
	if tx.Amount > 100000 { // $1000 in cents
		return fmt.Errorf("transaction amount exceeds limit")
	}

	// Simulate random failures (5% failure rate)
	if time.Now().UnixNano()%20 == 0 {
		return fmt.Errorf("processing failed due to network error")
	}

	return nil
}
