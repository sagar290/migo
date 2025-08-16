package migo

import (
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

type Runner struct {
	Db      *gorm.DB
	Config  *Config
	Tracker MigrationTracker
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

func NewMigo(cfg *Config, tracker *Tracker) (Migrator, error) {

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
		Db:      db,
		Config:  cfg,
		Tracker: tracker,
	}, nil
}

// Up migrate table
func (r *Runner) Up(ctx context.Context, db *gorm.DB) error {

	err := r.Tracker.InitTracker(ctx, db)
	if err != nil {
		return err
	}

	files := r.Tracker.GetMigrationFiles()

	err = UpMigrationFiles(ctx, db, files, r)
	if err != nil {
		return err
	}

	return nil
}

// Rollback migration according to the batch id
func (r *Runner) Rollback(ctx context.Context, db *gorm.DB) error {

	err := r.Tracker.InitTracker(ctx, db)
	if err != nil {
		return err
	}

	lastBatchId := r.Tracker.GetLastBatch()

	appliedFiles := r.Tracker.GetAppliedMigrationFileByBatchId(lastBatchId)

	err = DownMigrationFiles(ctx, db, appliedFiles, r)
	if err != nil {
		return err
	}

	return nil
}

// Refresh rollback all table and run migrate
func (r *Runner) Refresh(ctx context.Context, db *gorm.DB) error {

	err := r.Tracker.InitTracker(ctx, db)
	if err != nil {
		return err
	}

	appliedFiles := r.Tracker.GetAppliedMigrations()

	err = DownMigrationFiles(ctx, db, appliedFiles, r)
	if err != nil {
		return err
	}

	err = UpMigrationFiles(ctx, db, appliedFiles, r)
	if err != nil {
		return err
	}

	return nil
}

// Fresh drop all table and run migrate
func (r *Runner) Fresh(ctx context.Context, db *gorm.DB) error {

	//todo: will be different for mysql
	var tables []string
	db.Raw(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = ?
		  AND tablename != ?
	`, r.Config.GetSchemaName(), r.Config.GetMigrationTable()).Scan(&tables)

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			log.Printf("❌ Failed to drop table %s: %v", table, err)
		} else {
			log.Printf("✅ Dropped table %s", table)
		}
	}

	// fresh the migration table
	err := db.Exec(`TRUNCATE TABLE ? RESTART IDENTITY CASCADE`, r.Config.GetMigrationTable()).Error
	if err != nil {
		log.Fatalf("Failed to truncate table: %v", err)
	}

	err = r.Tracker.InitTracker(ctx, db)
	if err != nil {
		return err
	}

	files := r.Tracker.GetMigrationFiles()

	err = UpMigrationFiles(ctx, db, files, r)
	if err != nil {
		return err
	}

	return nil
}

// Status shows all applied migrations
func (r *Runner) Status(ctx context.Context, db *gorm.DB) ([]MigoMigration, error) {
	return nil, nil
}

func UpMigrationFiles(ctx context.Context, db *gorm.DB, files []string, r *Runner) error {
	for _, file := range files {
		queryText, err := r.Tracker.ExtractUpBlock(file)
		if err != nil {
			return err
		}

		if strings.TrimSpace(queryText) == "" {
			log.Println("⚠️ Skipping empty or no-up-block:", file)
			continue
		}

		if err := db.Exec(queryText).Error; err != nil {
			log.Printf("❌ Failed to execute %s: %v\n", file, err)
			continue
		}

		err = r.Tracker.AddMigrationInfo(ctx, db, file)
		if err != nil {
			return err
		}

		log.Printf("✅ %s", file)
	}
	return nil
}

func DownMigrationFiles(ctx context.Context, db *gorm.DB, appliedFiles []string, r *Runner) error {
	for _, file := range appliedFiles {
		queryText, err := r.Tracker.ExtractDownBlock(file)
		if err != nil {
			return err
		}

		if strings.TrimSpace(queryText) == "" {
			log.Println("⚠️ Skipping empty or no-up-block:", file)
			continue
		}

		if err := db.Exec(queryText).Error; err != nil {
			log.Printf("❌ Failed to execute %s: %v\n", file, err)
			continue
		}

		err = r.Tracker.RemoveMigrationInfo(ctx, db, file)
		if err != nil {
			return err
		}
	}
	return nil
}
