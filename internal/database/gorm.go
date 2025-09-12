package database

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/DuckDHD/BuyOrBye/internal/models"
)

// GormService provides GORM database functionality
type GormService struct {
	db *gorm.DB
}

// NewGormService creates a new GORM database service
func NewGormService() (*GormService, error) {
	// Get database configuration from environment
	dbname := os.Getenv("BLUEPRINT_DB_DATABASE")
	password := os.Getenv("BLUEPRINT_DB_PASSWORD")
	username := os.Getenv("BLUEPRINT_DB_USERNAME")
	port := os.Getenv("BLUEPRINT_DB_PORT")
	host := os.Getenv("BLUEPRINT_DB_HOST")

	// Build MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, dbname)

	// Configure GORM logger based on environment
	logLevel := logger.Info
	if os.Getenv("APP_ENV") == "production" {
		logLevel = logger.Warn
	}

	// Open GORM connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the schema for all models
	if err := db.AutoMigrate(
		&models.UserModel{},
		&models.RefreshTokenModel{},
		&models.ExpenseModel{},
		&models.IncomeModel{},
		&models.LoanModel{},
		&models.FinanceSummaryModel{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}
	
	// Run health domain migrations
	if err := RunHealthMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to migrate health schema: %w", err)
	}

	return &GormService{db: db}, nil
}

// GetDB returns the GORM database instance
func (gs *GormService) GetDB() *gorm.DB {
	return gs.db
}

// Health returns health status information for the database
func (gs *GormService) Health() map[string]string {
	stats := make(map[string]string)

	// Get underlying SQL DB for health check
	sqlDB, err := gs.db.DB()
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("failed to get sql db: %v", err)
		return stats
	}

	// Ping the database
	if err := sqlDB.Ping(); err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "GORM database connection is healthy"

	// Add connection pool stats
	dbStats := sqlDB.Stats()
	stats["open_connections"] = fmt.Sprintf("%d", dbStats.OpenConnections)
	stats["in_use"] = fmt.Sprintf("%d", dbStats.InUse)
	stats["idle"] = fmt.Sprintf("%d", dbStats.Idle)

	return stats
}

// Close closes the database connection
func (gs *GormService) Close() error {
	sqlDB, err := gs.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql db for closing: %w", err)
	}
	return sqlDB.Close()
}
