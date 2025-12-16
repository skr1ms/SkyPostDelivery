package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type (
	Config struct {
		App              `yaml:"app"`
		HTTP             `yaml:"http"`
		PG               `yaml:"postgres"`
		RabbitMQ         `yaml:"rabbitmq"`
		MinIO            `yaml:"minio"`
		OrchestratorGRPC `yaml:"orchestrator_grpc"`
		WebSocket        `yaml:"websocket"`
	}

	App struct {
		Name     string
		Version  string
		GinMode  string
		LogLevel string
	}

	HTTP struct {
		Port string
	}

	PG struct {
		URL string
	}

	RabbitMQ struct {
		URL string
	}

	MinIO struct {
		Endpoint      string
		RootUser      string
		RootPassword  string
		UseSSL        bool
		BucketRecords string
	}

	OrchestratorGRPC struct {
		URL string
	}

	WebSocket struct {
		BroadcastInterval int
	}
)

func New() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: App{
			Name:     getEnv("APP_NAME", "skypost-delivery"),
			Version:  getEnv("APP_VERSION", "1.0.0"),
			GinMode:  getEnv("GIN_MODE", "release"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		HTTP: HTTP{
			Port: getEnv("DRONE_SERVICE_HTTP_PORT", "8081"),
		},
		PG: PG{
			URL: createDSN(),
		},
		RabbitMQ: RabbitMQ{
			URL: getEnv("RABBITMQ_URL", "amqp://admin:admin@localhost:5672/"),
		},
		MinIO: MinIO{
			Endpoint:      getEnv("MINIO_ENDPOINT", "localhost:9000"),
			RootUser:      getEnv("MINIO_ROOT_USER", "admin"),
			RootPassword:  getEnv("MINIO_ROOT_PASSWORD", "admin"),
			UseSSL:        getEnv("MINIO_USE_SSL", "false") == "true",
			BucketRecords: getEnv("MINIO_BUCKET_RECORDS", "records"),
		},
		OrchestratorGRPC: OrchestratorGRPC{
			URL: getEnv("GRPC_GO_ORCHESTRATOR_URL", "localhost:50052"),
		},
		WebSocket: WebSocket{
			BroadcastInterval: getEnvInt("WEBSOCKET_BROADCAST_INTERVAL", 5),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func createDSN() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("POSTGRES_USER", "postgres"),
		getEnv("POSTGRES_PASSWORD", "postgres"),
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_PORT", "5432"),
		getEnv("POSTGRES_DB", "skypost-delivery"),
	)
}
