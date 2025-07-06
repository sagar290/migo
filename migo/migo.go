package migo

import (
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"time"
)

type Runner struct {
	Db     *gorm.DB
	Config *Config
}

type MigoMigration struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Migration string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	Batch     int       `gorm:"not null;default:1"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func EnsureMigrationTable(db *gorm.DB) error {
	return db.AutoMigrate(&MigoMigration{})
}

func NewMigo(cfg *Config) (Migrator, error) {

	var db *gorm.DB
	var err error

	switch cfg.DBType {
	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.DBURL), &gorm.Config{})
	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.DBURL), &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DBURL), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported DB type: %s", cfg.DBType)
	}

	if err != nil {
		return nil, err
	}

	err = EnsureMigrationTable(db)
	if err != nil {
		log.Panicln("failed to create migration table:", err)
	}

	return &Runner{
		Db:     db,
		Config: cfg,
	}, nil
}

func (r *Runner) Up(ctx context.Context, db *gorm.DB) error {

	return nil
}

func (r *Runner) Down(ctx context.Context, db *gorm.DB) error {
	return nil
}

func (r *Runner) Step(ctx context.Context, n int, db *gorm.DB) error {

	return nil
}

func (r *Runner) Reset(ctx context.Context, db *gorm.DB) error {

	return nil
}

func (r *Runner) Fresh(ctx context.Context, db *gorm.DB) error {

	return nil
}

func (r *Runner) Rollback(ctx context.Context, db *gorm.DB) error {
	return nil
}

func (r *Runner) Status(ctx context.Context, db *gorm.DB) ([]MigrationStatus, error) {
	return nil, nil
}
