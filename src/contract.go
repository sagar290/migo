package src

import (
	"context"
	"gorm.io/gorm"
)

type Migrator interface {
	Up(ctx context.Context) error
	Rollback(ctx context.Context) error
	Refresh(ctx context.Context) error
	Fresh(ctx context.Context) error
	Status(ctx context.Context) ([]MigoMigration, error)
}

type MigrationTracker interface {
	GetLastBatch() int
	PrepareAppliedMigrations(ctx context.Context, db *gorm.DB) error
	FilterNewMigrations() error
	ExtractUpBlock(file string) (string, error)
	ExtractDownBlock(file string) (string, error)
	InitTracker(ctx context.Context, db *gorm.DB) error
	GetMigrationFiles() []string
	GetAppliedMigrations() []string
	AddMigrationInfo(ctx context.Context, db *gorm.DB, file string) error
	RemoveMigrationInfo(ctx context.Context, db *gorm.DB, file string) error
	ListSqlFiles() error
	GetAppliedMigrationFileByBatchId(batchId int) []string
}
