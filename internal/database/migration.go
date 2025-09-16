package database

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/DuckDHD/BuyOrBye/internal/models"
)

func ensureIndex(db *gorm.DB, table, indexName, columns string) error {
	var cnt int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.statistics
		WHERE table_schema = DATABASE()
		  AND table_name = ?
		  AND index_name = ?`,
		table, indexName,
	).Scan(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	stmt := fmt.Sprintf("ALTER TABLE %s ADD INDEX %s (%s)", table, indexName, columns)
	return db.Exec(stmt).Error
}

func ensureUniqueConstraint(db *gorm.DB, table, constraintName, columns string) error {
	// Check by constraint name
	var cnt int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.table_constraints
		WHERE table_schema = DATABASE()
		  AND table_name = ?
		  AND constraint_name = ?
		  AND constraint_type = 'UNIQUE'`,
		table, constraintName,
	).Scan(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	// Add via unique index (portable); MySQL maps it to a unique constraint
	stmt := fmt.Sprintf("ALTER TABLE %s ADD UNIQUE %s (%s)", table, constraintName, columns)
	return db.Exec(stmt).Error
}

func ensureCheckConstraint(db *gorm.DB, table, constraintName, checkExpr string) error {
	// Many MySQL/MariaDB versions ignore or don’t support CHECK; try and ignore if not supported.
	var cnt int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.table_constraints
		WHERE table_schema = DATABASE()
		  AND table_name = ?
		  AND constraint_name = ?
		  AND constraint_type = 'CHECK'`,
		table, constraintName,
	).Scan(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	stmt := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s CHECK (%s)", table, constraintName, checkExpr)
	if err := db.Exec(stmt).Error; err != nil {
		// Gracefully ignore unsupported CHECK constraint errors
		return nil
	}
	return nil
}

// ensure a column exists; if not, add it with the provided definition
func ensureColumn(db *gorm.DB, table, column, typeDef string) error {
	var cnt int64
	if err := db.Raw(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?`,
		table, column,
	).Scan(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, typeDef)
	return db.Exec(stmt).Error
}

// rename a column if the desired column is missing but a legacy column exists
func renameColumnIfLegacy(db *gorm.DB, table, legacyCol, newCol, typeDef string) error {
	var hasNew, hasLegacy int64
	if err := db.Raw(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?`,
		table, newCol,
	).Scan(&hasNew).Error; err != nil {
		return err
	}
	if hasNew > 0 {
		return nil
	}
	if err := db.Raw(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?`,
		table, legacyCol,
	).Scan(&hasLegacy).Error; err != nil {
		return err
	}
	if hasLegacy == 0 {
		// nothing to rename
		return nil
	}
	// CHANGE COLUMN legacy -> new (portable; defines type too)
	stmt := fmt.Sprintf("ALTER TABLE %s CHANGE COLUMN %s %s %s", table, legacyCol, newCol, typeDef)
	return db.Exec(stmt).Error
}

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

	// Run decision domain migrations
	if err := RunDecisionMigrations(db); err != nil {
		return fmt.Errorf("decision migrations failed: %w", err)
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

// RunDecisionMigrations runs decision domain specific migrations
func RunDecisionMigrations(db *gorm.DB) error {
	// Run decision domain migrations with AutoMigrate
	err := db.AutoMigrate(
		&models.DecisionRecordModel{},
		&models.AIPromptLogModel{},
	)
	if err != nil {
		return fmt.Errorf("decision migration failed: %w", err)
	}

	// Create decision-specific composite indexes
	if err := createDecisionIndexes(db); err != nil {
		return fmt.Errorf("decision index creation failed: %w", err)
	}

	return nil
}

// createDecisionIndexes creates composite indexes for decision queries
func createDecisionIndexes(db *gorm.DB) error {
	// --- Align ai_prompt_logs columns before indexing ---

	// Legacy rename: provider -> ai_provider
	if err := renameColumnIfLegacy(db, "ai_prompt_logs", "provider", "ai_provider", "varchar(50) NOT NULL DEFAULT 'openai'"); err != nil {
		return fmt.Errorf("failed to rename provider->ai_provider: %w", err)
	}

	// Columns used by any ai_prompt_logs indexes
	toEnsure := []struct {
		col string
		def string
	}{
		{"ai_provider", "varchar(50) NOT NULL DEFAULT 'openai'"},
		{"success", "TINYINT(1) NOT NULL DEFAULT 0"},
		{"created_at", "DATETIME(3) NULL"},
		{"response_time_ms", "BIGINT NOT NULL DEFAULT 0"},
		{"tokens_total", "INT NOT NULL DEFAULT 0"},
		{"request_id", "varchar(255) NOT NULL DEFAULT ''"},
		{"intent_id", "varchar(255) NULL"},
	}

	for _, c := range toEnsure {
		if err := ensureColumn(db, "ai_prompt_logs", c.col, c.def); err != nil {
			return fmt.Errorf("ensure ai_prompt_logs.%s: %w", c.col, err)
		}
	}

	// --- decision_records indexes ---
	if err := ensureIndex(db, "decision_records", "idx_decisions_user_date", "user_id, created_at"); err != nil {
		return fmt.Errorf("failed to create decision index idx_decisions_user_date: %w", err)
	}
	if err := ensureIndex(db, "decision_records", "idx_decisions_user_category_decision", "user_id, category, decision"); err != nil {
		return fmt.Errorf("failed to create decision index idx_decisions_user_category_decision: %w", err)
	}
	if err := ensureIndex(db, "decision_records", "idx_decisions_user_recent", "user_id, decision, created_at"); err != nil {
		return fmt.Errorf("failed to create decision index idx_decisions_user_recent: %w", err)
	}
	if err := ensureIndex(db, "decision_records", "idx_decisions_category_outcome", "category, decision, confidence"); err != nil {
		return fmt.Errorf("failed to create decision index idx_decisions_category_outcome: %w", err)
	}

	// --- ai_prompt_logs indexes ---
	if err := ensureIndex(db, "ai_prompt_logs", "idx_ai_logs_user_date", "user_id, created_at"); err != nil {
		return fmt.Errorf("failed to create ai index idx_ai_logs_user_date: %w", err)
	}
	if err := ensureIndex(db, "ai_prompt_logs", "idx_ai_logs_provider_success", "ai_provider, success, created_at"); err != nil {
		return fmt.Errorf("failed to create ai index idx_ai_logs_provider_success: %w", err)
	}
	if err := ensureIndex(db, "ai_prompt_logs", "idx_ai_logs_performance", "success, response_time_ms, tokens_total"); err != nil {
		return fmt.Errorf("failed to create ai index idx_ai_logs_performance: %w", err)
	}
	if err := ensureIndex(db, "ai_prompt_logs", "idx_ai_logs_request_tracking", "request_id, intent_id, created_at"); err != nil {
		return fmt.Errorf("failed to create ai index idx_ai_logs_request_tracking: %w", err)
	}

	return nil
}

// createCompositeIndexes creates composite indexes for better query performance
func createCompositeIndexes(db *gorm.DB) error {
	// Health domain
	if err := ensureIndex(db, "medical_expenses", "idx_expenses_user_date", "user_id, date"); err != nil {
		return fmt.Errorf("failed idx_expenses_user_date: %w", err)
	}
	if err := ensureIndex(db, "medical_conditions", "idx_conditions_profile_severity", "profile_id, severity, is_active"); err != nil {
		return fmt.Errorf("failed idx_conditions_profile_severity: %w", err)
	}
	if err := ensureIndex(db, "insurance_policies", "idx_policies_user_active_type", "user_id, is_active, type"); err != nil {
		return fmt.Errorf("failed idx_policies_user_active_type: %w", err)
	}
	if err := ensureIndex(db, "health_profiles", "idx_health_profiles_user", "user_id, created_at"); err != nil {
		return fmt.Errorf("failed idx_health_profiles_user: %w", err)
	}

	// Finance domain (use created_at; there is no `date` column)
	if err := ensureIndex(db, "expenses", "idx_expenses_user_category_date", "user_id, category, created_at"); err != nil {
		return fmt.Errorf("failed idx_expenses_user_category_date: %w", err)
	}
	if err := ensureIndex(db, "incomes", "idx_incomes_user_active_date", "user_id, is_active, created_at"); err != nil {
		return fmt.Errorf("failed idx_incomes_user_active_date: %w", err)
	}

	// Medical expense analysis
	if err := ensureIndex(db, "medical_expenses", "idx_medical_expenses_coverage", "user_id, is_covered, insurance_payment"); err != nil {
		return fmt.Errorf("failed idx_medical_expenses_coverage: %w", err)
	}
	if err := ensureIndex(db, "medical_expenses", "idx_medical_expenses_recurring", "user_id, is_recurring, frequency"); err != nil {
		return fmt.Errorf("failed idx_medical_expenses_recurring: %w", err)
	}

	return nil
}

// createAdditionalConstraints creates additional database constraints
func createAdditionalConstraints(db *gorm.DB) error {
	// Unique constraints — create as UNIQUE constraints by name
	if err := ensureUniqueConstraint(db, "health_profiles", "unique_user_health_profile", "user_id"); err != nil {
		return fmt.Errorf("failed unique_user_health_profile: %w", err)
	}
	if err := ensureUniqueConstraint(db, "insurance_policies", "unique_policy_number", "policy_number"); err != nil {
		return fmt.Errorf("failed unique_policy_number: %w", err)
	}

	// CHECK constraints — best-effort (ignored on servers that don’t support/enforce)
	_ = ensureCheckConstraint(db, "health_profiles", "check_reasonable_bmi", "bmi >= 10 AND bmi <= 100")
	_ = ensureCheckConstraint(db, "medical_expenses", "check_positive_medical_expense", "amount > 0")
	_ = ensureCheckConstraint(db, "medical_expenses", "check_insurance_payment_reasonable", "insurance_payment <= amount")
	_ = ensureCheckConstraint(db, "insurance_policies", "check_deductible_reasonable", "deductible <= out_of_pocket_max")

	return nil
}

// MigrateHealthModelsOnly runs only health domain migrations (useful for development)
func MigrateHealthModelsOnly(db *gorm.DB) error {
	return runHealthDomainMigrations(db)
}

// RollbackHealthMigrations drops all health tables (for testing/cleanup)
func RollbackHealthMigrations(db *gorm.DB) error {
	// Drop health tables in reverse order
	tables := []string{
		"medical_expenses",
		"medical_conditions",
		"insurance_policies",
		"health_profiles",
	}

	for _, table := range tables {
		if db.Migrator().HasTable(table) {
			if err := db.Migrator().DropTable(table); err != nil {
				return fmt.Errorf("failed to drop table %s: %w", table, err)
			}
		}
	}

	return nil
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

	// Check decision tables
	decisionModels := []string{"decision_records", "ai_prompt_logs"}
	for _, table := range decisionModels {
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
		"decision_records", "ai_prompt_logs",
	}

	for _, table := range requiredTables {
		if !migrator.HasTable(table) {
			return fmt.Errorf("missing required table: %s", table)
		}
	}

	// Check critical columns exist
	criticalColumns := map[string][]string{
		"health_profiles":    {"user_id", "age", "gender", "height", "weight", "bmi"},
		"medical_conditions": {"user_id", "profile_id", "name", "severity", "is_active"},
		"medical_expenses":   {"user_id", "profile_id", "amount", "category", "date"},
		"insurance_policies": {"user_id", "policy_number", "type", "deductible", "out_of_pocket_max"},
		"decision_records":   {"user_id", "intent_id", "item_name", "item_cost", "category", "decision", "confidence"},
		"ai_prompt_logs":     {"user_id", "request_id", "ai_provider", "max_tokens", "tokens_input", "tokens_output", "success"},
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
	if err := RollbackHealthMigrations(db); err != nil {
		return fmt.Errorf("failed to drop existing health tables: %w", err)
	}

	// Run fresh migrations
	if err := RunAllMigrations(db); err != nil {
		return fmt.Errorf("failed to run test migrations: %w", err)
	}

	return nil
}
