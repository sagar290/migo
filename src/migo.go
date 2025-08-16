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

	err := DropTableByDialect(r.Config.GetSchemaName(), r.Config.GetMigrationTable())
	if err != nil {
		return err
	}

	// fresh the migration table
	query := fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE`, r.Config.GetMigrationTable())
	err = db.Exec(query).Error
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

func DropTableByDialect(schemaName, migrationsTable string) error {
	var tables []struct {
		Schema string
		Name   string
	}

	switch db.Dialector.Name() {

	// -------------------- PostgreSQL (also covers CockroachDB under "postgres") --------------------
	case "postgres":
		// list base tables in a schema (skip views), exclude migrations table
		if err := db.Raw(`
			SELECT table_schema AS schema, table_name AS name
			FROM information_schema.tables
			WHERE table_type = 'BASE TABLE'
			  AND table_schema = ?
			  AND table_name <> ?
		`, schemaName, migrationsTable).Scan(&tables).Error; err != nil {
			return err
		}

		for _, t := range tables {
			qualified := fmt.Sprintf(`%q.%q`, t.Schema, t.Name)
			q := fmt.Sprintf(`DROP TABLE IF EXISTS %s CASCADE`, qualified)
			if err := db.Exec(q).Error; err != nil {
				log.Printf("❌ Failed to drop %s: %v", qualified, err)
			} else {
				log.Printf("⛔️ Dropped %s", qualified)
			}
		}

	// -------------------- MySQL / MariaDB --------------------
	case "mysql":
		// disable FK checks so drop order doesn't matter
		if err := db.Exec(`SET FOREIGN_KEY_CHECKS = 0`).Error; err != nil {
			return err
		}
		defer db.Exec(`SET FOREIGN_KEY_CHECKS = 1`)

		// schemaName here is the database name
		if err := db.Raw(`
			SELECT table_schema AS schema, table_name AS name
			FROM information_schema.tables
			WHERE table_schema = ?
			  AND table_type = 'BASE TABLE'
			  AND table_name <> ?
		`, schemaName, migrationsTable).Scan(&tables).Error; err != nil {
			return err
		}

		for _, t := range tables {
			q := fmt.Sprintf("DROP TABLE IF EXISTS `%s`.`%s`", schemaName, t.Name)
			if err := db.Exec(q).Error; err != nil {
				log.Printf("❌ Failed to drop %s.%s: %v", schemaName, t.Name, err)
			} else {
				log.Printf("⛔️ Dropped %s.%s", schemaName, t.Name)
			}
		}

	// -------------------- SQLite --------------------
	case "sqlite":
		// turn off FKs to avoid dependency errors
		if err := db.Exec(`PRAGMA foreign_keys = OFF`).Error; err != nil {
			return err
		}
		defer db.Exec(`PRAGMA foreign_keys = ON`)

		// schemaName is ignored in SQLite; it’s a single-file DB
		if err := db.Raw(`
			SELECT name AS name
			FROM sqlite_master
			WHERE type = 'table'
			  AND name <> ?
			  AND name NOT LIKE 'sqlite_%'
		`, migrationsTable).Scan(&tables).Error; err != nil {
			return err
		}

		for _, t := range tables {
			q := fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, t.Name)
			if err := db.Exec(q).Error; err != nil {
				log.Printf("❌ Failed to drop %s: %v", t.Name, err)
			} else {
				log.Printf("⛔️Dropped %s", t.Name)
			}
		}

	// -------------------- SQL Server --------------------
	case "sqlserver":
		// Default schema is usually dbo; pass via schemaName.
		// Gather base tables for the schema (exclude views & migrations table)
		if err := db.Raw(`
			SELECT TABLE_SCHEMA AS schema, TABLE_NAME AS name
			FROM INFORMATION_SCHEMA.TABLES
			WHERE TABLE_TYPE = 'BASE TABLE'
			  AND TABLE_SCHEMA = ?
			  AND TABLE_NAME <> ?
		`, schemaName, migrationsTable).Scan(&tables).Error; err != nil {
			return err
		}

		// Disable constraints per table (SQL Server doesn't have DROP ... CASCADE)
		for _, t := range tables {
			qualified := fmt.Sprintf(`[%s].[%s]`, t.Schema, t.Name)

			// Disable all constraints to avoid FK issues
			if err := db.Exec(fmt.Sprintf(`ALTER TABLE %s NOCHECK CONSTRAINT ALL`, qualified)).Error; err != nil {
				// not fatal; try to drop anyway
				log.Printf("⚠️  Could not disable constraints on %s: %v", qualified, err)
			}

			// SQL Server 2016+ supports DROP TABLE IF EXISTS
			drop := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, qualified)
			if err := db.Exec(drop).Error; err != nil {
				log.Printf("❌ Failed to drop %s: %v", qualified, err)
			} else {
				log.Printf("⛔️ Dropped %s", qualified)
			}
		}

	default:
		return fmt.Errorf("unsupported dialect: %s", db.Dialector.Name())
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
