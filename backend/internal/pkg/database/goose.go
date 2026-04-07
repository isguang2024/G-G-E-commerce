// goose.go wires the goose migration runner. Phase 2a only runs the
// tenant baseline migration; the existing GORM AutoMigrate continues to
// own the rest of the schema until later phases retire it module-by-module.
package database

import (
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunGooseMigrations executes all SQL migrations under backend/migrations
// against the live *gorm.DB connection.
func RunGooseMigrations(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
