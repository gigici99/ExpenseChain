package db

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect opens the SQLite database at the given path and returns a GORM instance.
// Panics on connection failure — no point continuing without a DB.
func Connect(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, errors.New("db: DSN cannot be empty")
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "[DB] ", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("db: failed to open '%s': %w", dsn, err)
	}

	// keep a single connection — SQLite does not benefit from pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("db: failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)

	log.Printf("[DB] connected to %s", dsn)
	DB = db
	return db, nil
}

// MustConnect calls Connect and panics on error.
// Use at application startup where DB is required.
func MustConnect(dsn string) *gorm.DB {
	db, err := Connect(dsn)
	if err != nil {
		log.Fatalf("[DB] fatal: %v", err)
	}
	return db
}

// Close closes the underlying sql.DB connection pool.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("db: failed to retrieve sql.DB for close: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("db: close error: %w", err)
	}
	log.Println("[DB] connection closed")
	return nil
}
