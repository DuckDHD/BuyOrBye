package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// HealthService is an alias to the services.HealthService interface
// This handler acts as an adapter between DTOs and domain objects
type HealthService interface {
	// Profile operations
	CreateProfile(ctx context.Context, profile *domain.HealthProfile) error
	GetProfile(ctx context.Context, userID string) (*domain.HealthProfile, error)
	UpdateProfile(ctx context.Context, profile *domain.HealthProfile) error

	// Medical conditions
	AddCondition(ctx context.Context, condition *domain.MedicalCondition) error
	GetConditions(ctx context.Context, userID string) ([]domain.MedicalCondition, error)
	UpdateCondition(ctx context.Context, condition *domain.MedicalCondition) error
	RemoveCondition(ctx context.Context, userID, conditionID string) error

	// Medical expenses
	AddExpense(ctx context.Context, expense *domain.MedicalExpense) error
	GetExpenses(ctx context.Context, userID string) ([]domain.MedicalExpense, error)
	GetRecurringExpenses(ctx context.Context, userID string) ([]domain.MedicalExpense, error)

	// Insurance policies
	AddInsurancePolicy(ctx context.Context, policy *domain.InsurancePolicy) error
	GetActivePolicies(ctx context.Context, userID string) ([]domain.InsurancePolicy, error)
	UpdateDeductibleProgress(ctx context.Context, policyID string, amount float64) error

	// Calculations & Analysis
	CalculateHealthSummary(ctx context.Context, userID string) (*domain.HealthSummary, error)
}

// HealthHandler handles health-related HTTP requests
type HealthHandler struct {
	healthService HealthService
}

// NewHealthHandler creates a new health handler instance
func NewHealthHandler(healthService HealthService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

// getUserFromContext extracts user ID from JWT context
func (h *HealthHandler) getUserFromContext(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", fmt.Errorf("user not authenticated")
	}
	
	userIDStr, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("invalid user ID format")
	}
	
	return userIDStr, nil
}

// CreateProfile creates a new health profile
func (h *HealthHandler) CreateProfile(c *gin.Context) {
	var requestDTO types.CreateHealthProfileRequestDTO
	
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	// Get user from JWT context for authorization
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	// User ID comes from authenticated context, not from request body

	// Convert DTO to domain object
	profile := requestDTO.ToDomain(userID)

	// Create profile
	ctx := context.Background()
	if err := h.healthService.CreateProfile(ctx, profile); err != nil {
		if strings.Contains(err.Error(), "already has a health profile") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"message": "Profile created successfully"})
}

// GetProfile retrieves the user's health profile
func (h *HealthHandler) GetProfile(c *gin.Context) {
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	ctx := context.Background()
	responseDTO, err := h.healthService.GetProfile(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Health profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, responseDTO)
}

// UpdateProfile updates the user's health profile
func (h *HealthHandler) UpdateProfile(c *gin.Context) {
	var requestDTO types.UpdateHealthProfileRequestDTO
	
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	ctx := context.Background()

	// Get existing profile
	existingProfile, err := h.healthService.GetProfile(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Health profile not found"})
		return
	}

	// Apply updates to existing profile
	requestDTO.ApplyUpdates(existingProfile)

	// Update profile
	if err := h.healthService.UpdateProfile(ctx, existingProfile); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Health profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// AddCondition adds a new medical condition
func (h *HealthHandler) AddCondition(c *gin.Context) {
	var requestDTO types.CreateMedicalConditionRequestDTO
	
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	// User ID comes from authenticated context, not from request body

	// First, get user's profile to obtain profileID
	profile, err := h.healthService.GetProfile(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Profile not found. Please create health profile first."})
		return
	}

	// Convert DTO to domain object
	condition, err := requestDTO.ToDomain(userID, profile.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format: " + err.Error()})
		return
	}

	ctx := context.Background()
	if err := h.healthService.AddCondition(ctx, condition); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add condition: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"message": "Condition added successfully"})
}

// GetConditions retrieves all medical conditions for the user
func (h *HealthHandler) GetConditions(c *gin.Context) {
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	ctx := context.Background()
	response, err := h.healthService.GetConditions(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conditions: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateCondition updates a medical condition
func (h *HealthHandler) UpdateCondition(c *gin.Context) {
	conditionID := c.Param("id")
	if conditionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Condition ID is required"})
		return
	}
	
	var requestDTO types.UpdateMedicalConditionRequestDTO
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	ctx := context.Background()

	// Get existing conditions to find the one to update
	conditions, err := h.healthService.GetConditions(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve conditions"})
		return
	}

	// Find the condition to update
	var existingCondition *domain.MedicalCondition
	for i, condition := range conditions {
		if condition.ID == conditionID {
			existingCondition = &conditions[i]
			break
		}
	}

	if existingCondition == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Condition not found"})
		return
	}

	// Apply updates to existing condition
	requestDTO.ApplyUpdates(existingCondition)

	if err := h.healthService.UpdateCondition(ctx, existingCondition); err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update condition: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Condition updated successfully"})
}

// RemoveCondition removes a medical condition
func (h *HealthHandler) RemoveCondition(c *gin.Context) {
	conditionID := c.Param("id")
	if conditionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Condition ID is required"})
		return
	}
	
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	ctx := context.Background()
	if err := h.healthService.RemoveCondition(ctx, userID, conditionID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Condition not found"})
			return
		}
		if strings.Contains(err.Error(), "not authorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove condition: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Condition removed successfully"})
}

// AddExpense adds a new medical expense
func (h *HealthHandler) AddExpense(c *gin.Context) {
	var requestDTO types.CreateMedicalExpenseRequestDTO
	
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	// User ID comes from authenticated context, not from request body

	// First, get user's profile to obtain profileID
	profile, err := h.healthService.GetProfile(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Profile not found. Please create health profile first."})
		return
	}

	// Convert DTO to domain object
	expense, err := requestDTO.ToDomain(userID, profile.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format: " + err.Error()})
		return
	}

	ctx := context.Background()
	if err := h.healthService.AddExpense(ctx, expense); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add expense: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"message": "Expense added successfully"})
}

// GetExpenses retrieves all medical expenses for the user
func (h *HealthHandler) GetExpenses(c *gin.Context) {
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	ctx := context.Background()
	response, err := h.healthService.GetExpenses(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get expenses: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetRecurringExpenses retrieves recurring medical expenses for the user
func (h *HealthHandler) GetRecurringExpenses(c *gin.Context) {
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	ctx := context.Background()
	response, err := h.healthService.GetRecurringExpenses(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recurring expenses: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AddInsurancePolicy adds a new insurance policy
func (h *HealthHandler) AddInsurancePolicy(c *gin.Context) {
	var requestDTO types.CreateInsurancePolicyRequestDTO
	
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	// User ID comes from authenticated context, not from request body

	// First, get user's profile to obtain profileID
	profile, err := h.healthService.GetProfile(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Profile not found. Please create health profile first."})
		return
	}

	// Convert DTO to domain object
	policy, err := requestDTO.ToDomain(userID, profile.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format: " + err.Error()})
		return
	}

	ctx := context.Background()
	if err := h.healthService.AddInsurancePolicy(ctx, policy); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add policy: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"message": "Insurance policy added successfully"})
}

// GetActivePolicies retrieves active insurance policies for the user
func (h *HealthHandler) GetActivePolicies(c *gin.Context) {
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	ctx := context.Background()
	response, err := h.healthService.GetActivePolicies(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get policies: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateDeductibleProgress updates deductible progress for a policy
func (h *HealthHandler) UpdateDeductibleProgress(c *gin.Context) {
	policyID := c.Param("id")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy ID is required"})
		return
	}
	
	var requestDTO types.UpdateDeductibleRequestDTO
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	_, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	ctx := context.Background()
	if err := h.healthService.UpdateDeductibleProgress(ctx, policyID, requestDTO.Amount); err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Policy not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deductible progress: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Deductible progress updated successfully"})
}

// GetHealthSummary calculates and returns a comprehensive health summary
func (h *HealthHandler) GetHealthSummary(c *gin.Context) {
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	ctx := context.Background()
	responseDTO, err := h.healthService.CalculateHealthSummary(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Health profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate health summary: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, responseDTO)
}

// DeleteProfile deletes the user's health profile and all related data
func (h *HealthHandler) DeleteProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete profile not yet implemented"})
}

// CalculateRisk calculates and returns the user's health risk score  
func (h *HealthHandler) CalculateRisk(c *gin.Context) {
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	ctx := context.Background()
	summary, err := h.healthService.CalculateHealthSummary(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Health profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate risk: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"risk_score": summary.HealthRiskScore,
		"risk_level": summary.HealthRiskLevel,
	})
}

// Placeholder methods for missing handlers
func (h *HealthHandler) CreateCondition(c *gin.Context) {
	h.AddCondition(c) // Delegate to existing method
}

func (h *HealthHandler) GetCondition(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get single condition not yet implemented"})
}

func (h *HealthHandler) GetConditionsByProfile(c *gin.Context) {
	h.GetConditions(c) // Delegate to existing method  
}

func (h *HealthHandler) CreatePolicy(c *gin.Context) {
	h.AddInsurancePolicy(c) // Delegate to existing method
}

func (h *HealthHandler) GetPolicy(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get single policy not yet implemented"})
}

func (h *HealthHandler) UpdatePolicy(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Update policy not yet implemented"})
}

func (h *HealthHandler) DeletePolicy(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete policy not yet implemented"})
}

func (h *HealthHandler) GetPoliciesByProfile(c *gin.Context) {
	h.GetActivePolicies(c) // Delegate to existing method
}

func (h *HealthHandler) CreateExpense(c *gin.Context) {
	h.AddExpense(c) // Delegate to existing method
}

func (h *HealthHandler) GetExpense(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get single expense not yet implemented"})
}

func (h *HealthHandler) UpdateExpense(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Update expense not yet implemented"})
}

func (h *HealthHandler) DeleteExpense(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete expense not yet implemented"})
}

func (h *HealthHandler) GetExpensesByProfile(c *gin.Context) {
	h.GetExpenses(c) // Delegate to existing method
}