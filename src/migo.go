package src

import (
	"context"
	"fmt"
	"github.com/sagar290/migo/common"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

var db *gorm.DB

type Runner struct {
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
		Config:  cfg,
		Tracker: tracker,
	}, nil
}

// Up migrate table
func (r *Runner) Up(ctx context.Context) error {

	steps, _ := ctx.Value(common.StepsKey).(int)

	err := r.Tracker.InitTracker(ctx, db)
	if err != nil {
		return err
	}

	files := r.Tracker.GetMigrationFiles()

	// if steps provided limit the files
	if steps > 0 && steps < len(files) {
		files = files[:steps]
	}

	err = UpMigrationFiles(ctx, files, r)
	if err != nil {
		return err
	}

	return nil
}

// Rollback migration according to the batch id
func (r *Runner) Rollback(ctx context.Context) error {

	steps, _ := ctx.Value(common.StepsKey).(int)

	err := r.Tracker.InitTracker(ctx, db)
	if err != nil {
		return err
	}

	lastBatchId := r.Tracker.GetLastBatch()

	appliedFiles := r.Tracker.GetAppliedMigrationFileByBatchId(lastBatchId)

	// if steps provided limit the files
	if steps > 0 && steps < len(appliedFiles) {
		appliedFiles = appliedFiles[(len(appliedFiles) - steps):]
	}

	err = DownMigrationFiles(ctx, appliedFiles, r)

	if err != nil {
		return err
	}

	return nil
}

// Refresh rollback all table and run migrate
func (r *Runner) Refresh(ctx context.Context) error {

	err := r.Tracker.InitTracker(ctx, db)
	if err != nil {
		return err
	}

	appliedFiles := r.Tracker.GetAppliedMigrations()
	fmt.Println(appliedFiles)
	err = DownMigrationFiles(ctx, appliedFiles, r)
	if err != nil {
		return err
	}

	err = UpMigrationFiles(ctx, appliedFiles, r)
	if err != nil {
		return err
	}

	return nil
}

// Fresh drop all table and run migrate
func (r *Runner) Fresh(ctx context.Context) error {

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

	err = UpMigrationFiles(ctx, files, r)
	if err != nil {
		return err
	}

	return nil
}

// Status shows all applied migrations
func (r *Runner) Status(ctx context.Context) ([]MigoMigration, error) {
	return nil, nil
}

func UpMigrationFiles(ctx context.Context, files []string, r *Runner) error {
	dry, _ := ctx.Value(common.DryRunKey).(bool)
	if len(files) == 0 {
		log.Printf("🥂Nothing to migrate......")
		return nil
	}

	if dry {
		fmt.Printf("🔎 Dry run — %d migration(s) would run:\n", len(files))
	}

	for i, file := range files {
		queryText, err := r.Tracker.ExtractUpBlock(file)
		if err != nil {
			return err
		}

		if strings.TrimSpace(queryText) == "" {
			log.Println("⚠️ Skipping empty or no-up-block:", file)
			continue
		}

		if dry {
			fmt.Printf("  %2d) %s\n", i+1, queryText)
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

func DownMigrationFiles(ctx context.Context, appliedFiles []string, r *Runner) error {
	dry, _ := ctx.Value(common.DryRunKey).(bool)

	for i, file := range appliedFiles {
		queryText, err := r.Tracker.ExtractDownBlock(file)
		if err != nil {
			return err
		}

		if strings.TrimSpace(queryText) == "" {
			log.Println("⚠️ Skipping empty or no-down-block:", file)
			continue
		}

		if dry {
			fmt.Printf("  %2d) %s\n", i+1, queryText)
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

		log.Printf("⛔️%s", file)
	}
	return nil
}
