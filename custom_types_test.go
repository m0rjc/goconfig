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
		customStructType := reflect.TypeOf(CustomStruct{})
		mockStore := func(ctx context.Context, key string) (string, bool, error) {
			if key == "CUSTOM_STRUCT" {
				return "--Marker--", true, nil
			}
			return "", false, nil
		}
		mockParser := func(value string) (any, error) {
			return CustomStruct{Field1: value}, nil
		}
		mockHandler := process.NewCustomHandler(mockParser)

		t.Run("struct as value", func(t *testing.T) {
			type Config struct {
				Value CustomStruct `key:"CUSTOM_STRUCT"`
			}

			config := Config{Value: CustomStruct{Field1: ""}}
			err := Load(context.Background(), &config,
				WithCustomType(customStructType, mockHandler),
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
				WithCustomType(customStructType, mockHandler),
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
		customEnumType := reflect.TypeOf(CustomEnum(""))
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
				WithCustomType(customEnumType, mockHandler),
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
				WithCustomType(customEnumType, mockHandler),
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
				WithCustomType(customEnumType, mockHandler),
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
}
