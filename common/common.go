package common

import "os"

type ctxKey string

const (
	StepsKey  ctxKey = "steps"
	DryRunKey ctxKey = "dryRun"
)

func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
