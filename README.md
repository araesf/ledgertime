# Ledgertime - Modern Go Financial Ledger System

A production-ready financial ledger system built with Go, featuring card transaction processing, Kafka event streaming, and PostgreSQL persistence.

## 🏗️ Architecture

This system demonstrates modern Go backend development patterns:

- **Microservices Architecture**: Separate API server and Kafka consumer services
- **Event-Driven Design**: Kafka for asynchronous transaction processing
- **Clean Architecture**: Separation of concerns with internal packages
- **Database Integration**: PostgreSQL with proper connection pooling
- **Structured Logging**: JSON-based logging with contextual information
- **Configuration Management**: Environment-based configuration
- **Graceful Shutdown**: Proper resource cleanup and signal handling

## 📁 Project Structure

```
ledgertime/
├── cmd/                    # Application entry points
│   ├── api/               # HTTP API server
│   └── consumer/          # Kafka consumer service
├── internal/              # Private application code
│   ├── api/               # HTTP handlers and server
│   ├── config/            # Configuration management
│   ├── db/                # Database operations
│   ├── kafka/             # Kafka producer/consumer
│   ├── ledger/            # Core business logic
│   └── models/            # Data models
├── pkg/                   # Public libraries
│   └── logger/            # Structured logging
├── docker-compose.yml     # Development environment
└── go.mod                 # Go module definition
```

## 🚀 Features

### Core Functionality
- **User Management**: Create and manage users with multiple cards
- **Card Processing**: Handle multiple payment cards per user
- **Transaction Processing**: Process card payments with validation
- **Real-time Events**: Kafka-based event streaming
- **Data Persistence**: PostgreSQL with optimized queries
- **Transaction Summaries**: Aggregated spending analytics

### Technical Features
- **RESTful API**: Clean HTTP endpoints with proper status codes
- **Event Streaming**: Kafka integration for scalable processing
- **Database Migrations**: SQL schema with indexes and constraints
- **Structured Logging**: JSON logs with correlation IDs
- **Configuration**: Environment-based configuration management
- **Docker Support**: Complete containerized development environment
- **Graceful Shutdown**: Proper resource cleanup

## 🛠️ Technology Stack

- **Language**: Go 1.21
- **Database**: PostgreSQL 15
- **Message Queue**: Apache Kafka
- **HTTP Router**: Gorilla Mux
- **Logging**: Go's structured logging (slog)
- **Containerization**: Docker & Docker Compose

## 📊 API Endpoints

### Health Check
- `GET /health` - Service health status

### Users
- `POST /users` - Create a new user
- `GET /users/{id}` - Get user by ID

### Cards
- `POST /cards` - Register a new card
- `GET /cards/{cardNumber}` - Get card information

### Transactions
- `POST /transactions` - Process a card transaction
- `GET /users/{id}/transactions` - Get user's transaction history
- `GET /users/{id}/summary` - Get user's spending summary

## 🏃‍♂️ Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Git

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/araesf/ledgertime.git
   cd ledgertime
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Start infrastructure services**
   ```bash
   docker-compose up postgres kafka zookeeper -d
   ```

4. **Run the API server**
   ```bash
   go run cmd/api/main.go
   ```

5. **Run the Kafka consumer** (in another terminal)
   ```bash
   go run cmd/consumer/main.go
   ```

### Full Docker Setup

```bash
docker-compose up --build
```

This starts:
- PostgreSQL database with schema
- Kafka with Zookeeper
- API server on port 8080
- Kafka consumer service

## 📝 Example Usage

### Create a User
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'
```

### Register a Card
```bash
curl -X POST http://localhost:8080/cards \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid-here",
    "card_number": "4532-1234-5678-9012",
    "card_type": "visa"
  }'
```

### Process a Transaction
```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "card_number": "4532-1234-5678-9012",
    "amount": 2500,
    "merchant_name": "Coffee Shop",
    "category": "dining",
    "timestamp": "2024-01-15T10:30:00Z"
  }'
```

## 🔧 Configuration

Environment variables:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=ledgertime

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TRANSACTIONS_TOPIC=card-transactions
KAFKA_CONSUMER_GROUP=ledger-consumer

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## 🏢 Production Considerations

This system demonstrates enterprise-ready patterns:

- **Scalability**: Horizontal scaling with Kafka partitioning
- **Reliability**: Database transactions and error handling
- **Observability**: Structured logging and health checks
- **Security**: Input validation and SQL injection prevention
- **Performance**: Database indexes and connection pooling
- **Maintainability**: Clean architecture and separation of concerns

## 📈 Monitoring & Observability

- **Health Checks**: `/health` endpoint for load balancer integration
- **Structured Logs**: JSON format with correlation IDs
- **Database Metrics**: Connection pool monitoring
- **Kafka Metrics**: Consumer lag and throughput tracking

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/ledger
```

## 📄 License

This project is part of a portfolio demonstration and is available under the MIT License.

---

**Built with ❤️ using Go** - Demonstrating modern backend development practices for financial systems.
