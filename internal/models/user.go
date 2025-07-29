package models

import "time"

// User represents a user in the system
type User struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Card represents a user's payment card
type Card struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	CardNumber string    `json:"card_number" db:"card_number"`
	CardType   string    `json:"card_type" db:"card_type"` // visa, mastercard, etc
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// CardPayload represents incoming card transaction data
type CardPayload struct {
	CardNumber   string `json:"card_number"`
	Amount       int64  `json:"amount"` // Amount in cents
	MerchantName string `json:"merchant_name"`
	Category     string `json:"category"`  // groceries, gas, etc
	Timestamp    string `json:"timestamp"` // ISO 8601 format
}
