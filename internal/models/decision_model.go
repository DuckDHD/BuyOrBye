package models

import (
	"time"

	"github.com/google/uuid"
)

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

// API Response types
type DecisionResponse struct {
	Decision    *Decision `json:"decision"`
	FlipToYes   *string   `json:"flip_to_yes,omitempty"`
	ShareURL    *string   `json:"share_url,omitempty"`
	CreditsLeft int       `json:"credits_left"`
}
