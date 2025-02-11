package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	GRPCGracefulShutdownTimeout int64
	LogLevel                    string
	ServiceName                 string
	DBUrl                       string
	LocalStackEndpoint          string
}

var Envs = initConfig()

func initConfig() Config {
	if err := godotenv.Load(); err != nil {
		slog.Error("Error loading .env file", "err", err)
		os.Exit(1)
	}

	return Config{
		ServiceName:                 getEnv("SERVICE_NAME", "slack-connector"),
		LogLevel:                    getEnv("LOG_LEVEL", "info"),
		DBUrl:                       getEnv("DATABASE_URL", ""),
		GRPCGracefulShutdownTimeout: getEnvAsInt("GRPC_GRACEFUL_SHUTDOWN_TIMEOUT", 5),
		LocalStackEndpoint:          getEnv("LOCALSTACK_ENDPOINT", "http://localstack:4566"),
	}
}

// Gets the env by key or fallbacks
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}

		return i
	}

	return fallback
}
