package cmd

import (
	"context"
	"fmt"
	"github.com/sagar290/migo/common"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func UpScript(_ *cobra.Command, _ []string) {

	ctx := context.Background()

	ctx = context.WithValue(ctx, common.StepsKey, steps)
	ctx = context.WithValue(ctx, common.DryRunKey, dryRun)

	err := migoInstance.Up(ctx)
	if err != nil {
		panic(err)
	}
}

func DownScript(_ *cobra.Command, _ []string) {

	ctx := context.Background()

	ctx = context.WithValue(ctx, common.StepsKey, steps)
	ctx = context.WithValue(ctx, common.DryRunKey, dryRun)

	err := migoInstance.Rollback(ctx)
	if err != nil {
		panic(err)
	}
}

func RefreshScript(_ *cobra.Command, _ []string) {

	ctx := context.Background()

	ctx = context.WithValue(ctx, common.DryRunKey, dryRun)

	err := migoInstance.Refresh(ctx)
	if err != nil {
		panic(err)
	}
}

func FreshScript(_ *cobra.Command, _ []string) {

	ctx := context.Background()

	err := migoInstance.Fresh(ctx)
	if err != nil {
		panic(err)
	}
}

var fileTemplate = `[UP]

[/UP]

[DOWN]

[/DOWN]
`

func MakeScript(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatal("❌ Please provide a migration description")
	}

	description := strings.Join(args, "_")
	description = strings.ToLower(strings.ReplaceAll(description, " ", "_"))

	timestamp := time.Now().Format("20060102150405")

	fileName := fmt.Sprintf("%s_%s.sql", timestamp, description)

	fullPath := filepath.Join(configInstance.GetMigrationDir(), fileName)

	if err := os.WriteFile(fullPath, []byte(fileTemplate), 0644); err != nil {
		log.Fatalf("❌ Failed to create file: %v", err)
	}

	log.Printf("✅ Created migration files:\n   %s\n", fileName)
}
