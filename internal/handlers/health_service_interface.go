package handlers

import (
	"context"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// HealthServiceInterface defines the methods that the handlers expect from the health service
// This interface is defined in the handler package to follow dependency inversion principle
type HealthServiceInterface interface {
	// Profile operations
	CreateProfile(ctx context.Context, profile domain.HealthProfile) error
	GetProfile(ctx context.Context, profileID uint) (*domain.HealthProfile, error)
	GetProfilesByUserID(ctx context.Context, userID string) ([]domain.HealthProfile, error)
	UpdateProfile(ctx context.Context, profile domain.HealthProfile) error
	DeleteProfile(ctx context.Context, profileID uint) error
	
	// Health summary and risk analysis
	GetHealthSummary(ctx context.Context, profileID uint) (*domain.HealthSummary, error)
	CalculateRisk(ctx context.Context, profileID uint) (interface{}, error)
	
	// Medical condition operations
	CreateCondition(ctx context.Context, condition domain.MedicalCondition) error
	GetCondition(ctx context.Context, conditionID uint) (*domain.MedicalCondition, error)
	GetConditionsByProfile(ctx context.Context, profileID uint) ([]domain.MedicalCondition, error)
	UpdateCondition(ctx context.Context, condition domain.MedicalCondition) error
	RemoveCondition(ctx context.Context, conditionID uint) error
	
	// Insurance policy operations
	CreatePolicy(ctx context.Context, policy domain.InsurancePolicy) error
	GetPolicyByID(ctx context.Context, policyID uint) (*domain.InsurancePolicy, error)
	GetPoliciesByProfile(ctx context.Context, profileID uint) ([]domain.InsurancePolicy, error)
	UpdatePolicy(ctx context.Context, policy domain.InsurancePolicy) error
	DeletePolicy(ctx context.Context, policyID uint) error
	
	// Medical expense operations
	CreateExpense(ctx context.Context, expense domain.MedicalExpense) error
	GetExpense(ctx context.Context, expenseID uint) (*domain.MedicalExpense, error)
	GetExpensesByProfile(ctx context.Context, profileID uint) ([]domain.MedicalExpense, error)
	UpdateExpense(ctx context.Context, expense domain.MedicalExpense) error
	DeleteExpense(ctx context.Context, expenseID uint) error
}