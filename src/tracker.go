package migo

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type AppliedMigration struct {
	ID        uint
	Migration string
	Batch     int
	CreatedAt time.Time
}

type Tracker struct {
	AppliedMigrations map[string]MigoMigration
	MigrationFiles    []string
	LastBatch         int
	Config            *Config
}

func NewTracker(config *Config) *Tracker {
	return &Tracker{
		Config: config,
	}
}

func (t *Tracker) InitTracker(ctx context.Context, db *gorm.DB) error {
	err := t.PrepareAppliedMigrations(ctx, db)
	if err != nil {
		return fmt.Errorf("get last batch: %w", err)
	}

	err = t.ListSqlFiles()
	if err != nil {
		return fmt.Errorf("ListSqlFiles: %w", err)
	}

	err = t.FilterNewMigrations()
	if err != nil {
		return fmt.Errorf("FilterNewMigrations: %w", err)
	}

	return nil
}

func (t *Tracker) GetLastBatch() int {
	return t.LastBatch
}

func (t *Tracker) PrepareAppliedMigrations(ctx context.Context, db *gorm.DB) error {

	var records []MigoMigration

	// prepare last batch id
	err := db.Raw(`SELECT COALESCE(MAX(batch), 0) FROM ?`, t.Config.GetMigrationTable()).Scan(&t.LastBatch).Error
	if err != nil {
		return fmt.Errorf("get last batch: %w", err)
	}

	// prepare migration files
	results := db.WithContext(ctx).Table(t.Config.GetMigrationTable()).Select("migration").Scan(&t.AppliedMigrations)
	if results.Error != nil {
		return results.Error
	}

	for _, r := range records {
		t.AppliedMigrations[r.Migration] = r
	}

	return nil
}

func (t *Tracker) FilterNewMigrations() error {
	var filtered []string
	for _, file := range t.MigrationFiles {
		if _, ok := t.AppliedMigrations[file]; ok {
			filtered = append(filtered, file)
		}
	}

	t.MigrationFiles = filtered

	return nil
}

func (t *Tracker) ListSqlFiles() error {
	err := filepath.Walk(t.Config.MigrationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(info.Name()) == ".sql" {
			t.MigrationFiles = append(t.MigrationFiles, path)
		}

		return nil
	})

	if err != nil {
		return err
	}

	sort.Strings(t.MigrationFiles)

	return nil
}

func (t *Tracker) ExtractUpBlock(file string) (string, error) {

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", file, err)
	}

	var out bytes.Buffer

	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	insideUpBlock := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		switch trimmedLine {
		case "[UP]":
			insideUpBlock = true
		case "[/UP]":
			insideUpBlock = false
		case "[DOWN]":
			insideUpBlock = false
		default:
			if insideUpBlock {
				out.WriteString(line + "\n")
			}
		}
	}

	return out.String(), nil
}

func (t *Tracker) ExtractDownBlock(file string) (string, error) {

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", file, err)
	}

	var out bytes.Buffer

	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	insideUpBlock := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		switch trimmedLine {
		case "[DOWN]":
			insideUpBlock = true
		case "[/DOWN]":
			insideUpBlock = false
		case "[UP]":
			insideUpBlock = false
		default:
			if insideUpBlock {
				out.WriteString(line + "\n")
			}
		}
	}

	return out.String(), nil
}

func (t *Tracker) GetMigrationFiles() []string {
	return t.MigrationFiles
}

func (t *Tracker) GetAppliedMigrations() []string {

	var files []string

	for _, file := range t.MigrationFiles {
		files = append(files, file)
	}

	return files
}

func (t *Tracker) AddMigrationInfo(ctx context.Context, db *gorm.DB, file string) error {

	if err := db.WithContext(ctx).Create(&MigoMigration{
		Migration: file,
		Batch:     t.GetLastBatch() + 1,
	}).Error; err != nil {
		log.Fatalf("❌ Failed to add migration: %v", err)
		return err
	}

	return nil
}

func (t *Tracker) RemoveMigrationInfo(ctx context.Context, db *gorm.DB, file string) error {

	if err := db.Where("migration = ?", file).Delete(&MigoMigration{}).Error; err != nil {
		log.Fatalf("❌ Failed to delete migration: %v", err)
		return err
	}

	return nil
}

func (t *Tracker) GetAppliedMigrationFileByBatchId(batchId int) []string {

	var files []string

	for _, migration := range t.AppliedMigrations {
		if migration.Batch == batchId {
			files = append(files, migration.Migration)
		}
	}

	return files
}
