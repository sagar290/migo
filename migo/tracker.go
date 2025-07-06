package migo

type MigrationTracker interface {
	EnsureTable() error
	GetAppliedMigrations() ([]string, error)
	MarkAsApplied(name string, batch int) error
	MarkAsRolledBack(name string) error
	GetLastBatch() (int, error)
}
