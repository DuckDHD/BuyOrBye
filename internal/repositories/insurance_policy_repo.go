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

// insurancePolicyRepository implements services.InsurancePolicyRepository
type insurancePolicyRepository struct {
	db *gorm.DB
}

// NewInsurancePolicyRepository creates a new insurance policy repository
func NewInsurancePolicyRepository(db *gorm.DB) services.InsurancePolicyRepository {
	return &insurancePolicyRepository{db: db}
}

// Create creates a new insurance policy
func (r *insurancePolicyRepository) Create(ctx context.Context, policy *domain.InsurancePolicy) (*domain.InsurancePolicy, error) {
	// Convert ProfileID from string to uint
	profileID, err := strconv.ParseUint(policy.ProfileID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	model := &models.InsurancePolicyModel{}
	model.FromDomain(policy, uint(profileID))

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if containsStr(err.Error(), "UNIQUE constraint failed") ||
			containsStr(err.Error(), "Duplicate entry") ||
			containsStr(err.Error(), "unique constraint") {
			return nil, fmt.Errorf("insurance policy with number %s already exists: unique constraint violation", policy.PolicyNumber)
		}
		return nil, fmt.Errorf("failed to create insurance policy: %w", err)
	}

	return model.ToDomain(), nil
}

// GetByID retrieves an insurance policy by ID
func (r *insurancePolicyRepository) GetByID(ctx context.Context, id string) (*domain.InsurancePolicy, error) {
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid policy ID: %w", err)
	}

	var model models.InsurancePolicyModel
	
	if err := r.db.WithContext(ctx).First(&model, uint(idUint)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("insurance policy with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get insurance policy: %w", err)
	}

	return model.ToDomain(), nil
}

// Update updates an insurance policy
func (r *insurancePolicyRepository) Update(ctx context.Context, policy *domain.InsurancePolicy) (*domain.InsurancePolicy, error) {
	idUint, err := strconv.ParseUint(policy.ID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid policy ID: %w", err)
	}

	profileID, err := strconv.ParseUint(policy.ProfileID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	var model models.InsurancePolicyModel
	if err := r.db.WithContext(ctx).First(&model, uint(idUint)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("insurance policy with ID %s not found", policy.ID)
		}
		return nil, fmt.Errorf("failed to find insurance policy for update: %w", err)
	}

	// Update fields from domain
	model.FromDomain(policy, uint(profileID))
	model.ID = uint(idUint) // Preserve ID

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to update insurance policy: %w", err)
	}

	return model.ToDomain(), nil
}

// Delete performs soft delete on an insurance policy
func (r *insurancePolicyRepository) Delete(ctx context.Context, id string) error {
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid policy ID: %w", err)
	}

	result := r.db.WithContext(ctx).Delete(&models.InsurancePolicyModel{}, uint(idUint))
	if result.Error != nil {
		return fmt.Errorf("failed to delete insurance policy: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("insurance policy with ID %s not found", id)
	}

	return nil
}

// GetByUserID retrieves insurance policies by user ID
func (r *insurancePolicyRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.InsurancePolicy, error) {
	var models []models.InsurancePolicyModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get insurance policies: %w", err)
	}

	policies := make([]*domain.InsurancePolicy, len(models))
	for i, model := range models {
		policies[i] = model.ToDomain()
	}

	return policies, nil
}

// GetByType retrieves insurance policies by type
func (r *insurancePolicyRepository) GetByType(ctx context.Context, userID string, policyType string) ([]*domain.InsurancePolicy, error) {
	var models []models.InsurancePolicyModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, policyType).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get insurance policies by type: %w", err)
	}

	policies := make([]*domain.InsurancePolicy, len(models))
	for i, model := range models {
		policies[i] = model.ToDomain()
	}

	return policies, nil
}

// GetByPolicyNumber retrieves an insurance policy by policy number
func (r *insurancePolicyRepository) GetByPolicyNumber(ctx context.Context, policyNumber string) (*domain.InsurancePolicy, error) {
	var model models.InsurancePolicyModel
	
	if err := r.db.WithContext(ctx).
		Where("policy_number = ?", policyNumber).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("insurance policy with number %s not found", policyNumber)
		}
		return nil, fmt.Errorf("failed to get insurance policy by number: %w", err)
	}

	return model.ToDomain(), nil
}

// GetActivePolicies retrieves currently active insurance policies
func (r *insurancePolicyRepository) GetActivePolicies(ctx context.Context, userID string) ([]*domain.InsurancePolicy, error) {
	var models []models.InsurancePolicyModel
	now := time.Now()
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ? AND start_date <= ? AND end_date >= ?", userID, true, now, now).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get active insurance policies: %w", err)
	}

	policies := make([]*domain.InsurancePolicy, len(models))
	for i, model := range models {
		policies[i] = model.ToDomain()
	}

	return policies, nil
}

// GetByProfileID retrieves insurance policies by profile ID
func (r *insurancePolicyRepository) GetByProfileID(ctx context.Context, profileID string) ([]*domain.InsurancePolicy, error) {
	profileIDUint, err := strconv.ParseUint(profileID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	var models []models.InsurancePolicyModel
	
	if err := r.db.WithContext(ctx).
		Where("profile_id = ?", uint(profileIDUint)).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get insurance policies by profile: %w", err)
	}

	policies := make([]*domain.InsurancePolicy, len(models))
	for i, model := range models {
		policies[i] = model.ToDomain()
	}

	return policies, nil
}

// UpdateDeductibleProgress updates deductible and out-of-pocket progress
func (r *insurancePolicyRepository) UpdateDeductibleProgress(ctx context.Context, policyID string, deductibleMet, outOfPocketCurrent float64) (*domain.InsurancePolicy, error) {
	idUint, err := strconv.ParseUint(policyID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid policy ID: %w", err)
	}

	var model models.InsurancePolicyModel
	if err := r.db.WithContext(ctx).First(&model, uint(idUint)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("insurance policy with ID %s not found", policyID)
		}
		return nil, fmt.Errorf("failed to find insurance policy: %w", err)
	}

	// Update progress amounts with capping
	model.DeductibleMet = deductibleMet
	model.OutOfPocketCurrent = outOfPocketCurrent
	
	// Cap values at their maximums (this will be handled by the model's BeforeUpdate hook)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to update deductible progress: %w", err)
	}

	return model.ToDomain(), nil
}

// CalculateCoverageForExpense calculates insurance coverage for an expense
func (r *insurancePolicyRepository) CalculateCoverageForExpense(ctx context.Context, policyID string, expenseAmount float64) (*services.CoverageCalculation, error) {
	idUint, err := strconv.ParseUint(policyID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid policy ID: %w", err)
	}

	var model models.InsurancePolicyModel
	if err := r.db.WithContext(ctx).First(&model, uint(idUint)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("insurance policy with ID %s not found", policyID)
		}
		return nil, fmt.Errorf("failed to find insurance policy: %w", err)
	}

	// Convert to domain for calculation
	policy := model.ToDomain()
	
	// Calculate coverage using domain logic
	insuranceCoverage, outOfPocketAmount, newDeductibleMet := policy.CalculateCoverage(expenseAmount)
	
	// Calculate remaining amounts
	remainingDeductible := policy.GetRemainingDeductible()
	remainingOutOfPocket := policy.GetRemainingOutOfPocket()
	
	return &services.CoverageCalculation{
		InsurancePays:         insuranceCoverage,
		PatientPays:           outOfPocketAmount,
		NewDeductibleMet:      newDeductibleMet,
		NewOutOfPocketUsed:    policy.OutOfPocketCurrent + outOfPocketAmount,
		IsDeductibleMet:       newDeductibleMet >= policy.Deductible,
		IsOutOfPocketMaxMet:   policy.OutOfPocketCurrent + outOfPocketAmount >= policy.OutOfPocketMax,
		RemainingDeductible:   remainingDeductible,
		RemainingOutOfPocket:  remainingOutOfPocket,
	}, nil
}

// GetPoliciesByProvider retrieves insurance policies by provider
func (r *insurancePolicyRepository) GetPoliciesByProvider(ctx context.Context, userID string, provider string) ([]*domain.InsurancePolicy, error) {
	var models []models.InsurancePolicyModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get insurance policies by provider: %w", err)
	}

	policies := make([]*domain.InsurancePolicy, len(models))
	for i, model := range models {
		policies[i] = model.ToDomain()
	}

	return policies, nil
}