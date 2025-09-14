package domain

import (
	"fmt"
	"strings"
	"time"
)

// PurchaseIntent represents a user's intention to make a purchase
type PurchaseIntent struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ItemName    string    `json:"item_name"`
	ItemCost    float64   `json:"item_cost"`
	Category    string    `json:"category"`    // "electronics", "clothing", "food", "transport", "health", "entertainment", "other"
	Urgency     string    `json:"urgency"`     // "low", "medium", "high", "critical"
	Frequency   string    `json:"frequency"`   // "one_time", "monthly", "yearly"
	Purpose     string    `json:"purpose"`     // User's reason for purchase
	Alternative string    `json:"alternative"` // Cheaper alternative considered
	CreatedAt   time.Time `json:"created_at"`
}

// Valid categories for purchases
var validCategories = map[string]bool{
	"electronics":    true,
	"clothing":       true,
	"food":          true,
	"transport":     true,
	"health":        true,
	"entertainment": true,
	"other":         true,
}

// Valid urgency levels
var validUrgencies = map[string]bool{
	"low":      true,
	"medium":   true,
	"high":     true,
	"critical": true,
}

// Valid frequency options
var validFrequencies = map[string]bool{
	"one_time": true,
	"monthly":  true,
	"yearly":   true,
}

// Essential categories that take priority in decision making
var essentialCategories = map[string]bool{
	"health":    true,
	"food":      true,
	"transport": true,
}

// Validate performs comprehensive validation of the purchase intent
func (pi *PurchaseIntent) Validate() error {
	// Validate required fields
	if pi.ID == "" {
		return fmt.Errorf("ID is required")
	}

	if pi.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Validate item name
	if pi.ItemName == "" {
		return fmt.Errorf("item name is required")
	}

	if len(pi.ItemName) < 2 {
		return fmt.Errorf("item name must be at least 2 characters")
	}

	if len(pi.ItemName) > 200 {
		return fmt.Errorf("item name must not exceed 200 characters")
	}

	// Validate item cost
	if pi.ItemCost <= 0 {
		return fmt.Errorf("item cost must be greater than 0")
	}

	if pi.ItemCost > 1000000 {
		return fmt.Errorf("item cost must not exceed 1,000,000")
	}

	// Validate category
	if !validCategories[pi.Category] {
		return fmt.Errorf("invalid category: %s. Must be one of: electronics, clothing, food, transport, health, entertainment, other", pi.Category)
	}

	// Validate urgency
	if !validUrgencies[pi.Urgency] {
		return fmt.Errorf("invalid urgency: %s. Must be one of: low, medium, high, critical", pi.Urgency)
	}

	// Validate frequency
	if !validFrequencies[pi.Frequency] {
		return fmt.Errorf("invalid frequency: %s. Must be one of: one_time, monthly, yearly", pi.Frequency)
	}

	// Validate optional fields length
	if len(pi.Purpose) > 500 {
		return fmt.Errorf("purpose must not exceed 500 characters")
	}

	if len(pi.Alternative) > 200 {
		return fmt.Errorf("alternative must not exceed 200 characters")
	}

	return nil
}

// GetMonthlyImpact calculates the monthly financial impact of the purchase
func (pi *PurchaseIntent) GetMonthlyImpact() float64 {
	switch pi.Frequency {
	case "monthly":
		return pi.ItemCost
	case "yearly":
		return pi.ItemCost / 12.0
	case "one_time":
		return pi.ItemCost
	default:
		return pi.ItemCost
	}
}

// IsEssential returns true if the purchase is in an essential category
func (pi *PurchaseIntent) IsEssential() bool {
	return essentialCategories[pi.Category]
}

// GetUrgencyScore returns a numeric score for urgency (1-4, higher is more urgent)
func (pi *PurchaseIntent) GetUrgencyScore() int {
	switch pi.Urgency {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 1
	}
}

// IsCritical returns true if the purchase is marked as critical urgency
func (pi *PurchaseIntent) IsCritical() bool {
	return pi.Urgency == "critical"
}

// IsRecurring returns true if the purchase is not one-time
func (pi *PurchaseIntent) IsRecurring() bool {
	return pi.Frequency == "monthly" || pi.Frequency == "yearly"
}

// GetCategoryPriority returns priority score for the category (1-5, higher is more important)
func (pi *PurchaseIntent) GetCategoryPriority() int {
	switch pi.Category {
	case "health":
		return 5
	case "food":
		return 4
	case "transport":
		return 4
	case "electronics":
		return 2
	case "clothing":
		return 2
	case "entertainment":
		return 1
	case "other":
		return 2
	default:
		return 1
	}
}

// HasAlternative returns true if an alternative option was specified
func (pi *PurchaseIntent) HasAlternative() bool {
	return strings.TrimSpace(pi.Alternative) != ""
}

// HasPurpose returns true if a purpose was specified
func (pi *PurchaseIntent) HasPurpose() bool {
	return strings.TrimSpace(pi.Purpose) != ""
}

// String returns a string representation of the purchase intent
func (pi *PurchaseIntent) String() string {
	return fmt.Sprintf("PurchaseIntent{ID: %s, ItemName: %s, ItemCost: %.2f, Category: %s, Urgency: %s}",
		pi.ID, pi.ItemName, pi.ItemCost, pi.Category, pi.Urgency)
}