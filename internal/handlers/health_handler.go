package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/dtos"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// HealthHandler handles health-related HTTP requests
type HealthHandler struct {
	healthService services.HealthService
}

// NewHealthHandler creates a new health handler instance
func NewHealthHandler(healthService services.HealthService) *HealthHandler {
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
	var requestDTO dtos.CreateHealthProfileRequestDTO
	
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
	
	// Ensure user can only create profile for themselves
	if requestDTO.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot create profile for another user"})
		return
	}
	
	// Convert DTO to domain
	profile := requestDTO.ToDomain()
	
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
	profile, err := h.healthService.GetProfile(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Health profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile: " + err.Error()})
		return
	}
	
	// Convert domain to DTO
	var responseDTO dtos.HealthProfileResponseDTO
	responseDTO.FromDomain(profile)
	
	c.JSON(http.StatusOK, responseDTO)
}

// UpdateProfile updates the user's health profile
func (h *HealthHandler) UpdateProfile(c *gin.Context) {
	var requestDTO dtos.UpdateHealthProfileRequestDTO
	
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
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Health profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile: " + err.Error()})
		return
	}
	
	// Update fields that were provided
	if requestDTO.Age != 0 {
		existingProfile.Age = requestDTO.Age
	}
	if requestDTO.Gender != "" {
		existingProfile.Gender = requestDTO.Gender
	}
	if requestDTO.Height != 0 {
		existingProfile.Height = requestDTO.Height
	}
	if requestDTO.Weight != 0 {
		existingProfile.Weight = requestDTO.Weight
	}
	if requestDTO.FamilySize != 0 {
		existingProfile.FamilySize = requestDTO.FamilySize
	}
	existingProfile.HasChronicConditions = requestDTO.HasChronicConditions
	if requestDTO.EmergencyFundHealth >= 0 {
		existingProfile.EmergencyFundHealth = requestDTO.EmergencyFundHealth
	}
	
	// Update profile
	if err := h.healthService.UpdateProfile(ctx, existingProfile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// AddCondition adds a new medical condition
func (h *HealthHandler) AddCondition(c *gin.Context) {
	var requestDTO dtos.CreateMedicalConditionRequestDTO
	
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	// Ensure user can only add condition for themselves
	if requestDTO.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot add condition for another user"})
		return
	}
	
	// Convert DTO to domain
	condition := requestDTO.ToDomain()
	
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
	conditions, err := h.healthService.GetConditions(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conditions: " + err.Error()})
		return
	}
	
	// Convert domain to DTOs
	conditionDTOs := make([]dtos.MedicalConditionResponseDTO, len(conditions))
	for i, condition := range conditions {
		conditionDTOs[i].FromDomain(&condition)
	}
	
	response := dtos.MedicalConditionListResponseDTO{
		Conditions: conditionDTOs,
		Total:      len(conditionDTOs),
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
	
	var requestDTO dtos.UpdateMedicalConditionRequestDTO
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	// For now, we'll create a basic condition with the ID and user for update
	// In a real implementation, you'd first retrieve the existing condition
	condition := &domain.MedicalCondition{
		ID:     conditionID,
		UserID: userID,
	}
	
	// Update fields that were provided
	if requestDTO.Name != "" {
		condition.Name = requestDTO.Name
	}
	if requestDTO.Category != "" {
		condition.Category = requestDTO.Category
	}
	if requestDTO.Severity != "" {
		condition.Severity = requestDTO.Severity
	}
	condition.RequiresMedication = requestDTO.RequiresMedication
	if requestDTO.MonthlyMedCost >= 0 {
		condition.MonthlyMedCost = requestDTO.MonthlyMedCost
	}
	if requestDTO.RiskFactor >= 0 && requestDTO.RiskFactor <= 1 {
		condition.RiskFactor = requestDTO.RiskFactor
	}
	condition.IsActive = requestDTO.IsActive
	
	ctx := context.Background()
	if err := h.healthService.UpdateCondition(ctx, condition); err != nil {
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
	var requestDTO dtos.CreateMedicalExpenseRequestDTO
	
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	// Ensure user can only add expense for themselves
	if requestDTO.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot add expense for another user"})
		return
	}
	
	// Convert DTO to domain
	expense := requestDTO.ToDomain()
	
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
	expenses, err := h.healthService.GetExpenses(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get expenses: " + err.Error()})
		return
	}
	
	// Convert domain to DTOs
	expenseDTOs := make([]dtos.MedicalExpenseResponseDTO, len(expenses))
	for i, expense := range expenses {
		expenseDTOs[i].FromDomain(&expense)
	}
	
	response := dtos.MedicalExpenseListResponseDTO{
		Expenses: expenseDTOs,
		Total:    len(expenseDTOs),
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
	expenses, err := h.healthService.GetRecurringExpenses(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recurring expenses: " + err.Error()})
		return
	}
	
	// Convert domain to DTOs
	expenseDTOs := make([]dtos.MedicalExpenseResponseDTO, len(expenses))
	for i, expense := range expenses {
		expenseDTOs[i].FromDomain(&expense)
	}
	
	response := dtos.MedicalExpenseListResponseDTO{
		Expenses: expenseDTOs,
		Total:    len(expenseDTOs),
	}
	
	c.JSON(http.StatusOK, response)
}

// AddInsurancePolicy adds a new insurance policy
func (h *HealthHandler) AddInsurancePolicy(c *gin.Context) {
	var requestDTO dtos.CreateInsurancePolicyRequestDTO
	
	if err := c.ShouldBindJSON(&requestDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}
	
	userID, err := h.getUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	
	// Ensure user can only add policy for themselves
	if requestDTO.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot add policy for another user"})
		return
	}
	
	// Convert DTO to domain
	policy := requestDTO.ToDomain()
	
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
	policies, err := h.healthService.GetActivePolicies(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get policies: " + err.Error()})
		return
	}
	
	// Convert domain to DTOs
	policyDTOs := make([]dtos.InsurancePolicyResponseDTO, len(policies))
	for i, policy := range policies {
		policyDTOs[i].FromDomain(&policy)
	}
	
	response := dtos.InsurancePolicyListResponseDTO{
		Policies: policyDTOs,
		Total:    len(policyDTOs),
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
	
	var requestDTO dtos.UpdateDeductibleRequestDTO
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
	summary, err := h.healthService.CalculateHealthSummary(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Health profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate health summary: " + err.Error()})
		return
	}
	
	// Convert domain to DTO
	var responseDTO dtos.HealthSummaryResponseDTO
	responseDTO.FromDomain(summary)
	
	c.JSON(http.StatusOK, responseDTO)
}