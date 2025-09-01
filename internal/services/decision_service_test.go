package services

// import (
// 	"testing"

// 	"github.com/DuckDHD/BuyOrBye/internal/models"
// 	"github.com/stretchr/testify/assert"
// )

// func TestDecisionService_CalculateDecision(t *testing.T) {
// 	service := &DecisionService{}

// 	tests := []struct {
// 		name     string
// 		input    models.DecisionInput
// 		profile  *models.Profile
// 		costUSD  float64
// 		expected models.DecisionVerdict
// 	}{
// 		{
// 			name: "small necessity purchase",
// 			input: models.DecisionInput{
// 				ItemName:    "Coffee",
// 				Price:       5.0,
// 				Currency:    "USD",
// 				IsNecessity: true,
// 			},
// 			profile:  nil,
// 			costUSD:  5.0,
// 			expected: models.VerdictYes,
// 		},
// 		{
// 			name: "large purchase without profile",
// 			input: models.DecisionInput{
// 				ItemName: "Expensive gadget",
// 				Price:    500.0,
// 				Currency: "USD",
// 			},
// 			profile:  nil,
// 			costUSD:  500.0,
// 			expected: models.VerdictNo,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			verdict, _, _ := service.calculateDecision(tt.input, tt.profile, tt.costUSD)
// 			assert.Equal(t, tt.expected, verdict)
// 		})
// 	}
// }
