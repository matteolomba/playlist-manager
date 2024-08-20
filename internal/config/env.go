package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	LogLevel string
}

var Envs = initConfig()

// InitConfig initializes the configuration by reading the environment variables stored in the .env file.
func initConfig() Config {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		panic("CONFIG: Error loading .env file")
	}

	return Config{
		LogLevel: getEnv("LOG_LEVEL", "WARN"),
	}

}

// GetEnv returns the value of an environment variable if it exists, otherwise it returns the fallback value, provided as a parameter.
func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
