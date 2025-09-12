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
	// Health profiles indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_health_profiles_user_active ON health_profiles(user_id, created_at)").Error; err != nil {
		return fmt.Errorf("failed to create health_profiles user_active index: %w", err)
	}

	// Medical conditions indexes  
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_medical_conditions_user_category_active ON medical_conditions(user_id, category, is_active)").Error; err != nil {
		return fmt.Errorf("failed to create medical_conditions composite index: %w", err)
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_medical_conditions_profile_severity ON medical_conditions(profile_id, severity, is_active)").Error; err != nil {
		return fmt.Errorf("failed to create medical_conditions severity index: %w", err)
	}

	// Medical expenses indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_medical_expenses_user_date_category ON medical_expenses(user_id, date DESC, category)").Error; err != nil {
		return fmt.Errorf("failed to create medical_expenses date_category index: %w", err)
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_medical_expenses_profile_recurring ON medical_expenses(profile_id, is_recurring, frequency)").Error; err != nil {
		return fmt.Errorf("failed to create medical_expenses recurring index: %w", err)
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_medical_expenses_amount_covered ON medical_expenses(amount, is_covered, insurance_payment)").Error; err != nil {
		return fmt.Errorf("failed to create medical_expenses coverage index: %w", err)
	}

	// Insurance policies indexes
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_insurance_policies_user_active_dates ON insurance_policies(user_id, is_active, start_date, end_date)").Error; err != nil {
		return fmt.Errorf("failed to create insurance_policies active_dates index: %w", err)
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_insurance_policies_profile_type_active ON insurance_policies(profile_id, type, is_active)").Error; err != nil {
		return fmt.Errorf("failed to create insurance_policies type index: %w", err)
	}
	
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_insurance_policies_deductible_tracking ON insurance_policies(deductible_met, out_of_pocket_current, annual_deductible)").Error; err != nil {
		return fmt.Errorf("failed to create insurance_policies tracking index: %w", err)
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