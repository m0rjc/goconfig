package goconfigtools

import (
	"fmt"
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

func TestLoad_MinMaxValidation_Int(t *testing.T) {
	type PortConfig struct {
		Port int `key:"PORT" default:"8080" min:"1024" max:"65535"`
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
		errMsg    string
	}{
		{"valid value", "8080", false, ""},
		{"at minimum", "1024", false, ""},
		{"at maximum", "65535", false, ""},
		{"below minimum", "500", true, "value 500 is below minimum 1024"},
		{"above maximum", "70000", true, "value 70000 exceeds maximum 65535"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("PORT", tt.value)

			var cfg PortConfig
			err := Load(&cfg)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.errMsg)
				}
				if err.Error() != "invalid value for PORT: "+tt.errMsg {
					t.Errorf("Expected error %q, got %q", "invalid value for PORT: "+tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestLoad_MinMaxValidation_Uint(t *testing.T) {
	type BufferConfig struct {
		BufferSize uint `key:"BUFFER_SIZE" default:"1024" min:"512" max:"4096"`
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
		errMsg    string
	}{
		{"valid value", "1024", false, ""},
		{"below minimum", "100", true, "value 100 is below minimum 512"},
		{"above maximum", "5000", true, "value 5000 exceeds maximum 4096"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("BUFFER_SIZE", tt.value)

			var cfg BufferConfig
			err := Load(&cfg)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.errMsg)
				}
				if err.Error() != "invalid value for BUFFER_SIZE: "+tt.errMsg {
					t.Errorf("Expected error %q, got %q", "invalid value for BUFFER_SIZE: "+tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestLoad_MinMaxValidation_Float(t *testing.T) {
	type RatioConfig struct {
		LoadFactor float64 `key:"LOAD_FACTOR" default:"0.75" min:"0.0" max:"1.0"`
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
		errMsg    string
	}{
		{"valid value", "0.75", false, ""},
		{"at minimum", "0.0", false, ""},
		{"at maximum", "1.0", false, ""},
		{"below minimum", "-0.5", true, "value -0.500000 is below minimum 0.000000"},
		{"above maximum", "1.5", true, "value 1.500000 exceeds maximum 1.000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("LOAD_FACTOR", tt.value)

			var cfg RatioConfig
			err := Load(&cfg)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.errMsg)
				}
				if err.Error() != "invalid value for LOAD_FACTOR: "+tt.errMsg {
					t.Errorf("Expected error %q, got %q", "invalid value for LOAD_FACTOR: "+tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestLoad_InvalidMinMaxTags(t *testing.T) {
	type BadConfig struct {
		Port int `key:"PORT" default:"8080" min:"not-a-number"`
	}

	os.Clearenv()

	var cfg BadConfig
	err := Load(&cfg)
	if err == nil {
		t.Fatal("Expected error for invalid min tag")
	}
}

func TestLoad_WithValidator_RootField(t *testing.T) {
	type SimpleConfig struct {
		Port int `key:"PORT" default:"8080"`
	}

	tests := []struct {
		name      string
		value     string
		validator Validator
		shouldErr bool
		errMsg    string
	}{
		{
			name:  "passing validator",
			value: "8080",
			validator: func(value any) error {
				port := value.(int64)
				if port%10 != 0 {
					return fmt.Errorf("port must be multiple of 10")
				}
				return nil
			},
			shouldErr: false,
		},
		{
			name:  "failing validator",
			value: "8081",
			validator: func(value any) error {
				port := value.(int64)
				if port%10 != 0 {
					return fmt.Errorf("port must be multiple of 10")
				}
				return nil
			},
			shouldErr: true,
			errMsg:    "port must be multiple of 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("PORT", tt.value)

			var cfg SimpleConfig
			err := Load(&cfg, WithValidator("Port", tt.validator))

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.errMsg)
				}
				if err.Error() != "invalid value for PORT: "+tt.errMsg {
					t.Errorf("Expected error %q, got %q", "invalid value for PORT: "+tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestLoad_WithValidator_NestedField(t *testing.T) {
	type NestedConfig struct {
		Database struct {
			Host string `key:"DB_HOST" default:"localhost"`
		}
	}

	os.Clearenv()
	os.Setenv("DB_HOST", "192.168.1.1")

	var cfg NestedConfig
	err := Load(&cfg, WithValidator("Database.Host", func(value any) error {
		host := value.(string)
		if host == "192.168.1.1" {
			return fmt.Errorf("IP addresses not allowed")
		}
		return nil
	}))

	if err == nil {
		t.Fatal("Expected error for IP address")
	}

	if err.Error() != "invalid value for DB_HOST: IP addresses not allowed" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestLoad_WithValidator_MultipleValidators(t *testing.T) {
	type PortConfig struct {
		Port int `key:"PORT" default:"8080"`
	}

	os.Clearenv()
	os.Setenv("PORT", "8080")

	var cfg PortConfig
	err := Load(&cfg,
		WithValidator("Port", func(value any) error {
			port := value.(int64)
			if port < 1024 {
				return fmt.Errorf("port below 1024")
			}
			return nil
		}),
		WithValidator("Port", func(value any) error {
			port := value.(int64)
			if port%10 != 0 {
				return fmt.Errorf("port must be multiple of 10")
			}
			return nil
		}),
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test that first validator fails
	os.Clearenv()
	os.Setenv("PORT", "500")
	var cfg2 PortConfig
	err = Load(&cfg2,
		WithValidator("Port", func(value any) error {
			port := value.(int64)
			if port < 1024 {
				return fmt.Errorf("port below 1024")
			}
			return nil
		}),
		WithValidator("Port", func(value any) error {
			port := value.(int64)
			if port%10 != 0 {
				return fmt.Errorf("port must be multiple of 10")
			}
			return nil
		}),
	)

	if err == nil || err.Error() != "invalid value for PORT: port below 1024" {
		t.Errorf("Expected first validator to fail, got %v", err)
	}
}

func TestLoad_MinMaxAndCustomValidator(t *testing.T) {
	type PortConfig struct {
		Port int `key:"PORT" default:"8080" min:"1024" max:"65535"`
	}

	os.Clearenv()
	os.Setenv("PORT", "8085")

	var cfg PortConfig
	err := Load(&cfg, WithValidator("Port", func(value any) error {
		port := value.(int64)
		if port%10 != 0 {
			return fmt.Errorf("port must be multiple of 10")
		}
		return nil
	}))

	// Should fail custom validator (not multiple of 10)
	if err == nil {
		t.Fatal("Expected error for non-multiple of 10")
	}
	if err.Error() != "invalid value for PORT: port must be multiple of 10" {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test min validation
	os.Clearenv()
	os.Setenv("PORT", "500")
	var cfg2 PortConfig
	err = Load(&cfg2)
	if err == nil || err.Error() != "invalid value for PORT: value 500 is below minimum 1024" {
		t.Errorf("Expected min validation to fail, got %v", err)
	}
}

func TestLoad_BackwardCompatibility(t *testing.T) {
	// Ensure that existing Load(&config) calls work without options
	os.Clearenv()
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("WEBHOOK_TIMEOUT", "30s")

	var cfg Config
	if err := Load(&cfg); err != nil {
		t.Fatalf("Load without options failed: %v", err)
	}

	if cfg.AI.APIKey != "test-key" {
		t.Errorf("Expected APIKey to be 'test-key', got %q", cfg.AI.APIKey)
	}
}

func TestLoad_PatternValidation_SimplePattern(t *testing.T) {
	type UsernameConfig struct {
		Username string `key:"USERNAME" pattern:"^[a-zA-Z0-9_]+$"`
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
		errMsg    string
	}{
		{"valid alphanumeric", "user123", false, ""},
		{"valid with underscore", "user_name", false, ""},
		{"invalid with space", "user name", true, "value user name does not match pattern ^[a-zA-Z0-9_]+$"},
		{"invalid with dash", "user-name", true, "value user-name does not match pattern ^[a-zA-Z0-9_]+$"},
		{"invalid with special char", "user@name", true, "value user@name does not match pattern ^[a-zA-Z0-9_]+$"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("USERNAME", tt.value)

			var cfg UsernameConfig
			err := Load(&cfg)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.errMsg)
				}
				if err.Error() != "invalid value for USERNAME: "+tt.errMsg {
					t.Errorf("Expected error %q, got %q", "invalid value for USERNAME: "+tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if cfg.Username != tt.value {
					t.Errorf("Expected Username to be %q, got %q", tt.value, cfg.Username)
				}
			}
		})
	}
}

func TestLoad_PatternValidation_EmailPattern(t *testing.T) {
	type EmailConfig struct {
		Email string `key:"EMAIL" pattern:"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"valid simple email", "user@example.com", false},
		{"valid email with dots", "user.name@example.com", false},
		{"valid email with plus", "user+tag@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"invalid no at sign", "userexample.com", true},
		{"invalid no domain", "user@", true},
		{"invalid no tld", "user@example", true},
		{"invalid spaces", "user @example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("EMAIL", tt.value)

			var cfg EmailConfig
			err := Load(&cfg)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error for value %q, got nil", tt.value)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error for value %q, got %v", tt.value, err)
				}
				if cfg.Email != tt.value {
					t.Errorf("Expected Email to be %q, got %q", tt.value, cfg.Email)
				}
			}
		})
	}
}

func TestLoad_PatternValidation_URLPattern(t *testing.T) {
	type URLConfig struct {
		WebhookURL string `key:"WEBHOOK_URL" pattern:"^https?://[a-zA-Z0-9.-]+.*$"`
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com", false},
		{"valid with path", "https://example.com/webhook", false},
		{"valid with query", "https://example.com/webhook?token=123", false},
		{"invalid ftp", "ftp://example.com", true},
		{"invalid no protocol", "example.com", true},
		{"invalid ws protocol", "ws://example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("WEBHOOK_URL", tt.value)

			var cfg URLConfig
			err := Load(&cfg)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error for value %q, got nil", tt.value)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error for value %q, got %v", tt.value, err)
				}
				if cfg.WebhookURL != tt.value {
					t.Errorf("Expected WebhookURL to be %q, got %q", tt.value, cfg.WebhookURL)
				}
			}
		})
	}
}

func TestLoad_PatternValidation_InvalidRegex(t *testing.T) {
	type BadPatternConfig struct {
		Field string `key:"FIELD" pattern:"[invalid(regex"`
	}

	os.Clearenv()
	os.Setenv("FIELD", "value")

	var cfg BadPatternConfig
	err := Load(&cfg)
	if err == nil {
		t.Fatal("Expected error for invalid regex pattern")
	}

	// Should contain an error about invalid pattern
	if err.Error() == "" {
		t.Error("Expected non-empty error message for invalid regex")
	}
}

func TestLoad_PatternValidation_OnNonStringType(t *testing.T) {
	type InvalidConfig struct {
		Port int `key:"PORT" pattern:"^[0-9]+$"`
	}

	os.Clearenv()
	os.Setenv("PORT", "8080")

	var cfg InvalidConfig
	err := Load(&cfg)
	if err == nil {
		t.Fatal("Expected error for pattern tag on non-string type")
	}

	// The error message includes information about the pattern tag not being supported
	expectedMsg := "pattern tag not supported for type int"
	if err.Error() != "invalid pattern tag value \"\" for field Port: "+expectedMsg {
		t.Errorf("Expected error %q, got %q", "invalid pattern tag value \"\" for field Port: "+expectedMsg, err.Error())
	}
}

func TestLoad_PatternValidation_WithDefaultValue(t *testing.T) {
	type ConfigWithDefault struct {
		APIKey string `key:"API_KEY" default:"test_key_123" pattern:"^[a-z_0-9]+$"`
	}

	tests := []struct {
		name      string
		setValue  bool
		value     string
		shouldErr bool
	}{
		{"default value matches", false, "", false},
		{"env value matches", true, "prod_key_456", false},
		{"env value doesn't match", true, "INVALID-KEY", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.setValue {
				os.Setenv("API_KEY", tt.value)
			}

			var cfg ConfigWithDefault
			err := Load(&cfg)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error for value %q, got nil", tt.value)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				expectedValue := tt.value
				if !tt.setValue {
					expectedValue = "test_key_123"
				}
				if cfg.APIKey != expectedValue {
					t.Errorf("Expected APIKey to be %q, got %q", expectedValue, cfg.APIKey)
				}
			}
		})
	}
}

func TestLoad_PatternValidation_NestedStruct(t *testing.T) {
	type DatabaseConfig struct {
		ConnectionString string `key:"DB_CONN" pattern:"^[a-zA-Z]+://.*$"`
	}

	type AppConfig struct {
		Database DatabaseConfig
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"valid postgres", "postgresql://localhost:5432/db", false},
		{"valid mysql", "mysql://localhost:3306/db", false},
		{"invalid no protocol", "localhost:5432/db", true},
		{"invalid ends with colon", "postgres:", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("DB_CONN", tt.value)

			var cfg AppConfig
			err := Load(&cfg)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error for value %q, got nil", tt.value)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error for value %q, got %v", tt.value, err)
				}
				if cfg.Database.ConnectionString != tt.value {
					t.Errorf("Expected ConnectionString to be %q, got %q", tt.value, cfg.Database.ConnectionString)
				}
			}
		})
	}
}

func TestLoad_PatternValidation_WithMinMaxAndCustomValidator(t *testing.T) {
	type ComplexConfig struct {
		Port     int    `key:"PORT" default:"8080" min:"1024" max:"65535"`
		Hostname string `key:"HOSTNAME" pattern:"^[a-z0-9-]+$"`
	}

	os.Clearenv()
	os.Setenv("PORT", "8080")
	os.Setenv("HOSTNAME", "web-server-01")

	var cfg ComplexConfig
	err := Load(&cfg, WithValidator("Port", func(value any) error {
		port := value.(int64)
		if port%10 != 0 {
			return fmt.Errorf("port must be multiple of 10")
		}
		return nil
	}))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected Port to be 8080, got %d", cfg.Port)
	}

	if cfg.Hostname != "web-server-01" {
		t.Errorf("Expected Hostname to be 'web-server-01', got %q", cfg.Hostname)
	}

	// Test pattern validation failure
	os.Clearenv()
	os.Setenv("PORT", "8080")
	os.Setenv("HOSTNAME", "UPPERCASE")

	var cfg2 ComplexConfig
	err = Load(&cfg2)
	if err == nil {
		t.Fatal("Expected error for invalid hostname pattern")
	}

	// Test port validation failure
	os.Clearenv()
	os.Setenv("PORT", "500")
	os.Setenv("HOSTNAME", "valid-host")

	var cfg3 ComplexConfig
	err = Load(&cfg3)
	if err == nil {
		t.Fatal("Expected error for port below minimum")
	}
}

func TestLoad_PatternValidation_CaseSensitive(t *testing.T) {
	type CaseSensitiveConfig struct {
		Code string `key:"CODE" pattern:"^[a-z]{3}$"`
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"lowercase matches", "abc", false},
		{"uppercase doesn't match", "ABC", true},
		{"mixed case doesn't match", "Abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("CODE", tt.value)

			var cfg CaseSensitiveConfig
			err := Load(&cfg)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error for value %q, got nil", tt.value)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error for value %q, got %v", tt.value, err)
				}
			}
		})
	}
}

func TestLoad_PatternValidation_SpecialCharacters(t *testing.T) {
	type TokenConfig struct {
		// Pattern that allows alphanumeric, underscores, and hyphens
		Token string `key:"TOKEN" pattern:"^[a-zA-Z0-9_-]{8,}$"`
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"valid token with letters and numbers", "abcd1234", false},
		{"valid token with underscore", "test_token_123", false},
		{"valid token with hyphen", "test-token-123", false},
		{"valid mixed case", "TestToken123", false},
		{"invalid with spaces", "test token", true},
		{"invalid with special char", "test@token", true},
		{"invalid too short", "test", true},
		{"invalid with dot", "test.token", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("TOKEN", tt.value)

			var cfg TokenConfig
			err := Load(&cfg)

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error for value %q, got nil", tt.value)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error for value %q, got %v", tt.value, err)
				}
				if cfg.Token != tt.value {
					t.Errorf("Expected Token to be %q, got %q", tt.value, cfg.Token)
				}
			}
		})
	}
}
