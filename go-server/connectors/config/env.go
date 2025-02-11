package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

/*
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
*/

// Env is the expected config values from the process's environment
type ApplicationEnvironment string

const (
	Dev        = "dev"
	Staging    = "staging"
	Production = "production"
)

type Env struct {
	AppEnv   ApplicationEnvironment `default:"dev" split_words:"true"`
	Name     string                 `envconfig:"SERVICE_NAME" required:"true"`
	LogLevel string                 `envconfig:"LOG_LEVEL" required:"true" split_words:"true"`

	PostgresHost       string `required:"true" split_words:"true"`
	PostgresPort       string `required:"true" split_words:"true"`
	PostgresPoolSize   int    `required:"true" split_words:"true"`
	PostgresSecureMode bool   `required:"true" split_words:"true"`
	PostgresUser       string `required:"true" split_words:"true"`
	PostgresPassword   string `required:"true" split_words:"true"`
	PostgresDatabase   string `required:"true" split_words:"true"`
	PostgresDebug      bool   `default:"false" split_words:"true"`

	AWSRegion            string `envconfig:"AWS_REGION" required:"true" split_words:"true"`
	AWSEndpoint          string `envconfig:"AWS_ENDPOINT" required:"true" split_words:"true"`
	AWSForcePathStyle    bool   `envconfig:"AWS_FORCE_PATH_STYLE" required:"true" split_words:"true"`
	AWSCredentialsID     string `envconfig:"AWS_CREDENTIALS_ID" required:"true" split_words:"true"`
	AWSCredentialsSecret string `envconfig:"AWS_CREDENTIALS_SECRET" required:"true" split_words:"true"`
	AWSCredentialsToken  string `envconfig:"AWS_CREDENTIALS_TOKEN" required:"false" split_words:"true"`

	RPCGracefulShutdownTimeout int    `envconfig:"RPC_GRACEFUL_SHUTDOWN_TIMEOUT" required:"true" split_words:"true"`
	RPCPort                    string `envconfig:"RPC_PORT" required:"true" split_words:"true"`
}

func LoadEnv(env *Env) error {
	// try to load from .env first
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}

	err := godotenv.Load(".env." + appEnv + ".local")
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok || !errors.Is(perr.Unwrap(), os.ErrNotExist) {
			return err
		}
	}

	if appEnv != "test" {
		godotenv.Load(".env.local")
	}
	err = godotenv.Load(".env." + appEnv)
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok || !errors.Is(perr.Unwrap(), os.ErrNotExist) {
			return err
		}
	}
	err = godotenv.Load() // The Original .env
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok || !errors.Is(perr.Unwrap(), os.ErrNotExist) {
			return err
		}
	}

	if err := envconfig.Process("", env); err != nil {
		return err
	}

	return nil
}
