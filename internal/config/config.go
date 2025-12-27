package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	App      AppConfig
	Log      LogConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Storage  StorageConfig
	Image    ImageConfig
	AI       AIConfig
	CORS     CORSConfig
	Rate     RateConfig
}

// AppConfig holds application basic configuration
type AppConfig struct {
	Env   string
	Port  string
	Debug bool
	Name  string
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level    string
	Format   string
	Output   string
	FilePath string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret        string
	AccessExpire  time.Duration
	RefreshExpire time.Duration
	Issuer        string
}

// StorageConfig holds file storage configuration
type StorageConfig struct {
	Type         string
	Path         string
	BaseURL      string
	MaxSize      int64
	AllowedTypes []string
}

// ImageConfig holds image processing configuration
type ImageConfig struct {
	MaxDimension   int
	Quality        int
	ThumbSmWidth   int
	ThumbSmHeight  int
	ThumbSmQuality int
	ThumbMdWidth   int
	ThumbMdHeight  int
	ThumbMdQuality int
	ThumbLgWidth   int
	ThumbLgHeight  int
	ThumbLgQuality int
}

// AIConfig holds AI service configuration
type AIConfig struct {
	ServiceURL string
	Timeout    time.Duration
	Retry      int
	APIKey     string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled        bool
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

// RateConfig holds rate limiting configuration
type RateConfig struct {
	Enabled     bool
	Requests    int
	Burst       int
	UploadLimit int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:   getEnv("APP_ENV", "development"),
			Port:  getEnv("APP_PORT", "8080"),
			Debug: getEnvBool("APP_DEBUG", true),
			Name:  getEnv("APP_NAME", "QuanPhotos"),
		},
		Log: LogConfig{
			Level:    getEnv("LOG_LEVEL", "info"),
			Format:   getEnv("LOG_FORMAT", "json"),
			Output:   getEnv("LOG_OUTPUT", "stdout"),
			FilePath: getEnv("LOG_FILE_PATH", "./logs/app.log"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			Name:            getEnv("DB_NAME", "quanphotos"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME", 300)) * time.Second,
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", ""),
			AccessExpire:  time.Duration(getEnvInt("JWT_ACCESS_EXPIRE", 3600)) * time.Second,
			RefreshExpire: time.Duration(getEnvInt("JWT_REFRESH_EXPIRE", 604800)) * time.Second,
			Issuer:        getEnv("JWT_ISSUER", "quanphotos"),
		},
		Storage: StorageConfig{
			Type:         getEnv("STORAGE_TYPE", "local"),
			Path:         getEnv("STORAGE_PATH", "./uploads"),
			BaseURL:      getEnv("STORAGE_BASE_URL", ""),
			MaxSize:      getEnvInt64("STORAGE_MAX_SIZE", 52428800),
			AllowedTypes: getEnvSlice("STORAGE_ALLOWED_TYPES", []string{"jpg", "jpeg", "png"}),
		},
		Image: ImageConfig{
			MaxDimension:   getEnvInt("IMAGE_MAX_DIMENSION", 4096),
			Quality:        getEnvInt("IMAGE_QUALITY", 92),
			ThumbSmWidth:   getEnvInt("THUMB_SM_WIDTH", 300),
			ThumbSmHeight:  getEnvInt("THUMB_SM_HEIGHT", 200),
			ThumbSmQuality: getEnvInt("THUMB_SM_QUALITY", 80),
			ThumbMdWidth:   getEnvInt("THUMB_MD_WIDTH", 800),
			ThumbMdHeight:  getEnvInt("THUMB_MD_HEIGHT", 533),
			ThumbMdQuality: getEnvInt("THUMB_MD_QUALITY", 85),
			ThumbLgWidth:   getEnvInt("THUMB_LG_WIDTH", 1600),
			ThumbLgHeight:  getEnvInt("THUMB_LG_HEIGHT", 1067),
			ThumbLgQuality: getEnvInt("THUMB_LG_QUALITY", 90),
		},
		AI: AIConfig{
			ServiceURL: getEnv("AI_SERVICE_URL", "http://localhost:8000"),
			Timeout:    time.Duration(getEnvInt("AI_SERVICE_TIMEOUT", 30)) * time.Second,
			Retry:      getEnvInt("AI_SERVICE_RETRY", 3),
			APIKey:     getEnv("AI_SERVICE_API_KEY", ""),
		},
		CORS: CORSConfig{
			Enabled:        getEnvBool("CORS_ENABLED", true),
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
			AllowedMethods: getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Authorization", "Content-Type", "Accept-Language"}),
			MaxAge:         getEnvInt("CORS_MAX_AGE", 86400),
		},
		Rate: RateConfig{
			Enabled:     getEnvBool("RATE_LIMIT_ENABLED", true),
			Requests:    getEnvInt("RATE_LIMIT_REQUESTS", 100),
			Burst:       getEnvInt("RATE_LIMIT_BURST", 20),
			UploadLimit: getEnvInt("UPLOAD_RATE_LIMIT", 10),
		},
	}

	return cfg, nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// Helper functions to get environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
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

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
