package repositories

import (
	"context"
	"fmt"
	"sync"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// inMemoryFinanceSummaryRepository provides a simple in-memory implementation
// This is a temporary implementation until a proper database-backed version is created
type inMemoryFinanceSummaryRepository struct {
	summaries map[string]domain.FinanceSummary
	mu        sync.RWMutex
}

// NewFinanceSummaryRepository creates a new in-memory finance summary repository
func NewFinanceSummaryRepository() services.FinanceSummaryRepository {
	return &inMemoryFinanceSummaryRepository{
		summaries: make(map[string]domain.FinanceSummary),
	}
}

// SaveFinanceSummary saves a finance summary to memory
func (r *inMemoryFinanceSummaryRepository) SaveFinanceSummary(ctx context.Context, summary domain.FinanceSummary) error {
	if summary.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.summaries[summary.UserID] = summary
	return nil
}

// GetFinanceSummaryByUserID retrieves a finance summary by user ID
func (r *inMemoryFinanceSummaryRepository) GetFinanceSummaryByUserID(ctx context.Context, userID string) (domain.FinanceSummary, error) {
	if userID == "" {
		return domain.FinanceSummary{}, fmt.Errorf("user ID is required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	summary, exists := r.summaries[userID]
	if !exists {
		return domain.FinanceSummary{}, domain.ErrFinanceSummaryNotFound
	}

	return summary, nil
}

// UpdateFinanceSummary updates an existing finance summary
func (r *inMemoryFinanceSummaryRepository) UpdateFinanceSummary(ctx context.Context, summary domain.FinanceSummary) error {
	if summary.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.summaries[summary.UserID]; !exists {
		return domain.ErrFinanceSummaryNotFound
	}

	r.summaries[summary.UserID] = summary
	return nil
}

// DeleteFinanceSummary removes a finance summary from memory
func (r *inMemoryFinanceSummaryRepository) DeleteFinanceSummary(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.summaries[userID]; !exists {
		return domain.ErrFinanceSummaryNotFound
	}

	delete(r.summaries, userID)
	return nil
}

// GetFinanceSummariesByHealthStatus returns summaries filtered by health status
func (r *inMemoryFinanceSummaryRepository) GetFinanceSummariesByHealthStatus(ctx context.Context, healthStatus string) ([]domain.FinanceSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.FinanceSummary
	for _, summary := range r.summaries {
		if summary.FinancialHealth == healthStatus {
			result = append(result, summary)
		}
	}

	return result, nil
}

// GetUsersWithHighDebtRatio returns summaries with debt ratio above threshold
func (r *inMemoryFinanceSummaryRepository) GetUsersWithHighDebtRatio(ctx context.Context, threshold float64) ([]domain.FinanceSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.FinanceSummary
	for _, summary := range r.summaries {
		if summary.DebtToIncomeRatio > threshold {
			result = append(result, summary)
		}
	}

	return result, nil
}

// GetUsersWithLowSavingsRate returns summaries with savings rate below threshold
func (r *inMemoryFinanceSummaryRepository) GetUsersWithLowSavingsRate(ctx context.Context, threshold float64) ([]domain.FinanceSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.FinanceSummary
	for _, summary := range r.summaries {
		if summary.SavingsRate < threshold {
			result = append(result, summary)
		}
	}

	return result, nil
}