package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type (
	Config struct {
		App           `yaml:"app"`
		HTTP          `yaml:"http"`
		GRPC          `yaml:"grpc"`
		PG            `yaml:"postgres"`
		Redis         `yaml:"redis"`
		JWT           `yaml:"jwt"`
		QR            `yaml:"qr"`
		MinIO         `yaml:"minio"`
		SMSAero       `yaml:"smsaero"`
		RabbitMQ      `yaml:"rabbitmq"`
		Firebase      `yaml:"firebase"`
		AdminPanelURL `yaml:"admin_panel_url"`
		FirstAdmin    `yaml:"full_admin"`
		SecondAdmin   `yaml:"second_admin"`
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

	GRPC struct {
		Port string
	}

	PG struct {
		URL            string
		MigrationsPath string
	}

	Redis struct {
		URL      string
		Password string
		DB       int
	}

	JWT struct {
		AccessSecret  string
		RefreshSecret string
		AccessTTL     time.Duration
		RefreshTTL    time.Duration
	}

	QR struct {
		HMACSecret string
	}

	MinIO struct {
		Endpoint      string
		AccessKey     string
		SecretKey     string
		UseSSL        bool
		BucketQR      string
		BucketRecords string
		PublicURL     string
	}

	SMSAero struct {
		BaseURL string
		Email   string
		APIKey  string
	}

	RabbitMQ struct {
		URL string
	}

	Firebase struct {
		CredentialsFile string
		ProjectID       string
	}

	AdminPanelURL struct {
		URL string
	}

	FirstAdmin struct {
		FullName  string
		Email     string
		Phone     string
		Password  string
		CreatedAt string
	}
	SecondAdmin struct {
		FullName  string
		Email     string
		Phone     string
		Password  string
		CreatedAt string
	}
)

func New() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: App{
			Name:     getEnv("APP_NAME", "skypost-delivery"),
			Version:  getEnv("APP_VERSION", "1.0.0"),
			GinMode:  getEnv("GIN_MODE", "realese"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		HTTP: HTTP{
			Port: getEnv("GO_ORCHESTRATOR_HTTP_PORT", "8080"),
		},
		GRPC: GRPC{
			Port: getEnv("GO_ORCHESTRATOR_GRPC_PORT", "50053"),
		},
		PG: PG{
			URL:            createDSN(),
			MigrationsPath: getEnv("MIGRATIONS_PATH", "migrations"),
		},
		Redis: Redis{
			URL:      getEnv("REDIS_URL", "redis://localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWT{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "your-secret-key-change-in-production"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key-change-in-production"),
			AccessTTL:     1 * 24 * time.Hour,
			RefreshTTL:    3 * 24 * time.Hour,
		},
		QR: QR{
			HMACSecret: getEnv("QR_HMAC_SECRET", "your-hmac-secret-key-change-in-production"),
		},
		MinIO: MinIO{
			Endpoint:      getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:     getEnv("MINIO_ROOT_USER", "minioadmin"),
			SecretKey:     getEnv("MINIO_ROOT_PASSWORD", "minioadmin"),
			UseSSL:        getEnv("MINIO_USE_SSL", "false") == "true",
			BucketQR:      getEnv("MINIO_BUCKET_QR_CODES", "qr-codes"),
			BucketRecords: getEnv("MINIO_BUCKET_RECORDS", "records"),
			PublicURL:     getEnv("MINIO_PUBLIC_URL", "http://localhost:9000"),
		},
		SMSAero: SMSAero{
			BaseURL: getEnv("SMSAERO_URL", ""),
			Email:   getEnv("SMSAERO_EMAIL", ""),
			APIKey:  getEnv("SMSAERO_API_KEY", ""),
		},
		RabbitMQ: RabbitMQ{
			URL: getEnv("RABBITMQ_URL", "amqp://admin:admin@localhost:5672/"),
		},
		Firebase: Firebase{
			CredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE_IN_DOCKER", ""),
			ProjectID:       getEnv("FIREBASE_PROJECT_ID", ""),
		},
		AdminPanelURL: AdminPanelURL{
			URL: getEnv("ADMIN_PANEL_URL", "http://localhost:3000"),
		},
		FirstAdmin: FirstAdmin{
			FullName:  getEnv("ADMIN_FULLNAME", ""),
			Email:     getEnv("ADMIN_EMAIL", ""),
			Phone:     getEnv("ADMIN_PHONE", ""),
			Password:  getEnv("ADMIN_PASSWORD", ""),
			CreatedAt: getEnv("ADMIN_CREATED_AT", ""),
		},
		SecondAdmin: SecondAdmin{
			FullName:  getEnv("SECOND_ADMIN_FULLNAME", ""),
			Email:     getEnv("SECOND_ADMIN_EMAIL", ""),
			Phone:     getEnv("SECOND_ADMIN_PHONE", ""),
			Password:  getEnv("SECOND_ADMIN_PASSWORD", ""),
			CreatedAt: getEnv("SECOND_ADMIN_CREATED_AT", ""),
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
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return intValue
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
