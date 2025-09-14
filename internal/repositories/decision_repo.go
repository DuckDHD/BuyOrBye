package repositories

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// DecisionRepository implements the DecisionRepository interface using GORM
type DecisionRepository struct {
	db *gorm.DB
}

// NewDecisionRepository creates a new instance of DecisionRepository
func NewDecisionRepository(db *gorm.DB) services.DecisionRepository {
	return &DecisionRepository{
		db: db,
	}
}

// SaveDecision saves a decision outcome and purchase intent to the database
func (r *DecisionRepository) SaveDecision(ctx context.Context, outcome domain.DecisionOutcome, intent domain.PurchaseIntent) error {
	// Create model instance
	record := &models.DecisionRecordModel{}
	
	// Convert domain to model
	if err := record.FromDomain(outcome, intent); err != nil {
		return fmt.Errorf("failed to convert decision to model: %w", err)
	}
	
	// Save to database
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("failed to save decision record: %w", err)
	}
	
	return nil
}

// GetDecisionByID retrieves a single decision by its ID
func (r *DecisionRepository) GetDecisionByID(ctx context.Context, decisionID string) (*domain.DecisionOutcome, error) {
	// Parse string ID to uint
	id, err := strconv.ParseUint(decisionID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid decision ID format: %w", err)
	}
	
	var record models.DecisionRecordModel
	if err := r.db.WithContext(ctx).First(&record, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("decision not found with ID %s", decisionID)
		}
		return nil, fmt.Errorf("failed to get decision: %w", err)
	}
	
	// Convert model to domain
	outcome, err := record.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert model to domain: %w", err)
	}
	
	return outcome, nil
}

// GetDecisionHistory retrieves decision history for a user, ordered by creation date (newest first)
func (r *DecisionRepository) GetDecisionHistory(ctx context.Context, userID string, limit, offset int) ([]*domain.DecisionOutcome, error) {
	var records []models.DecisionRecordModel
	
	// Calculate date filter (last 30 days by default)
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at > ?", userID, thirtyDaysAgo).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)
	
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get decision history: %w", err)
	}
	
	// Convert models to domain objects
	decisions := make([]*domain.DecisionOutcome, len(records))
	for i, record := range records {
		outcome, err := record.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert record %d to domain: %w", i, err)
		}
		decisions[i] = outcome
	}
	
	return decisions, nil
}

// GetDecisionsByCategory retrieves decisions for a specific category within a time range
func (r *DecisionRepository) GetDecisionsByCategory(ctx context.Context, userID string, category string, daysBack int) ([]*domain.DecisionOutcome, error) {
	var records []models.DecisionRecordModel
	
	// Calculate date filter
	cutoffDate := time.Now().Add(-time.Duration(daysBack) * 24 * time.Hour)
	
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND category = ? AND created_at > ?", userID, category, cutoffDate).
		Order("created_at DESC")
	
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get decisions by category: %w", err)
	}
	
	// Convert models to domain objects
	decisions := make([]*domain.DecisionOutcome, len(records))
	for i, record := range records {
		outcome, err := record.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert record %d to domain: %w", i, err)
		}
		decisions[i] = outcome
	}
	
	return decisions, nil
}

// GetDecisionStats calculates comprehensive statistics for user decisions within a time range
func (r *DecisionRepository) GetDecisionStats(ctx context.Context, userID string, daysBack int) (*services.DecisionStats, error) {
	// Calculate date filter
	cutoffDate := time.Now().Add(-time.Duration(daysBack) * 24 * time.Hour)
	
	stats := &services.DecisionStats{}
	
	// Get total counts by decision type
	var totalCount int64
	if err := r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Where("user_id = ? AND created_at > ?", userID, cutoffDate).
		Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats.TotalDecisions = totalCount
	
	if totalCount == 0 {
		// Return empty stats for users with no decisions
		return stats, nil
	}
	
	// Count by decision type
	var buyCount, waitCount, byeCount int64
	
	r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Where("user_id = ? AND created_at > ? AND decision = ?", userID, cutoffDate, "BUY").
		Count(&buyCount)
	
	r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Where("user_id = ? AND created_at > ? AND decision = ?", userID, cutoffDate, "WAIT").
		Count(&waitCount)
	
	r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Where("user_id = ? AND created_at > ? AND decision = ?", userID, cutoffDate, "BYE").
		Count(&byeCount)
	
	stats.TotalBuyDecisions = buyCount
	stats.TotalWaitDecisions = waitCount
	stats.TotalByeDecisions = byeCount
	
	// Calculate averages
	var avgConfidenceResult struct {
		AvgConfidence float64
	}
	
	if err := r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Select("AVG(confidence) as avg_confidence").
		Where("user_id = ? AND created_at > ?", userID, cutoffDate).
		Scan(&avgConfidenceResult).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate average confidence: %w", err)
	}
	stats.AverageConfidence = avgConfidenceResult.AvgConfidence
	
	// Calculate average BUY cost and total spent (only for BUY decisions)
	if buyCount > 0 {
		var buyStatsResult struct {
			AvgCost   float64
			TotalCost float64
		}
		
		if err := r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
			Select("AVG(item_cost) as avg_cost, SUM(item_cost) as total_cost").
			Where("user_id = ? AND created_at > ? AND decision = ?", userID, cutoffDate, "BUY").
			Scan(&buyStatsResult).Error; err != nil {
			return nil, fmt.Errorf("failed to calculate buy statistics: %w", err)
		}
		stats.AverageBuyCost = buyStatsResult.AvgCost
		stats.TotalAmountSpent = buyStatsResult.TotalCost
	}
	
	// Find most frequent category
	var categoryResult struct {
		Category string
		Count    int64
	}
	
	if err := r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Select("category, COUNT(*) as count").
		Where("user_id = ? AND created_at > ?", userID, cutoffDate).
		Group("category").
		Order("count DESC").
		Limit(1).
		Scan(&categoryResult).Error; err != nil {
		// It's okay if no category is found
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to find most frequent category: %w", err)
		}
	}
	stats.MostFrequentCategory = categoryResult.Category
	
	// Calculate high confidence rate (decisions with confidence >= 0.8)
	var highConfidenceCount int64
	r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Where("user_id = ? AND created_at > ? AND confidence >= ?", userID, cutoffDate, 0.8).
		Count(&highConfidenceCount)
	
	if totalCount > 0 {
		stats.HighConfidenceRate = float64(highConfidenceCount) / float64(totalCount)
		stats.BuySuccessRate = float64(buyCount) / float64(totalCount)
	}
	
	return stats, nil
}

// GetRecentDecisions retrieves recent decisions for a user within specified days
func (r *DecisionRepository) GetRecentDecisions(ctx context.Context, userID string, daysBack int) ([]*domain.DecisionOutcome, error) {
	var records []models.DecisionRecordModel
	
	// Calculate date filter
	cutoffDate := time.Now().Add(-time.Duration(daysBack) * 24 * time.Hour)
	
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at > ?", userID, cutoffDate).
		Order("created_at DESC")
	
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent decisions: %w", err)
	}
	
	// Convert models to domain objects
	decisions := make([]*domain.DecisionOutcome, len(records))
	for i, record := range records {
		outcome, err := record.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert record %d to domain: %w", i, err)
		}
		decisions[i] = outcome
	}
	
	return decisions, nil
}

// GetDecisionTrends calculates decision trends and patterns over time
func (r *DecisionRepository) GetDecisionTrends(ctx context.Context, userID string, daysBack int) (*services.DecisionTrends, error) {
	// Calculate date filter
	cutoffDate := time.Now().Add(-time.Duration(daysBack) * 24 * time.Hour)
	
	trends := &services.DecisionTrends{
		DailyDecisionCounts:  make(map[string]int),
		CategoryDistribution: make(map[string]int),
		DecisionDistribution: make(map[string]int),
		ConfidenceTrend:      make([]float64, 0),
		SpendingTrend:        make([]float64, 0),
	}
	
	// Get daily decision counts
	var dailyResults []struct {
		Date  string
		Count int
	}
	
	if err := r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("user_id = ? AND created_at > ?", userID, cutoffDate).
		Group("DATE(created_at)").
		Order("date").
		Scan(&dailyResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get daily decision counts: %w", err)
	}
	
	for _, result := range dailyResults {
		trends.DailyDecisionCounts[result.Date] = result.Count
	}
	
	// Get category distribution
	var categoryResults []struct {
		Category string
		Count    int
	}
	
	if err := r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Select("category, COUNT(*) as count").
		Where("user_id = ? AND created_at > ?", userID, cutoffDate).
		Group("category").
		Scan(&categoryResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get category distribution: %w", err)
	}
	
	for _, result := range categoryResults {
		trends.CategoryDistribution[result.Category] = result.Count
	}
	
	// Get decision type distribution
	var decisionResults []struct {
		Decision string
		Count    int
	}
	
	if err := r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Select("decision, COUNT(*) as count").
		Where("user_id = ? AND created_at > ?", userID, cutoffDate).
		Group("decision").
		Scan(&decisionResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get decision distribution: %w", err)
	}
	
	for _, result := range decisionResults {
		trends.DecisionDistribution[result.Decision] = result.Count
	}
	
	// Get daily confidence trends
	var confidenceResults []struct {
		Date           string
		AvgConfidence  float64
	}
	
	if err := r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Select("DATE(created_at) as date, AVG(confidence) as avg_confidence").
		Where("user_id = ? AND created_at > ?", userID, cutoffDate).
		Group("DATE(created_at)").
		Order("date").
		Scan(&confidenceResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get confidence trends: %w", err)
	}
	
	for _, result := range confidenceResults {
		trends.ConfidenceTrend = append(trends.ConfidenceTrend, result.AvgConfidence)
	}
	
	// Get daily spending trends (only for BUY decisions)
	var spendingResults []struct {
		Date         string
		TotalSpent   float64
	}
	
	if err := r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Select("DATE(created_at) as date, SUM(item_cost) as total_spent").
		Where("user_id = ? AND created_at > ? AND decision = ?", userID, cutoffDate, "BUY").
		Group("DATE(created_at)").
		Order("date").
		Scan(&spendingResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get spending trends: %w", err)
	}
	
	for _, result := range spendingResults {
		trends.SpendingTrend = append(trends.SpendingTrend, result.TotalSpent)
	}
	
	return trends, nil
}

// Additional helper methods for repository functionality

// GetDecisionCount returns the total number of decisions for a user
func (r *DecisionRepository) GetDecisionCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.DecisionRecordModel{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get decision count: %w", err)
	}
	return count, nil
}

// GetLatestDecision returns the most recent decision for a user
func (r *DecisionRepository) GetLatestDecision(ctx context.Context, userID string) (*domain.DecisionOutcome, error) {
	var record models.DecisionRecordModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no decisions found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to get latest decision: %w", err)
	}
	
	outcome, err := record.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert latest decision to domain: %w", err)
	}
	
	return outcome, nil
}

// DeleteDecision removes a decision record (for cleanup or testing purposes)
func (r *DecisionRepository) DeleteDecision(ctx context.Context, decisionID string) error {
	// Parse string ID to uint
	id, err := strconv.ParseUint(decisionID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid decision ID format: %w", err)
	}
	
	result := r.db.WithContext(ctx).Delete(&models.DecisionRecordModel{}, uint(id))
	if result.Error != nil {
		return fmt.Errorf("failed to delete decision: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("decision not found with ID %s", decisionID)
	}
	
	return nil
}

// GetDecisionsByDateRange retrieves decisions within a specific date range
func (r *DecisionRepository) GetDecisionsByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*domain.DecisionOutcome, error) {
	var records []models.DecisionRecordModel
	
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Order("created_at DESC")
	
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get decisions by date range: %w", err)
	}
	
	// Convert models to domain objects
	decisions := make([]*domain.DecisionOutcome, len(records))
	for i, record := range records {
		outcome, err := record.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to convert record %d to domain: %w", i, err)
		}
		decisions[i] = outcome
	}
	
	return decisions, nil
}