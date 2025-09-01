package domain

import (
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/google/uuid"
)

// Enhanced models for multi-step validation
type ValidationStep string

const (
	StepFinancialProfile ValidationStep = "financial_profile"
	StepDebtAssessment   ValidationStep = "debt_assessment"
	StepHealthContext    ValidationStep = "health_context"
	StepTransportation   ValidationStep = "transportation"
	StepPurchaseReason   ValidationStep = "purchase_reason"
	StepComplete         ValidationStep = "complete"
)

type ValidationState struct {
	ID                  uuid.UUID                   `json:"id"`
	SessionID           uuid.UUID                   `json:"session_id"`
	UserID              *uuid.UUID                  `json:"user_id,omitempty"`
	CurrentStep         ValidationStep              `json:"current_step"`
	FinancialValidation *models.FinancialValidation `json:"financial_validation,omitempty"`
	DebtAssessment      *models.DebtAssessment      `json:"debt_assessment,omitempty"`
	HealthContext       *models.HealthContext       `json:"health_context,omitempty"`
	TransportationInfo  *models.TransportationInfo  `json:"transportation_info,omitempty"`
	PurchaseReason      *models.PurchaseReason      `json:"purchase_reason,omitempty"`
	OriginalInput       *models.DecisionInput       `json:"original_input,omitempty"`
	CreatedAt           time.Time                   `json:"created_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
}

type ValidationResponse struct {
	State     *ValidationState         `json:"state"`
	NextStep  ValidationStep           `json:"next_step"`
	Questions []string                 `json:"questions,omitempty"`
	Complete  bool                     `json:"complete"`
	Decision  *models.DecisionResponse `json:"decision,omitempty"`
}
