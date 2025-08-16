package src

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"strings"
)

type Config struct {
	DBType        string `mapstructure:"db_type"`
	DBURL         string `mapstructure:"db_url"`
	MigrationsDir string `mapstructure:"migrations_dir"`
	LogLevel      string `mapstructure:"log_level"`
}

func LoadConfig() (*Config, error) {

	v := viper.New()

	v.SetConfigName("migo")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	v.AutomaticEnv()
	v.SetEnvPrefix("MIGO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		log.Println("⚠️ migo.yaml not found or failed to load:", err)
	}

	sub := v.Sub("migo")
	if sub == nil {
		log.Println("⚠️ No 'migo' section found in config file")
		sub = v
	}

	var cfg *Config
	if err := sub.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.DBType == "" || cfg.DBURL == "" {
		return nil, fmt.Errorf("missing required env: MIGO_DB_TYPE or MIGO_DB_URL")
	}

	return cfg, nil
}

func (cfg *Config) GetMigrationTable() string {
	return "migo_migrations"
}

func (cfg *Config) GetSchemaName() string {
	return "public"
}
