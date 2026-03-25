package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// findDotenvFile walks from startDir upward looking for a file named .env.
func findDotenvFile(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for range 20 {
		p := filepath.Join(dir, ".env")
		fi, err := os.Stat(p)
		if err == nil && !fi.IsDir() {
			return filepath.Clean(p), nil
		}
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

func tryLoadDotenvFrom(startDir string) bool {
	p, err := findDotenvFile(startDir)
	if err != nil {
		return false
	}
	if err := godotenv.Load(p); err != nil {
		log.Fatalf("Invalid .env file at %s: %v\n(dotenv only allows KEY=value lines; move SQL or other text out of .env.)", p, err)
	}
	return true
}

// loadDotenv loads the first valid .env: walk up from cwd, then walk up from
// the executable directory (so `go build` in cmd/server and running ./server.exe
// still finds the repo-root .env).
func loadDotenv() bool {
	if wd, err := os.Getwd(); err == nil {
		if tryLoadDotenvFrom(wd) {
			return true
		}
	}
	if exe, err := os.Executable(); err == nil {
		if tryLoadDotenvFrom(filepath.Dir(exe)) {
			return true
		}
	}
	return false
}

// WebAssetsRoot returns the directory that contains web/templates and web/static,
// or "" if not found. Walks upward from cwd and from the executable directory
// so the HTML UI works when the binary is run from cmd/server.
func WebAssetsRoot() string {
	try := func(startDir string) string {
		dir, err := filepath.Abs(startDir)
		if err != nil {
			return ""
		}
		for range 20 {
			tpl := filepath.Join(dir, "web", "templates")
			fi, err := os.Stat(tpl)
			if err == nil && fi.IsDir() {
				return filepath.Clean(dir)
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		return ""
	}
	if wd, err := os.Getwd(); err == nil {
		if s := try(wd); s != "" {
			return s
		}
	}
	if exe, err := os.Executable(); err == nil {
		if s := try(filepath.Dir(exe)); s != "" {
			return s
		}
	}
	return ""
}

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
	if !loadDotenv() {
		log.Fatal("No .env file found. Copy .env.example to .env at the project root (or next to the server executable) and fill in your values.")
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
