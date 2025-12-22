package goconfig

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestCompositeStore(t *testing.T) {
	ctx := context.Background()

	store1 := func(ctx context.Context, key string) (string, bool, error) {
		if key == "KEY1" {
			return "VAL1", true, nil
		}
		return "", false, nil
	}

	store2 := func(ctx context.Context, key string) (string, bool, error) {
		if key == "KEY2" {
			return "VAL2", true, nil
		}
		if key == "KEY1" {
			return "VAL2-OVERRIDE", true, nil
		}
		return "", false, nil
	}

	errStore := func(ctx context.Context, key string) (string, bool, error) {
		return "", false, errors.New("store error")
	}

	t.Run("first store hits", func(t *testing.T) {
		composite := CompositeStore(store1, store2)
		val, present, err := composite(ctx, "KEY1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !present {
			t.Fatal("expected value to be present")
		}
		if val != "VAL1" {
			t.Errorf("expected VAL1, got %s", val)
		}
	})

	t.Run("second store hits", func(t *testing.T) {
		composite := CompositeStore(store1, store2)
		val, present, err := composite(ctx, "KEY2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !present {
			t.Fatal("expected value to be present")
		}
		if val != "VAL2" {
			t.Errorf("expected VAL2, got %s", val)
		}
	})

	t.Run("no hits", func(t *testing.T) {
		composite := CompositeStore(store1, store2)
		_, present, err := composite(ctx, "MISSING")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if present {
			t.Fatal("expected value to be not present")
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		composite := CompositeStore(errStore, store1)
		_, _, err := composite(ctx, "KEY1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "store error" {
			t.Errorf("expected 'store error', got '%v'", err)
		}
	})

	t.Run("stops at first error", func(t *testing.T) {
		// Even if store1 would have the value, if errStore is first it should return error
		composite := CompositeStore(errStore, store1)
		_, _, err := composite(ctx, "KEY1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("empty composite", func(t *testing.T) {
		composite := CompositeStore()
		_, present, err := composite(ctx, "ANY")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if present {
			t.Fatal("expected value to be not present")
		}
	})
}

func TestEnvironmentKeyStore(t *testing.T) {
	ctx := context.Background()
	key := "GOCONFIG_TEST_KEY"
	val := "test_value"

	// Ensure it's clean
	os.Unsetenv(key)

	t.Run("not present", func(t *testing.T) {
		_, present, err := EnvironmentKeyStore(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if present {
			t.Error("expected not present")
		}
	})

	t.Run("present", func(t *testing.T) {
		os.Setenv(key, val)
		defer os.Unsetenv(key)

		gotVal, present, err := EnvironmentKeyStore(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !present {
			t.Error("expected present")
		}
		if gotVal != val {
			t.Errorf("expected %s, got %s", val, gotVal)
		}
	})

	t.Run("present but empty", func(t *testing.T) {
		os.Setenv(key, "")
		defer os.Unsetenv(key)

		gotVal, present, err := EnvironmentKeyStore(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !present {
			t.Error("expected present even if empty")
		}
		if gotVal != "" {
			t.Errorf("expected empty string, got %s", gotVal)
		}
	})
}
