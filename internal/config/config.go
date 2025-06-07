package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config struct holds all configuration values for the application
type Config struct {
	AppName   string
	AppPort   string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSslMode string
	APIKey    string
	LogLevel  string
	JWTKey    string
}

type TestConfig struct {
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSslMode string
	JWTKey    string
}

// LoadConfig loads environment variables from .env file and stores them in the Config struct
func LoadConfig() *Config {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	config := &Config{
		AppName:   getEnv("APP_NAME", "MyGoApp"),
		AppPort:   getEnv("APP_PORT", "8080"),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "postgres"),
		DBPass:    getEnv("DB_PASSWORD", ""),
		DBName:    getEnv("DB_NAME", "myproject_db"),
		DBSslMode: getEnv("SSL_MODE", ""),
		APIKey:    getEnv("API_KEY", ""),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		JWTKey:    getEnv("JWT_KEY", ""),
	}

	return config
}

func LoadTestConfig() *TestConfig {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	testConfig := &TestConfig{
		DBHost:    getEnv("TEST_DB_HOST", "localhost"),
		DBPort:    getEnv("TEST_DB_PORT", "5432"),
		DBUser:    getEnv("TEST_DB_USER", "postgres"),
		DBPass:    getEnv("TEST_DB_PASSWORD", ""),
		DBName:    getEnv("TEST_DB_NAME", "myproject_db"),
		DBSslMode: getEnv("TEST_SSL_MODE", ""),
		JWTKey:    getEnv("TEST_JWT_KEY", ""),
	}

	return testConfig
}

// getEnv reads an environment variable and returns its value or a default value if not set
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
