package goconfigtools

import (
	"os"
	"testing"
	"time"
)

type WebhookConfig struct {
	Path    string        `key:"WEBHOOK_PATH" default:"webhook"`
	Timeout time.Duration `key:"WEBHOOK_TIMEOUT"` // Optional, no default
}

type AIConfig struct {
	APIKey string `key:"OPENAI_API_KEY" required:"true"` // Required field
	Model  string `key:"OPENAI_MODEL" default:"gpt-4"`
}

type Config struct {
	AI       AIConfig
	WebHook  WebhookConfig
	EnableAI bool `key:"ENABLE_AI" default:"false"`
}

func TestLoad_WithDefaults(t *testing.T) {
	// Clear environment
	os.Clearenv()

	// Set required fields only
	os.Setenv("OPENAI_API_KEY", "test-key-123")

	var cfg Config
	if err := Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Check that defaults were applied
	if cfg.WebHook.Path != "webhook" {
		t.Errorf("Expected WebHook.Path to be 'webhook', got %q", cfg.WebHook.Path)
	}

	if cfg.AI.Model != "gpt-4" {
		t.Errorf("Expected AI.Model to be 'gpt-4', got %q", cfg.AI.Model)
	}

	if cfg.EnableAI != false {
		t.Errorf("Expected EnableAI to be false, got %v", cfg.EnableAI)
	}

	// Check that env values were set
	if cfg.AI.APIKey != "test-key-123" {
		t.Errorf("Expected AI.APIKey to be 'test-key-123', got %q", cfg.AI.APIKey)
	}

	// Check that optional field was left at zero value
	if cfg.WebHook.Timeout != 0 {
		t.Errorf("Expected WebHook.Timeout to be 0 (zero value), got %v", cfg.WebHook.Timeout)
	}
}

func TestLoad_WithOverrides(t *testing.T) {
	// Clear environment
	os.Clearenv()

	// Set all values explicitly
	os.Setenv("OPENAI_API_KEY", "custom-key")
	os.Setenv("OPENAI_MODEL", "gpt-3.5-turbo")
	os.Setenv("WEBHOOK_PATH", "/custom/path")
	os.Setenv("WEBHOOK_TIMEOUT", "1m")
	os.Setenv("ENABLE_AI", "true")

	var cfg Config
	if err := Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Check that all values were overridden
	if cfg.WebHook.Path != "/custom/path" {
		t.Errorf("Expected WebHook.Path to be '/custom/path', got %q", cfg.WebHook.Path)
	}

	if cfg.AI.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected AI.Model to be 'gpt-3.5-turbo', got %q", cfg.AI.Model)
	}

	if cfg.EnableAI != true {
		t.Errorf("Expected EnableAI to be true, got %v", cfg.EnableAI)
	}

	if cfg.WebHook.Timeout != time.Minute {
		t.Errorf("Expected WebHook.Timeout to be 1m, got %v", cfg.WebHook.Timeout)
	}
}

func TestLoad_MissingRequiredField(t *testing.T) {
	// Clear environment
	os.Clearenv()

	// Don't set the required field
	var cfg Config
	err := Load(&cfg)
	if err == nil {
		t.Fatal("Expected error for missing required field OPENAI_API_KEY")
	}
}

func TestLoad_OptionalFields(t *testing.T) {
	type OptionalConfig struct {
		Required string `key:"REQUIRED_FIELD" required:"true"`
		Optional string `key:"OPTIONAL_FIELD"`
		WithDef  string `key:"WITH_DEFAULT" default:"default-value"`
	}

	os.Clearenv()
	os.Setenv("REQUIRED_FIELD", "set")

	var cfg OptionalConfig
	if err := Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Required != "set" {
		t.Errorf("Expected Required to be 'set', got %q", cfg.Required)
	}

	if cfg.Optional != "" {
		t.Errorf("Expected Optional to be empty (zero value), got %q", cfg.Optional)
	}

	if cfg.WithDef != "default-value" {
		t.Errorf("Expected WithDef to be 'default-value', got %q", cfg.WithDef)
	}
}

func TestLoad_PreInitializedDefaults(t *testing.T) {
	type DefaultConfig struct {
		EnvOverride  string `key:"ENV_OVERRIDE"`
		TagOverride  string `key:"TAG_OVERRIDE" default:"tag-default"`
		CodedDefault string `key:"CODED_DEFAULT"`
		Required     string `key:"REQUIRED" required:"true"`
	}

	os.Clearenv()
	os.Setenv("ENV_OVERRIDE", "from-env")
	os.Setenv("REQUIRED", "required-value")

	// Pre-initialize with coded defaults
	cfg := DefaultConfig{
		EnvOverride:  "coded-default-1",
		TagOverride:  "coded-default-2",
		CodedDefault: "coded-default-3",
	}

	if err := Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Env variable should override coded default
	if cfg.EnvOverride != "from-env" {
		t.Errorf("Expected EnvOverride to be 'from-env', got %q", cfg.EnvOverride)
	}

	// Tag default should override coded default
	if cfg.TagOverride != "tag-default" {
		t.Errorf("Expected TagOverride to be 'tag-default', got %q", cfg.TagOverride)
	}

	// Coded default should be preserved when no env or tag default
	if cfg.CodedDefault != "coded-default-3" {
		t.Errorf("Expected CodedDefault to be 'coded-default-3', got %q", cfg.CodedDefault)
	}

	if cfg.Required != "required-value" {
		t.Errorf("Expected Required to be 'required-value', got %q", cfg.Required)
	}
}

func TestLoad_InvalidTypes(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		errorMsg string
	}{
		{
			name: "invalid bool",
			setup: func() {
				os.Clearenv()
				os.Setenv("OPENAI_API_KEY", "key")
				os.Setenv("WEBHOOK_TIMEOUT", "30s")
				os.Setenv("ENABLE_AI", "not-a-bool")
			},
			errorMsg: "invalid bool",
		},
		{
			name: "invalid duration",
			setup: func() {
				os.Clearenv()
				os.Setenv("OPENAI_API_KEY", "key")
				os.Setenv("WEBHOOK_TIMEOUT", "not-a-duration")
			},
			errorMsg: "invalid duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			var cfg Config
			err := Load(&cfg)
			if err == nil {
				t.Fatalf("Expected error containing %q, got nil", tt.errorMsg)
			}
		})
	}
}

func TestLoad_NotAPointer(t *testing.T) {
	var cfg Config
	err := Load(cfg) // Not a pointer
	if err == nil {
		t.Fatal("Expected error when config is not a pointer")
	}
}

func TestLoad_IntAndUintTypes(t *testing.T) {
	type IntConfig struct {
		Port     int   `key:"PORT" default:"8080"`
		MaxConns int64 `key:"MAX_CONNS" default:"100"`
		BuffSize uint  `key:"BUFF_SIZE" default:"1024"`
	}

	os.Clearenv()

	var cfg IntConfig
	if err := Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected Port to be 8080, got %d", cfg.Port)
	}

	if cfg.MaxConns != 100 {
		t.Errorf("Expected MaxConns to be 100, got %d", cfg.MaxConns)
	}

	if cfg.BuffSize != 1024 {
		t.Errorf("Expected BuffSize to be 1024, got %d", cfg.BuffSize)
	}
}

func TestLoad_FloatTypes(t *testing.T) {
	type FloatConfig struct {
		Temperature float64 `key:"TEMPERATURE" default:"0.7"`
		Threshold   float32 `key:"THRESHOLD" default:"0.5"`
	}

	os.Clearenv()

	var cfg FloatConfig
	if err := Load(&cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Temperature != 0.7 {
		t.Errorf("Expected Temperature to be 0.7, got %f", cfg.Temperature)
	}

	if cfg.Threshold != 0.5 {
		t.Errorf("Expected Threshold to be 0.5, got %f", cfg.Threshold)
	}
}
