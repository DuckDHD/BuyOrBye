package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Email       string     `json:"email" db:"email"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	Locale      string     `json:"locale" db:"locale"`
	Currency    string     `json:"currency" db:"currency"`
	IsPro       bool       `json:"is_pro" db:"is_pro"`
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at"`
}

type Profile struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	UserID        uuid.UUID      `json:"user_id" db:"user_id"`
	Income        float64        `json:"income" db:"income"`
	FixedExpenses float64        `json:"fixed_expenses" db:"fixed_expenses"`
	DebtsMin      float64        `json:"debts_min" db:"debts_min"`
	SavingsLiquid float64        `json:"savings_liquid" db:"savings_liquid"`
	Guardrails    GuardrailsJSON `json:"guardrails" db:"guardrails_json"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}
