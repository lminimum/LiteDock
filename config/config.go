package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/lminimum/LiteDock/pkg/errors"
)

type (
	// Config holds all application configuration settings.
	Config struct {
		App     App
		HTTP    HTTP
		Log     Log
		DB      DB
		Metrics Metrics
		Swagger Swagger
	}

	// App holds application metadata configuration.
	App struct {
		Name    string `env:"APP_NAME,required"`
		Version string `env:"APP_VERSION,required"`
	}

	// HTTP holds HTTP server configuration.
	HTTP struct {
		Port           string `env:"HTTP_PORT,required"`
		UsePreforkMode bool   `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
	}

	// Log holds logging configuration settings.
	Log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	// DB holds database connection and pool configuration.
	DB struct {
		Type      string `env:"DB_TYPE" envDefault:"sqlite"` // postgres, mysql, sqlite
		PoolMax   int    `env:"DB_POOL_MAX,required"`
		URL       string `env:"DB_URL,required"`
		SQLiteDSN string `env:"SQLITE_DSN"`
	}

	// Metrics holds application metrics collection configuration.
	Metrics struct {
		Enabled bool `env:"METRICS_ENABLED" envDefault:"true"`
	}

	// Swagger holds API documentation configuration.
	Swagger struct {
		Enabled bool `env:"SWAGGER_ENABLED" envDefault:"false"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, errors.Wrap(err, "Config.New.Parse")
	}

	return cfg, nil
}
