package testutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	mysqlcontainer "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/DuckDHD/BuyOrBye/internal/models"
)

// TestDatabase represents a test database container setup
type TestDatabase struct {
	Container testcontainers.Container
	DB        *gorm.DB
	DSN       string
}

// SetupMySQLTestContainer creates and starts a MySQL testcontainer for integration tests
func SetupMySQLTestContainer(t *testing.T) *TestDatabase {
	ctx := context.Background()

	// Create MySQL container
	mysqlContainer, err := mysqlcontainer.RunContainer(ctx,
		testcontainers.WithImage("mysql:8.0"),
		mysqlcontainer.WithDatabase("buyorbye_test"),
		mysqlcontainer.WithUsername("testuser"),
		mysqlcontainer.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("ready for connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start MySQL container: %v", err)
	}

	// Get connection string
	dsn, err := mysqlContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Adjust DSN for GORM (add parseTime=true)
	dsn += "?parseTime=true&loc=Local"

	// Create GORM connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Reduce logging noise in tests
	})
	if err != nil {
		// Clean up container if DB connection fails
		if termErr := mysqlContainer.Terminate(ctx); termErr != nil {
			t.Logf("Failed to terminate container after DB connection failure: %v", termErr)
		}
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate all models
	err = AutoMigrateAllModels(db)
	if err != nil {
		// Clean up on migration failure
		if termErr := mysqlContainer.Terminate(ctx); termErr != nil {
			t.Logf("Failed to terminate container after migration failure: %v", termErr)
		}
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return &TestDatabase{
		Container: mysqlContainer,
		DB:        db,
		DSN:       dsn,
	}
}

// AutoMigrateAllModels runs auto-migration for all project models
func AutoMigrateAllModels(db *gorm.DB) error {
	return db.AutoMigrate(
		// Health domain models
		&models.HealthProfileModel{},
		&models.MedicalConditionModel{},
		&models.MedicalExpenseModel{},
		&models.InsurancePolicyModel{},

		// Decision domain models
		&models.DecisionRecordModel{},
		&models.AIPromptLogModel{},

		// Auth/User models (if they exist)
		&models.UserModel{},
		&models.RefreshTokenModel{},

		// Finance domain models
		&models.IncomeModel{},
		&models.ExpenseModel{},
		&models.LoanModel{},
		&models.FinanceSummaryModel{},
	)
}

// Cleanup terminates the test container
func (td *TestDatabase) Cleanup(t *testing.T) {
	ctx := context.Background()
	if err := td.Container.Terminate(ctx); err != nil {
		t.Logf("Failed to terminate test container: %v", err)
	}
}

// ResetDatabase cleans all tables for test isolation
func (td *TestDatabase) ResetDatabase(t *testing.T) {
	// List of all tables to clean (in order to respect foreign key constraints)
	tables := []string{
		"ai_prompt_logs",
		"decision_records",
		"medical_expenses",
		"medical_conditions",
		"insurance_policies",
		"health_profiles",
		"refresh_tokens",
		"users",
		"expenses",
		"incomes",
		"loans",
		"finance_summaries",
	}

	// Disable foreign key checks
	td.DB.Exec("SET FOREIGN_KEY_CHECKS = 0")

	// Truncate all tables
	for _, table := range tables {
		if err := td.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table)).Error; err != nil {
			// Log error but don't fail test - table might not exist
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}

	// Re-enable foreign key checks
	td.DB.Exec("SET FOREIGN_KEY_CHECKS = 1")
}

// CreateTestData creates common test data for repositories
func (td *TestDatabase) CreateTestData(t *testing.T) {
	// Create a test user
	user := &models.UserModel{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$10$example.hash.for.testing",
		IsActive:     true,
	}
	
	if err := td.DB.Create(user).Error; err != nil {
		t.Logf("Warning: Failed to create test user: %v", err)
	}

	// Create test health profile
	healthProfile := &models.HealthProfileModel{
		UserID:     "test-user-123",
		Age:        30,
		Gender:     "male",
		Height:     175.0,
		Weight:     70.0,
		FamilySize: 2,
	}
	
	if err := td.DB.Create(healthProfile).Error; err != nil {
		t.Logf("Warning: Failed to create test health profile: %v", err)
	}
}

// GetTestDSN returns a test database DSN for non-container tests
func GetTestDSN() string {
	// For CI/CD environments or when MySQL is available locally
	return "testuser:testpass@tcp(localhost:3306)/buyorbye_test?parseTime=true&loc=Local"
}

// IsContainerTestsEnabled checks if container tests should run
func IsContainerTestsEnabled() bool {
	// You can set this based on environment variables or build tags
	// For now, we'll always enable container tests
	return true
}

// SetupTestDB creates either a container-based or regular test database
func SetupTestDB(t *testing.T) *TestDatabase {
	if IsContainerTestsEnabled() {
		return SetupMySQLTestContainer(t)
	}
	
	// Fallback to regular connection (assumes MySQL is running locally)
	db, err := gorm.Open(mysql.Open(GetTestDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping test: local MySQL not available: %v", err)
	}
	
	// Auto-migrate models
	err = AutoMigrateAllModels(db)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	
	return &TestDatabase{
		Container: nil, // No container for local DB
		DB:        db,
		DSN:       GetTestDSN(),
	}
}

// Helper functions for test data creation

// CreateTestDecisionRecord creates a test decision record
func CreateTestDecisionRecord(db *gorm.DB, userID, intentID string) *models.DecisionRecordModel {
	record := &models.DecisionRecordModel{
		UserID:          userID,
		IntentID:        intentID,
		ItemName:        "Test Laptop",
		ItemCost:        999.99,
		Category:        "electronics",
		Urgency:         "medium",
		Frequency:       "one_time",
		Purpose:         "Work laptop replacement",
		Alternative:     "Consider refurbished options",
		Decision:        "BUY",
		Confidence:      0.85,
		PrimaryReason:   "Good value within budget",
		WaitPeriod:      0,
		MaxBudget:       1200.0,
		ProcessingTime:  2500,
		DecisionSource:  "business_rules",
	}
	
	db.Create(record)
	return record
}

// CreateTestAIPromptLog creates a test AI prompt log
func CreateTestAIPromptLog(db *gorm.DB, userID, requestID, intentID string) *models.AIPromptLogModel {
	record := &models.AIPromptLogModel{
		UserID:           userID,
		RequestID:        requestID,
		IntentID:         intentID,
		AIProvider:       "openai",
		AIModel:          "gpt-4o-mini",
		Temperature:      0.7,
		SystemContext:    "You are a purchase decision assistant",
		UserContext:      "User has budget of $1000",
		PurchaseDetails:  "Item: Laptop, Cost: $800",
		DecisionCriteria: "Consider budget and necessity",
		ResponseFormat:   "JSON format with decision and confidence",
		MaxTokens:        500,
		TokensInput:      200,
		TokensOutput:     150,
		TokensTotal:      350,
		RawResponse:      `{"decision": "BUY", "confidence": 0.85}`,
		ParsedDecision:   "BUY",
		ParsedConfidence: 0.85,
		ResponseTimeMs:   2000,
		ProcessingTimeMs: 2500,
		Success:          true,
		EstimatedCostUSD: 0.001,
	}
	
	db.Create(record)
	return record
}

// BatchCreateDecisionRecords creates multiple decision records for testing
func BatchCreateDecisionRecords(db *gorm.DB, userID string, count int) []*models.DecisionRecordModel {
	records := make([]*models.DecisionRecordModel, count)
	decisions := []string{"BUY", "WAIT", "BYE"}
	categories := []string{"electronics", "clothing", "food", "transport"}
	
	for i := 0; i < count; i++ {
		records[i] = &models.DecisionRecordModel{
			UserID:         userID,
			IntentID:       fmt.Sprintf("intent-%d", i+1),
			ItemName:       fmt.Sprintf("Test Item %d", i+1),
			ItemCost:       float64(100 + (i * 50)),
			Category:       categories[i%len(categories)],
			Urgency:        "medium",
			Frequency:      "one_time",
			Purpose:        fmt.Sprintf("Test purpose %d", i+1),
			Decision:       decisions[i%len(decisions)],
			Confidence:     0.5 + (float64(i%5) * 0.1), // 0.5 to 0.9
			PrimaryReason:  fmt.Sprintf("Test reason %d", i+1),
			ProcessingTime: int64(1000 + (i * 100)),
		}
		
		// Stagger creation times
		records[i].CreatedAt = time.Now().Add(-time.Duration(i*24) * time.Hour)
	}
	
	// Batch create
	db.Create(&records)
	return records
}

// BatchCreateAIPromptLogs creates multiple AI prompt logs for testing
func BatchCreateAIPromptLogs(db *gorm.DB, userID string, count int) []*models.AIPromptLogModel {
	records := make([]*models.AIPromptLogModel, count)
	providers := []string{"openai", "anthropic"}
	decisions := []string{"BUY", "WAIT", "BYE"}
	
	for i := 0; i < count; i++ {
		success := i%4 != 0 // 75% success rate
		
		records[i] = &models.AIPromptLogModel{
			UserID:           userID,
			RequestID:        fmt.Sprintf("req-%d", i+1),
			IntentID:         fmt.Sprintf("intent-%d", i+1),
			AIProvider:       providers[i%len(providers)],
			AIModel:          "gpt-4o-mini",
			Temperature:      0.7,
			SystemContext:    "Test system context",
			UserContext:      fmt.Sprintf("Test user context %d", i+1),
			PurchaseDetails:  fmt.Sprintf("Test purchase %d", i+1),
			DecisionCriteria: "Test criteria",
			ResponseFormat:   "JSON",
			MaxTokens:        500,
			TokensInput:      100 + (i * 10),
			TokensOutput:     50 + (i * 5),
			ResponseTimeMs:   int64(1000 + (i * 100)),
			ProcessingTimeMs: int64(1200 + (i * 100)),
			Success:          success,
		}
		
		if success {
			records[i].TokensOutput = 50 + (i * 5)
			records[i].TokensTotal = records[i].TokensInput + records[i].TokensOutput
			records[i].RawResponse = fmt.Sprintf(`{"decision": "%s", "confidence": 0.8}`, decisions[i%len(decisions)])
			records[i].ParsedDecision = decisions[i%len(decisions)]
			records[i].ParsedConfidence = 0.8
			records[i].EstimatedCostUSD = 0.001 + (float64(i) * 0.0001)
		} else {
			records[i].ErrorMessage = "API timeout"
			records[i].StatusCode = 500
		}
		
		// Stagger creation times
		records[i].CreatedAt = time.Now().Add(-time.Duration(i*6) * time.Hour)
	}
	
	// Batch create
	db.Create(&records)
	return records
}