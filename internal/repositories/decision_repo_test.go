package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
)

func setupDecisionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate decision models
	err = db.AutoMigrate(
		&models.DecisionRecordModel{},
	)
	require.NoError(t, err)

	return db
}

func createTestDecisionOutcome(userID, intentID string) *domain.DecisionOutcome {
	return &domain.DecisionOutcome{
		UserID:        userID,
		IntentID:      intentID,
		Decision:      "BUY",
		Confidence:    0.85,
		PrimaryReason: "Good value for money and within budget",
		Factors: []domain.DecisionFactor{
			{
				Category:    "financial",
				Impact:      "positive",
				Weight:      0.8,
				Description: "Price is below budget threshold",
			},
			{
				Category:    "practical",
				Impact:      "positive",
				Weight:      0.7,
				Description: "Item fulfills immediate need",
			},
		},
		Recommendations: []string{
			"Purchase from vendor with best warranty",
			"Consider buying during sale period",
		},
		WaitPeriod:     0,
		MaxBudget:      1200.0,
		CreatedAt:      time.Now(),
		ProcessingTime: 2500,
	}
}

func createTestPurchaseIntent(itemName string, cost float64, category string) domain.PurchaseIntent {
	return domain.PurchaseIntent{
		ItemName:    itemName,
		ItemCost:    cost,
		Category:    category,
		Urgency:     "medium",
		Frequency:   "one_time",
		Purpose:     "Essential item for daily use",
		Alternative: "Consider similar items from other brands",
	}
}

// PHASE 1: Repository Tests (RED) - Test cases that should initially fail

func TestDecisionRepository_SaveDecision_Success(t *testing.T) {
	db := setupDecisionTestDB(t)
	repo := NewDecisionRepository(db)
	ctx := context.Background()

	outcome := createTestDecisionOutcome("user-123", "intent-456")
	intent := createTestPurchaseIntent("Laptop", 999.99, "electronics")

	err := repo.SaveDecision(ctx, *outcome, intent)

	assert.NoError(t, err)

	// Verify record was saved to database
	var count int64
	db.Model(&models.DecisionRecordModel{}).Where("user_id = ? AND intent_id = ?", "user-123", "intent-456").Count(&count)
	assert.Equal(t, int64(1), count)

	// Verify data integrity
	var record models.DecisionRecordModel
	db.Where("user_id = ? AND intent_id = ?", "user-123", "intent-456").First(&record)
	assert.Equal(t, "user-123", record.UserID)
	assert.Equal(t, "intent-456", record.IntentID)
	assert.Equal(t, "BUY", record.Decision)
	assert.Equal(t, 0.85, record.Confidence)
	assert.Equal(t, "Laptop", record.ItemName)
	assert.Equal(t, 999.99, record.ItemCost)
	assert.Equal(t, "electronics", record.Category)
	assert.NotEmpty(t, record.FactorsJSON)
	assert.NotEmpty(t, record.RecommendationsJSON)
}

func TestDecisionRepository_SaveDecision_DuplicateIntentIDFails(t *testing.T) {
	db := setupDecisionTestDB(t)
	repo := NewDecisionRepository(db)
	ctx := context.Background()

	outcome1 := createTestDecisionOutcome("user-123", "intent-456")
	outcome2 := createTestDecisionOutcome("user-456", "intent-456") // Same intent ID
	intent := createTestPurchaseIntent("Laptop", 999.99, "electronics")

	// First save should succeed
	err := repo.SaveDecision(ctx, *outcome1, intent)
	assert.NoError(t, err)

	// Second save with same intent ID should fail due to unique constraint
	err = repo.SaveDecision(ctx, *outcome2, intent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UNIQUE constraint")
}

func TestDecisionRepository_SaveDecision_InvalidDecisionFails(t *testing.T) {
	db := setupDecisionTestDB(t)
	repo := NewDecisionRepository(db)
	ctx := context.Background()

	outcome := createTestDecisionOutcome("user-123", "intent-456")
	outcome.Decision = "INVALID" // Invalid decision value
	intent := createTestPurchaseIntent("Laptop", 999.99, "electronics")

	err := repo.SaveDecision(ctx, *outcome, intent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid decision")
}

func TestDecisionRepository_GetDecisionHistory_OrderByCreatedDesc(t *testing.T) {
	db := setupDecisionTestDB(t)
	repo := NewDecisionRepository(db)
	ctx := context.Background()

	// Create multiple decisions with different timestamps
	decisions := []struct {
		intentID  string
		decision  string
		createdAt time.Time
	}{
		{"intent-1", "BUY", time.Now().Add(-3 * 24 * time.Hour)},  // 3 days ago
		{"intent-2", "WAIT", time.Now().Add(-1 * 24 * time.Hour)}, // 1 day ago  
		{"intent-3", "BYE", time.Now().Add(-5 * 24 * time.Hour)},  // 5 days ago
	}

	// Save decisions
	for _, d := range decisions {
		outcome := createTestDecisionOutcome("user-123", d.intentID)
		outcome.Decision = d.decision
		outcome.CreatedAt = d.createdAt
		intent := createTestPurchaseIntent("Test Item", 100.0, "electronics")
		
		err := repo.SaveDecision(ctx, *outcome, intent)
		require.NoError(t, err)
	}

	// Get history
	history, err := repo.GetDecisionHistory(ctx, "user-123", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, history, 3)

	// Verify order (most recent first)
	assert.Equal(t, "intent-2", history[0].IntentID) // 1 day ago (most recent)
	assert.Equal(t, "intent-1", history[1].IntentID) // 3 days ago
	assert.Equal(t, "intent-3", history[2].IntentID) // 5 days ago (oldest)

	// Verify decision values
	assert.Equal(t, "WAIT", history[0].Decision)
	assert.Equal(t, "BUY", history[1].Decision)
	assert.Equal(t, "BYE", history[2].Decision)
}

func TestDecisionRepository_GetDecisionHistory_WithPagination(t *testing.T) {
	db := setupDecisionTestDB(t)
	repo := NewDecisionRepository(db)
	ctx := context.Background()

	// Create 5 decisions
	for i := 0; i < 5; i++ {
		outcome := createTestDecisionOutcome("user-123", "intent-"+string(rune(i+'1')))
		outcome.CreatedAt = time.Now().Add(time.Duration(-i) * time.Hour)
		intent := createTestPurchaseIntent("Test Item", 100.0, "electronics")
		
		err := repo.SaveDecision(ctx, *outcome, intent)
		require.NoError(t, err)
	}

	// Test pagination: Get first 3 records
	history1, err := repo.GetDecisionHistory(ctx, "user-123", 3, 0)
	assert.NoError(t, err)
	assert.Len(t, history1, 3)

	// Test pagination: Get next 2 records
	history2, err := repo.GetDecisionHistory(ctx, "user-123", 3, 3)
	assert.NoError(t, err)
	assert.Len(t, history2, 2)

	// Verify no overlap
	intentIDs1 := make(map[string]bool)
	for _, decision := range history1 {
		intentIDs1[decision.IntentID] = true
	}

	for _, decision := range history2 {
		assert.False(t, intentIDs1[decision.IntentID], "Found overlapping intent ID: %s", decision.IntentID)
	}
}

func TestDecisionRepository_GetDecisionsByCategory(t *testing.T) {
	db := setupDecisionTestDB(t)
	repo := NewDecisionRepository(db)
	ctx := context.Background()

	// Create decisions for different categories
	categories := []string{"electronics", "clothing", "food", "electronics"}
	for i, category := range categories {
		outcome := createTestDecisionOutcome("user-123", "intent-"+string(rune(i+'1')))
		intent := createTestPurchaseIntent("Test Item", 100.0, category)
		
		err := repo.SaveDecision(ctx, *outcome, intent)
		require.NoError(t, err)
	}

	// Get decisions for electronics category only
	electronicsDecisions, err := repo.GetDecisionsByCategory(ctx, "user-123", "electronics", 30)
	assert.NoError(t, err)
	assert.Len(t, electronicsDecisions, 2) // Should have 2 electronics decisions

	for _, decision := range electronicsDecisions {
		// Note: Category is not in DecisionOutcome, but we can verify UserID
		assert.Equal(t, "user-123", decision.UserID)
	}

	// Get decisions for clothing category
	clothingDecisions, err := repo.GetDecisionsByCategory(ctx, "user-123", "clothing", 30)
	assert.NoError(t, err)
	assert.Len(t, clothingDecisions, 1) // Should have 1 clothing decision

	// Get decisions for non-existent category
	noneDecisions, err := repo.GetDecisionsByCategory(ctx, "user-123", "books", 30)
	assert.NoError(t, err)
	assert.Len(t, noneDecisions, 0) // Should have no decisions
}

func TestDecisionRepository_GetDecisionStats_Calculations(t *testing.T) {
	db := setupDecisionTestDB(t)
	repo := NewDecisionRepository(db)
	ctx := context.Background()

	// Create test decisions with known statistics
	testData := []struct {
		decision   string
		confidence float64
		cost       float64
		category   string
	}{
		{"BUY", 0.9, 500.0, "electronics"},
		{"BUY", 0.8, 200.0, "clothing"},
		{"WAIT", 0.7, 1000.0, "electronics"},
		{"BYE", 0.6, 50.0, "food"},
		{"BUY", 0.85, 300.0, "electronics"},
	}

	for i, data := range testData {
		outcome := createTestDecisionOutcome("user-123", "intent-"+string(rune(i+'1')))
		outcome.Decision = data.decision
		outcome.Confidence = data.confidence
		intent := createTestPurchaseIntent("Test Item", data.cost, data.category)
		
		err := repo.SaveDecision(ctx, *outcome, intent)
		require.NoError(t, err)
	}

	// Get statistics
	stats, err := repo.GetDecisionStats(ctx, "user-123", 30)

	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// Verify counts
	assert.Equal(t, int64(3), stats.TotalBuyDecisions)   // 3 BUY decisions
	assert.Equal(t, int64(1), stats.TotalWaitDecisions)  // 1 WAIT decision
	assert.Equal(t, int64(1), stats.TotalByeDecisions)   // 1 BYE decision
	assert.Equal(t, int64(5), stats.TotalDecisions)      // 5 total decisions

	// Verify averages (approximately)
	expectedAvgConfidence := (0.9 + 0.8 + 0.7 + 0.6 + 0.85) / 5
	assert.InDelta(t, expectedAvgConfidence, stats.AverageConfidence, 0.01)

	expectedAvgBuyCost := (500.0 + 200.0 + 300.0) / 3 // Only BUY decisions
	assert.InDelta(t, expectedAvgBuyCost, stats.AverageBuyCost, 0.01)

	// Verify total spent (only BUY decisions)
	expectedTotalSpent := 500.0 + 200.0 + 300.0
	assert.Equal(t, expectedTotalSpent, stats.TotalAmountSpent)

	// Verify most frequent category
	assert.Equal(t, "electronics", stats.MostFrequentCategory) // 3 electronics decisions
}

func TestDecisionRepository_GetDecisionStats_EmptyResults(t *testing.T) {
	db := setupDecisionTestDB(t)
	repo := NewDecisionRepository(db)
	ctx := context.Background()

	// Get stats for user with no decisions
	stats, err := repo.GetDecisionStats(ctx, "non-existent-user", 30)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalDecisions)
	assert.Equal(t, int64(0), stats.TotalBuyDecisions)
	assert.Equal(t, int64(0), stats.TotalWaitDecisions)
	assert.Equal(t, int64(0), stats.TotalByeDecisions)
	assert.Equal(t, 0.0, stats.AverageConfidence)
	assert.Equal(t, 0.0, stats.AverageBuyCost)
	assert.Equal(t, 0.0, stats.TotalAmountSpent)
	assert.Equal(t, "", stats.MostFrequentCategory)
}

func TestDecisionRepository_GetDecisionHistory_FilterByDateRange(t *testing.T) {
	db := setupDecisionTestDB(t)
	repo := NewDecisionRepository(db)
	ctx := context.Background()

	now := time.Now()
	
	// Create decisions with different timestamps
	timestamps := []time.Time{
		now.Add(-40 * 24 * time.Hour), // 40 days ago (outside 30-day range)
		now.Add(-20 * 24 * time.Hour), // 20 days ago (within 30-day range)
		now.Add(-10 * 24 * time.Hour), // 10 days ago (within 30-day range)
		now.Add(-5 * 24 * time.Hour),  // 5 days ago (within 30-day range)
	}

	for i, timestamp := range timestamps {
		outcome := createTestDecisionOutcome("user-123", "intent-"+string(rune(i+'1')))
		outcome.CreatedAt = timestamp
		intent := createTestPurchaseIntent("Test Item", 100.0, "electronics")
		
		err := repo.SaveDecision(ctx, *outcome, intent)
		require.NoError(t, err)
	}

	// Get history for last 30 days
	history, err := repo.GetDecisionHistory(ctx, "user-123", 10, 0)
	
	assert.NoError(t, err)
	assert.Len(t, history, 3) // Should exclude the 40-day-old decision

	// Verify all returned decisions are within the date range
	thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)
	for _, decision := range history {
		assert.True(t, decision.CreatedAt.After(thirtyDaysAgo), 
			"Decision timestamp %v should be after %v", decision.CreatedAt, thirtyDaysAgo)
	}
}

func TestDecisionRepository_GetDecisionsByCategory_RecentOnly(t *testing.T) {
	db := setupDecisionTestDB(t)
	repo := NewDecisionRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create decisions - some recent, some old
	testData := []struct {
		intentID  string
		category  string
		createdAt time.Time
	}{
		{"intent-1", "electronics", now.Add(-10 * 24 * time.Hour)}, // Recent
		{"intent-2", "electronics", now.Add(-40 * 24 * time.Hour)}, // Old
		{"intent-3", "electronics", now.Add(-5 * 24 * time.Hour)},  // Recent
	}

	for _, data := range testData {
		outcome := createTestDecisionOutcome("user-123", data.intentID)
		outcome.CreatedAt = data.createdAt
		intent := createTestPurchaseIntent("Test Item", 100.0, data.category)
		
		err := repo.SaveDecision(ctx, *outcome, intent)
		require.NoError(t, err)
	}

	// Get electronics decisions from last 30 days
	decisions, err := repo.GetDecisionsByCategory(ctx, "user-123", "electronics", 30)

	assert.NoError(t, err)
	assert.Len(t, decisions, 2) // Should only return recent decisions

	// Verify returned decisions are recent
	thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)
	for _, decision := range decisions {
		assert.True(t, decision.CreatedAt.After(thirtyDaysAgo))
	}
}

func TestDecisionRepository_GetDecisionStats_DateRangeFilter(t *testing.T) {
	db := setupDecisionTestDB(t)
	repo := NewDecisionRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create decisions - some recent, some old
	testData := []struct {
		decision  string
		createdAt time.Time
		cost      float64
	}{
		{"BUY", now.Add(-10 * 24 * time.Hour), 100.0}, // Recent
		{"BUY", now.Add(-40 * 24 * time.Hour), 200.0}, // Old (should be excluded)
		{"WAIT", now.Add(-5 * 24 * time.Hour), 300.0}, // Recent
	}

	for i, data := range testData {
		outcome := createTestDecisionOutcome("user-123", "intent-"+string(rune(i+'1')))
		outcome.Decision = data.decision
		outcome.CreatedAt = data.createdAt
		intent := createTestPurchaseIntent("Test Item", data.cost, "electronics")
		
		err := repo.SaveDecision(ctx, *outcome, intent)
		require.NoError(t, err)
	}

	// Get stats for last 30 days
	stats, err := repo.GetDecisionStats(ctx, "user-123", 30)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), stats.TotalDecisions)      // Only recent decisions
	assert.Equal(t, int64(1), stats.TotalBuyDecisions)   // Only 1 recent BUY
	assert.Equal(t, int64(1), stats.TotalWaitDecisions)  // Only 1 recent WAIT
	assert.Equal(t, 100.0, stats.TotalAmountSpent)       // Only recent BUY cost
}