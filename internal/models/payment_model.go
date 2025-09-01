package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	UserID         uuid.UUID       `json:"user_id" db:"user_id"`
	Processor      string          `json:"processor" db:"processor"`
	Plan           string          `json:"plan" db:"plan"`
	Amount         float64         `json:"amount" db:"amount"`
	Currency       string          `json:"currency" db:"currency"`
	Status         PaymentStatus   `json:"status" db:"status"`
	RawWebhookJSON json.RawMessage `json:"raw_webhook_json" db:"raw_webhook_json"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)
