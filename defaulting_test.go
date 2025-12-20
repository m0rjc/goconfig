package goconfigtools

// This test file corresponds to the Defaulting section in README.md

import (
	"context"
	"errors"
	"testing"
)

// TestDefaultingBehavior_NoRequiredTags tests the defaulting behavior when no required tags are present
func TestDefaultingBehavior_NoRequiredTags(t *testing.T) {
	type Config struct {
		Value string `key:"FOO"`
	}

	tests := []struct {
		name          string
		envValue      string
		envPresent    bool
		initialValue  string
		expectedValue string
		expectedError bool
		checkSentinel error
	}{
		{
			name:          "export FOO=bar - Value becomes bar",
			envValue:      "bar",
			envPresent:    true,
			initialValue:  "initial",
			expectedValue: "bar",
			expectedError: false,
		},
		{
			name:          "export FOO= - Value becomes empty",
			envValue:      "",
			envPresent:    true,
			initialValue:  "initial",
			expectedValue: "",
			expectedError: false,
		},
		{
			name:          "unset FOO - Value remains unchanged",
			envValue:      "",
			envPresent:    false,
			initialValue:  "initial",
			expectedValue: "initial",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyStore := func(ctx context.Context, key string) (string, bool, error) {
				if key == "FOO" {
					return tt.envValue, tt.envPresent, nil
				}
				return "", false, nil
			}

			cfg := Config{Value: tt.initialValue}
			err := Load(context.Background(), &cfg, WithKeyStore(keyStore))

			if tt.expectedError && err == nil {
				t.Fatal("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Fatalf("Expected no error but got: %v", err)
			}

			if cfg.Value != tt.expectedValue {
				t.Errorf("Expected Value to be %q, got %q", tt.expectedValue, cfg.Value)
			}
		})
	}
}

// TestDefaultingBehavior_KeyRequiredTag tests the defaulting behavior when keyRequired="true"
func TestDefaultingBehavior_KeyRequiredTag(t *testing.T) {
	type Config struct {
		Value string `key:"FOO" keyRequired:"true"`
	}

	tests := []struct {
		name          string
		envValue      string
		envPresent    bool
		initialValue  string
		expectedValue string
		expectedError bool
		checkSentinel error
	}{
		{
			name:          "export FOO=bar - Value becomes bar",
			envValue:      "bar",
			envPresent:    true,
			initialValue:  "initial",
			expectedValue: "bar",
			expectedError: false,
		},
		{
			name:          "export FOO= - Value becomes empty",
			envValue:      "",
			envPresent:    true,
			initialValue:  "initial",
			expectedValue: "",
			expectedError: false,
		},
		{
			name:          "unset FOO - Error (ErrMissingConfigKey)",
			envValue:      "",
			envPresent:    false,
			initialValue:  "initial",
			expectedValue: "initial",
			expectedError: true,
			checkSentinel: ErrMissingConfigKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyStore := func(ctx context.Context, key string) (string, bool, error) {
				if key == "FOO" {
					return tt.envValue, tt.envPresent, nil
				}
				return "", false, nil
			}

			cfg := Config{Value: tt.initialValue}
			err := Load(context.Background(), &cfg, WithKeyStore(keyStore))

			if tt.expectedError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if tt.checkSentinel != nil {
					if !errors.Is(err, tt.checkSentinel) {
						t.Errorf("Expected error to wrap %v, but errors.Is returned false. Got error: %v", tt.checkSentinel, err)
					}
				}
				// Value should remain unchanged on error
				if cfg.Value != tt.expectedValue {
					t.Errorf("Expected Value to remain %q on error, got %q", tt.expectedValue, cfg.Value)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error but got: %v", err)
				}
				if cfg.Value != tt.expectedValue {
					t.Errorf("Expected Value to be %q, got %q", tt.expectedValue, cfg.Value)
				}
			}
		})
	}
}

// TestDefaultingBehavior_RequiredTag tests the defaulting behavior when required="true"
func TestDefaultingBehavior_RequiredTag(t *testing.T) {
	type Config struct {
		Value string `key:"FOO" required:"true"`
	}

	tests := []struct {
		name          string
		envValue      string
		envPresent    bool
		initialValue  string
		expectedValue string
		expectedError bool
		checkSentinel error
	}{
		{
			name:          "export FOO=bar - Value becomes bar",
			envValue:      "bar",
			envPresent:    true,
			initialValue:  "initial",
			expectedValue: "bar",
			expectedError: false,
		},
		{
			name:          "export FOO= - Error (ErrMissingValue)",
			envValue:      "",
			envPresent:    true,
			initialValue:  "initial",
			expectedValue: "initial",
			expectedError: true,
			checkSentinel: ErrMissingValue,
		},
		{
			name:          "unset FOO - Error (ErrMissingConfigKey)",
			envValue:      "",
			envPresent:    false,
			initialValue:  "initial",
			expectedValue: "initial",
			expectedError: true,
			checkSentinel: ErrMissingConfigKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyStore := func(ctx context.Context, key string) (string, bool, error) {
				if key == "FOO" {
					return tt.envValue, tt.envPresent, nil
				}
				return "", false, nil
			}

			cfg := Config{Value: tt.initialValue}
			err := Load(context.Background(), &cfg, WithKeyStore(keyStore))

			if tt.expectedError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if tt.checkSentinel != nil {
					if !errors.Is(err, tt.checkSentinel) {
						t.Errorf("Expected error to wrap %v, but errors.Is returned false. Got error: %v", tt.checkSentinel, err)
					}
				}
				// Value should remain unchanged on error
				if cfg.Value != tt.expectedValue {
					t.Errorf("Expected Value to remain %q on error, got %q", tt.expectedValue, cfg.Value)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error but got: %v", err)
				}
				if cfg.Value != tt.expectedValue {
					t.Errorf("Expected Value to be %q, got %q", tt.expectedValue, cfg.Value)
				}
			}
		})
	}
}

// TestDefaultingBehavior_WithDefaults tests the interaction of defaults with the behavior grid
func TestDefaultingBehavior_WithDefaults(t *testing.T) {
	type Config struct {
		NoTag       string `key:"NO_TAG" default:"default_value"`
		KeyRequired string `key:"KEY_REQ" default:"default_value" keyRequired:"true"`
		Required    string `key:"REQUIRED" default:"default_value" required:"true"`
	}

	tests := []struct {
		name          string
		setupKeyStore func(ctx context.Context, key string) (string, bool, error)
		expected      Config
		expectedError bool
	}{
		{
			name: "All keys set to non-empty - all use env values",
			setupKeyStore: func(ctx context.Context, key string) (string, bool, error) {
				values := map[string]string{
					"NO_TAG":   "env_value",
					"KEY_REQ":  "env_value",
					"REQUIRED": "env_value",
				}
				if val, ok := values[key]; ok {
					return val, true, nil
				}
				return "", false, nil
			},
			expected: Config{
				NoTag:       "env_value",
				KeyRequired: "env_value",
				Required:    "env_value",
			},
			expectedError: false,
		},
		{
			name: "All keys set to empty - required fails, others use empty",
			setupKeyStore: func(ctx context.Context, key string) (string, bool, error) {
				values := map[string]string{
					"NO_TAG":   "",
					"KEY_REQ":  "",
					"REQUIRED": "",
				}
				if _, ok := values[key]; ok {
					return "", true, nil
				}
				return "", false, nil
			},
			expectedError: true, // Required field will fail
		},
		{
			name: "All keys unset - use defaults where possible, fail where required",
			setupKeyStore: func(ctx context.Context, key string) (string, bool, error) {
				return "", false, nil
			},
			expected: Config{
				NoTag:       "default_value",
				KeyRequired: "default_value",
				Required:    "default_value",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{}
			err := Load(context.Background(), &cfg, WithKeyStore(tt.setupKeyStore))

			if tt.expectedError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error but got: %v", err)
				}
				if cfg != tt.expected {
					t.Errorf("Expected config %+v, got %+v", tt.expected, cfg)
				}
			}
		})
	}
}

// TestDefaultingBehavior_MultipleFields tests multiple fields with different behaviors
func TestDefaultingBehavior_MultipleFields(t *testing.T) {
	type Config struct {
		Field1 string `key:"FIELD1"`                    // No required tags
		Field2 string `key:"FIELD2" keyRequired:"true"` // keyRequired
		Field3 string `key:"FIELD3" required:"true"`    // required
		Field4 string `key:"FIELD4" default:"default4"` // default, no required
	}

	// Test successful load with all requirements met
	t.Run("All requirements met", func(t *testing.T) {
		keyStore := func(ctx context.Context, key string) (string, bool, error) {
			values := map[string]string{
				"FIELD1": "value1",
				"FIELD2": "value2",
				"FIELD3": "value3",
			}
			if val, ok := values[key]; ok {
				return val, true, nil
			}
			return "", false, nil
		}

		cfg := Config{Field1: "initial1"}
		err := Load(context.Background(), &cfg, WithKeyStore(keyStore))
		if err != nil {
			t.Fatalf("Expected no error but got: %v", err)
		}

		if cfg.Field1 != "value1" {
			t.Errorf("Expected Field1 to be 'value1', got %q", cfg.Field1)
		}
		if cfg.Field2 != "value2" {
			t.Errorf("Expected Field2 to be 'value2', got %q", cfg.Field2)
		}
		if cfg.Field3 != "value3" {
			t.Errorf("Expected Field3 to be 'value3', got %q", cfg.Field3)
		}
		if cfg.Field4 != "default4" {
			t.Errorf("Expected Field4 to be 'default4', got %q", cfg.Field4)
		}
	})

	// Test that FIELD1 (no required) preserves initial value when unset
	t.Run("FIELD1 unset preserves initial value", func(t *testing.T) {
		keyStore := func(ctx context.Context, key string) (string, bool, error) {
			values := map[string]string{
				"FIELD2": "value2",
				"FIELD3": "value3",
			}
			if val, ok := values[key]; ok {
				return val, true, nil
			}
			return "", false, nil
		}

		cfg := Config{Field1: "preserved"}
		err := Load(context.Background(), &cfg, WithKeyStore(keyStore))
		if err != nil {
			t.Fatalf("Expected no error but got: %v", err)
		}

		if cfg.Field1 != "preserved" {
			t.Errorf("Expected Field1 to remain 'preserved', got %q", cfg.Field1)
		}
	})

	// Test that FIELD2 (keyRequired) fails when unset
	t.Run("FIELD2 keyRequired fails when unset", func(t *testing.T) {
		keyStore := func(ctx context.Context, key string) (string, bool, error) {
			values := map[string]string{
				"FIELD1": "value1",
				"FIELD3": "value3",
			}
			if val, ok := values[key]; ok {
				return val, true, nil
			}
			return "", false, nil
		}

		cfg := Config{}
		err := Load(context.Background(), &cfg, WithKeyStore(keyStore))
		if err == nil {
			t.Fatal("Expected error for missing FIELD2")
		}
		if !errors.Is(err, ErrMissingConfigKey) {
			t.Errorf("Expected ErrMissingConfigKey, got: %v", err)
		}
	})

	// Test that FIELD2 (keyRequired) accepts empty value
	t.Run("FIELD2 keyRequired accepts empty value", func(t *testing.T) {
		keyStore := func(ctx context.Context, key string) (string, bool, error) {
			values := map[string]string{
				"FIELD1": "value1",
				"FIELD2": "",
				"FIELD3": "value3",
			}
			if val, ok := values[key]; ok {
				return val, true, nil
			}
			return "", false, nil
		}

		cfg := Config{Field2: "initial"}
		err := Load(context.Background(), &cfg, WithKeyStore(keyStore))
		if err != nil {
			t.Fatalf("Expected no error but got: %v", err)
		}

		if cfg.Field2 != "" {
			t.Errorf("Expected Field2 to be empty, got %q", cfg.Field2)
		}
	})

	// Test that FIELD3 (required) fails when value is empty
	t.Run("FIELD3 required fails when value is empty", func(t *testing.T) {
		keyStore := func(ctx context.Context, key string) (string, bool, error) {
			values := map[string]string{
				"FIELD1": "value1",
				"FIELD2": "value2",
				"FIELD3": "",
			}
			if val, ok := values[key]; ok {
				return val, true, nil
			}
			return "", false, nil
		}

		cfg := Config{}
		err := Load(context.Background(), &cfg, WithKeyStore(keyStore))
		if err == nil {
			t.Fatal("Expected error for empty FIELD3")
		}
		if !errors.Is(err, ErrMissingValue) {
			t.Errorf("Expected ErrMissingValue, got: %v", err)
		}
	})

	// Test that FIELD3 (required) fails when unset
	t.Run("FIELD3 required fails when unset", func(t *testing.T) {
		keyStore := func(ctx context.Context, key string) (string, bool, error) {
			values := map[string]string{
				"FIELD1": "value1",
				"FIELD2": "value2",
			}
			if val, ok := values[key]; ok {
				return val, true, nil
			}
			return "", false, nil
		}

		cfg := Config{}
		err := Load(context.Background(), &cfg, WithKeyStore(keyStore))
		if err == nil {
			t.Fatal("Expected error for missing FIELD3")
		}
		if !errors.Is(err, ErrMissingConfigKey) {
			t.Errorf("Expected ErrMissingConfigKey, got: %v", err)
		}
	})
}

// TestDefaultingBehavior_NestedStructs tests defaulting behavior in nested structs
func TestDefaultingBehavior_NestedStructs(t *testing.T) {
	type Database struct {
		Host     string `key:"DB_HOST"`
		Port     string `key:"DB_PORT" keyRequired:"true"`
		Password string `key:"DB_PASS" required:"true"`
	}

	type Config struct {
		Database Database
		AppName  string `key:"APP_NAME" default:"myapp"`
	}

	t.Run("All nested requirements met", func(t *testing.T) {
		keyStore := func(ctx context.Context, key string) (string, bool, error) {
			values := map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
				"DB_PASS": "secret",
			}
			if val, ok := values[key]; ok {
				return val, true, nil
			}
			return "", false, nil
		}

		cfg := Config{}
		err := Load(context.Background(), &cfg, WithKeyStore(keyStore))
		if err != nil {
			t.Fatalf("Expected no error but got: %v", err)
		}

		if cfg.Database.Host != "localhost" {
			t.Errorf("Expected Host to be 'localhost', got %q", cfg.Database.Host)
		}
		if cfg.Database.Port != "5432" {
			t.Errorf("Expected Port to be '5432', got %q", cfg.Database.Port)
		}
		if cfg.Database.Password != "secret" {
			t.Errorf("Expected Password to be 'secret', got %q", cfg.Database.Password)
		}
		if cfg.AppName != "myapp" {
			t.Errorf("Expected AppName to be 'myapp', got %q", cfg.AppName)
		}
	})

	t.Run("Nested keyRequired fails when missing", func(t *testing.T) {
		keyStore := func(ctx context.Context, key string) (string, bool, error) {
			values := map[string]string{
				"DB_HOST": "localhost",
				"DB_PASS": "secret",
			}
			if val, ok := values[key]; ok {
				return val, true, nil
			}
			return "", false, nil
		}

		cfg := Config{}
		err := Load(context.Background(), &cfg, WithKeyStore(keyStore))
		if err == nil {
			t.Fatal("Expected error for missing DB_PORT")
		}
		if !errors.Is(err, ErrMissingConfigKey) {
			t.Errorf("Expected ErrMissingConfigKey, got: %v", err)
		}
	})

	t.Run("Nested required fails when empty", func(t *testing.T) {
		keyStore := func(ctx context.Context, key string) (string, bool, error) {
			values := map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
				"DB_PASS": "",
			}
			if val, ok := values[key]; ok {
				return val, true, nil
			}
			return "", false, nil
		}

		cfg := Config{}
		err := Load(context.Background(), &cfg, WithKeyStore(keyStore))
		if err == nil {
			t.Fatal("Expected error for empty DB_PASS")
		}
		if !errors.Is(err, ErrMissingValue) {
			t.Errorf("Expected ErrMissingValue, got: %v", err)
		}
	})
}

// TestDefaultingBehavior_TypedFields tests defaulting behavior with various types
func TestDefaultingBehavior_TypedFields(t *testing.T) {
	type Config struct {
		IntField  int    `key:"INT_FIELD"`
		BoolField bool   `key:"BOOL_FIELD"`
		StrField  string `key:"STR_FIELD" required:"true"`
	}

	t.Run("Unset int and bool preserve zero values", func(t *testing.T) {
		keyStore := func(ctx context.Context, key string) (string, bool, error) {
			if key == "STR_FIELD" {
				return "value", true, nil
			}
			return "", false, nil
		}

		cfg := Config{IntField: 42, BoolField: true}
		err := Load(context.Background(), &cfg, WithKeyStore(keyStore))
		if err != nil {
			t.Fatalf("Expected no error but got: %v", err)
		}

		// When unset, values should remain unchanged
		if cfg.IntField != 42 {
			t.Errorf("Expected IntField to remain 42, got %d", cfg.IntField)
		}
		if cfg.BoolField != true {
			t.Errorf("Expected BoolField to remain true, got %v", cfg.BoolField)
		}
	})

	t.Run("Set empty string for required field fails", func(t *testing.T) {
		keyStore := func(ctx context.Context, key string) (string, bool, error) {
			if key == "STR_FIELD" {
				return "", true, nil
			}
			return "", false, nil
		}

		cfg := Config{}
		err := Load(context.Background(), &cfg, WithKeyStore(keyStore))
		if err == nil {
			t.Fatal("Expected error for empty required field")
		}
		if !errors.Is(err, ErrMissingValue) {
			t.Errorf("Expected ErrMissingValue, got: %v", err)
		}
	})
}
