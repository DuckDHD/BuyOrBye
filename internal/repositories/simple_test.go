package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
)

func setupSimpleTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.DecisionRecordModel{},
		&models.AIPromptLogModel{},
	)
	require.NoError(t, err)

	return db
}

func TestDecisionRepository_BasicOperations(t *testing.T) {
	db := setupSimpleTestDB(t)
	repo := NewDecisionRepository(db)
	ctx := context.Background()

	// Create test decision outcome
	outcome := domain.DecisionOutcome{
		UserID:        "test-user-123",
		IntentID:      "test-intent-456",
		Decision:      "BUY",
		Confidence:    0.85,
		PrimaryReason: "Good value for money",
	}

	// Create test purchase intent
	intent := domain.PurchaseIntent{
		ItemName: "Test Laptop",
		ItemCost: 999.99,
		Category: "electronics",
		Urgency:  "medium",
	}

	// Test saving decision
	err := repo.SaveDecision(ctx, outcome, intent)
	assert.NoError(t, err)

	// Test getting decision stats
	stats, err := repo.GetDecisionStats(ctx, "test-user-123", 30)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.TotalDecisions)
	assert.Equal(t, int64(1), stats.TotalBuyDecisions)

	// Test getting decision history
	history, err := repo.GetDecisionHistory(ctx, "test-user-123", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, "test-user-123", history[0].UserID)
	assert.Equal(t, "BUY", history[0].Decision)
}

func TestPromptLogRepository_BasicOperations(t *testing.T) {
	db := setupSimpleTestDB(t)
	repo := NewPromptLogRepository(db)
	ctx := context.Background()

	// Create test AI prompt
	prompt := domain.AIPrompt{
		SystemContext:    "You are a purchase decision assistant",
		UserContext:      "User has budget of $1000",
		PurchaseDetails:  "Item: Laptop, Cost: $800",
		DecisionCriteria: "Consider budget and necessity",
		ResponseFormat:   "JSON format",
		MaxTokens:        500,
		Temperature:      0.7,
	}

	// Test logging prompt
	logID, err := repo.LogPrompt(ctx, prompt, "test-user-123", "test-request-456", "test-intent-789")
	assert.NoError(t, err)
	assert.NotEmpty(t, logID)

	// Test updating with response
	response := domain.AIResponse{
		RawResponse: `{"decision": "BUY", "confidence": 0.85}`,
		Decision:    "BUY",
		Confidence:  0.85,
		TokensUsed:  150,
	}

	err = repo.UpdateWithResponse(ctx, logID, response, 2500)
	assert.NoError(t, err)

	// Test getting token usage stats
	stats, err := repo.GetTokenUsageStats(ctx, "test-user-123", 24)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.SuccessfulRequests)
	assert.Greater(t, stats.TotalTokens, int64(0))

	// Test generating prompt hash
	hash := repo.GeneratePromptHash(prompt)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64) // SHA-256 hash should be 64 characters
}