package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPurchaseIntent_Validate_ValidData_Success(t *testing.T) {
	intent := PurchaseIntent{
		ID:          "test-id-123",
		UserID:      "user-456",
		ItemName:    "Test Item",
		ItemCost:    100.50,
		Category:    "electronics",
		Urgency:     "medium",
		Frequency:   "one_time",
		Purpose:     "Work laptop for development",
		Alternative: "Used laptop for half price",
		CreatedAt:   time.Now(),
	}

	err := intent.Validate()
	assert.NoError(t, err)
}

func TestPurchaseIntent_Validate_EmptyItemName_ReturnsError(t *testing.T) {
	intent := PurchaseIntent{
		ID:        "test-id",
		UserID:    "user-id",
		ItemName:  "", // Empty name should fail
		ItemCost:  100.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "item name")
}

func TestPurchaseIntent_Validate_ItemNameTooShort_ReturnsError(t *testing.T) {
	intent := PurchaseIntent{
		ID:        "test-id",
		UserID:    "user-id",
		ItemName:  "A", // Too short (less than 2 chars)
		ItemCost:  100.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "item name must be at least 2 characters")
}

func TestPurchaseIntent_Validate_ItemNameTooLong_ReturnsError(t *testing.T) {
	longName := string(make([]byte, 201)) // 201 characters, exceeds 200 limit
	for i := range longName {
		longName = string(longName[:i]) + "a" + string(longName[i+1:])
	}
	
	intent := PurchaseIntent{
		ID:        "test-id",
		UserID:    "user-id",
		ItemName:  longName,
		ItemCost:  100.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "item name must not exceed 200 characters")
}

func TestPurchaseIntent_Validate_ZeroCost_ReturnsError(t *testing.T) {
	intent := PurchaseIntent{
		ID:        "test-id",
		UserID:    "user-id",
		ItemName:  "Test Item",
		ItemCost:  0.0, // Zero cost should fail
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "item cost must be greater than 0")
}

func TestPurchaseIntent_Validate_NegativeCost_ReturnsError(t *testing.T) {
	intent := PurchaseIntent{
		ID:        "test-id",
		UserID:    "user-id",
		ItemName:  "Test Item",
		ItemCost:  -50.0, // Negative cost should fail
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "item cost must be greater than 0")
}

func TestPurchaseIntent_Validate_ExcessiveCost_ReturnsError(t *testing.T) {
	intent := PurchaseIntent{
		ID:        "test-id",
		UserID:    "user-id",
		ItemName:  "Test Item",
		ItemCost:  1000001.0, // Exceeds 1M limit
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "item cost must not exceed 1,000,000")
}

func TestPurchaseIntent_Validate_InvalidCategory_ReturnsError(t *testing.T) {
	intent := PurchaseIntent{
		ID:        "test-id",
		UserID:    "user-id",
		ItemName:  "Test Item",
		ItemCost:  100.0,
		Category:  "invalid-category", // Invalid category
		Urgency:   "medium",
		Frequency: "one_time",
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid category")
}

func TestPurchaseIntent_Validate_ValidCategories_Success(t *testing.T) {
	validCategories := []string{
		"electronics", "clothing", "food", "transport", "health", "entertainment", "other",
	}

	for _, category := range validCategories {
		t.Run("category_"+category, func(t *testing.T) {
			intent := PurchaseIntent{
				ID:        "test-id",
				UserID:    "user-id",
				ItemName:  "Test Item",
				ItemCost:  100.0,
				Category:  category,
				Urgency:   "medium",
				Frequency: "one_time",
			}

			err := intent.Validate()
			assert.NoError(t, err, "Category %s should be valid", category)
		})
	}
}

func TestPurchaseIntent_Validate_InvalidUrgency_ReturnsError(t *testing.T) {
	intent := PurchaseIntent{
		ID:        "test-id",
		UserID:    "user-id",
		ItemName:  "Test Item",
		ItemCost:  100.0,
		Category:  "electronics",
		Urgency:   "invalid-urgency", // Invalid urgency
		Frequency: "one_time",
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid urgency")
}

func TestPurchaseIntent_Validate_ValidUrgencies_Success(t *testing.T) {
	validUrgencies := []string{"low", "medium", "high", "critical"}

	for _, urgency := range validUrgencies {
		t.Run("urgency_"+urgency, func(t *testing.T) {
			intent := PurchaseIntent{
				ID:        "test-id",
				UserID:    "user-id",
				ItemName:  "Test Item",
				ItemCost:  100.0,
				Category:  "electronics",
				Urgency:   urgency,
				Frequency: "one_time",
			}

			err := intent.Validate()
			assert.NoError(t, err, "Urgency %s should be valid", urgency)
		})
	}
}

func TestPurchaseIntent_Validate_InvalidFrequency_ReturnsError(t *testing.T) {
	intent := PurchaseIntent{
		ID:        "test-id",
		UserID:    "user-id",
		ItemName:  "Test Item",
		ItemCost:  100.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "invalid-frequency", // Invalid frequency
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid frequency")
}

func TestPurchaseIntent_Validate_ValidFrequencies_Success(t *testing.T) {
	validFrequencies := []string{"one_time", "monthly", "yearly"}

	for _, frequency := range validFrequencies {
		t.Run("frequency_"+frequency, func(t *testing.T) {
			intent := PurchaseIntent{
				ID:        "test-id",
				UserID:    "user-id",
				ItemName:  "Test Item",
				ItemCost:  100.0,
				Category:  "electronics",
				Urgency:   "medium",
				Frequency: frequency,
			}

			err := intent.Validate()
			assert.NoError(t, err, "Frequency %s should be valid", frequency)
		})
	}
}

func TestPurchaseIntent_Validate_PurposeTooLong_ReturnsError(t *testing.T) {
	longPurpose := string(make([]byte, 501)) // 501 characters, exceeds 500 limit
	for i := range longPurpose {
		longPurpose = string(longPurpose[:i]) + "a" + string(longPurpose[i+1:])
	}
	
	intent := PurchaseIntent{
		ID:        "test-id",
		UserID:    "user-id",
		ItemName:  "Test Item",
		ItemCost:  100.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
		Purpose:   longPurpose,
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "purpose must not exceed 500 characters")
}

func TestPurchaseIntent_Validate_AlternativeTooLong_ReturnsError(t *testing.T) {
	longAlternative := string(make([]byte, 201)) // 201 characters, exceeds 200 limit
	for i := range longAlternative {
		longAlternative = string(longAlternative[:i]) + "a" + string(longAlternative[i+1:])
	}
	
	intent := PurchaseIntent{
		ID:          "test-id",
		UserID:      "user-id",
		ItemName:    "Test Item",
		ItemCost:    100.0,
		Category:    "electronics",
		Urgency:     "medium",
		Frequency:   "one_time",
		Alternative: longAlternative,
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "alternative must not exceed 200 characters")
}

func TestPurchaseIntent_Validate_EmptyUserID_ReturnsError(t *testing.T) {
	intent := PurchaseIntent{
		ID:        "test-id",
		UserID:    "", // Empty UserID should fail
		ItemName:  "Test Item",
		ItemCost:  100.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestPurchaseIntent_Validate_EmptyID_ReturnsError(t *testing.T) {
	intent := PurchaseIntent{
		ID:        "", // Empty ID should fail
		UserID:    "user-id",
		ItemName:  "Test Item",
		ItemCost:  100.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	err := intent.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ID is required")
}

func TestPurchaseIntent_GetMonthlyImpact_OneTime_ReturnsCost(t *testing.T) {
	intent := PurchaseIntent{
		ItemCost:  1200.0,
		Frequency: "one_time",
	}

	impact := intent.GetMonthlyImpact()
	assert.Equal(t, 1200.0, impact)
}

func TestPurchaseIntent_GetMonthlyImpact_Monthly_ReturnsCost(t *testing.T) {
	intent := PurchaseIntent{
		ItemCost:  300.0,
		Frequency: "monthly",
	}

	impact := intent.GetMonthlyImpact()
	assert.Equal(t, 300.0, impact)
}

func TestPurchaseIntent_GetMonthlyImpact_Yearly_ReturnsYearlyCostDividedBy12(t *testing.T) {
	intent := PurchaseIntent{
		ItemCost:  1200.0,
		Frequency: "yearly",
	}

	impact := intent.GetMonthlyImpact()
	assert.Equal(t, 100.0, impact)
}

func TestPurchaseIntent_IsEssential_Health_ReturnsTrue(t *testing.T) {
	intent := PurchaseIntent{
		Category: "health",
	}

	assert.True(t, intent.IsEssential())
}

func TestPurchaseIntent_IsEssential_Food_ReturnsTrue(t *testing.T) {
	intent := PurchaseIntent{
		Category: "food",
	}

	assert.True(t, intent.IsEssential())
}

func TestPurchaseIntent_IsEssential_Transport_ReturnsTrue(t *testing.T) {
	intent := PurchaseIntent{
		Category: "transport",
	}

	assert.True(t, intent.IsEssential())
}

func TestPurchaseIntent_IsEssential_Entertainment_ReturnsFalse(t *testing.T) {
	intent := PurchaseIntent{
		Category: "entertainment",
	}

	assert.False(t, intent.IsEssential())
}

func TestPurchaseIntent_IsEssential_Electronics_ReturnsFalse(t *testing.T) {
	intent := PurchaseIntent{
		Category: "electronics",
	}

	assert.False(t, intent.IsEssential())
}