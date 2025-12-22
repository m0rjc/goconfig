package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/m0rjc/goconfig"
)

type APIKey string
type APIEndpoint string
type DatabaseHost string

// ServerConfig demonstrates validation for server settings
type ServerConfig struct {
	// Port must be in the unprivileged range
	Port int `key:"SERVER_PORT" default:"8080" min:"1024" max:"65535"`

	// MaxConnections must be reasonable
	MaxConnections int `key:"MAX_CONNECTIONS" default:"1000" min:"1" max:"100000"`

	// ReadTimeout and WriteTimeout with min/max durations
	ReadTimeout  time.Duration `key:"READ_TIMEOUT" default:"30s" min:"1s" max:"5m"`
	WriteTimeout time.Duration `key:"WRITE_TIMEOUT" default:"30s" min:"1s" max:"5m"`

	// Hostname must match pattern (alphanumeric, dots, hyphens)
	Hostname string `key:"HOSTNAME" default:"localhost" pattern:"^[a-zA-Z0-9.-]+$"`
}

// RateLimitConfig demonstrates float validation
type RateLimitConfig struct {
	// RequestsPerSecond limits API calls
	RequestsPerSecond float64 `key:"RATE_LIMIT_RPS" default:"100" min:"0.1" max:"10000"`

	// BurstMultiplier controls burst allowance (e.g., 1.5 = 50% burst)
	BurstMultiplier float64 `key:"RATE_LIMIT_BURST" default:"1.5" min:"1.0" max:"5.0"`

	// LoadFactor for scaling (0.0 to 1.0)
	LoadFactor float64 `key:"LOAD_FACTOR" default:"0.75" min:"0.0" max:"1.0"`
}

// DatabaseConfig demonstrates pattern validation
type DatabaseConfig struct {
	// Host can be hostname or IP
	Host DatabaseHost `key:"DB_HOST" default:"localhost"`

	// Port in standard database range
	Port int `key:"DB_PORT" default:"5432" min:"1024" max:"65535"`

	// Username must be alphanumeric
	Username string `key:"DB_USER" default:"postgres" pattern:"^[a-zA-Z0-9_]+$"`

	// MaxConnections to database
	MaxConnections int `key:"DB_MAX_CONNS" default:"25" min:"1" max:"1000"`

	// Connection timeout
	ConnectTimeout time.Duration `key:"DB_TIMEOUT" default:"10s" min:"1s" max:"1m"`
}

// APIConfig demonstrates custom validation
type APIConfig struct {
	// API key with custom validation
	APIKey APIKey `key:"API_KEY" required:"true"`

	// Endpoint URL with custom validation
	Endpoint APIEndpoint `key:"API_ENDPOINT" default:"https://api.example.com"`

	// Retry settings
	MaxRetries   int           `key:"API_MAX_RETRIES" default:"3" min:"0" max:"10"`
	RetryBackoff time.Duration `key:"API_RETRY_BACKOFF" default:"1s" min:"100ms" max:"30s"`
}

// Config is the root configuration
type Config struct {
	Server    ServerConfig
	RateLimit RateLimitConfig
	Database  DatabaseConfig
	API       APIConfig
}

// main runs the validation example.
// If you want this sample to output a complete configuration then you must at least set the environment
// variable API_KEY to a 20 character string starting `sk-`
//
//	export API_KEY=sk-456789012345678901
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	var config Config

	// Load configuration with custom validators
	err := goconfig.Load(context.Background(), &config,
		// Validate API key format (must start with "sk-" and be at least 20 chars)
		goconfig.WithCustomType[APIKey](goconfig.NewCustomHandler(
			func(rawValue string) (APIKey, error) { return APIKey(rawValue), nil },
			func(value APIKey) error {
				key := string(value)
				if !strings.HasPrefix(key, "sk-") {
					return fmt.Errorf("API key must start with 'sk-'")
				}
				if len(key) < 20 {
					return fmt.Errorf("API key must be at least 20 characters long")
				}
				return nil
			})),

		// Validate API endpoint is a valid URL with https
		goconfig.WithCustomType[APIEndpoint](goconfig.NewCustomHandler(
			func(rawValue string) (APIEndpoint, error) { return APIEndpoint(rawValue), nil },
			func(value APIEndpoint) error {
				endpoint := string(value)
				if !strings.HasPrefix(endpoint, "https://") {
					return fmt.Errorf("API endpoint must use HTTPS")
				}
				return nil
			})),

		// Validate database host is not a loopback address in production
		goconfig.WithCustomType[DatabaseHost](goconfig.NewCustomHandler(
			func(rawValue string) (DatabaseHost, error) { return DatabaseHost(rawValue), nil },
			func(value DatabaseHost) error {
				host := string(value)
				ip := net.ParseIP(host)
				if ip != nil && ip.IsLoopback() {
					// This is just an example - you might want to allow loopback in dev
					fmt.Printf("Warning: Database host %s is a loopback address\n", host)
				}
				return nil
			})),
	)

	if err != nil {
		// goconfig provides support for structured logging of errors. Collected errors are logged individually.
		// This allows all errors in configuration to be seen at once, rather than a whack-a-mole approach seeing
		// them one by one.
		goconfig.LogError(logger, err, goconfig.WithLogMessage("configuration_error"))

		// An error would normally be fatal. This repository has a GitHub Action which verifies that the
		// sample code runs. We must exit(0) for an expected error, but will exit(1) for any other error.
		var validationErr *goconfig.ConfigErrors
		if errors.As(err, &validationErr) {
			os.Exit(0)
		}
		os.Exit(1)
	}

	// Configuration is valid - print it
	printConfig(config)
}

func printConfig(config Config) {
	fmt.Println("Configuration loaded and validated successfully:")
	fmt.Println()

	fmt.Println("Server Configuration:")
	fmt.Printf("  Port:           %d (range: 1024-65535)\n", config.Server.Port)
	fmt.Printf("  MaxConnections: %d (range: 1-100000)\n", config.Server.MaxConnections)
	fmt.Printf("  ReadTimeout:    %v (range: 1s-5m)\n", config.Server.ReadTimeout)
	fmt.Printf("  WriteTimeout:   %v (range: 1s-5m)\n", config.Server.WriteTimeout)
	fmt.Printf("  Hostname:       %s (pattern: ^[a-zA-Z0-9.-]+$)\n", config.Server.Hostname)
	fmt.Println()

	fmt.Println("Rate Limit Configuration:")
	fmt.Printf("  RequestsPerSec: %.1f (range: 0.1-10000)\n", config.RateLimit.RequestsPerSecond)
	fmt.Printf("  BurstMultiplier:%.2f (range: 1.0-5.0)\n", config.RateLimit.BurstMultiplier)
	fmt.Printf("  LoadFactor:     %.2f (range: 0.0-1.0)\n", config.RateLimit.LoadFactor)
	fmt.Println()

	fmt.Println("Database Configuration:")
	fmt.Printf("  Host:           %s\n", config.Database.Host)
	fmt.Printf("  Port:           %d (range: 1024-65535)\n", config.Database.Port)
	fmt.Printf("  Username:       %s (pattern: ^[a-zA-Z0-9_]+$)\n", config.Database.Username)
	fmt.Printf("  MaxConnections: %d (range: 1-1000)\n", config.Database.MaxConnections)
	fmt.Printf("  ConnectTimeout: %v (range: 1s-1m)\n", config.Database.ConnectTimeout)
	fmt.Println()

	fmt.Println("API Configuration:")
	fmt.Printf("  APIKey:         %s (custom: must start with 'sk-', min 20 chars)\n", maskKey(string(config.API.APIKey)))
	fmt.Printf("  Endpoint:       %s (custom: must use HTTPS)\n", config.API.Endpoint)
	fmt.Printf("  MaxRetries:     %d (range: 0-10)\n", config.API.MaxRetries)
	fmt.Printf("  RetryBackoff:   %v (range: 100ms-30s)\n", config.API.RetryBackoff)
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
