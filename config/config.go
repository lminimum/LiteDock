package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/lminimum/LiteDock/pkg/errors"
)

type (
	// Config
	Config struct {
		App     App
		HTTP    HTTP
		Log     Log
		DB      DB
		Metrics Metrics
		Swagger Swagger
	}

	// App
	App struct {
		Name    string `env:"APP_NAME,required"`
		Version string `env:"APP_VERSION,required"`
	}

	// HTTP
	HTTP struct {
		Port           string `env:"HTTP_PORT,required"`
		UsePreforkMode bool   `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
	}

	// Log
	Log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	// DB
	DB struct {
		Type      string `env:"DB_TYPE" envDefault:"sqlite"` // postgres, mysql, sqlite
		PoolMax   int    `env:"DB_POOL_MAX,required"`
		URL       string `env:"DB_URL,required"`
		SQLiteDSN string `env:"SQLITE_DSN"`
	}

	// Metrics
	Metrics struct {
		Enabled bool `env:"METRICS_ENABLED" envDefault:"true"`
	}

	// Swagger
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
