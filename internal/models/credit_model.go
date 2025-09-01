package models

import (
	"time"

	"github.com/google/uuid"
)

type CreditTransaction struct {
	ID        uuid.UUID             `json:"id" db:"id"`
	UserID    uuid.UUID             `json:"user_id" db:"user_id"`
	Delta     int                   `json:"delta" db:"delta"`
	Reason    CreditTransactionType `json:"reason" db:"reason"`
	RefID     *string               `json:"ref_id,omitempty" db:"ref_id"`
	CreatedAt time.Time             `json:"created_at" db:"created_at"`
}

type CreditTransactionType string

const (
	CreditPurchase CreditTransactionType = "purchase"
	CreditReferral CreditTransactionType = "referral"
	CreditGrant    CreditTransactionType = "grant"
)

type UserCredits struct {
	UserID       uuid.UUID `json:"user_id"`
	TotalCredits int       `json:"total_credits"`
	UsedToday    int       `json:"used_today"`
	ResetAt      time.Time `json:"reset_at"`
}
