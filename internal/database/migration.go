package database

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/DuckDHD/BuyOrBye/internal/models"
)

// RunAllMigrations runs all database migrations in the correct order
func RunAllMigrations(db *gorm.DB) error {
	// Run core system migrations first
	if err := runCoreMigrations(db); err != nil {
		return fmt.Errorf("core migrations failed: %w", err)
	}

	// Run health domain migrations
	if err := runHealthDomainMigrations(db); err != nil {
		return fmt.Errorf("health migrations failed: %w", err)
	}

	// Add composite indexes for performance
	if err := createCompositeIndexes(db); err != nil {
		return fmt.Errorf("index creation failed: %w", err)
	}

	// Create additional constraints
	if err := createAdditionalConstraints(db); err != nil {
		return fmt.Errorf("constraint creation failed: %w", err)
	}

	return nil
}

// runCoreMigrations runs core system table migrations
func runCoreMigrations(db *gorm.DB) error {
	// Auto-migrate core models in dependency order
	if err := db.AutoMigrate(
		&models.UserModel{},
		&models.RefreshTokenModel{},
		&models.ExpenseModel{},
		&models.IncomeModel{},
		&models.LoanModel{},
		&models.FinanceSummaryModel{},
	); err != nil {
		return fmt.Errorf("failed to auto-migrate core models: %w", err)
	}

	return nil
}

// runHealthDomainMigrations runs health domain specific migrations
func runHealthDomainMigrations(db *gorm.DB) error {
	// Run health domain migrations
	err := db.AutoMigrate(
		&models.HealthProfileModel{},
		&models.MedicalConditionModel{},
		&models.MedicalExpenseModel{},
		&models.InsurancePolicyModel{},
	)
	if err != nil {
		return fmt.Errorf("health migration failed: %w", err)
	}

	return nil
}

// createCompositeIndexes creates composite indexes for better query performance
func createCompositeIndexes(db *gorm.DB) error {
	// Health-related composite indexes
	indexes := []struct {
		name  string
		query string
	}{
		// Medical expenses indexes for common queries
		{
			name:  "idx_expenses_user_date",
			query: "CREATE INDEX IF NOT EXISTS idx_expenses_user_date ON medical_expenses(user_id, date DESC)",
		},
		{
			name:  "idx_conditions_profile_severity",
			query: "CREATE INDEX IF NOT EXISTS idx_conditions_profile_severity ON medical_conditions(profile_id, severity, is_active)",
		},
		// Insurance policies for coverage lookups
		{
			name:  "idx_policies_user_active_type",
			query: "CREATE INDEX IF NOT EXISTS idx_policies_user_active_type ON insurance_policies(user_id, is_active, type)",
		},
		// Health profiles for user lookups
		{
			name:  "idx_health_profiles_user",
			query: "CREATE INDEX IF NOT EXISTS idx_health_profiles_user ON health_profiles(user_id, created_at)",
		},
		// Financial data indexes for affordability calculations
		{
			name:  "idx_expenses_user_category_date",
			query: "CREATE INDEX IF NOT EXISTS idx_expenses_user_category_date ON expenses(user_id, category, date DESC)",
		},
		{
			name:  "idx_incomes_user_active_date",
			query: "CREATE INDEX IF NOT EXISTS idx_incomes_user_active_date ON incomes(user_id, is_active, date DESC)",
		},
		// Medical expenses coverage analysis
		{
			name:  "idx_medical_expenses_coverage",
			query: "CREATE INDEX IF NOT EXISTS idx_medical_expenses_coverage ON medical_expenses(user_id, is_covered, insurance_payment)",
		},
		// Recurring expenses tracking
		{
			name:  "idx_medical_expenses_recurring",
			query: "CREATE INDEX IF NOT EXISTS idx_medical_expenses_recurring ON medical_expenses(user_id, is_recurring, frequency)",
		},
	}

	for _, idx := range indexes {
		if err := db.Exec(idx.query).Error; err != nil {
			return fmt.Errorf("failed to create index %s: %w", idx.name, err)
		}
	}

	return nil
}

// createAdditionalConstraints creates additional database constraints
func createAdditionalConstraints(db *gorm.DB) error {
	constraints := []struct {
		name  string
		query string
	}{
		// Ensure one health profile per user
		{
			name:  "unique_user_health_profile",
			query: "ALTER TABLE health_profiles ADD CONSTRAINT unique_user_health_profile UNIQUE (user_id)",
		},
		// Ensure policy numbers are globally unique
		{
			name:  "unique_policy_number",
			query: "ALTER TABLE insurance_policies ADD CONSTRAINT unique_policy_number UNIQUE (policy_number)",
		},
		// Ensure reasonable BMI constraints
		{
			name:  "check_reasonable_bmi",
			query: "ALTER TABLE health_profiles ADD CONSTRAINT check_reasonable_bmi CHECK (bmi >= 10 AND bmi <= 100)",
		},
		// Ensure positive amounts for expenses
		{
			name:  "check_positive_medical_expense",
			query: "ALTER TABLE medical_expenses ADD CONSTRAINT check_positive_medical_expense CHECK (amount > 0)",
		},
		// Ensure insurance payment doesn't exceed expense amount
		{
			name:  "check_insurance_payment_reasonable",
			query: "ALTER TABLE medical_expenses ADD CONSTRAINT check_insurance_payment_reasonable CHECK (insurance_payment <= amount)",
		},
		// Ensure deductible doesn't exceed out-of-pocket max
		{
			name:  "check_deductible_reasonable",
			query: "ALTER TABLE insurance_policies ADD CONSTRAINT check_deductible_reasonable CHECK (deductible <= out_of_pocket_max)",
		},
	}

	for _, constraint := range constraints {
		if err := db.Exec(constraint.query).Error; err != nil {
			// Many databases don't support IF NOT EXISTS for constraints
			// So we ignore errors for existing constraints
			if !isConstraintExistsError(err) {
				return fmt.Errorf("failed to create constraint %s: %w", constraint.name, err)
			}
		}
	}

	return nil
}

// MigrateHealthModelsOnly runs only health domain migrations (useful for development)
func MigrateHealthModelsOnly(db *gorm.DB) error {
	return RunHealthMigrations(db)
}

// RollbackHealthMigrations drops all health tables (for testing/cleanup)
func RollbackHealthMigrations(db *gorm.DB) error {
	return DropHealthTables(db)
}

// GetMigrationStatus returns the status of all migrations
func GetMigrationStatus(db *gorm.DB) map[string]bool {
	status := make(map[string]bool)

	// Check core tables
	coreModels := []string{"users", "refresh_tokens", "expenses", "incomes", "loans", "finance_summaries"}
	for _, table := range coreModels {
		status[table] = db.Migrator().HasTable(table)
	}

	// Check health tables
	healthModels := []string{"health_profiles", "medical_conditions", "medical_expenses", "insurance_policies"}
	for _, table := range healthModels {
		status[table] = db.Migrator().HasTable(table)
	}

	return status
}

// ValidateMigrationIntegrity checks that all expected tables and constraints exist
func ValidateMigrationIntegrity(db *gorm.DB) error {
	migrator := db.Migrator()

	// Required tables
	requiredTables := []string{
		"users", "refresh_tokens", "expenses", "incomes", "loans", "finance_summaries",
		"health_profiles", "medical_conditions", "medical_expenses", "insurance_policies",
	}

	for _, table := range requiredTables {
		if !migrator.HasTable(table) {
			return fmt.Errorf("missing required table: %s", table)
		}
	}

	// Check critical columns exist
	criticalColumns := map[string][]string{
		"health_profiles": {"user_id", "age", "gender", "height", "weight", "bmi"},
		"medical_conditions": {"user_id", "profile_id", "name", "severity", "is_active"},
		"medical_expenses": {"user_id", "profile_id", "amount", "category", "date"},
		"insurance_policies": {"user_id", "policy_number", "type", "deductible", "out_of_pocket_max"},
	}

	for table, columns := range criticalColumns {
		for _, column := range columns {
			if !migrator.HasColumn(table, column) {
				return fmt.Errorf("missing required column %s.%s", table, column)
			}
		}
	}

	return nil
}

// SetupTestDatabase prepares database for testing with clean migrations
func SetupTestDatabase(db *gorm.DB) error {
	// Drop existing tables to ensure clean state
	if err := DropHealthTables(db); err != nil {
		return fmt.Errorf("failed to drop existing health tables: %w", err)
	}

	// Run fresh migrations
	if err := RunAllMigrations(db); err != nil {
		return fmt.Errorf("failed to run test migrations: %w", err)
	}

	return nil
}