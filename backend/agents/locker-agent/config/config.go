package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	entityError "github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity/error"
)

type Config struct {
	HTTP         HTTPConfig
	Orchestrator OrchestratorConfig
	Arduino      ArduinoConfig
	Display      DisplayConfig
	Camera       CameraConfig
	Serial       SerialConfig
	Log          LogConfig
	Service      ServiceConfig
}

type HTTPConfig struct {
	Port string
}

type OrchestratorConfig struct {
	URL        string
	Timeout    time.Duration
	RetryCount int
}

type ArduinoConfig struct {
	Port     string
	Baudrate int
	MockMode bool
}

type DisplayConfig struct {
	Port     string
	Baudrate int
	MockMode bool
}

type CameraConfig struct {
	Index          int
	Width          int
	Height         int
	FPS            int
	ScanIntervalMS int
	MockMode       bool
}

type SerialConfig struct {
	TimeoutMS      int
	WriteTimeoutMS int
}

type LogConfig struct {
	Level  string
	Format string
}

type ServiceConfig struct {
	LockerAgentID string
	GinMode       string
}

func New() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT", "8000"),
		},
		Orchestrator: OrchestratorConfig{
			URL:        getEnv("ORCHESTRATOR_URL", "http://localhost:8080/api/v1"),
			Timeout:    time.Duration(getEnvAsInt("ORCHESTRATOR_TIMEOUT_SEC", 10)) * time.Second,
			RetryCount: getEnvAsInt("ORCHESTRATOR_RETRY_COUNT", 3),
		},
		Arduino: ArduinoConfig{
			Port:     getEnv("ARDUINO_PORT", "/dev/ttyUSB0"),
			Baudrate: getEnvAsInt("ARDUINO_BAUDRATE", 9600),
			MockMode: getEnvAsBool("ARDUINO_MOCK_MODE", false),
		},
		Display: DisplayConfig{
			Port:     getEnv("DISPLAY_PORT", "/dev/ttyUSB1"),
			Baudrate: getEnvAsInt("DISPLAY_BAUDRATE", 115200),
			MockMode: getEnvAsBool("DISPLAY_MOCK_MODE", false),
		},
		Camera: CameraConfig{
			Index:          getEnvAsInt("CAMERA_INDEX", 0),
			Width:          getEnvAsInt("CAMERA_WIDTH", 640),
			Height:         getEnvAsInt("CAMERA_HEIGHT", 480),
			FPS:            getEnvAsInt("CAMERA_FPS", 30),
			ScanIntervalMS: getEnvAsInt("QR_SCAN_INTERVAL_MS", 100),
			MockMode:       getEnvAsBool("CAMERA_MOCK_MODE", false),
		},
		Serial: SerialConfig{
			TimeoutMS:      getEnvAsInt("SERIAL_TIMEOUT_MS", 1000),
			WriteTimeoutMS: getEnvAsInt("SERIAL_WRITE_TIMEOUT_MS", 500),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Service: ServiceConfig{
			LockerAgentID: getEnv("LOCKER_AGENT_ID", ""),
			GinMode:       getEnv("GIN_MODE", "release"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("Config - New - validate: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.HTTP.Port == "" {
		return entityError.ErrConfigMissingRequired
	}
	if c.Orchestrator.URL == "" {
		return entityError.ErrConfigMissingRequired
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
