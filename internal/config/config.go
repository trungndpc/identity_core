package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort int

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	InternalAPIKey     string
	AdminAPIKey        string
	AdminUsername      string
	AdminPassword      string
	AdminTenantCode    string
	JWTSecret          string
	JWTExpiryHours     int
	UserJWTSecret      string
	UserJWTExpiryHours int
	AllowLegacyUserID  bool
	CORSAllowedOrigins []string

	ZaloAuthDevMode  bool
	ZaloAppSecretKey string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("APP_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid APP_PORT: %w", err)
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	appEnv := getEnv("APP_ENV", "development")

	jwtExpiryHours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY_HOURS: %w", err)
	}
	userJWTExpiryHours, err := strconv.Atoi(getEnv("USER_JWT_EXPIRY_HOURS", "1"))
	if err != nil {
		return nil, fmt.Errorf("invalid USER_JWT_EXPIRY_HOURS: %w", err)
	}

	jwtSecret := getEnv("JWT_SECRET", "change-me-jwt-secret")
	userJWTSecret := getEnv("USER_JWT_SECRET", "")
	if userJWTSecret == "" {
		userJWTSecret = jwtSecret
	}
	legacyDefault := "true"
	if appEnv == "production" {
		legacyDefault = "false"
	}

	cfg := &Config{
		AppEnv:             appEnv,
		AppPort:            port,
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             dbPort,
		DBUser:             getEnv("DB_USER", "identity"),
		DBPassword:         getEnv("DB_PASSWORD", "identity"),
		DBName:             getEnv("DB_NAME", "identity_core"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		InternalAPIKey:     getEnv("INTERNAL_API_KEY", ""),
		AdminAPIKey:        getEnv("ADMIN_API_KEY", ""),
		AdminUsername:      getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:      getEnv("ADMIN_PASSWORD", "admin123"),
		AdminTenantCode:    getEnv("ADMIN_TENANT_CODE", "support"),
		JWTSecret:          jwtSecret,
		JWTExpiryHours:     jwtExpiryHours,
		UserJWTSecret:      userJWTSecret,
		UserJWTExpiryHours: userJWTExpiryHours,
		AllowLegacyUserID:  getEnv("ALLOW_LEGACY_USER_ID_HEADER", legacyDefault) == "true",
		CORSAllowedOrigins: parseCORSAllowedOrigins(getEnv("CORS_ALLOWED_ORIGINS", ""), appEnv),
		ZaloAuthDevMode:    getEnv("ZALO_AUTH_DEV_MODE", "true") == "true",
		ZaloAppSecretKey:   getEnv("ZALO_APP_SECRET_KEY", ""),
	}

	return cfg, nil
}

func parseCORSAllowedOrigins(value, appEnv string) []string {
	if value != "" {
		return strings.Split(value, ",")
	}
	if appEnv != "production" {
		return []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		}
	}
	return nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
