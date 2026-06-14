package database

import (
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"aetheris/internal/notification"
)

func Open(databaseURL string) (*gorm.DB, error) {
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		return gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	}

	// Default to SQLite (remove sqlite:// prefix if present)
	dbFile := strings.TrimPrefix(databaseURL, "sqlite://")
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Optimize SQLite performance and safety settings
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA foreign_keys=ON;")
	db.Exec("PRAGMA busy_timeout=5000;")

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return notification.AutoMigrate(db)
}
