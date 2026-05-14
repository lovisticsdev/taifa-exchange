package config

import (
	"os"
	"time"
)

type Config struct {
	ServiceName string
	Environment string
	HTTP        HTTPConfig
}

type HTTPConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func Load() Config {
	return Config{
		ServiceName: envString("TAIFA_EXCHANGE_SERVICE_NAME", "taifa-exchange"),
		Environment: envString("TAIFA_EXCHANGE_ENV", "local"),
		HTTP: HTTPConfig{
			Addr:            envString("TAIFA_EXCHANGE_HTTP_ADDR", ":8081"),
			ReadTimeout:     envDuration("TAIFA_EXCHANGE_HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    envDuration("TAIFA_EXCHANGE_HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     envDuration("TAIFA_EXCHANGE_HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: envDuration("TAIFA_EXCHANGE_HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
	}
}

func envString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
