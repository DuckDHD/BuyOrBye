package handlers

import (
	"context"

	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// DecisionServiceInterface defines the methods that the decision handler expects from the decision service
// This interface is defined in the handler package to follow dependency inversion principle
type DecisionServiceInterface interface {
	// MakeDecision processes a purchase intent and returns a decision outcome
	MakeDecision(ctx context.Context, intent types.PurchaseIntentDTO) (*types.DecisionResponseDTO, error)

	// GetDecisionHistory retrieves past decisions for a user within the specified number of days
	GetDecisionHistory(ctx context.Context, userID string, days int) (*types.DecisionHistoryDTO, error)
}