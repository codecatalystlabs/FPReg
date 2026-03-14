package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort  string
	AppEnv   string
	GinMode  string
	BasePath string

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
	if err := godotenv.Load(); err != nil {
		if exe, e := os.Executable(); e == nil {
			if err2 := godotenv.Load(filepath.Join(filepath.Dir(exe), ".env")); err2 != nil {
				log.Fatal("No .env file found. Copy .env.example to .env and fill in your values.")
			}
		} else {
			log.Fatal("No .env file found. Copy .env.example to .env and fill in your values.")
		}
	}

	cfg := &Config{
		AppPort:  os.Getenv("APP_PORT"),
		AppEnv:   os.Getenv("APP_ENV"),
		GinMode:  os.Getenv("GIN_MODE"),
		BasePath: strings.TrimRight(os.Getenv("BASE_PATH"), "/"),

		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),
		DBTimezone: os.Getenv("DB_TIMEZONE"),

		JWTSecret:              os.Getenv("JWT_SECRET"),
		JWTAccessExpiryMinutes: envInt("JWT_ACCESS_EXPIRY_MINUTES"),
		JWTRefreshExpiryHours:  envInt("JWT_REFRESH_EXPIRY_HOURS"),

		SeedAdminEmail:    os.Getenv("SEED_ADMIN_EMAIL"),
		SeedAdminPassword: os.Getenv("SEED_ADMIN_PASSWORD"),
		SeedAdminName:     os.Getenv("SEED_ADMIN_NAME"),
	}

	required := map[string]string{
		"APP_PORT":    cfg.AppPort,
		"DB_HOST":     cfg.DBHost,
		"DB_PORT":     cfg.DBPort,
		"DB_USER":     cfg.DBUser,
		"DB_NAME":     cfg.DBName,
		"JWT_SECRET":  cfg.JWTSecret,
	}
	for key, val := range required {
		if val == "" {
			log.Fatalf("Required env var %s is missing or empty. Check your .env file.", key)
		}
	}

	return cfg
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

func envInt(key string) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return 0
}
