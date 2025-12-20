package goconfigtools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
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
	if err := Load(context.Background(), &cfg); err != nil {
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
	if err := Load(context.Background(), &cfg); err != nil {
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
	err := Load(context.Background(), &cfg)
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
	if err := Load(context.Background(), &cfg); err != nil {
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

	if err := Load(context.Background(), &cfg); err != nil {
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
			err := Load(context.Background(), &cfg)
			if err == nil {
				t.Fatalf("Expected error containing %q, got nil", tt.errorMsg)
			}
		})
	}
}

func TestLoad_NotAPointer(t *testing.T) {
	var cfg Config
	err := Load(context.Background(), cfg) // Not a pointer
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
	if err := Load(context.Background(), &cfg); err != nil {
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
	if err := Load(context.Background(), &cfg); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Temperature != 0.7 {
		t.Errorf("Expected Temperature to be 0.7, got %f", cfg.Temperature)
	}

	if cfg.Threshold != 0.5 {
		t.Errorf("Expected Threshold to be 0.5, got %f", cfg.Threshold)
	}
}

func TestLoad_BackwardCompatibility(t *testing.T) {
	// Ensure that existing Load(&config) calls work without options
	os.Clearenv()
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("WEBHOOK_TIMEOUT", "30s")

	var cfg Config
	if err := Load(context.Background(), &cfg); err != nil {
		t.Fatalf("Load without options failed: %v", err)
	}

	if cfg.AI.APIKey != "test-key" {
		t.Errorf("Expected APIKey to be 'test-key', got %q", cfg.AI.APIKey)
	}
}

// Validation Framework Tests
// These tests verify that the validation framework is integrated correctly,
// using mock validators to avoid testing specific validation logic (which is in validation_test.go)

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
			err := Load(context.Background(), &cfg, WithValidator("Port", tt.validator))

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error containing %q, got nil", tt.errMsg)
				}
				if err.Error() != "PORT: "+tt.errMsg {
					t.Errorf("Expected error %q, got %q", "PORT: "+tt.errMsg, err.Error())
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
	err := Load(context.Background(), &cfg, WithValidator("Database.Host", func(value any) error {
		host := value.(string)
		if host == "192.168.1.1" {
			return fmt.Errorf("IP addresses not allowed")
		}
		return nil
	}))

	if err == nil {
		t.Fatal("Expected error for IP address")
	}

	if err.Error() != "DB_HOST: IP addresses not allowed" {
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
	err := Load(context.Background(), &cfg,
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
	err = Load(context.Background(), &cfg2,
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

	if err == nil || err.Error() != "PORT: port below 1024" {
		t.Errorf("Expected first validator to fail, got %v", err)
	}
}

// Integration test: Verify builtin validators (min/max/pattern) work at different field positions
func TestLoad_BuiltinValidators_RootField(t *testing.T) {
	type PortConfig struct {
		Port int `key:"PORT" default:"8080" min:"1024" max:"65535"`
	}

	// Test valid value
	os.Clearenv()
	os.Setenv("PORT", "8080")

	var cfg PortConfig
	err := Load(context.Background(), &cfg)
	if err != nil {
		t.Fatalf("Expected no error for valid value, got %v", err)
	}

	// Test below minimum
	os.Clearenv()
	os.Setenv("PORT", "500")
	var cfg2 PortConfig
	err = Load(context.Background(), &cfg2)
	if err == nil {
		t.Fatal("Expected error for value below minimum")
	}
	if err.Error() != "PORT: value 500 is below minimum 1024" {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test above maximum
	os.Clearenv()
	os.Setenv("PORT", "70000")
	var cfg3 PortConfig
	err = Load(context.Background(), &cfg3)
	if err == nil {
		t.Fatal("Expected error for value above maximum")
	}
	if err.Error() != "PORT: value 70000 exceeds maximum 65535" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestLoad_BuiltinValidators_NestedField(t *testing.T) {
	type DatabaseConfig struct {
		ConnectionString string `key:"DB_CONN" pattern:"^[a-zA-Z]+://.*$"`
	}

	type AppConfig struct {
		Database DatabaseConfig
	}

	// Test valid value
	os.Clearenv()
	os.Setenv("DB_CONN", "postgresql://localhost:5432/db")

	var cfg AppConfig
	err := Load(context.Background(), &cfg)
	if err != nil {
		t.Fatalf("Expected no error for valid value, got %v", err)
	}

	if cfg.Database.ConnectionString != "postgresql://localhost:5432/db" {
		t.Errorf("Expected ConnectionString to be set correctly, got %q", cfg.Database.ConnectionString)
	}

	// Test invalid value
	os.Clearenv()
	os.Setenv("DB_CONN", "localhost:5432/db")
	var cfg2 AppConfig
	err = Load(context.Background(), &cfg2)
	if err == nil {
		t.Fatal("Expected error for invalid pattern")
	}
}

func TestLoad_BuiltinValidators_WithDefaultValue(t *testing.T) {
	type ConfigWithDefault struct {
		APIKey string `key:"API_KEY" default:"test_key_123" pattern:"^[a-z_0-9]+$"`
	}

	// Test that default value is validated
	os.Clearenv()

	var cfg ConfigWithDefault
	err := Load(context.Background(), &cfg)
	if err != nil {
		t.Fatalf("Expected no error for valid default value, got %v", err)
	}

	if cfg.APIKey != "test_key_123" {
		t.Errorf("Expected APIKey to be 'test_key_123', got %q", cfg.APIKey)
	}

	// Test that env value is validated
	os.Clearenv()
	os.Setenv("API_KEY", "INVALID-KEY")
	var cfg2 ConfigWithDefault
	err = Load(context.Background(), &cfg2)
	if err == nil {
		t.Fatal("Expected error for invalid env value")
	}
}

func TestLoad_BuiltinValidators_WithCustomValidator(t *testing.T) {
	type PortConfig struct {
		Port int `key:"PORT" default:"8080" min:"1024" max:"65535"`
	}

	os.Clearenv()
	os.Setenv("PORT", "8085")

	var cfg PortConfig
	err := Load(context.Background(), &cfg, WithValidator("Port", func(value any) error {
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
	if err.Error() != "PORT: port must be multiple of 10" {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test min validation still works
	os.Clearenv()
	os.Setenv("PORT", "500")
	var cfg2 PortConfig
	err = Load(context.Background(), &cfg2)
	if err == nil || err.Error() != "PORT: value 500 is below minimum 1024" {
		t.Errorf("Expected min validation to fail, got %v", err)
	}
}

func TestLoad_BuiltinValidators_InvalidTags(t *testing.T) {
	type BadMinConfig struct {
		Port int `key:"PORT" default:"8080" min:"not-a-number"`
	}

	os.Clearenv()

	var cfg BadMinConfig
	err := Load(context.Background(), &cfg)
	if err == nil {
		t.Fatal("Expected error for invalid min tag")
	}

	type BadMaxConfig struct {
		Port int `key:"PORT" default:"8080" max:"not-a-number"`
	}

	var cfg2 BadMaxConfig
	err = Load(context.Background(), &cfg2)
	if err == nil {
		t.Fatal("Expected error for invalid max tag")
	}

	type BadPatternConfig struct {
		Field string `key:"FIELD" pattern:"[invalid(regex"`
	}

	os.Setenv("FIELD", "value")
	var cfg3 BadPatternConfig
	err = Load(context.Background(), &cfg3)
	if err == nil {
		t.Fatal("Expected error for invalid regex pattern")
	}
}

func TestLoad_BuiltinValidators_UnsupportedTypes(t *testing.T) {
	type InvalidConfig struct {
		Port int `key:"PORT" pattern:"^[0-9]+$"`
	}

	os.Clearenv()
	os.Setenv("PORT", "8080")

	var cfg InvalidConfig
	err := Load(context.Background(), &cfg)
	if err == nil {
		t.Fatal("Expected error for pattern tag on non-string type")
	}

	expectedMsg := "pattern tag not supported for type int"
	if err.Error() != "invalid pattern tag value \"^[0-9]+$\" for field Port: "+expectedMsg {
		t.Errorf("Expected error %q, got %q", "invalid pattern tag value \"^[0-9]+$\" for field Port: "+expectedMsg, err.Error())
	}
}

func TestLoad_WithValidatorFactory(t *testing.T) {
	type EmailConfig struct {
		Email    string `key:"EMAIL" email:"true"`
		Username string `key:"USERNAME"`
	}

	// Custom factory that validates email fields based on a custom tag
	emailFactory := func(fieldType reflect.StructField, registry ValidatorRegistry) error {
		if fieldType.Tag.Get("email") == "true" {
			registry(func(value any) error {
				email := value.(string)
				if email == "" || !contains(email, "@") {
					return fmt.Errorf("invalid email format")
				}
				return nil
			})
		}
		return nil
	}

	// Test valid email
	os.Clearenv()
	os.Setenv("EMAIL", "user@example.com")
	os.Setenv("USERNAME", "testuser")

	var cfg EmailConfig
	err := Load(context.Background(), &cfg, WithValidatorFactory(emailFactory))
	if err != nil {
		t.Fatalf("Expected no error for valid email, got %v", err)
	}

	if cfg.Email != "user@example.com" {
		t.Errorf("Expected Email to be 'user@example.com', got %q", cfg.Email)
	}

	// Test invalid email
	os.Clearenv()
	os.Setenv("EMAIL", "invalid-email")
	os.Setenv("USERNAME", "testuser")

	var cfg2 EmailConfig
	err = Load(context.Background(), &cfg2, WithValidatorFactory(emailFactory))
	if err == nil {
		t.Fatal("Expected error for invalid email")
	}

	if err.Error() != "EMAIL: invalid email format" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestLoad_WithValidatorFactory_MultipleFactories(t *testing.T) {
	type Config struct {
		Email    string `key:"EMAIL" email:"true"`
		Username string `key:"USERNAME" alphanum:"true"`
	}

	emailFactory := func(fieldType reflect.StructField, registry ValidatorRegistry) error {
		if fieldType.Tag.Get("email") == "true" {
			registry(func(value any) error {
				email := value.(string)
				if !contains(email, "@") {
					return fmt.Errorf("must contain @")
				}
				return nil
			})
		}
		return nil
	}

	alphanumFactory := func(fieldType reflect.StructField, registry ValidatorRegistry) error {
		if fieldType.Tag.Get("alphanum") == "true" {
			registry(func(value any) error {
				username := value.(string)
				for _, c := range username {
					if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
						return fmt.Errorf("must be alphanumeric")
					}
				}
				return nil
			})
		}
		return nil
	}

	// Test both validators pass
	os.Clearenv()
	os.Setenv("EMAIL", "user@example.com")
	os.Setenv("USERNAME", "user123")

	var cfg Config
	err := Load(context.Background(), &cfg, WithValidatorFactory(emailFactory), WithValidatorFactory(alphanumFactory))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test email validator fails
	os.Clearenv()
	os.Setenv("EMAIL", "invalid")
	os.Setenv("USERNAME", "user123")

	var cfg2 Config
	err = Load(context.Background(), &cfg2, WithValidatorFactory(emailFactory), WithValidatorFactory(alphanumFactory))
	if err == nil {
		t.Fatal("Expected error for invalid email")
	}

	// Test username validator fails
	os.Clearenv()
	os.Setenv("EMAIL", "user@example.com")
	os.Setenv("USERNAME", "user-123")

	var cfg3 Config
	err = Load(context.Background(), &cfg3, WithValidatorFactory(emailFactory), WithValidatorFactory(alphanumFactory))
	if err == nil {
		t.Fatal("Expected error for invalid username")
	}
}

// TestLoad_MultipleRuntimeErrors tests that multiple runtime errors are collected and reported together
func TestLoad_MultipleRuntimeErrors(t *testing.T) {
	type Config struct {
		RequiredField string `key:"REQ" keyRequired:"true"`
		BadInt        int    `key:"BAD_INT"`
		OutOfRange    int    `key:"OUT_OF_RANGE" min:"10" max:"100"`
	}

	os.Clearenv()
	os.Setenv("BAD_INT", "not-a-number")
	os.Setenv("OUT_OF_RANGE", "500")

	var cfg Config
	err := Load(context.Background(), &cfg)

	// Should return a ConfigErrors type
	configErr, ok := err.(*ConfigErrors)
	if !ok {
		t.Fatalf("Expected *ConfigErrors, got %T", err)
	}

	if configErr.Len() != 3 {
		t.Errorf("Expected 3 errors, got %d", configErr.Len())
	}

	// Check error message format
	errMsg := err.Error()
	if !contains(errMsg, "REQ") {
		t.Error("Error should mention REQ")
	}
	if !contains(errMsg, "BAD_INT") {
		t.Error("Error should mention BAD_INT")
	}
	if !contains(errMsg, "OUT_OF_RANGE") {
		t.Error("Error should mention OUT_OF_RANGE")
	}

	// Verify error message contains expected substrings
	if !contains(errMsg, ErrMissingConfigKey.Error()) {
		t.Error("Error should mention required environment variable")
	}
	if !contains(errMsg, "ParseInt") {
		t.Errorf("Error should mention invalid int value. Got %s", errMsg)
	}
	if !contains(errMsg, "exceeds maximum") {
		t.Error("Error should mention exceeds maximum")
	}
}

// TestLoad_NestedStructWithMultipleErrors tests error collection in nested structs
func TestLoad_NestedStructWithMultipleErrors(t *testing.T) {
	type Database struct {
		Port int    `key:"DB_PORT" min:"1024" max:"65535"`
		Host string `key:"DB_HOST" required:"true"`
	}

	type Config struct {
		Database Database
		APIKey   string `key:"API_KEY" required:"true"`
	}

	os.Clearenv()
	os.Setenv("DB_PORT", "500") // Below minimum

	var cfg Config
	err := Load(context.Background(), &cfg)

	configErr, ok := err.(*ConfigErrors)
	if !ok {
		t.Fatalf("Expected *ConfigErrors, got %T", err)
	}

	if configErr.Len() != 3 {
		t.Errorf("Expected 3 errors (DB_PORT, DB_HOST, API_KEY), got %d", configErr.Len())
	}

	errMsg := err.Error()
	if !contains(errMsg, "DB_PORT") {
		t.Error("Error should mention DB_PORT")
	}
	if !contains(errMsg, "DB_HOST") {
		t.Error("Error should mention DB_HOST")
	}
	if !contains(errMsg, "API_KEY") {
		t.Error("Error should mention API_KEY")
	}
}

// TestConfigErrors_Unwrap tests the Unwrap method for error inspection
func TestConfigErrors_Unwrap(t *testing.T) {
	type Config struct {
		Field1 string `key:"FIELD1" required:"true"`
		Field2 string `key:"FIELD2" required:"true"`
	}

	os.Clearenv()

	var cfg Config
	err := Load(context.Background(), &cfg)

	configErr, ok := err.(*ConfigErrors)
	if !ok {
		t.Fatalf("Expected *ConfigErrors, got %T", err)
	}

	unwrapped := configErr.Unwrap()
	if len(unwrapped) != 2 {
		t.Errorf("Expected 2 unwrapped errors, got %d", len(unwrapped))
	}

	// Verify the unwrapped errors are not nil
	for i, e := range unwrapped {
		if e == nil {
			t.Errorf("Unwrapped error %d should not be nil", i)
		}
	}
}

// TestLoad_ConfigErrorsStillFailFast tests that configuration errors (bad tags, invalid validators) still fail fast
func TestLoad_ConfigErrorsStillFailFast(t *testing.T) {
	type BadConfig struct {
		Port int `key:"PORT" min:"not-a-number"`
	}

	os.Clearenv()
	os.Setenv("PORT", "8080")

	var cfg BadConfig
	err := Load(context.Background(), &cfg)

	// Should get a regular error, not ConfigErrors, since this is a configuration error
	if err == nil {
		t.Fatal("Expected error for invalid min tag")
	}

	_, ok := err.(*ConfigErrors)
	if ok {
		t.Error("Configuration errors should not be wrapped in ConfigErrors, they should fail fast")
	}

	// Verify it's a configuration error about the invalid tag
	if !contains(err.Error(), "invalid min tag") {
		t.Errorf("Expected configuration error about invalid min tag, got: %v", err)
	}
}

// TestLoad_MultipleValidationErrors tests that all validators run even when some fail
func TestLoad_MultipleValidationErrors(t *testing.T) {
	type Config struct {
		Port1 int `key:"PORT1" min:"1024" max:"65535"`
		Port2 int `key:"PORT2" min:"1024" max:"65535"`
		Port3 int `key:"PORT3" min:"1024" max:"65535"`
	}

	os.Clearenv()
	os.Setenv("PORT1", "500")   // Below minimum
	os.Setenv("PORT2", "70000") // Above maximum
	os.Setenv("PORT3", "abc")   // Invalid type

	var cfg Config
	err := Load(context.Background(), &cfg)

	configErr, ok := err.(*ConfigErrors)
	if !ok {
		t.Fatalf("Expected *ConfigErrors, got %T", err)
	}

	if configErr.Len() != 3 {
		t.Errorf("Expected 3 errors, got %d: %v", configErr.Len(), err)
	}

	errMsg := err.Error()
	// All three errors should be reported
	if !contains(errMsg, "PORT1") || !contains(errMsg, "PORT2") || !contains(errMsg, "PORT3") {
		t.Errorf("All three port errors should be reported, got: %v", errMsg)
	}
}

func TestLoad_CustomKeystore_ReadsFromStore(t *testing.T) {
	type Config struct {
		Value string `key:"custom_key" required:"true"`
	}

	keyStore := func(ctx context.Context, key string) (string, bool, error) {
		return key + ":OK", true, nil
	}

	config := Config{
		Value: "Value was not modified",
	}

	err := Load(context.Background(), &config, WithKeyStore(keyStore))
	if err != nil {
		t.Fatalf("Error loading custom keystore: %v", err)
	}

	if config.Value != "custom_key:OK" {
		t.Fatalf("Error loading custom keystore: expected custom_key:OK, got %s", config.Value)
	}
}

func TestLoad_CustomKeystore_PassesBackError(t *testing.T) {
	type Config struct {
		Field1 string `key:"FIELD1" required:"true"`
		Field2 string `key:"FIELD2" required:"true"`
		Field3 string `key:"FIELD3" required:"true"`
	}

	expectedError := errors.New("keystore connection failed")

	keyStore := func(ctx context.Context, key string) (string, bool, error) {
		// First field succeeds
		if key == "FIELD1" {
			return "value1", true, nil
		}
		// Second field returns a fatal keystore error
		if key == "FIELD2" {
			return "", false, expectedError
		}
		// Third field should never be reached
		return "value3", true, nil
	}

	config := Config{
		Field1: "initial1",
		Field2: "initial2",
		Field3: "initial3",
	}

	err := Load(context.Background(), &config, WithKeyStore(keyStore))
	if err == nil {
		t.Fatalf("No error was returned when loading the custom store")
	}

	// Verify the error is NOT a ConfigErrors - keystore errors should be fatal
	if _, isConfigErrors := err.(*ConfigErrors); isConfigErrors {
		t.Fatalf("Expected keystore error to be fatal, not collected in ConfigErrors")
	}

	// Verify the error is (or wraps) the expected error
	if !errors.Is(err, expectedError) {
		t.Fatalf("Expected error to be or wrap keystore error, got: %v", err)
	}

	// Verify Field1 was updated (it succeeded before the error)
	if config.Field1 != "value1" {
		t.Fatalf("Expected Field1 to be 'value1', got %s", config.Field1)
	}

	// Verify Field2 was not modified (error occurred)
	if config.Field2 != "initial2" {
		t.Fatalf("Field2 should not have been modified, got %s", config.Field2)
	}

	// Verify Field3 was not modified (processing stopped)
	if config.Field3 != "initial3" {
		t.Fatalf("Field3 should not have been modified (processing should have stopped), got %s", config.Field3)
	}
}

func TestLoad_JsonInterface_Parses(t *testing.T) {
	type Config struct {
		Value map[string]interface{} `key:"OPENAI_MODEL_PARAMS"`
	}

	keyStore := func(ctx context.Context, key string) (string, bool, error) {
		if key == "OPENAI_MODEL_PARAMS" {
			return "{\"marker\":\"present\"}", true, nil
		}
		return "", false, fmt.Errorf("bad key passed %s", key)
	}

	config := Config{}
	err := Load(context.Background(), &config, WithKeyStore(keyStore))
	if err != nil {
		t.Fatalf("Error loading custom keystore: %v", err)
	}

	if config.Value == nil {
		t.Fatalf("Value was not loaded correctly")
	}

	if config.Value["marker"] != "present" {
		t.Fatalf("Value was not loaded correctly")
	}
}

func TestLoad_JsonTypes_Parses(t *testing.T) {
	type TypedJsonStruct struct {
		Marker string `json:"marker"`
	}

	type Config struct {
		Value TypedJsonStruct `key:"OPENAI_MODEL_PARAMS"`
	}

	keyStore := func(ctx context.Context, key string) (string, bool, error) {
		if key == "OPENAI_MODEL_PARAMS" {
			return "{\"marker\":\"present\"}", true, nil
		}
		return "", false, fmt.Errorf("bad key passed %s", key)
	}

	config := Config{}
	err := Load(context.Background(), &config, WithKeyStore(keyStore))
	if err != nil {
		t.Fatalf("Error loading custom keystore: %v", err)
	}

	if config.Value.Marker != "present" {
		t.Fatalf("Value was not loaded correctly")
	}
}

func TestLoad_JsonPointerTypes_Parses(t *testing.T) {
	type TypedJsonStruct struct {
		Marker string `json:"marker"`
	}

	type Config struct {
		Value *TypedJsonStruct `key:"OPENAI_MODEL_PARAMS"`
	}

	keyStore := func(ctx context.Context, key string) (string, bool, error) {
		if key == "OPENAI_MODEL_PARAMS" {
			return "{\"marker\":\"present\"}", true, nil
		}
		return "", false, fmt.Errorf("bad key passed %s", key)
	}

	config := Config{}
	err := Load(context.Background(), &config, WithKeyStore(keyStore))
	if err != nil {
		t.Fatalf("Error loading custom keystore: %v", err)
	}

	if config.Value.Marker != "present" {
		t.Fatalf("Value was not loaded correctly")
	}
}

// Helper function for string contains check
// TODO: Use Go's standard function instead of this.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
