package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort          string
	DatabaseURL         string
	JWTSecret           string
	JWTExpiration       time.Duration
	MaxUploadSize       int64
	FileStoragePath     string
	CORSAllowedOrigin   string
	EnableDevRegister   bool
}

func Load() (Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	expiration, err := time.ParseDuration(envOr("JWT_EXPIRATION", "1h"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_EXPIRATION: %w", err)
	}

	maxUpload, err := strconv.ParseInt(envOr("MAX_UPLOAD_SIZE", "52428800"), 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf("invalid MAX_UPLOAD_SIZE: %w", err)
	}

	return Config{
		ServerPort:        envOr("SERVER_PORT", "8080"),
		DatabaseURL:       dbURL,
		JWTSecret:         jwtSecret,
		JWTExpiration:     expiration,
		MaxUploadSize:     maxUpload,
		FileStoragePath:   envOr("FILE_STORAGE_PATH", "./data/uploads"),
		CORSAllowedOrigin: envOr("CORS_ALLOWED_ORIGIN", "http://localhost:5175"),
		EnableDevRegister: envOr("ENABLE_DEV_REGISTER", "false") == "true",
	}, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
