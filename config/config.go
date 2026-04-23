package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/lminimum/LiteDock/pkg/errors"
)

type (
	Config struct {
		App     App
		HTTP    HTTP
		Log     Log
		DB      DB
		Metrics Metrics
		Swagger Swagger
		Cache   Cache
	}

	App struct {
		Name    string `env:"APP_NAME,required"`
		Version string `env:"APP_VERSION,required"`
	}

	HTTP struct {
		Port           string `env:"HTTP_PORT,required"`
		UsePreforkMode bool   `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
	}

	Log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	DB struct {
		Type      string `env:"DB_TYPE" envDefault:"sqlite"`
		PoolMax   int    `env:"DB_POOL_MAX,required"`
		URL       string `env:"DB_URL,required"`
		SQLiteDSN string `env:"SQLITE_DSN"`
	}

	Metrics struct {
		Enabled bool `env:"METRICS_ENABLED" envDefault:"true"`
	}

	Swagger struct {
		Enabled bool `env:"SWAGGER_ENABLED" envDefault:"false"`
	}

	Cache struct {
		ContainerTTL time.Duration `env:"CACHE_CONTAINER_TTL" envDefault:"30s"`
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
