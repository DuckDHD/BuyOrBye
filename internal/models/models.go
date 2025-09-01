package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
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

type GuardrailsJSON map[string]interface{}

// Implement database/sql interfaces for custom JSON type
func (g GuardrailsJSON) Value() (driver.Value, error) {
	if g == nil {
		return nil, nil
	}
	return json.Marshal(g)
}

func (g *GuardrailsJSON) Scan(value interface{}) error {
	if value == nil {
		*g = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, g)
	case string:
		return json.Unmarshal([]byte(v), g)
	default:
		return fmt.Errorf("cannot scan %T into GuardrailsJSON", value)
	}
}

type Decision struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	UserID    *uuid.UUID      `json:"user_id,omitempty" db:"user_id"`
	SessionID uuid.UUID       `json:"session_id" db:"session_id"`
	Input     DecisionInput   `json:"input" db:"input_json"`
	Verdict   DecisionVerdict `json:"verdict" db:"verdict"`
	Score     float64         `json:"score" db:"score"`
	Reasons   ReasonsJSON     `json:"reasons" db:"reasons_json"`
	Cost      float64         `json:"cost" db:"cost"`
	Currency  string          `json:"currency" db:"currency"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	Shared    bool            `json:"shared" db:"shared"`
	ShareURL  *string         `json:"share_url,omitempty" db:"share_url"`
}

type DecisionVerdict string

const (
	VerdictYes DecisionVerdict = "YES"
	VerdictNo  DecisionVerdict = "NO"
)

type DecisionInput struct {
	ItemName    string  `json:"item_name"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Category    string  `json:"category,omitempty"`
	IsNecessity bool    `json:"is_necessity"`
	Description string  `json:"description,omitempty"`
}

type ReasonsJSON []string

func (r ReasonsJSON) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return json.Marshal(r)
}

func (r *ReasonsJSON) Scan(value interface{}) error {
	if value == nil {
		*r = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, r)
	case string:
		return json.Unmarshal([]byte(v), r)
	default:
		return fmt.Errorf("cannot scan %T into ReasonsJSON", value)
	}
}

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

type Session struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	Anonymous bool       `json:"anonymous" db:"anon"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
}

type Referral struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	InviterUserID uuid.UUID      `json:"inviter_user_id" db:"inviter_user_id"`
	InviteeEmail  string         `json:"invitee_email" db:"invitee_email"`
	Status        ReferralStatus `json:"status" db:"status"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
}

type ReferralStatus string

const (
	ReferralPending   ReferralStatus = "pending"
	ReferralCompleted ReferralStatus = "completed"
	ReferralExpired   ReferralStatus = "expired"
)

// API Response types
type DecisionResponse struct {
	Decision    *Decision `json:"decision"`
	FlipToYes   *string   `json:"flip_to_yes,omitempty"`
	ShareURL    *string   `json:"share_url,omitempty"`
	CreditsLeft int       `json:"credits_left"`
}

type UserCredits struct {
	UserID       uuid.UUID `json:"user_id"`
	TotalCredits int       `json:"total_credits"`
	UsedToday    int       `json:"used_today"`
	ResetAt      time.Time `json:"reset_at"`
}
