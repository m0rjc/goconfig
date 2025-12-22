package goconfig

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/m0rjc/goconfig/process"
)

func TestLoad_Basic(t *testing.T) {
	type Config struct {
		Port int    `key:"PORT" default:"8080"`
		Host string `key:"HOST" default:"localhost"`
	}

	mockStore := func(ctx context.Context, key string) (string, bool, error) {
		if key == "PORT" {
			return "9000", true, nil
		}
		return "", false, nil
	}

	var cfg Config
	err := Load(context.Background(), &cfg, WithKeyStore(mockStore))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Port != 9000 {
		t.Errorf("Expected Port 9000, got %d", cfg.Port)
	}
	if cfg.Host != "localhost" {
		t.Errorf("Expected Host localhost, got %s", cfg.Host)
	}
}

func TestLoad_Nested(t *testing.T) {
	type DatabaseConfig struct {
		URL string `key:"DB_URL" required:"true"`
	}
	type Config struct {
		DB DatabaseConfig
	}

	mockStore := func(ctx context.Context, key string) (string, bool, error) {
		if key == "DB_URL" {
			return "postgres://localhost:5432", true, nil
		}
		return "", false, nil
	}

	var cfg Config
	err := Load(context.Background(), &cfg, WithKeyStore(mockStore))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.DB.URL != "postgres://localhost:5432" {
		t.Errorf("Expected DB URL postgres://localhost:5432, got %s", cfg.DB.URL)
	}

	t.Run("Pointer to nested struct", func(t *testing.T) {
		type ConfigPtr struct {
			DB *DatabaseConfig
		}
		var cfgPtr ConfigPtr
		err := Load(context.Background(), &cfgPtr, WithKeyStore(mockStore))
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfgPtr.DB == nil {
			t.Fatal("Expected cfgPtr.DB to be non-nil")
		}
		if cfgPtr.DB.URL != "postgres://localhost:5432" {
			t.Errorf("Expected DB URL postgres://localhost:5432, got %s", cfgPtr.DB.URL)
		}
	})
}

func TestLoad_Pointers(t *testing.T) {
	type Config struct {
		Port *int    `key:"PORT"`
		Name *string `key:"NAME" default:"guest"`
	}

	mockStore := func(ctx context.Context, key string) (string, bool, error) {
		if key == "PORT" {
			return "8080", true, nil
		}
		return "", false, nil
	}

	var cfg Config
	err := Load(context.Background(), &cfg, WithKeyStore(mockStore))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Port == nil {
		t.Fatal("Expected Port to be non-nil")
	}
	if *cfg.Port != 8080 {
		t.Errorf("Expected *Port 8080, got %d", *cfg.Port)
	}
	if cfg.Name == nil {
		t.Fatal("Expected Name to be non-nil")
	}
	if *cfg.Name != "guest" {
		t.Errorf("Expected *Name guest, got %s", *cfg.Name)
	}
}

func TestLoad_Defaulting(t *testing.T) {
	// Referring to defaulting.md table
	type Config struct {
		Optional    string `key:"OPTIONAL"`
		KeyRequired string `key:"KEY_REQUIRED" keyRequired:"true"`
		Required    string `key:"REQUIRED" required:"true"`
		WithDefault string `key:"WITH_DEFAULT" default:"def"`
	}

	tests := []struct {
		name      string
		store     map[string]string
		present   map[string]bool
		expectErr error
		verify    func(t *testing.T, cfg Config)
	}{
		{
			name: "all set",
			store: map[string]string{
				"OPTIONAL":     "opt",
				"KEY_REQUIRED": "key",
				"REQUIRED":     "req",
				"WITH_DEFAULT": "val",
			},
			present: map[string]bool{
				"OPTIONAL":     true,
				"KEY_REQUIRED": true,
				"REQUIRED":     true,
				"WITH_DEFAULT": true,
			},
			verify: func(t *testing.T, cfg Config) {
				if cfg.Optional != "opt" || cfg.KeyRequired != "key" || cfg.Required != "req" || cfg.WithDefault != "val" {
					t.Errorf("Unexpected values: %+v", cfg)
				}
			},
		},
		{
			name: "set to empty",
			store: map[string]string{
				"OPTIONAL":     "",
				"KEY_REQUIRED": "",
				"REQUIRED":     "",
				"WITH_DEFAULT": "",
			},
			present: map[string]bool{
				"OPTIONAL":     true,
				"KEY_REQUIRED": true,
				"REQUIRED":     true,
				"WITH_DEFAULT": true,
			},
			expectErr: ErrMissingValue, // REQUIRED is set to empty
		},
		{
			name:      "unset",
			store:     map[string]string{},
			present:   map[string]bool{},
			expectErr: ErrMissingConfigKey, // KEY_REQUIRED and REQUIRED are unset
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := func(ctx context.Context, key string) (string, bool, error) {
				val := tt.store[key]
				present := tt.present[key]
				return val, present, nil
			}

			var cfg Config
			err := Load(context.Background(), &cfg, WithKeyStore(mockStore))
			if tt.expectErr != nil {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", tt.expectErr)
				}
				if !errors.Is(err, tt.expectErr) {
					t.Errorf("Expected error %v, got %v", tt.expectErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Load failed: %v", err)
				}
				tt.verify(t, cfg)
			}
		})
	}
}

func TestLoad_Options(t *testing.T) {
	type Config struct {
		Port int `key:"PORT"`
	}

	mockStore := func(ctx context.Context, key string) (string, bool, error) {
		if key == "PORT" {
			return "8080", true, nil
		}
		return "", false, nil
	}

	t.Run("Custom Parser", func(t *testing.T) {
		type Port int
		type Config struct {
			Port Port `key:"PORT"`
		}
		var cfg Config
		// Custom parser for the custom Port type
		handler := process.NewCustomHandler(func(rawValue string) (Port, error) {
			return Port(9000), nil
		})

		err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[Port](handler))
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if cfg.Port != 9000 {
			t.Errorf("Expected Port 9000, got %d", cfg.Port)
		}
	})

	t.Run("Custom Validator", func(t *testing.T) {
		type Port int
		type Config struct {
			Port Port `key:"PORT"`
		}
		var cfg Config
		handler := process.NewCustomHandler(func(rawValue string) (Port, error) {
			v, err := strconv.Atoi(rawValue)
			return Port(v), err
		}, func(value Port) error {
			if value != 8080 {
				return errors.New("wrong port")
			}
			return nil
		})

		err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[Port](handler))
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
	})
}

func TestLoad_EnvVars(t *testing.T) {
	os.Setenv("TEST_PORT", "7000")
	defer os.Unsetenv("TEST_PORT")

	type Config struct {
		Port int `key:"TEST_PORT"`
	}

	var cfg Config
	err := Load(context.Background(), &cfg)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Port != 7000 {
		t.Errorf("Expected Port 7000, got %d", cfg.Port)
	}
}

func TestLoad_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("Failure to instantiate pipeline", func(t *testing.T) {
		// Use a type that internal/process doesn't support (like a channel)
		type Config struct {
			Chan chan int `key:"CHAN"`
		}
		var cfg Config
		mockStore := func(ctx context.Context, key string) (string, bool, error) {
			return "something", true, nil
		}
		err := Load(ctx, &cfg, WithKeyStore(mockStore))
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "setting up field process Chan") {
			t.Errorf("Expected setup error, got: %v", err)
		}
	})

	t.Run("Failure in getCustomValidators", func(t *testing.T) {
		mockStore := func(ctx context.Context, key string) (string, bool, error) {
			return "8080", true, nil
		}
		// Validator factories are no longer supported in this way
		// Instead we use custom types with handlers
		type CustomPort int
		type CustomConfig struct {
			Port CustomPort `key:"PORT"`
		}
		var customCfg CustomConfig

		failingHandler := process.NewCustomHandler(func(rawValue string) (CustomPort, error) {
			return 0, errors.New("factory failure")
		})

		err := Load(ctx, &customCfg, WithKeyStore(mockStore), WithCustomType[CustomPort](failingHandler))
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "factory failure") {
			t.Errorf("Expected factory error, got: %v", err)
		}
	})

	t.Run("Pipeline failure (invalid input)", func(t *testing.T) {
		type Config struct {
			Port int `key:"PORT"`
		}
		var cfg Config
		mockStore := func(ctx context.Context, key string) (string, bool, error) {
			if key == "PORT" {
				return "foo", true, nil
			}
			return "", false, nil
		}
		err := Load(ctx, &cfg, WithKeyStore(mockStore))
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		var cfgErrs *ConfigErrors
		if !errors.As(err, &cfgErrs) {
			t.Fatalf("Expected ConfigErrors, got %T", err)
		}
		if cfgErrs.Len() != 1 {
			t.Errorf("Expected 1 error, got %d", cfgErrs.Len())
		}
		if cfgErrs.Errors[0].Key != "PORT" {
			t.Errorf("Expected error for PORT, got %s", cfgErrs.Errors[0].Key)
		}
	})

	t.Run("Nested struct setup failure (shortcuts out)", func(t *testing.T) {
		type Inner struct {
			Chan chan int `key:"CHAN"`
		}
		type Config struct {
			Inner Inner
		}
		var cfg Config
		mockStore := func(ctx context.Context, key string) (string, bool, error) {
			return "something", true, nil
		}
		err := Load(ctx, &cfg, WithKeyStore(mockStore))
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "setting up field process Inner.Chan") {
			t.Errorf("Expected setup error for nested field, got: %v", err)
		}
	})

	t.Run("Nested struct value failure (collects errors)", func(t *testing.T) {
		type Inner struct {
			Port int `key:"PORT"`
		}
		type Config struct {
			Inner Inner
		}
		var cfg Config
		mockStore := func(ctx context.Context, key string) (string, bool, error) {
			if key == "PORT" {
				return "foo", true, nil
			}
			return "", false, nil
		}
		err := Load(ctx, &cfg, WithKeyStore(mockStore))
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		var cfgErrs *ConfigErrors
		if !errors.As(err, &cfgErrs) {
			t.Fatalf("Expected ConfigErrors, got %T", err)
		}
		if cfgErrs.Len() != 1 {
			t.Errorf("Expected 1 error, got %d", cfgErrs.Len())
		}
		if cfgErrs.Errors[0].Key != "PORT" {
			t.Errorf("Expected error for PORT, got %s", cfgErrs.Errors[0].Key)
		}
	})

	t.Run("Private fields", func(t *testing.T) {
		t.Run("Private field without key tag is ignored", func(t *testing.T) {
			type Config struct {
				secret string
				Port   int `key:"PORT" default:"8080"`
			}
			var cfg Config
			err := Load(ctx, &cfg, WithKeyStore(func(ctx context.Context, key string) (string, bool, error) {
				return "", false, nil
			}))
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if cfg.secret != "" {
				t.Errorf("Expected private field to be empty, got %q", cfg.secret)
			}
		})

		t.Run("Private field with key tag returns setup error", func(t *testing.T) {
			type Config struct {
				secret string `key:"SECRET"`
			}
			var cfg Config
			err := Load(ctx, &cfg)
			if err == nil {
				t.Fatal("Expected setup error for private field with key tag, got nil")
			}
			if !strings.Contains(err.Error(), "unexported") {
				t.Errorf("Expected error message to mention unexported, got %q", err.Error())
			}
		})
	})

	t.Run("Keystore error (fast-fail)", func(t *testing.T) {
		type Config struct {
			Port int `key:"PORT"`
		}
		var cfg Config
		keystoreErr := errors.New("keystore failure")
		mockStore := func(ctx context.Context, key string) (string, bool, error) {
			return "", false, keystoreErr
		}
		err := Load(ctx, &cfg, WithKeyStore(mockStore))
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !errors.Is(err, keystoreErr) {
			t.Errorf("Expected error %v, got %v", keystoreErr, err)
		}
	})

	t.Run("Invalid config argument", func(t *testing.T) {
		t.Run("Non-pointer", func(t *testing.T) {
			type Config struct {
				Port int `key:"PORT"`
			}
			var cfg Config
			err := Load(ctx, cfg) // Passing value, not pointer
			if err == nil {
				t.Fatal("Expected error when passing non-pointer, got nil")
			}
			if !strings.Contains(err.Error(), "must be a pointer to a struct") {
				t.Errorf("Expected 'must be a pointer to a struct' error, got: %v", err)
			}
		})

		t.Run("Pointer to non-struct", func(t *testing.T) {
			var port int
			err := Load(ctx, &port) // Passing pointer to int
			if err == nil {
				t.Fatal("Expected error when passing pointer to non-struct, got nil")
			}
			if !strings.Contains(err.Error(), "must be a pointer to a struct") {
				t.Errorf("Expected 'must be a pointer to a struct' error, got: %v", err)
			}
		})

		t.Run("Nil pointer", func(t *testing.T) {
			type Config struct {
				Port int `key:"PORT"`
			}
			var cfg *Config = nil
			err := Load(ctx, cfg)
			if err == nil {
				t.Fatal("Expected error when passing nil pointer, got nil")
			}
			if !strings.Contains(err.Error(), "must be a pointer to a struct") {
				t.Errorf("Expected 'must be a pointer to a struct' error, got: %v", err)
			}
		})

		t.Run("Literal nil", func(t *testing.T) {
			err := Load(ctx, nil)
			if err == nil {
				t.Fatal("Expected error when passing literal nil, got nil")
			}
			if !strings.Contains(err.Error(), "must be a pointer to a struct") {
				t.Errorf("Expected 'must be a pointer to a struct' error, got: %v", err)
			}
		})
	})
}
