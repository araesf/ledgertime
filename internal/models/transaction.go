package models

import (
	"time"
)

// Transaction represents a double-entry ledger transaction
type Transaction struct {
	ID          string    `json:"id" db:"id"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	Description string    `json:"description" db:"description"`
	Reference   string    `json:"reference,omitempty" db:"reference"` // Card number, invoice ID, etc.
	Entries     []Entry   `json:"entries"`
}

// Entry represents a single debit or credit entry in a transaction
type Entry struct {
	ID            string `json:"id" db:"id"`
	TransactionID string `json:"transaction_id" db:"transaction_id"`
	AccountID     string `json:"account_id" db:"account_id"`
	Amount        int64  `json:"amount" db:"amount"` // Amount in cents to avoid floating point issues
	Type          string `json:"type" db:"type"`     // "debit" or "credit"
	Description   string `json:"description,omitempty" db:"description"`
}

// CardPayload represents the raw data coming from card transactions
// This will be transformed into a Transaction
type CardPayload struct {
	CardNumber   string `json:"card_number"`
	Name         string `json:"name"`
	PurchaseDate string `json:"purchase_date"`
	Price        int    `json:"price"` // Price in cents
	ObjectBought string `json:"object"`
}
