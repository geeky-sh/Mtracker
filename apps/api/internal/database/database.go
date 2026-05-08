package database

import (
	"github.com/aash/mtracker/apps/api/internal/config"
	"github.com/aash/mtracker/apps/api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openDialector builds the GORM dialector; overridable in tests.
var openDialector = func(dsn string) gorm.Dialector {
	return postgres.Open(dsn)
}

func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(openDialector(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Activity{},
		&models.ActivityLog{},
	); err != nil {
		return nil, err
	}

	// Unique constraint: one log per activity per day
	db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_activity_logs_activity_date
		ON activity_logs (activity_id, logged_date)
	`)

	return db, nil
}
