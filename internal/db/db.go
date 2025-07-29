package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/araesf/ledgertime/internal/config"
	"github.com/araesf/ledgertime/internal/models"
	"github.com/araesf/ledgertime/pkg/logger"
	_ "github.com/lib/pq"
)

// DB wraps the database connection with additional methods
type DB struct {
	*sql.DB
	logger *logger.Logger
}

// Connect creates a new database connection
func Connect(cfg config.DatabaseConfig, log *logger.Logger) (*DB, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Database connection established")

	return &DB{
		DB:     db,
		logger: log,
	}, nil
}

// User operations
func (db *DB) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (id, name, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := db.Exec(query, user.ID, user.Name, user.Email, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		db.logger.Error("Failed to create user", "error", err, "user_id", user.ID)
		return fmt.Errorf("failed to create user: %w", err)
	}

	db.logger.Info("User created", "user_id", user.ID)
	return nil
}

func (db *DB) GetUser(id string) (*models.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users WHERE id = $1`

	user := &models.User{}
	err := db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Card operations
func (db *DB) CreateCard(card *models.Card) error {
	query := `
		INSERT INTO cards (id, user_id, card_number, card_type, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.Exec(query, card.ID, card.UserID, card.CardNumber, card.CardType, card.IsActive, card.CreatedAt)
	if err != nil {
		db.logger.Error("Failed to create card", "error", err, "card_id", card.ID)
		return fmt.Errorf("failed to create card: %w", err)
	}

	db.logger.Info("Card created", "card_id", card.ID, "user_id", card.UserID)
	return nil
}

func (db *DB) GetCardByNumber(cardNumber string) (*models.Card, error) {
	query := `
		SELECT id, user_id, card_number, card_type, is_active, created_at
		FROM cards WHERE card_number = $1 AND is_active = true`

	card := &models.Card{}
	err := db.QueryRow(query, cardNumber).Scan(
		&card.ID, &card.UserID, &card.CardNumber, &card.CardType, &card.IsActive, &card.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found: %s", cardNumber)
		}
		return nil, fmt.Errorf("failed to get card: %w", err)
	}

	return card, nil
}

// Transaction operations
func (db *DB) CreateTransaction(tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (id, user_id, card_id, amount, merchant_name, category, description, status, timestamp, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := db.Exec(query,
		tx.ID, tx.UserID, tx.CardID, tx.Amount, tx.MerchantName,
		tx.Category, tx.Description, tx.Status, tx.Timestamp,
		tx.CreatedAt, tx.UpdatedAt,
	)
	if err != nil {
		db.logger.Error("Failed to create transaction", "error", err, "transaction_id", tx.ID)
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	db.logger.Info("Transaction created", "transaction_id", tx.ID, "user_id", tx.UserID, "amount", tx.Amount)
	return nil
}

func (db *DB) GetTransactionsByUser(userID string, limit, offset int) ([]*models.Transaction, error) {
	query := `
		SELECT id, user_id, card_id, amount, merchant_name, category, description, status, timestamp, created_at, updated_at
		FROM transactions 
		WHERE user_id = $1 
		ORDER BY timestamp DESC 
		LIMIT $2 OFFSET $3`

	rows, err := db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		tx := &models.Transaction{}
		err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.CardID, &tx.Amount, &tx.MerchantName,
			&tx.Category, &tx.Description, &tx.Status, &tx.Timestamp,
			&tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (db *DB) GetTransactionSummary(userID string) (*models.TransactionSummary, error) {
	query := `
		SELECT 
			user_id,
			COALESCE(SUM(amount), 0) as total_amount,
			COUNT(*) as total_count,
			COALESCE(AVG(amount), 0) as avg_amount,
			COALESCE(MODE() WITHIN GROUP (ORDER BY category), '') as top_category,
			COALESCE(MODE() WITHIN GROUP (ORDER BY merchant_name), '') as top_merchant
		FROM transactions 
		WHERE user_id = $1 AND status = 'completed'
		GROUP BY user_id`

	summary := &models.TransactionSummary{}
	err := db.QueryRow(query, userID).Scan(
		&summary.UserID, &summary.TotalAmount, &summary.TotalCount,
		&summary.AvgAmount, &summary.TopCategory, &summary.TopMerchant,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return &models.TransactionSummary{UserID: userID}, nil
		}
		return nil, fmt.Errorf("failed to get transaction summary: %w", err)
	}

	return summary, nil
}
