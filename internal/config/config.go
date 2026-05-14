package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName string
	Environment string
	HTTP        HTTPConfig
	Database    DatabaseConfig
	TaifaID     TaifaIDConfig
}

type HTTPConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	DSN            string
	MinConns       int32
	MaxConns       int32
	ConnectTimeout time.Duration
}

type TaifaIDConfig struct {
	BaseURL string
	Timeout time.Duration
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
		Database: DatabaseConfig{
			DSN:            envString("TAIFA_EXCHANGE_DATABASE_DSN", ""),
			MinConns:       envInt32("TAIFA_EXCHANGE_DATABASE_MIN_CONNS", 1),
			MaxConns:       envInt32("TAIFA_EXCHANGE_DATABASE_MAX_CONNS", 5),
			ConnectTimeout: envDuration("TAIFA_EXCHANGE_DATABASE_CONNECT_TIMEOUT", 5*time.Second),
		},
		TaifaID: TaifaIDConfig{
			BaseURL: envString("TAIFA_EXCHANGE_TAIFA_ID_BASE_URL", ""),
			Timeout: envDuration(
				"TAIFA_EXCHANGE_TAIFA_ID_TIMEOUT",
				10*time.Second,
			),
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

func envInt32(key string, fallback int32) int32 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return fallback
	}

	return int32(parsed)
}
