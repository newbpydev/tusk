package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration values.
// It uses struct tags to specify the environment variable names.
type Config struct {
	DBURL        string `env:"DB_URL"`
	Port         string `env:"PORT"`
	AppEnv       string `env:"APP_ENV"`
	LogLevel     string `env:"LOG_LEVEL"`
	LogDir       string `env:"LOG_DIR"`
	LogToFile    bool   `env:"LOG_TO_FILE"`
	LogToConsole bool   `env:"LOG_TO_CONSOLE"`
}

// Load loads the configuration from environment variables and returns a Config struct.
// It uses the godotenv package to load variables from a .env file if it exists.
func Load() *Config {
	_ = godotenv.Load() // Load .env file if it exists, ignore errors if file not found

	return &Config{
		DBURL:        getEnv("DB_URL", "postgres://postgres:password@localhost:5432/tuskapp"),
		Port:         getEnv("PORT", "8080"),
		AppEnv:       getEnv("APP_ENV", "development"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		LogDir:       getEnv("LOG_DIR", ""),
		LogToFile:    getBoolEnv("LOG_TO_FILE", true),
		LogToConsole: getBoolEnv("LOG_TO_CONSOLE", true),
	}
}

// getEnv retrieves the value of the environment variable with the given key.
// If the variable is not set, it returns the provided fallback value.
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

// getBoolEnv retrieves a boolean value from an environment variable.
// It returns the fallback value if the environment variable is not set
// or if it cannot be parsed as a boolean.
func getBoolEnv(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	switch val {
	case "true", "1", "yes", "y", "on":
		return true
	case "false", "0", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
