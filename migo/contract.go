package migo

import (
	"context"
	"gorm.io/gorm"
)

type Migrator interface {
	Up(ctx context.Context, db *gorm.DB) error
	Rollback(ctx context.Context, db *gorm.DB) error
	Refresh(ctx context.Context, db *gorm.DB) error
	Fresh(ctx context.Context, db *gorm.DB) error
	Status(ctx context.Context, db *gorm.DB) ([]MigoMigration, error)
}

type MigrationTracker interface {
	GetLastBatch() int
	PrepareAppliedMigrations(ctx context.Context, db *gorm.DB) error
	FilterNewMigrations() error
	ExtractUpBlock(file string) (string, error)
	ExtractDownBlock(file string) (string, error)
	InitTracker(ctx context.Context, db *gorm.DB) error
	GetMigrationFiles() []string
	GetAppliedMigrations() map[string]MigoMigration
	AddMigrationInfo(ctx context.Context, db *gorm.DB, file string) error
	RemoveMigrationInfo(ctx context.Context, db *gorm.DB, file string) error
	ListSqlFiles() error
	GetAppliedMigrationFileByBatchId(batchId int) []string
}
