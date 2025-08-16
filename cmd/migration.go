package cmd

import (
	"context"
	"github.com/sagar290/migo/common"
	"github.com/spf13/cobra"
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
