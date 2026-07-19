package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	SQLite     SQLiteConfig
	JWT        JWTConfig
	Accounting AccountingConfig
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type DatabaseConfig struct {
	URL             string        `mapstructure:"url"`
	MaxConns        int32         `mapstructure:"max_conns"`
	MinConns        int32         `mapstructure:"min_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `mapstructure:"max_conn_idle_time"`
}

type SQLiteConfig struct {
	Path string `mapstructure:"path"`
	Mode bool   `mapstructure:"mode"` // true = use SQLite instead of PostgreSQL
}

type JWTConfig struct {
	Secret   string        `mapstructure:"secret"`
	TTLHours time.Duration `mapstructure:"ttl_hours"`
}

type AccountingConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.env", "development")
	v.SetDefault("database.max_conns", 25)
	v.SetDefault("database.min_conns", 5)
	v.SetDefault("database.max_conn_lifetime", "1h")
	v.SetDefault("database.max_conn_idle_time", "30m")
	v.SetDefault("sqlite.path", "./local.db")
	v.SetDefault("sqlite.mode", false)
	v.SetDefault("jwt.ttl_hours", "8h")
	v.SetDefault("accounting.enabled", false)

	// Bind env vars
	_ = v.BindEnv("database.url", "DATABASE_URL")
	_ = v.BindEnv("sqlite.path", "SQLITE_PATH")
	_ = v.BindEnv("sqlite.mode", "SQLITE_MODE")
	_ = v.BindEnv("jwt.secret", "JWT_SECRET")
	_ = v.BindEnv("jwt.ttl_hours", "JWT_TTL_HOURS")
	_ = v.BindEnv("server.port", "SERVER_PORT")
	_ = v.BindEnv("server.env", "SERVER_ENV")
	_ = v.BindEnv("accounting.enabled", "ACCOUNTING_ENABLED")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if !cfg.SQLite.Mode && cfg.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required when SQLITE_MODE is false")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	return nil
}

func (c *Config) IsDev() bool {
	return c.Server.Env == "development"
}

func (c *Config) IsProd() bool {
	return c.Server.Env == "production"
}
