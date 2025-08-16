package cmd

import (
	"github.com/sagar290/migo/src"
	"github.com/spf13/cobra"
)

var migoInstance src.Migrator
var (
	steps  int
	dryRun bool
)

var RootCmd = &cobra.Command{
	Use:   "migo",
	Short: "A lightweight database migration tool for Go",
	Long: `
Migo is a simple and flexible database migration tool built for Go developers. 
It helps you create, run, and manage database migrations with ease, 
using a workflow inspired by tools like Laravel migrations. 

You can use 'migo' commands to initialize migrations, apply them, roll them back, 
and inspect migration history in your database.
	`,
}

var UpCommand = &cobra.Command{
	Use:   "up",
	Short: "Run pending database migrations",
	Long: `
The 'up' command applies all pending migrations to your database.

It will execute migration files in sequential order, track applied batches, 
and ensure your database schema stays up to date. 

Examples:
  migo up
  migo up --steps=1     # Run only the next migration
  migo up --dry-run     # Preview pending migrations without applying
	`,
	Run: UpScript,
}

var DownCommand = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last batch of migrations",
	Long: `
Reverts the most recent migrations applied to the database. Use --steps to rollback only N migrations.
	`,
	Run: DownScript,
}

func Init(migo src.Migrator) {

	UpCommand.Flags().IntVar(&steps, "steps", 0, "Number of migrations to run (0 = all)")
	UpCommand.Flags().BoolVar(&dryRun, "dry-run", false, "Preview pending migrations without applying")

	DownCommand.Flags().IntVar(&steps, "steps", 0, "Number of migrations to run (0 = all)")
	DownCommand.Flags().BoolVar(&dryRun, "dry-run", false, "Preview pending migrations without applying")

	RootCmd.AddCommand(UpCommand)
	RootCmd.AddCommand(DownCommand)
	migoInstance = migo
}
