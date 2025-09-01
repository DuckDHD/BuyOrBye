package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/config"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentService struct {
	db     *pgxpool.Pool
	config *config.Config
}

func NewPaymentService(cfg *config.Config) *PaymentService {
	return &PaymentService{
		config: cfg,
	}
}

func (s *PaymentService) CreatePayment(ctx context.Context, userID uuid.UUID, plan string, amount float64, currency string) (*models.Payment, error) {
	payment := &models.Payment{
		ID:        uuid.New(),
		UserID:    userID,
		Processor: "paddle",
		Plan:      plan,
		Amount:    amount,
		Currency:  currency,
		Status:    models.PaymentPending,
		CreatedAt: time.Now(),
	}

	query := `
		INSERT INTO payments (id, user_id, processor, plan, amount, currency, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := s.db.Exec(ctx, query,
		payment.ID, payment.UserID, payment.Processor,
		payment.Plan, payment.Amount, payment.Currency,
		payment.Status, payment.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, payload []byte) error {
	// Parse Paddle webhook
	var webhook struct {
		EventType string          `json:"event_type"`
		Data      json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(payload, &webhook); err != nil {
		return fmt.Errorf("failed to parse webhook: %w", err)
	}

	switch webhook.EventType {
	case "payment.succeeded":
		return s.handlePaymentSucceeded(ctx, webhook.Data)
	case "subscription.created":
		return s.handleSubscriptionCreated(ctx, webhook.Data)
	default:
		// Log unknown webhook type
		fmt.Printf("Unknown webhook type: %s\n", webhook.EventType)
	}

	return nil
}

func (s *PaymentService) handlePaymentSucceeded(ctx context.Context, data json.RawMessage) error {
	// Implementation for handling successful payments
	// This would update the payment status and grant credits/pro status
	return nil
}

func (s *PaymentService) handleSubscriptionCreated(ctx context.Context, data json.RawMessage) error {
	// Implementation for handling new subscriptions
	// This would update user to pro status
	return nil
}
