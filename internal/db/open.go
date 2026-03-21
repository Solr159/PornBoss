package db

import (
	"fmt"

	"pornboss/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// Open initialises a GORM-backed SQLite database and applies schema migrations.
func Open(path string) (*gorm.DB, error) {
	driverName := registerSQLiteFunctions()
	db, err := gorm.Open(sqlite.New(sqlite.Config{DriverName: driverName, DSN: path}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrent read/write performance.
	if err := db.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
		return nil, fmt.Errorf("enable WAL: %w", err)
	}
	if err := db.Exec("PRAGMA foreign_keys=ON;").Error; err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := db.AutoMigrate(
		&models.Directory{},
		&models.Video{},
		&models.Tag{},
		&models.VideoTag{},
		&models.Config{},
		&models.Jav{},
		&models.JavTag{},
		&models.JavIdol{},
		&models.JavTagMap{},
		&models.JavIdolMap{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	if err := db.Exec("UPDATE video SET play_count = 0 WHERE play_count IS NULL").Error; err != nil {
		return nil, fmt.Errorf("backfill play_count: %w", err)
	}
	return db, nil
}
