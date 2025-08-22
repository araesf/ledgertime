package graphql

import (
	"github.com/araesf/ledgertime/internal/db"
	"github.com/araesf/ledgertime/internal/ledger"
	"github.com/araesf/ledgertime/internal/models"
	"github.com/araesf/ledgertime/pkg/logger"
	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"time"
)

type Resolver struct {
	db            *db.DB
	ledgerService *ledger.Service
	logger        *logger.Logger
}

func NewResolver(database *db.DB, ledgerService *ledger.Service, log *logger.Logger) *Resolver {
	return &Resolver{
		db:            database,
		ledgerService: ledgerService,
		logger:        log,
	}
}

func (r *Resolver) BuildSchema() (graphql.Schema, error) {
	// User Type
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id":         &graphql.Field{Type: graphql.String},
			"name":       &graphql.Field{Type: graphql.String},
			"email":      &graphql.Field{Type: graphql.String},
			"created_at": &graphql.Field{Type: graphql.DateTime},
			"updated_at": &graphql.Field{Type: graphql.DateTime},
		},
	})

	// Card Type
	cardType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Card",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.String},
			"user_id":     &graphql.Field{Type: graphql.String},
			"card_number": &graphql.Field{Type: graphql.String},
			"card_type":   &graphql.Field{Type: graphql.String},
			"is_active":   &graphql.Field{Type: graphql.Boolean},
			"created_at":  &graphql.Field{Type: graphql.DateTime},
		},
	})

	// Transaction Type
	transactionType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Transaction",
		Fields: graphql.Fields{
			"id":               &graphql.Field{Type: graphql.String},
			"user_id":          &graphql.Field{Type: graphql.String},
			"card_id":          &graphql.Field{Type: graphql.String},
			"amount":           &graphql.Field{Type: graphql.Int},
			"merchant_name":    &graphql.Field{Type: graphql.String},
			"merchant_city":    &graphql.Field{Type: graphql.String},
			"merchant_country": &graphql.Field{Type: graphql.String},
			"mcc":              &graphql.Field{Type: graphql.String},
			"auth_code":        &graphql.Field{Type: graphql.String},
			"status":           &graphql.Field{Type: graphql.String},
			"timestamp":        &graphql.Field{Type: graphql.DateTime},
		},
	})

	// Summary Type
	summaryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "UserSummary",
		Fields: graphql.Fields{
			"user_id":            &graphql.Field{Type: graphql.String},
			"total_transactions": &graphql.Field{Type: graphql.Int},
			"total_amount":       &graphql.Field{Type: graphql.Int},
			"category_breakdown": &graphql.Field{Type: graphql.NewList(graphql.String)},
		},
	})

	// Query Type
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"user": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: r.getUserResolver,
			},
			"card": &graphql.Field{
				Type: cardType,
				Args: graphql.FieldConfigArgument{
					"card_number": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: r.getCardResolver,
			},
			"transactions": &graphql.Field{
				Type: graphql.NewList(transactionType),
				Args: graphql.FieldConfigArgument{
					"user_id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"limit": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 10,
					},
					"offset": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 0,
					},
				},
				Resolve: r.getTransactionsResolver,
			},
			"userSummary": &graphql.Field{
				Type: summaryType,
				Args: graphql.FieldConfigArgument{
					"user_id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: r.getUserSummaryResolver,
			},
		},
	})

	// Mutation Type
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createUser": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: r.createUserResolver,
			},
			"createCard": &graphql.Field{
				Type: cardType,
				Args: graphql.FieldConfigArgument{
					"user_id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"card_number": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"card_type": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: r.createCardResolver,
			},
			"processTransaction": &graphql.Field{
				Type: transactionType,
				Args: graphql.FieldConfigArgument{
					"card_number": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"amount": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"merchant_name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"merchant_city": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"merchant_country": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"mcc": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: r.processTransactionResolver,
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
}

// Query Resolvers
func (r *Resolver) getUserResolver(p graphql.ResolveParams) (interface{}, error) {
	id := p.Args["id"].(string)
	return r.db.GetUser(id)
}

func (r *Resolver) getCardResolver(p graphql.ResolveParams) (interface{}, error) {
	cardNumber := p.Args["card_number"].(string)
	return r.db.GetCardByNumber(cardNumber)
}

func (r *Resolver) getTransactionsResolver(p graphql.ResolveParams) (interface{}, error) {
	userID := p.Args["user_id"].(string)
	limit := p.Args["limit"].(int)
	offset := p.Args["offset"].(int)
	return r.ledgerService.GetUserTransactions(userID, limit, offset)
}

func (r *Resolver) getUserSummaryResolver(p graphql.ResolveParams) (interface{}, error) {
	userID := p.Args["user_id"].(string)
	return r.ledgerService.GetUserSummary(userID)
}

// Mutation Resolvers
func (r *Resolver) createUserResolver(p graphql.ResolveParams) (interface{}, error) {
	name := p.Args["name"].(string)
	email := p.Args["email"].(string)
	
	now := time.Now()
	user := &models.User{
		ID:        uuid.New().String(),
		Name:      name,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	if err := r.db.CreateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Resolver) createCardResolver(p graphql.ResolveParams) (interface{}, error) {
	userID := p.Args["user_id"].(string)
	cardNumber := p.Args["card_number"].(string)
	cardType := p.Args["card_type"].(string)
	
	card := &models.Card{
		ID:         uuid.New().String(),
		UserID:     userID,
		CardNumber: cardNumber,
		CardType:   cardType,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}
	
	if err := r.db.CreateCard(card); err != nil {
		return nil, err
	}
	return card, nil
}

func (r *Resolver) processTransactionResolver(p graphql.ResolveParams) (interface{}, error) {
	payload := models.CardPayload{
		CardNumber:   p.Args["card_number"].(string),
		Amount:       int64(p.Args["amount"].(int)),
		MerchantName: p.Args["merchant_name"].(string),
		Category:     "general", // default category
		Timestamp:    time.Now().Format(time.RFC3339),
	}
	
	if mcc, ok := p.Args["mcc"].(string); ok {
		// Map MCC to category
		payload.Category = mapMCCToCategory(mcc)
	}
	
	return r.ledgerService.ProcessCardPayload(payload)
}

func mapMCCToCategory(mcc string) string {
	// Simple MCC to category mapping
	switch mcc {
	case "5411", "5412":
		return "groceries"
	case "5541", "5542":
		return "gas"
	case "5812", "5814":
		return "dining"
	default:
		return "general"
	}
}