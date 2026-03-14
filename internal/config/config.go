package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	AppEnv  string
	GinMode string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBTimezone string

	JWTSecret              string
	JWTAccessExpiryMinutes int
	JWTRefreshExpiryHours  int

	SeedAdminEmail    string
	SeedAdminPassword string
	SeedAdminName     string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),
		GinMode: getEnv("GIN_MODE", "debug"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "fpreg"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		DBTimezone: getEnv("DB_TIMEZONE", "Africa/Kampala"),

		JWTSecret:              getEnv("JWT_SECRET", "default-dev-secret-change-me"),
		JWTAccessExpiryMinutes: getEnvInt("JWT_ACCESS_EXPIRY_MINUTES", 30),
		JWTRefreshExpiryHours:  getEnvInt("JWT_REFRESH_EXPIRY_HOURS", 168),

		SeedAdminEmail:    getEnv("SEED_ADMIN_EMAIL", "admin@moh.go.ug"),
		SeedAdminPassword: getEnv("SEED_ADMIN_PASSWORD", "ChangeMe@2026!"),
		SeedAdminName:     getEnv("SEED_ADMIN_NAME", "System Administrator"),
	}
}

func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" port=" + c.DBPort +
		" sslmode=" + c.DBSSLMode +
		" TimeZone=" + c.DBTimezone
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
