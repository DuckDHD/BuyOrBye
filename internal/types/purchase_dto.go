package types

import (
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// Purchase DTOs

/*
Request PurchaseCreateDTO dto
Request to create a new purchase with item details and categorization
*/
type PurchaseCreateDTO struct {
	UserID      string  `json:"-"` // Set by handler, not from request body
	Name        string  `json:"name" binding:"required,min=2,max=100" example:"iPhone 15 Pro"`
	Amount      float64 `json:"amount" binding:"required,gt=0,max=1000000" example:"1199.00"`
	Category    string  `json:"category" binding:"required,oneof=electronics clothing food transport health entertainment home other" example:"electronics"`
	Description string  `json:"description" binding:"max=500" example:"Latest iPhone with advanced camera features"`
	Priority    string  `json:"priority" binding:"required,oneof=low medium high critical" example:"medium"`
	Urgency     string  `json:"urgency" binding:"required,oneof=low medium high critical" example:"low"`
	Frequency   string  `json:"frequency" binding:"required,oneof=one_time monthly yearly" example:"one_time"`
	Store       string  `json:"store" binding:"max=100" example:"Apple Store"`
	Brand       string  `json:"brand" binding:"max=50" example:"Apple"`
	Model       string  `json:"model" binding:"max=100" example:"iPhone 15 Pro 256GB"`
}

/*
Request PurchaseUpdateDTO dto
Request to update an existing purchase with optional fields
*/
type PurchaseUpdateDTO struct {
	Name        *string  `json:"name,omitempty" binding:"omitempty,min=2,max=100" example:"iPhone 15 Pro Max"`
	Amount      *float64 `json:"amount,omitempty" binding:"omitempty,gt=0,max=1000000" example:"1299.00"`
	Category    *string  `json:"category,omitempty" binding:"omitempty,oneof=electronics clothing food transport health entertainment home other" example:"electronics"`
	Description *string  `json:"description,omitempty" binding:"omitempty,max=500" example:"Upgraded model with larger screen"`
	Priority    *string  `json:"priority,omitempty" binding:"omitempty,oneof=low medium high critical" example:"high"`
	Urgency     *string  `json:"urgency,omitempty" binding:"omitempty,oneof=low medium high critical" example:"medium"`
	Frequency   *string  `json:"frequency,omitempty" binding:"omitempty,oneof=one_time monthly yearly" example:"one_time"`
	Store       *string  `json:"store,omitempty" binding:"omitempty,max=100" example:"Best Buy"`
	Brand       *string  `json:"brand,omitempty" binding:"omitempty,max=50" example:"Apple"`
	Model       *string  `json:"model,omitempty" binding:"omitempty,max=100" example:"iPhone 15 Pro Max 512GB"`
	Status      *string  `json:"status,omitempty" binding:"omitempty,oneof=pending approved rejected completed cancelled" example:"approved"`
}

/*
Response PurchaseItemDTO dto
Purchase item details for list views and summary displays
*/
type PurchaseItemDTO struct {
	ID          string    `json:"id" example:"purchase-123"`
	Name        string    `json:"name" example:"iPhone 15 Pro"`
	Amount      float64   `json:"amount" example:"1199.00"`
	Category    string    `json:"category" example:"electronics"`
	Priority    string    `json:"priority" example:"medium"`
	Urgency     string    `json:"urgency" example:"low"`
	Status      string    `json:"status" example:"pending"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	DecisionID  string    `json:"decision_id,omitempty" example:"decision-456"`
	Decision    string    `json:"decision,omitempty" example:"approved"`
	Confidence  float64   `json:"confidence,omitempty" example:"0.85"`
}

/*
Response PurchaseDetailDTO dto
Complete purchase details for individual purchase views
*/
type PurchaseDetailDTO struct {
	ID          string    `json:"id" example:"purchase-123"`
	UserID      string    `json:"user_id" example:"user-456"`
	Name        string    `json:"name" example:"iPhone 15 Pro"`
	Amount      float64   `json:"amount" example:"1199.00"`
	Category    string    `json:"category" example:"electronics"`
	Description string    `json:"description" example:"Latest iPhone with advanced camera features"`
	Priority    string    `json:"priority" example:"medium"`
	Urgency     string    `json:"urgency" example:"low"`
	Frequency   string    `json:"frequency" example:"one_time"`
	Store       string    `json:"store" example:"Apple Store"`
	Brand       string    `json:"brand" example:"Apple"`
	Model       string    `json:"model" example:"iPhone 15 Pro 256GB"`
	Status      string    `json:"status" example:"pending"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
	DecisionID  string    `json:"decision_id,omitempty" example:"decision-456"`
	Decision    string    `json:"decision,omitempty" example:"approved"`
	Confidence  float64   `json:"confidence,omitempty" example:"0.85"`
	Reasoning   string    `json:"reasoning,omitempty" example:"Good financial fit with current budget"`
}

/*
Response PurchaseListDTO dto
List of purchases with pagination and filtering information
*/
type PurchaseListDTO struct {
	Purchases    []PurchaseItemDTO `json:"purchases"`
	Total        int               `json:"total"`
	Page         int               `json:"page"`
	PerPage      int               `json:"per_page"`
	TotalPages   int               `json:"total_pages"`
	HasNext      bool              `json:"has_next"`
	HasPrevious  bool              `json:"has_previous"`
	FilterBy     string            `json:"filter_by,omitempty"`
	FilterValue  string            `json:"filter_value,omitempty"`
	SortBy       string            `json:"sort_by,omitempty"`
	SortOrder    string            `json:"sort_order,omitempty"`
	Error        *ErrorResponseDTO `json:"error,omitempty"`
}

/*
Response PurchaseStatsDTO dto
Purchase statistics and analytics for dashboard and reports
*/
type PurchaseStatsDTO struct {
	UserID                string    `json:"user_id" example:"user-456"`
	TotalPurchases        int       `json:"total_purchases" example:"25"`
	TotalAmount           float64   `json:"total_amount" example:"15750.00"`
	AverageAmount         float64   `json:"average_amount" example:"630.00"`
	MonthlySpending       float64   `json:"monthly_spending" example:"2500.00"`
	TopCategory           string    `json:"top_category" example:"electronics"`
	TopCategoryAmount     float64   `json:"top_category_amount" example:"5200.00"`
	PendingPurchases      int       `json:"pending_purchases" example:"3"`
	ApprovedPurchases     int       `json:"approved_purchases" example:"18"`
	RejectedPurchases     int       `json:"rejected_purchases" example:"4"`
	BudgetUtilization     float64   `json:"budget_utilization" example:"0.65"`
	AverageConfidence     float64   `json:"average_confidence" example:"0.78"`
	LastPurchaseDate      time.Time `json:"last_purchase_date" example:"2024-01-15T10:30:00Z"`
	UpdatedAt             time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

/*
Response PurchaseCategoryStatsDTO dto
Purchase statistics grouped by category
*/
type PurchaseCategoryStatsDTO struct {
	Category      string  `json:"category" example:"electronics"`
	Count         int     `json:"count" example:"8"`
	TotalAmount   float64 `json:"total_amount" example:"5200.00"`
	AverageAmount float64 `json:"average_amount" example:"650.00"`
	Percentage    float64 `json:"percentage" example:"0.33"`
}

/*
Filter PurchaseFilterDTO dto
Purchase filtering and search parameters
*/
type PurchaseFilterDTO struct {
	UserID      string    `json:"-"` // Set by handler
	Category    string    `json:"category,omitempty" binding:"omitempty,oneof=electronics clothing food transport health entertainment home other"`
	Status      string    `json:"status,omitempty" binding:"omitempty,oneof=pending approved rejected completed cancelled"`
	Priority    string    `json:"priority,omitempty" binding:"omitempty,oneof=low medium high critical"`
	Urgency     string    `json:"urgency,omitempty" binding:"omitempty,oneof=low medium high critical"`
	MinAmount   *float64  `json:"min_amount,omitempty" binding:"omitempty,gte=0"`
	MaxAmount   *float64  `json:"max_amount,omitempty" binding:"omitempty,gte=0"`
	DateFrom    *time.Time `json:"date_from,omitempty"`
	DateTo      *time.Time `json:"date_to,omitempty"`
	Search      string    `json:"search,omitempty" binding:"omitempty,max=100"`
	Page        int       `json:"page,omitempty" binding:"omitempty,min=1" example:"1"`
	PerPage     int       `json:"per_page,omitempty" binding:"omitempty,min=1,max=100" example:"20"`
	SortBy      string    `json:"sort_by,omitempty" binding:"omitempty,oneof=name amount created_at updated_at priority urgency"`
	SortOrder   string    `json:"sort_order,omitempty" binding:"omitempty,oneof=asc desc"`
}

// Conversion Methods - DTO to Domain

// ToDomain converts PurchaseCreateDTO to domain.PurchaseIntent
func (dto PurchaseCreateDTO) ToDomain() *domain.PurchaseIntent {
	return &domain.PurchaseIntent{
		UserID:      dto.UserID,
		ItemName:    dto.Name,
		ItemCost:    dto.Amount,
		Category:    dto.Category,
		Urgency:     dto.Urgency,
		Frequency:   dto.Frequency,
		Purpose:     dto.Description,
		Alternative: "", // This would come from the Alternative field if added
		CreatedAt:   time.Now(),
	}
}

// Conversion Methods - Domain to DTO

// FromPurchaseIntent converts domain.PurchaseIntent to PurchaseDetailDTO
func FromPurchaseDetail(intent *domain.PurchaseIntent) PurchaseDetailDTO {
	return PurchaseDetailDTO{
		ID:          intent.ID,
		UserID:      intent.UserID,
		Name:        intent.ItemName,
		Amount:      intent.ItemCost,
		Category:    intent.Category,
		Description: intent.Purpose,
		Priority:    "", // Not available in PurchaseIntent, would need to be calculated
		Urgency:     intent.Urgency,
		Frequency:   intent.Frequency,
		Store:       "", // Not available in PurchaseIntent
		Brand:       "", // Not available in PurchaseIntent
		Model:       "", // Not available in PurchaseIntent
		Status:      "", // Not available in PurchaseIntent
		CreatedAt:   intent.CreatedAt,
		UpdatedAt:   intent.CreatedAt, // PurchaseIntent doesn't have UpdatedAt
		DecisionID:  "", // Would come from related decision record
		Decision:    "", // Would come from related decision record
		Confidence:  0,  // Would come from related decision record
		Reasoning:   "", // Would come from related decision record
	}
}

// FromPurchaseIntent converts domain.PurchaseIntent to PurchaseItemDTO
func FromPurchaseItem(intent *domain.PurchaseIntent) PurchaseItemDTO {
	return PurchaseItemDTO{
		ID:         intent.ID,
		Name:       intent.ItemName,
		Amount:     intent.ItemCost,
		Category:   intent.Category,
		Priority:   "", // Would need to be calculated from urgency/category
		Urgency:    intent.Urgency,
		Status:     "", // Not available in PurchaseIntent
		CreatedAt:  intent.CreatedAt,
		DecisionID: "", // Would come from related decision record
		Decision:   "", // Would come from related decision record
		Confidence: 0,  // Would come from related decision record
	}
}

// FromPurchaseIntentList converts slice of domain.PurchaseIntent to slice of PurchaseItemDTO
func FromPurchaseList(intents []*domain.PurchaseIntent) []PurchaseItemDTO {
	items := make([]PurchaseItemDTO, len(intents))
	for i, intent := range intents {
		items[i] = FromPurchaseItem(intent)
	}
	return items
}

// Update Methods - Apply Updates to Domain Structs

// ApplyUpdates applies PurchaseUpdateDTO fields to domain.PurchaseIntent
func (dto PurchaseUpdateDTO) ApplyUpdates(intent *domain.PurchaseIntent) {
	if dto.Name != nil {
		intent.ItemName = *dto.Name
	}
	if dto.Amount != nil {
		intent.ItemCost = *dto.Amount
	}
	if dto.Category != nil {
		intent.Category = *dto.Category
	}
	if dto.Description != nil {
		intent.Purpose = *dto.Description
	}
	if dto.Urgency != nil {
		intent.Urgency = *dto.Urgency
	}
	if dto.Frequency != nil {
		intent.Frequency = *dto.Frequency
	}
	// Note: PurchaseIntent doesn't have fields for Store, Brand, Model, Status, Priority
	// These would need to be handled separately or the domain model would need to be extended
}

// Validation and Helper Methods

// ValidateFilter validates and sets defaults for PurchaseFilterDTO
func (dto *PurchaseFilterDTO) ValidateAndSetDefaults() {
	if dto.Page == 0 {
		dto.Page = 1
	}
	if dto.PerPage == 0 {
		dto.PerPage = 20
	}
	if dto.PerPage > 100 {
		dto.PerPage = 100
	}
	if dto.SortBy == "" {
		dto.SortBy = "created_at"
	}
	if dto.SortOrder == "" {
		dto.SortOrder = "desc"
	}

	// Validate amount range
	if dto.MinAmount != nil && dto.MaxAmount != nil && *dto.MinAmount > *dto.MaxAmount {
		// Swap values if min > max
		dto.MinAmount, dto.MaxAmount = dto.MaxAmount, dto.MinAmount
	}

	// Validate date range
	if dto.DateFrom != nil && dto.DateTo != nil && dto.DateFrom.After(*dto.DateTo) {
		// Swap values if from > to
		dto.DateFrom, dto.DateTo = dto.DateTo, dto.DateFrom
	}
}

// CalculatePagination calculates pagination info for PurchaseListDTO
func CalculatePurchasePagination(total, page, perPage int) (totalPages int, hasNext, hasPrevious bool) {
	if perPage == 0 {
		perPage = 20
	}
	if page == 0 {
		page = 1
	}

	totalPages = (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	hasNext = page < totalPages
	hasPrevious = page > 1

	return totalPages, hasNext, hasPrevious
}

// CreatePurchaseListResponse creates a PurchaseListDTO with pagination
func CreatePurchaseListResponse(intents []*domain.PurchaseIntent, total, page, perPage int, filter *PurchaseFilterDTO) *PurchaseListDTO {
	items := FromPurchaseList(intents)
	totalPages, hasNext, hasPrevious := CalculatePurchasePagination(total, page, perPage)

	response := &PurchaseListDTO{
		Purchases:   items,
		Total:       total,
		Page:        page,
		PerPage:     perPage,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}

	// Add filter information if provided
	if filter != nil {
		response.SortBy = filter.SortBy
		response.SortOrder = filter.SortOrder

		// Set filter information based on applied filters
		if filter.Category != "" {
			response.FilterBy = "category"
			response.FilterValue = filter.Category
		} else if filter.Status != "" {
			response.FilterBy = "status"
			response.FilterValue = filter.Status
		} else if filter.Priority != "" {
			response.FilterBy = "priority"
			response.FilterValue = filter.Priority
		} else if filter.Search != "" {
			response.FilterBy = "search"
			response.FilterValue = filter.Search
		}
	}

	return response
}

// Additional FromDomain functions for completeness

// FromDomainPurchaseIntentToCreate creates PurchaseCreateDTO from domain.PurchaseIntent
func FromDomainPurchaseIntentToCreate(intent *domain.PurchaseIntent) PurchaseCreateDTO {
	return PurchaseCreateDTO{
		UserID:      intent.UserID,
		Name:        intent.ItemName,
		Amount:      intent.ItemCost,
		Category:    intent.Category,
		Description: intent.Purpose,
		Priority:    "", // Not available in PurchaseIntent
		Urgency:     intent.Urgency,
		Frequency:   intent.Frequency,
		Store:       "", // Not available in PurchaseIntent
		Brand:       "", // Not available in PurchaseIntent
		Model:       "", // Not available in PurchaseIntent
	}
}

// FromDomainPurchaseIntentListToCreate creates slice of PurchaseCreateDTO from slice of domain.PurchaseIntent
func FromDomainPurchaseIntentListToCreate(intents []*domain.PurchaseIntent) []PurchaseCreateDTO {
	dtos := make([]PurchaseCreateDTO, len(intents))
	for i, intent := range intents {
		dtos[i] = FromDomainPurchaseIntentToCreate(intent)
	}
	return dtos
}

// FromDomainPurchaseDetailList creates slice of PurchaseDetailDTO from slice of domain.PurchaseIntent
func FromDomainPurchaseDetailList(intents []*domain.PurchaseIntent) []PurchaseDetailDTO {
	dtos := make([]PurchaseDetailDTO, len(intents))
	for i, intent := range intents {
		dtos[i] = FromPurchaseDetail(intent)
	}
	return dtos
}

// ToDomainPurchaseIntent converts PurchaseCreateDTO to domain.PurchaseIntent (alias for clarity)
func (dto PurchaseCreateDTO) ToDomainPurchaseIntent() *domain.PurchaseIntent {
	return dto.ToDomain()
}

// Validation methods

// Validate validates PurchaseCreateDTO fields
func (dto PurchaseCreateDTO) Validate() error {
	if dto.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(dto.Name) < 2 || len(dto.Name) > 100 {
		return fmt.Errorf("name must be between 2 and 100 characters")
	}
	if dto.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if dto.Amount > 1000000 {
		return fmt.Errorf("amount must not exceed 1,000,000")
	}
	if dto.Category == "" {
		return fmt.Errorf("category is required")
	}

	validCategories := map[string]bool{
		"electronics": true, "clothing": true, "food": true, "transport": true,
		"health": true, "entertainment": true, "home": true, "other": true,
	}
	if !validCategories[dto.Category] {
		return fmt.Errorf("invalid category: %s", dto.Category)
	}

	if dto.Priority != "" {
		validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
		if !validPriorities[dto.Priority] {
			return fmt.Errorf("invalid priority: %s", dto.Priority)
		}
	}

	if dto.Urgency == "" {
		return fmt.Errorf("urgency is required")
	}
	validUrgencies := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
	if !validUrgencies[dto.Urgency] {
		return fmt.Errorf("invalid urgency: %s", dto.Urgency)
	}

	if dto.Frequency == "" {
		return fmt.Errorf("frequency is required")
	}
	validFrequencies := map[string]bool{"one_time": true, "monthly": true, "yearly": true}
	if !validFrequencies[dto.Frequency] {
		return fmt.Errorf("invalid frequency: %s", dto.Frequency)
	}

	if len(dto.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}
	if len(dto.Store) > 100 {
		return fmt.Errorf("store must not exceed 100 characters")
	}
	if len(dto.Brand) > 50 {
		return fmt.Errorf("brand must not exceed 50 characters")
	}
	if len(dto.Model) > 100 {
		return fmt.Errorf("model must not exceed 100 characters")
	}

	return nil
}

// Validate validates PurchaseUpdateDTO fields
func (dto PurchaseUpdateDTO) Validate() error {
	if dto.Name != nil {
		if len(*dto.Name) < 2 || len(*dto.Name) > 100 {
			return fmt.Errorf("name must be between 2 and 100 characters")
		}
	}
	if dto.Amount != nil {
		if *dto.Amount <= 0 {
			return fmt.Errorf("amount must be greater than 0")
		}
		if *dto.Amount > 1000000 {
			return fmt.Errorf("amount must not exceed 1,000,000")
		}
	}
	if dto.Category != nil {
		validCategories := map[string]bool{
			"electronics": true, "clothing": true, "food": true, "transport": true,
			"health": true, "entertainment": true, "home": true, "other": true,
		}
		if !validCategories[*dto.Category] {
			return fmt.Errorf("invalid category: %s", *dto.Category)
		}
	}
	if dto.Priority != nil {
		validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
		if !validPriorities[*dto.Priority] {
			return fmt.Errorf("invalid priority: %s", *dto.Priority)
		}
	}
	if dto.Urgency != nil {
		validUrgencies := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
		if !validUrgencies[*dto.Urgency] {
			return fmt.Errorf("invalid urgency: %s", *dto.Urgency)
		}
	}
	if dto.Frequency != nil {
		validFrequencies := map[string]bool{"one_time": true, "monthly": true, "yearly": true}
		if !validFrequencies[*dto.Frequency] {
			return fmt.Errorf("invalid frequency: %s", *dto.Frequency)
		}
	}
	if dto.Status != nil {
		validStatuses := map[string]bool{
			"pending": true, "approved": true, "rejected": true, "completed": true, "cancelled": true,
		}
		if !validStatuses[*dto.Status] {
			return fmt.Errorf("invalid status: %s", *dto.Status)
		}
	}
	if dto.Description != nil && len(*dto.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}
	if dto.Store != nil && len(*dto.Store) > 100 {
		return fmt.Errorf("store must not exceed 100 characters")
	}
	if dto.Brand != nil && len(*dto.Brand) > 50 {
		return fmt.Errorf("brand must not exceed 50 characters")
	}
	if dto.Model != nil && len(*dto.Model) > 100 {
		return fmt.Errorf("model must not exceed 100 characters")
	}

	return nil
}