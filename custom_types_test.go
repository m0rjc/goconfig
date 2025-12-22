package goconfig

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

// NonConvertibleHandler is a TypedHandler[string] that returns an int from Build
type NonConvertibleHandler struct{}

func (n NonConvertibleHandler) GetParser() FieldProcessor[string] { return nil }
func (n NonConvertibleHandler) GetWrapper() Wrapper[string]       { return nil }
func (n NonConvertibleHandler) Build(tags reflect.StructTag) (FieldProcessor[any], error) {
	return func(rawValue string) (any, error) {
		return struct{ A int }{A: 123}, nil // Returns a struct instead of string
	}, nil
}

// Explore what can be done with custom types
func TestLoad_WithCustomTypes(t *testing.T) {
	t.Run("A type handler can be registered for a custom struct", func(t *testing.T) {
		type CustomStruct struct {
			Field1 string
		}
		mockStore := func(ctx context.Context, key string) (string, bool, error) {
			if key == "CUSTOM_STRUCT" {
				return "--Marker--", true, nil
			}
			return "", false, nil
		}
		mockParser := func(value string) (CustomStruct, error) {
			return CustomStruct{Field1: value}, nil
		}
		mockHandler := NewCustomHandler[CustomStruct](mockParser)

		t.Run("struct as value", func(t *testing.T) {
			type Config struct {
				Value CustomStruct `key:"CUSTOM_STRUCT"`
			}

			config := Config{Value: CustomStruct{Field1: ""}}
			err := Load(context.Background(), &config,
				WithCustomType[CustomStruct](mockHandler),
				WithKeyStore(mockStore))
			if err != nil {
				t.Fatalf("Load failed: %v", err)
			}
			if config.Value.Field1 != "--Marker--" {
				t.Errorf("Expected Field1 to be set to --Marker--, got %s", config.Value.Field1)
			}
		})

		t.Run("struct as pointer", func(t *testing.T) {
			type Config struct {
				Value *CustomStruct `key:"CUSTOM_STRUCT"`
			}

			config := Config{Value: nil}
			err := Load(context.Background(), &config,
				WithCustomType[CustomStruct](mockHandler),
				WithKeyStore(mockStore))
			if err != nil {
				t.Fatalf("Load failed: %v", err)
			}
			if config.Value == nil {
				t.Fatal("Expected config.Value to be non-nil")
			}
			if config.Value.Field1 != "--Marker--" {
				t.Errorf("Expected Field1 to be set to --Marker--, got %s", config.Value.Field1)
			}
		})
	})

	t.Run("Custom type can be registered for a custom enum (string kind)", func(t *testing.T) {
		type CustomEnum string
		const (
			CustomEnum1 CustomEnum = "--Marker--1--"
			CustomEnum2 CustomEnum = "--Marker--2--"
		)
		mockStore := func(ctx context.Context, key string) (string, bool, error) {
			if key == "CUSTOM_ENUM_1" {
				return string(CustomEnum1), true, nil
			}
			if key == "CUSTOM_ENUM_2" {
				return string(CustomEnum2), true, nil
			}
			if key == "SOME_OTHER_KEY" {
				return "foo", true, nil
			}
			return "", false, nil
		}
		expectedError := errors.New("expected CustomEnum")
		mockHandler := NewCustomHandler(func(value string) (CustomEnum, error) {
			return CustomEnum(value), nil
		}, func(value CustomEnum) error {
			if value != CustomEnum1 && value != CustomEnum2 {
				return expectedError
			}
			return nil
		})

		t.Run("string enum as value", func(t *testing.T) {
			type Config struct {
				Value  CustomEnum `key:"CUSTOM_ENUM_1"`
				Value2 CustomEnum `key:"CUSTOM_ENUM_2"`
				Other  string     `key:"SOME_OTHER_KEY"`
			}

			config := Config{}
			err := Load(context.Background(), &config,
				WithCustomType[CustomEnum](mockHandler),
				WithKeyStore(mockStore))
			if err != nil {
				t.Fatalf("Load failed: %v", err)
			}
			if config.Value != CustomEnum1 {
				t.Errorf("Expected Field1 to be set to %s, got %s", CustomEnum1, config.Value)
			}
			if config.Value2 != CustomEnum2 {
				t.Errorf("Expected Field1 to be set to %s, got %s", CustomEnum2, config.Value)
			}
			if config.Other != "foo" {
				t.Errorf("Expected Other to be set to foo, got %s", config.Other)
			}
		})

		t.Run("the validator is called", func(t *testing.T) {
			type Config struct {
				Value CustomEnum `key:"SOME_OTHER_KEY"`
			}

			config := Config{}
			err := Load(context.Background(), &config,
				WithCustomType[CustomEnum](mockHandler),
				WithKeyStore(mockStore))
			if err == nil {
				t.Fatal("Load should have failed")
			}
			if !errors.Is(err, expectedError) {
				t.Errorf("Expected validator error, got: %v", err)
			}
		})

		t.Run("string enum as pointer", func(t *testing.T) {
			type Config struct {
				Value *CustomEnum `key:"CUSTOM_ENUM_1"`
				Other string      `key:"SOME_OTHER_KEY"`
			}

			config := Config{}
			err := Load(context.Background(), &config,
				WithCustomType[CustomEnum](mockHandler),
				WithKeyStore(mockStore))
			if err != nil {
				t.Fatalf("Load failed: %v", err)
			}
			if config.Value == nil {
				t.Fatal("Expected config.Value to be non-nil")
			}
			if *config.Value != CustomEnum1 {
				t.Errorf("Expected Field1 to be set to %s, got %s", CustomEnum1, *config.Value)
			}
			if config.Other != "foo" {
				t.Errorf("Expected Other to be set to foo, got %s", config.Other)
			}
		})
	})

	t.Run("WithCustomType can add validation to an existing type", func(t *testing.T) {
		type Config struct {
			Port int `key:"PORT" min:"1000"`
		}

		mockStore := func(ctx context.Context, key string) (string, bool, error) {
			if key == "PORT" {
				return "1024", true, nil
			}
			return "", false, nil
		}

		// Modification: must be even
		t.Run("Adding validator to int", func(t *testing.T) {
			// Reuse the standard int handler logic and add a validator
			base := NewTypedIntHandler[int]()

			mod, err := PrependValidators(base, func(v int) error {
				if v%2 != 0 {
					return errors.New("must be even")
				}
				return nil
			})
			if err != nil {
				t.Fatalf("Failed to create modified handler: %v", err)
			}

			var cfg Config
			err = Load(context.Background(), &cfg,
				WithKeyStore(mockStore),
				WithCustomType[int](mod))

			if err != nil {
				t.Fatalf("Load failed: %v", err)
			}
			if cfg.Port != 1024 {
				t.Errorf("Expected 1024, got %d", cfg.Port)
			}

			// Test failure
			mockStoreOdd := func(ctx context.Context, key string) (string, bool, error) {
				return "1025", true, nil
			}
			err = Load(context.Background(), &cfg,
				WithKeyStore(mockStoreOdd),
				WithCustomType[int](mod))
			if err == nil {
				t.Fatal("Expected error")
			}
			if !strings.Contains(err.Error(), "must be even") {
				t.Errorf("Expected error to contain 'must be even', got %v", err)
			}
		})
		t.Run("NewEnumHandler provides validation for string enums", func(t *testing.T) {
			type MyEnum string
			const (
				ValA MyEnum = "A"
				ValB MyEnum = "B"
			)
			handler := NewEnumHandler(ValA, ValB)

			type Config struct {
				Value MyEnum `key:"ENUM"`
			}

			t.Run("Valid value", func(t *testing.T) {
				mockStore := func(ctx context.Context, key string) (string, bool, error) {
					return "A", true, nil
				}
				var cfg Config
				err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[MyEnum](handler))
				if err != nil {
					t.Fatalf("Load failed: %v", err)
				}
				if cfg.Value != ValA {
					t.Errorf("Expected A, got %s", cfg.Value)
				}
			})

			t.Run("Invalid value", func(t *testing.T) {
				mockStore := func(ctx context.Context, key string) (string, bool, error) {
					return "C", true, nil
				}
				var cfg Config
				err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[MyEnum](handler))
				if err == nil {
					t.Fatal("Expected error")
				}
				if !strings.Contains(err.Error(), "invalid value: C") {
					t.Errorf("Expected error message to contain 'invalid value: C', got %v", err)
				}
			})
		})

		t.Run("ReplaceParser can change the parsing logic while keeping validators", func(t *testing.T) {
			// Base handler for int with a range validator (via tag or manual)
			// We'll use NewTypedIntHandler which has range validation support
			base := NewTypedIntHandler[int]()

			// Replace parser to multiply input by 2
			mod, err := ReplaceParser(base, func(rawValue string) (int, error) {
				v, err := strconv.Atoi(rawValue)
				if err != nil {
					return 0, err
				}
				return v * 2, nil
			})
			if err != nil {
				t.Fatalf("ReplaceParser failed: %v", err)
			}

			type Config struct {
				Value int `key:"VAL" max:"10"`
			}

			t.Run("Success", func(t *testing.T) {
				mockStore := func(ctx context.Context, key string) (string, bool, error) {
					return "4", true, nil // 4 * 2 = 8, which is <= 10
				}
				var cfg Config
				err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[int](mod))
				if err != nil {
					t.Fatalf("Load failed: %v", err)
				}
				if cfg.Value != 8 {
					t.Errorf("Expected 8, got %d", cfg.Value)
				}
			})

			t.Run("Validator still works", func(t *testing.T) {
				mockStore := func(ctx context.Context, key string) (string, bool, error) {
					return "11", true, nil // 11 * 2 = 22, which is > 10 * 2 = 20
				}
				var cfg Config
				err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[int](mod))
				if err == nil {
					t.Fatal("Expected error")
				}
				if !strings.Contains(err.Error(), "above maximum") {
					t.Errorf("Expected range validation error, got %v", err)
				}
			})
		})

		t.Run("Triggering type conversion error at assignment", func(t *testing.T) {
			// This test demonstrates how it is possible to trigger the error at line 193 in config.go
			// by providing a custom TypedHandler that returns an incorrect type from its Build method.
			type Config struct {
				Value string `key:"VAL"`
			}
			mockStore := func(ctx context.Context, key string) (string, bool, error) {
				return "anything", true, nil
			}

			var cfg Config
			err := Load(context.Background(), &cfg,
				WithKeyStore(mockStore),
				WithCustomType[string](NonConvertibleHandler{}))

			if err == nil {
				t.Fatal("Expected error")
			}
			if !strings.Contains(err.Error(), "cannot be converted to string") {
				t.Errorf("Expected conversion error, got: %v", err)
			}
		})

		t.Run("Standard typed handlers", func(t *testing.T) {
			t.Run("String handler with pattern", func(t *testing.T) {
				handler := NewTypedStringHandler()
				type Config struct {
					Value string `key:"S" pattern:"^foo.*$"`
				}
				var cfg Config
				mockStore := func(ctx context.Context, key string) (string, bool, error) {
					return "bar", true, nil
				}
				err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[string](handler))
				if err == nil {
					t.Fatal("Expected error")
				}
			})

			t.Run("Uint handler", func(t *testing.T) {
				handler := NewTypedUintHandler[uint32]()
				type Config struct {
					Value uint32 `key:"U" max:"100"`
				}
				var cfg Config
				mockStore := func(ctx context.Context, key string) (string, bool, error) {
					return "101", true, nil
				}
				err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[uint32](handler))
				if err == nil {
					t.Fatal("Expected error")
				}
			})

			t.Run("Float handler", func(t *testing.T) {
				handler := NewTypedFloatHandler[float64]()
				type Config struct {
					Value float64 `key:"F" min:"0.5"`
				}
				var cfg Config
				mockStore := func(ctx context.Context, key string) (string, bool, error) {
					return "0.4", true, nil
				}
				err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[float64](handler))
				if err == nil {
					t.Fatal("Expected error")
				}
			})

			t.Run("Duration handler", func(t *testing.T) {
				handler := NewTypedDurationHandler()
				type Config struct {
					Value time.Duration `key:"D" min:"1s"`
				}
				var cfg Config
				mockStore := func(ctx context.Context, key string) (string, bool, error) {
					return "500ms", true, nil
				}
				err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[time.Duration](handler))
				if err == nil {
					t.Fatal("Expected error")
				}
			})
		})
	})
}
