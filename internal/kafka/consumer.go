package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/araesf/ledgertime/internal/config"
	"github.com/araesf/ledgertime/internal/ledger"
	"github.com/araesf/ledgertime/internal/models"
	"github.com/araesf/ledgertime/pkg/logger"
)

// Consumer handles Kafka message consumption
type Consumer struct {
	reader        *kafka.Reader
	ledgerService *ledger.Service
	logger        *logger.Logger
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg config.KafkaConfig, ledgerService *ledger.Service, log *logger.Logger) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		Topic:       cfg.TransactionsTopic,
		GroupID:     cfg.ConsumerGroup,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
		StartOffset: kafka.LastOffset,
	})

	return &Consumer{
		reader:        reader,
		ledgerService: ledgerService,
		logger:        log,
	}, nil
}

// Start begins consuming messages from Kafka
func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("Starting Kafka consumer")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Consumer context cancelled, shutting down")
			return nil
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.Error("Failed to read message", "error", err)
				continue
			}

			if err := c.processMessage(message); err != nil {
				c.logger.Error("Failed to process message", "error", err, "offset", message.Offset)
			}
		}
	}
}

// processMessage processes a single Kafka message
func (c *Consumer) processMessage(message kafka.Message) error {
	c.logger.Info("Processing message", "offset", message.Offset, "partition", message.Partition)

	// Parse the card payload
	var payload models.CardPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Process the transaction
	transaction, err := c.ledgerService.ProcessCardPayload(payload)
	if err != nil {
		return fmt.Errorf("failed to process card payload: %w", err)
	}

	c.logger.Info("Transaction processed successfully", 
		"transaction_id", transaction.ID, 
		"user_id", transaction.UserID,
		"amount", transaction.Amount,
	)

	return nil
}

// Close closes the Kafka consumer
func (c *Consumer) Close() error {
	c.logger.Info("Closing Kafka consumer")
	return c.reader.Close()
}

// Producer handles Kafka message production
type Producer struct {
	writer *kafka.Writer
	logger *logger.Logger
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg config.KafkaConfig, log *logger.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.TransactionsTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    cfg.BatchSize,
		BatchTimeout: 10 * time.Millisecond,
	}

	return &Producer{
		writer: writer,
		logger: log,
	}
}

// PublishCardPayload publishes a card payload to Kafka
func (p *Producer) PublishCardPayload(ctx context.Context, payload models.CardPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(payload.CardNumber),
		Value: data,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		p.logger.Error("Failed to publish message", "error", err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Info("Message published", "card_number", payload.CardNumber, "amount", payload.Amount)
	return nil
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	p.logger.Info("Closing Kafka producer")
	return p.writer.Close()
}
