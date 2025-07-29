package models

import (
	"time"
)

// Transaction represents a financial transaction in the ledger
type Transaction struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	CardID       string    `json:"card_id" db:"card_id"`
	Amount       int64     `json:"amount" db:"amount"` // Amount in cents
	MerchantName string    `json:"merchant_name" db:"merchant_name"`
	Category     string    `json:"category" db:"category"`
	Description  string    `json:"description" db:"description"`
	Status       string    `json:"status" db:"status"` // pending, completed, failed
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// TransactionStatus constants
const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
)

// TransactionSummary represents aggregated transaction data
type TransactionSummary struct {
	UserID      string `json:"user_id"`
	TotalAmount int64  `json:"total_amount"`
	TotalCount  int    `json:"total_count"`
	AvgAmount   int64  `json:"avg_amount"`
	TopCategory string `json:"top_category"`
	TopMerchant string `json:"top_merchant"`
}
