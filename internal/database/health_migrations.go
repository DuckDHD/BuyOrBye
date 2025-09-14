package database

import (
	"fmt"

	"github.com/DuckDHD/BuyOrBye/internal/models"
	"gorm.io/gorm"
)

// RunHealthMigrations runs all health-related database migrations
func RunHealthMigrations(db *gorm.DB) error {
	// Auto-migrate health models in dependency order
	// HealthProfile must be created first as others reference it
	if err := db.AutoMigrate(
		&models.HealthProfileModel{},
		&models.MedicalConditionModel{},
		&models.MedicalExpenseModel{},
		&models.InsurancePolicyModel{},
	); err != nil {
		return fmt.Errorf("failed to auto-migrate health models: %w", err)
	}

	// Create custom indexes for better query performance
	if err := createHealthIndexes(db); err != nil {
		return fmt.Errorf("failed to create health indexes: %w", err)
	}

	return nil
}

// createHealthIndexes creates custom composite indexes for health tables
func createHealthIndexes(db *gorm.DB) error {
	indexes := []struct {
		name  string
		query string
	}{
		{
			name:  "idx_health_profiles_user_active",
			query: "CREATE INDEX idx_health_profiles_user_active ON health_profiles(user_id, created_at)",
		},
		{
			name:  "idx_medical_conditions_user_category_active",
			query: "CREATE INDEX idx_medical_conditions_user_category_active ON medical_conditions(user_id, category, is_active)",
		},
		{
			name:  "idx_medical_conditions_profile_severity",
			query: "CREATE INDEX idx_medical_conditions_profile_severity ON medical_conditions(profile_id, severity, is_active)",
		},
		{
			name:  "idx_medical_expenses_user_date_category",
			query: "CREATE INDEX idx_medical_expenses_user_date_category ON medical_expenses(user_id, date DESC, category)",
		},
		{
			name:  "idx_medical_expenses_profile_recurring",
			query: "CREATE INDEX idx_medical_expenses_profile_recurring ON medical_expenses(profile_id, is_recurring, frequency)",
		},
		{
			name:  "idx_medical_expenses_amount_covered",
			query: "CREATE INDEX idx_medical_expenses_amount_covered ON medical_expenses(amount, is_covered, insurance_payment)",
		},
		{
			name:  "idx_insurance_policies_user_active_dates",
			query: "CREATE INDEX idx_insurance_policies_user_active_dates ON insurance_policies(user_id, is_active, start_date, end_date)",
		},
		{
			name:  "idx_insurance_policies_profile_type_active",
			query: "CREATE INDEX idx_insurance_policies_profile_type_active ON insurance_policies(profile_id, type, is_active)",
		},
		{
			name:  "idx_insurance_policies_deductible_tracking",
			query: "CREATE INDEX idx_insurance_policies_deductible_tracking ON insurance_policies(deductible_met, out_of_pocket_current, annual_deductible)",
		},
	}

	for _, idx := range indexes {
		if err := db.Exec(idx.query).Error; err != nil {
			// Ignore errors for existing indexes (MySQL doesn't support IF NOT EXISTS before 8.0.1)
			if !isIndexExistsError(err) {
				return fmt.Errorf("failed to create index %s: %w", idx.name, err)
			}
		}
	}

	return nil
}

// DropHealthTables drops all health-related tables (for testing/cleanup)
func DropHealthTables(db *gorm.DB) error {
	// Drop in reverse dependency order
	tables := []string{
		"insurance_policies",
		"medical_expenses", 
		"medical_conditions",
		"health_profiles",
	}
	
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}
	
	return nil
}

// CreateHealthConstraints creates additional foreign key constraints for health tables
func CreateHealthConstraints(db *gorm.DB) error {
	// Note: GORM's AutoMigrate should handle most constraints, 
	// but we can add additional ones here if needed
	
	// Ensure health profiles have unique user constraint
	if err := db.Exec("ALTER TABLE health_profiles ADD CONSTRAINT unique_user_profile UNIQUE (user_id)").Error; err != nil {
		// Ignore error if constraint already exists
		if !isConstraintExistsError(err) {
			return fmt.Errorf("failed to create unique user profile constraint: %w", err)
		}
	}
	
	// Ensure policy numbers are globally unique
	if err := db.Exec("ALTER TABLE insurance_policies ADD CONSTRAINT unique_policy_number UNIQUE (policy_number)").Error; err != nil {
		// Ignore error if constraint already exists
		if !isConstraintExistsError(err) {
			return fmt.Errorf("failed to create unique policy number constraint: %w", err)
		}
	}
	
	return nil
}

// isConstraintExistsError checks if the error is due to constraint already existing
func isConstraintExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// MySQL error patterns for existing constraints
	return contains(errStr, "Duplicate key name") || 
		   contains(errStr, "already exists") ||
		   contains(errStr, "Duplicate entry")
}

// isIndexExistsError checks if the error is due to index already existing
func isIndexExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// MySQL error patterns for existing indexes
	return contains(errStr, "Duplicate key name") ||
		   contains(errStr, "already exists") ||
		   contains(errStr, "Error 1061")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}