package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl                       string
	GRPCGracefulShutdownTimeout int64
}

var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()

	return Config{
		DBUrl:                       getEnv("DATABASE_URL", ""),
		GRPCGracefulShutdownTimeout: getEnvAsInt("GRPC_GRACEFUL_SHUTDOWN_TIMEOUT", 10),
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
