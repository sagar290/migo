package migo

import (
	"context"
	"gorm.io/gorm"
)

type Migrator interface {
	Up(ctx context.Context, db *gorm.DB) error
	Down(ctx context.Context, db *gorm.DB) error
	Rollback(ctx context.Context, db *gorm.DB) error
	Step(ctx context.Context, n int, db *gorm.DB) error
	Reset(ctx context.Context, db *gorm.DB) error
	Fresh(ctx context.Context, db *gorm.DB) error
	Status(ctx context.Context, db *gorm.DB) ([]MigrationStatus, error)
}
