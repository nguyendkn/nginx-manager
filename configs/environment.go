package configs

import (
	"os"
	"strconv"
	"strings"
)

// Environment holds all environment configuration
type Environment struct {
	// Server configuration
	Port string `json:"port"`
	Host string `json:"host"`

	// Application configuration
	AppName        string `json:"app_name"`
	AppVersion     string `json:"app_version"`
	AppEnvironment string `json:"app_environment"`

	// Gin configuration
	GinMode string `json:"gin_mode"`

	// CORS configuration
	CORSAllowedOrigins []string `json:"cors_allowed_origins"`
	CORSAllowedMethods []string `json:"cors_allowed_methods"`
	CORSAllowedHeaders []string `json:"cors_allowed_headers"`

	// Logging configuration
	LogLevel    string `json:"log_level"`
	LogEncoding string `json:"log_encoding"`
}

// LoadEnvironment loads environment variables into Environment struct
func LoadEnvironment() *Environment {
	env := &Environment{
		// Server configuration
		Port: getEnvWithDefault("PORT", "8080"),
		Host: getEnvWithDefault("HOST", "0.0.0.0"),

		// Application configuration
		AppName:        getEnvWithDefault("APP_NAME", "c-agents"),
		AppVersion:     getEnvWithDefault("APP_VERSION", "1.0.0"),
		AppEnvironment: getEnvWithDefault("APP_ENV", "development"),

		// Gin configuration
		GinMode: getEnvWithDefault("GIN_MODE", "debug"),

		// CORS configuration
		CORSAllowedOrigins: getEnvSliceWithDefault("CORS_ALLOWED_ORIGINS", []string{"*"}),
		CORSAllowedMethods: getEnvSliceWithDefault("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		CORSAllowedHeaders: getEnvSliceWithDefault("CORS_ALLOWED_HEADERS", []string{"*"}),

		// Logging configuration
		LogLevel:    getEnvWithDefault("LOG_LEVEL", "info"),
		LogEncoding: getEnvWithDefault("LOG_ENCODING", "console"),
	}

	return env
}

// Helper functions for environment variable parsing

// getEnvWithDefault gets environment variable with default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvSliceWithDefault gets environment variable as slice with default value
func getEnvSliceWithDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Trim spaces from each element
		parts := strings.Split(value, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		return parts
	}
	return defaultValue
}

// getEnvIntWithDefault gets environment variable as int with default value
func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBoolWithDefault gets environment variable as bool with default value
func getEnvBoolWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// Getter methods for easy access

// GetPort returns the server port
func (e *Environment) GetPort() string {
	return e.Port
}

// GetHost returns the server host
func (e *Environment) GetHost() string {
	return e.Host
}

// GetServerAddress returns the full server address
func (e *Environment) GetServerAddress() string {
	return e.Host + ":" + e.Port
}

// IsProduction returns true if running in production environment
func (e *Environment) IsProduction() bool {
	return strings.ToLower(e.AppEnvironment) == "production"
}

// IsDevelopment returns true if running in development environment
func (e *Environment) IsDevelopment() bool {
	return strings.ToLower(e.AppEnvironment) == "development"
}

// IsTest returns true if running in test environment
func (e *Environment) IsTest() bool {
	return strings.ToLower(e.AppEnvironment) == "test"
}

// CORS Configuration Getters

// GetCORSAllowedOrigins returns CORS allowed origins
func (e *Environment) GetCORSAllowedOrigins() []string {
	return e.CORSAllowedOrigins
}

// GetCORSAllowedMethods returns CORS allowed methods
func (e *Environment) GetCORSAllowedMethods() []string {
	return e.CORSAllowedMethods
}

// GetCORSAllowedHeaders returns CORS allowed headers
func (e *Environment) GetCORSAllowedHeaders() []string {
	return e.CORSAllowedHeaders
}

// Logging Configuration Getters

// GetLogLevel returns the log level
func (e *Environment) GetLogLevel() string {
	return e.LogLevel
}

// GetLogEncoding returns the log encoding
func (e *Environment) GetLogEncoding() string {
	return e.LogEncoding
}

// Application Configuration Getters

// GetAppName returns the application name
func (e *Environment) GetAppName() string {
	return e.AppName
}

// GetAppVersion returns the application version
func (e *Environment) GetAppVersion() string {
	return e.AppVersion
}

// GetAppEnvironment returns the application environment
func (e *Environment) GetAppEnvironment() string {
	return e.AppEnvironment
}

// GetGinMode returns the Gin mode
func (e *Environment) GetGinMode() string {
	return e.GinMode
}

// Validation methods

// Validate validates the environment configuration
func (e *Environment) Validate() error {
	// Add validation logic here if needed
	return nil
}

// String returns a string representation of the environment (without sensitive data)
func (e *Environment) String() string {
	return "Environment{" +
		"Port:" + e.Port +
		", Host:" + e.Host +
		", AppName:" + e.AppName +
		", AppVersion:" + e.AppVersion +
		", AppEnvironment:" + e.AppEnvironment +
		", GinMode:" + e.GinMode +
		", LogLevel:" + e.LogLevel +
		", LogEncoding:" + e.LogEncoding +
		"}"
}
