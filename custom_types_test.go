package goconfig

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/m0rjc/goconfig/process"
)

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
		mockHandler := process.NewCustomHandler[CustomStruct](mockParser)

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
		mockHandler := process.NewCustomHandler(func(value string) (CustomEnum, error) {
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
			base := process.NewTypedIntHandler(reflect.TypeOf(int(0)).Bits())
			mod := process.NewCustomHandler[int](func(s string) (int, error) {
				v, err := base.GetParser()(s)
				return int(v), err
			}, func(v int) error {
				if v%2 != 0 {
					return errors.New("must be even")
				}
				return nil
			})

			var cfg Config
			err := Load(context.Background(), &cfg,
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
			if err == nil || !reflect.TypeOf(err).AssignableTo(reflect.TypeOf(&ConfigErrors{})) {
				t.Fatalf("Expected ConfigErrors, got %v", err)
			}
			if !errors.Is(err, errors.New("must be even")) {
				// ConfigErrors.Error() contains the string
				if !reflect.ValueOf(err).MethodByName("HasErrors").Call(nil)[0].Bool() {
					t.Fatal("Expected errors")
				}
			}
		})
	})
}
