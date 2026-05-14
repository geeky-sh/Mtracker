package database

import (
	"fmt"

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
		&models.UserIdentity{},
		&models.Activity{},
		&models.ActivityLog{},
	); err != nil {
		return nil, err
	}

	// Backfill user_identities from the legacy google_id column (if it still exists).
	db.Exec(fmt.Sprintf(`
		INSERT INTO user_identities (id, user_id, provider, provider_user_id, created_at, updated_at)
		SELECT gen_random_uuid(), id, '%s', google_id, now(), now()
		FROM users
		WHERE google_id IS NOT NULL AND google_id != ''
		ON CONFLICT DO NOTHING
	`, models.ProviderGoogle))
	db.Exec(`ALTER TABLE users DROP COLUMN IF EXISTS google_id`)

	// Unique constraint: one identity per provider per external user.
	db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_user_identities_provider
		ON user_identities (provider, provider_user_id)
	`)

	// Unique constraint: one log per activity per day.
	db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_activity_logs_activity_date
		ON activity_logs (activity_id, logged_date)
	`)

	return db, nil
}
