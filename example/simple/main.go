package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/m0rjc/goconfig"
)

// WebhookConfig holds webhook-related configuration
type WebhookConfig struct {
	Path    string        `key:"WEBHOOK_PATH" default:"webhook"`
	Timeout time.Duration `key:"WEBHOOK_TIMEOUT"` // No default here
}

// AIConfig holds AI-related configuration (OpenAI, conversation state, etc.)
type AIConfig struct {
	APIKey string `key:"OPENAI_API_KEY"`
	Model  string `key:"OPENAI_MODEL" default:"gpt-4"`
}

// Config is the root configuration struct
type Config struct {
	AI       AIConfig
	WebHook  WebhookConfig
	EnableAI bool `key:"ENABLE_AI" default:"false"`
}

func main() {
	var config Config

	// Load configuration from environment variables
	if err := goconfig.Load(context.Background(), &config); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Print the loaded configuration
	fmt.Println("Configuration loaded successfully:")
	fmt.Printf("  AI.APIKey: %s\n", maskAPIKey(config.AI.APIKey))
	fmt.Printf("  AI.Model: %s\n", config.AI.Model)
	fmt.Printf("  WebHook.Path: %s\n", config.WebHook.Path)
	fmt.Printf("  WebHook.Timeout: %v\n", config.WebHook.Timeout)
	fmt.Printf("  EnableAI: %v\n", config.EnableAI)
}

// maskAPIKey masks the API key for display purposes
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
