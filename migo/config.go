package migo

import (
	"errors"
	"github.com/sagar290/migo/common"
)

type Config struct {
	DBType        string `mapstructure:"db_type"`
	DBURL         string `mapstructure:"db_url"`
	MigrationsDir string `mapstructure:"migrations_dir"`
	LogLevel      string `mapstructure:"log_level"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DBType:        common.GetEnv("MIGO_DB_TYPE", ""),
		DBURL:         common.GetEnv("MIGO_DB_URL", ""),
		MigrationsDir: common.GetEnv("MIGO_MIGRATIONS_DIR", "migrations"),
		LogLevel:      common.GetEnv("MIGO_LOG_LEVEL", "info"),
	}

	if cfg.DBType == "" || cfg.DBURL == "" {
		return nil, errors.New("missing required env: MIGO_DB_TYPE or MIGO_DB_URL")
	}

	return cfg, nil
}

func (cfg *Config) GetMigrationTable() string {
	return "migo_migrations"
}

func (cfg *Config) GetSchemaName() string {
	return "public"
}
