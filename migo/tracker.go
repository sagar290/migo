package migo

import (
	"bufio"
	"bytes"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type MigrationTracker interface {
	//EnsureTable() error
	//GetAppliedMigrations() ([]string, error)
	//MarkAsApplied(name string, batch int) error
	//MarkAsRolledBack(name string) error
	GetLastBatch(db *gorm.DB) (int, error)
	PrepareAppliedMigrations(db *gorm.DB) error
	FilterNewMigrations() error
	ExtractUpBlock(sql []byte) string
	ExtractDownBlock(sql []byte) string
}

type AppliedMigration struct {
	ID        uint
	Migration string
	Batch     int
	CreatedAt time.Time
}

type Tracker struct {
	AppliedMigrations map[string]MigoMigration
	MigrationFiles    []string
	Config            *Config
}

//
//func (t *Tracker) EnsureTable() error {}
//
//func (t *Tracker) GetAppliedMigrations() ([]string, error) {
//
//	return t.AppliedMigrations, nil
//}
//
//func (t *Tracker) MarkAsApplied(name string, batch int) error {
//	t.AppliedMigrations = append(t.AppliedMigrations, name)
//	return nil
//}
//
//func (t *Tracker) MarkAsRolledBack(name string) error {
//
//	t.AppliedMigrations = append(t.AppliedMigrations, name)
//}

func (t *Tracker) GetLastBatch(db *gorm.DB) (int, error) {
	var lastBatch int
	db.Raw(`SELECT COALESCE(MAX(batch), 0) FROM ?`, t.Config.GetMigrationTable()).Scan(&lastBatch)
	return lastBatch, nil
}

func (t *Tracker) PrepareAppliedMigrations(db *gorm.DB) error {
	var records []MigoMigration
	results := db.Table(t.Config.GetMigrationTable()).Select("migration").Scan(&t.AppliedMigrations)
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

func (t *Tracker) ExtractUpBlock(sql []byte) string {
	var out bytes.Buffer

	scanner := bufio.NewScanner(strings.NewReader(string(sql)))

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

	return out.String()
}

func (t *Tracker) ExtractDownBlock(sql []byte) string {
	var out bytes.Buffer

	scanner := bufio.NewScanner(strings.NewReader(string(sql)))

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

	return out.String()
}
