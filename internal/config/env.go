package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL  string `env:"DB_URL"`
	Port   string `env:"PORT"`
	AppEnv string `env:"APP_ENV"`
}

func Load() *Config {
	_ = godotenv.Load() // Load .env file if it exists

	return &Config{
		DBURL:  getEnv("DB_URL", "postgres://postgres:password@localhost:5432/tuskapp"),
		Port:   getEnv("PORT", "8080"),
		AppEnv: getEnv("APP_ENV", "development"),
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
