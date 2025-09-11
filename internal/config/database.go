package config

import (
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseService provides database connection management
type DatabaseService interface {
	// GetDB returns the GORM database instance
	GetDB() *gorm.DB

	// Health returns database health status
	Health() map[string]string

	// Close closes the database connection
	Close() error
}

// NewDatabaseService creates a new database service from configuration
func NewDatabaseService(config *DatabaseConfig, logConfig *LoggingConfig) (DatabaseService, error) {
	db, err := setupDatabase(config, logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
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

	return &databaseService{
		db:     db,
		config: config,
	}, nil
}

type databaseService struct {
	db     *gorm.DB
	config *DatabaseConfig
}

func (d *databaseService) GetDB() *gorm.DB {
	return d.db
}

func (d *databaseService) Health() map[string]string {
	stats := make(map[string]string)

	// Get underlying sql.DB for health check
	sqlDB, err := d.db.DB()
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("failed to get sql.DB: %v", err)
		return stats
	}

	// Ping the database
	if err := sqlDB.Ping(); err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db ping failed: %v", err)
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "Database connection is healthy"

	// Get database stats
	dbStats := sqlDB.Stats()
	stats["open_connections"] = fmt.Sprintf("%d", dbStats.OpenConnections)
	stats["in_use"] = fmt.Sprintf("%d", dbStats.InUse)
	stats["idle"] = fmt.Sprintf("%d", dbStats.Idle)

	// Add configured limits for reference
	stats["max_open_configured"] = fmt.Sprintf("%d", d.config.MaxOpenConns)
	stats["max_idle_configured"] = fmt.Sprintf("%d", d.config.MaxIdleConns)

	// Evaluate stats for health message
	if dbStats.OpenConnections > int(float64(d.config.MaxOpenConns)*0.8) {
		stats["message"] = "Database experiencing heavy load"
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "Database has high wait events, potential bottlenecks"
	}

	return stats
}

func (d *databaseService) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// setupDatabase initializes the database connection
func setupDatabase(config *DatabaseConfig, logConfig *LoggingConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	if config.IsMySQL() {
		dialector = mysql.Open(config.GetDSN())
	} else {
		dialector = sqlite.Open(config.GetDSN())
	}

	// Configure GORM logger based on logging config
	gormLogger := logger.Default
	switch logConfig.Level {
	case "debug":
		gormLogger = logger.Default.LogMode(logger.Info)
	case "info":
		gormLogger = logger.Default.LogMode(logger.Warn)
	case "warn", "error":
		gormLogger = logger.Default.LogMode(logger.Error)
	default:
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// Open database connection
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool for MySQL
	if config.IsMySQL() {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get sql.DB: %w", err)
		}

		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)

		if config.ConnMaxLifetime > 0 {
			sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
		}
	}

	return db, nil
}

// GetGormConfig returns GORM configuration for the given environment
func GetGormConfig(environment string) *gorm.Config {
	switch environment {
	case "production":
		return &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		}
	case "test":
		return &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		}
	default: // development
		return &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		}
	}
}
