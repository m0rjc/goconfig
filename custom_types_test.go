package goconfig

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestNewCustomType(t *testing.T) {
	type CustomString string
	type Config struct {
		Val CustomString `key:"VAL"`
	}

	handler := NewCustomType(
		func(rawValue string) (CustomString, error) {
			return CustomString("prefix-" + rawValue), nil
		},
		func(value CustomString) error {
			if len(value) < 10 {
				return errors.New("too short")
			}
			return nil
		},
	)

	mockStore := func(ctx context.Context, key string) (string, bool, error) {
		if key == "VAL" {
			return "12345", true, nil
		}
		return "", false, nil
	}

	var cfg Config
	err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[CustomString](handler))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Val != "prefix-12345" {
		t.Errorf("Expected prefix-12345, got %s", cfg.Val)
	}

	// Test validation failure
	mockStoreFail := func(ctx context.Context, key string) (string, bool, error) {
		if key == "VAL" {
			return "1", true, nil
		}
		return "", false, nil
	}
	err = Load(context.Background(), &cfg, WithKeyStore(mockStoreFail), WithCustomType[CustomString](handler))
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}
}

func TestNewStringEnumType(t *testing.T) {
	type Mode string
	const (
		ModeDev  Mode = "dev"
		ModeProd Mode = "prod"
	)

	type Config struct {
		AppMode Mode `key:"MODE"`
	}

	handler := NewStringEnumType(ModeDev, ModeProd)

	tests := []struct {
		name      string
		val       string
		expectErr bool
	}{
		{"valid dev", "dev", false},
		{"valid prod", "prod", false},
		{"invalid", "other", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := func(ctx context.Context, key string) (string, bool, error) {
				return tt.val, true, nil
			}
			var cfg Config
			err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[Mode](handler))
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Load failed: %v", err)
				}
				if string(cfg.AppMode) != tt.val {
					t.Errorf("Expected %s, got %s", tt.val, cfg.AppMode)
				}
			}
		})
	}
}

func TestAddValidators(t *testing.T) {
	type Config struct {
		Val int `key:"VAL"`
	}

	baseHandler := DefaultIntegerType[int]()
	handler := AddValidators(baseHandler, func(value int) error {
		if value%2 != 0 {
			return errors.New("must be even")
		}
		return nil
	})

	mockStore := func(ctx context.Context, key string) (string, bool, error) {
		return "42", true, nil
	}

	var cfg Config
	err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[int](handler))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Val != 42 {
		t.Errorf("Expected 42, got %d", cfg.Val)
	}

	mockStoreFail := func(ctx context.Context, key string) (string, bool, error) {
		return "43", true, nil
	}
	err = Load(context.Background(), &cfg, WithKeyStore(mockStoreFail), WithCustomType[int](handler))
	if err == nil {
		t.Fatal("Expected error for odd number, got nil")
	}
}

func TestAddDynamicValidation(t *testing.T) {
	type Config struct {
		Val string `key:"VAL" check:"true"`
	}

	baseHandler := DefaultStringType()
	handler := AddDynamicValidation(baseHandler, func(tags reflect.StructTag, inputProcess FieldProcessor[string]) (FieldProcessor[string], error) {
		if tags.Get("check") == "true" {
			return func(rawValue string) (string, error) {
				val, err := inputProcess(rawValue)
				if err != nil {
					return val, err
				}
				if val == "forbidden" {
					return val, errors.New("forbidden value")
				}
				return val, nil
			}, nil
		}
		return inputProcess, nil
	})

	mockStore := func(ctx context.Context, key string) (string, bool, error) {
		return "forbidden", true, nil
	}

	var cfg Config
	err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[string](handler))
	if err == nil {
		t.Fatal("Expected dynamic validation error, got nil")
	}
}

func TestCastCustomType(t *testing.T) {
	type MyInt int
	type Config struct {
		Val MyInt `key:"VAL"`
	}

	baseHandler := DefaultIntegerType[int]()
	// Cast int handler to MyInt handler
	handler := CastCustomType[int, MyInt](baseHandler)

	mockStore := func(ctx context.Context, key string) (string, bool, error) {
		return "100", true, nil
	}

	var cfg Config
	err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[MyInt](handler))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Val != 100 {
		t.Errorf("Expected 100, got %d", cfg.Val)
	}
}

func TestDefaultTypeHandlers(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		handler := DefaultStringType()
		p, _ := handler.BuildPipeline("")
		val, _ := p("hello")
		if any(val).(string) != "hello" {
			t.Errorf("Expected hello, got %s", val)
		}
	})

	t.Run("Integer", func(t *testing.T) {
		handler := DefaultIntegerType[int32]()
		p, _ := handler.BuildPipeline("")
		val, _ := p("123")
		if any(val).(int32) != 123 {
			t.Errorf("Expected 123, got %v", val)
		}
	})

	t.Run("UnsignedInteger", func(t *testing.T) {
		handler := DefaultUnsignedIntegerType[uint16]()
		p, _ := handler.BuildPipeline("")
		val, _ := p("456")
		if any(val).(uint16) != 456 {
			t.Errorf("Expected 456, got %v", val)
		}
	})

	t.Run("Float", func(t *testing.T) {
		handler := DefaultFloatIntegerType[float64]()
		p, _ := handler.BuildPipeline("")
		val, _ := p("1.23")
		if any(val).(float64) != 1.23 {
			t.Errorf("Expected 1.23, got %v", val)
		}
	})

	t.Run("Duration", func(t *testing.T) {
		handler := DefaultDurationType()
		p, _ := handler.BuildPipeline("")
		val, _ := p("10s")
		if any(val).(time.Duration) != 10*time.Second {
			t.Errorf("Expected 10s, got %v", val)
		}
	})
}

func TestRegisterCustomType(t *testing.T) {
	type GlobalCustom string
	type Config struct {
		Val GlobalCustom `key:"VAL"`
	}

	RegisterCustomType[GlobalCustom](NewCustomType(func(rawValue string) (GlobalCustom, error) {
		return GlobalCustom("global-" + rawValue), nil
	}))

	mockStore := func(ctx context.Context, key string) (string, bool, error) {
		return "test", true, nil
	}

	var cfg Config
	err := Load(context.Background(), &cfg, WithKeyStore(mockStore))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Val != "global-test" {
		t.Errorf("Expected global-test, got %s", cfg.Val)
	}
}

func TestDefaultIntegerType_Tags(t *testing.T) {
	type Config struct {
		Val int `key:"VAL" min:"10" max:"20"`
	}

	handler := DefaultIntegerType[int]()

	mockStore := func(ctx context.Context, key string) (string, bool, error) {
		return "15", true, nil
	}

	var cfg Config
	err := Load(context.Background(), &cfg, WithKeyStore(mockStore), WithCustomType[int](handler))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Test min failure
	mockStoreMin := func(ctx context.Context, key string) (string, bool, error) {
		return "5", true, nil
	}
	err = Load(context.Background(), &cfg, WithKeyStore(mockStoreMin), WithCustomType[int](handler))
	if err == nil {
		t.Fatal("Expected min validation error, got nil")
	}

	// Test max failure
	mockStoreMax := func(ctx context.Context, key string) (string, bool, error) {
		return "25", true, nil
	}
	err = Load(context.Background(), &cfg, WithKeyStore(mockStoreMax), WithCustomType[int](handler))
	if err == nil {
		t.Fatal("Expected max validation error, got nil")
	}
}
