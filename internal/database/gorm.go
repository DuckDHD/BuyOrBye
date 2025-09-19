package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/DuckDHD/BuyOrBye/internal/config"
)

// GormService provides GORM database functionality
type GormService struct {
	db *gorm.DB
}

// NewGormService creates a new GORM database service using config
func NewGormService() (*GormService, error) {
	// Load configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := SetupDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	// Run all migrations (includes core and health domains)
	if err := RunAllMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &GormService{db: db}, nil
}

// SetupDatabase initializes the database connection using configuration
func SetupDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)

	// Configure GORM logger based on environment
	logLevel := logger.Info
	if cfg.App.Environment == "production" {
		logLevel = logger.Warn
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	return db, nil
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
