package readpipeline

import (
	"reflect"
	"testing"
)

func TestPatternValidator(t *testing.T) {
	// Base processor that just returns the input string
	baseProcessor := func(rawValue string) (string, error) {
		return rawValue, nil
	}

	t.Run("invalid pattern spec in setup", func(t *testing.T) {
		tags := reflect.StructTag(`pattern:"["`) // Invalid regex
		_, err := WrapProcessUsingPatternTag(tags, baseProcessor)
		if err == nil {
			t.Error("expected error for invalid pattern spec, got nil")
		}
	})

	t.Run("valid pattern and matching input", func(t *testing.T) {
		tags := reflect.StructTag(`pattern:"^[a-z]+$"`)
		processor, err := WrapProcessUsingPatternTag(tags, baseProcessor)
		if err != nil {
			t.Fatalf("unexpected error in setup: %v", err)
		}

		_, err = processor("hello")
		if err != nil {
			t.Errorf("expected match for 'hello', got error: %v", err)
		}
	})

	t.Run("valid pattern and non-matching input", func(t *testing.T) {
		tags := reflect.StructTag(`pattern:"^[a-z]+$"`)
		processor, err := WrapProcessUsingPatternTag(tags, baseProcessor)
		if err != nil {
			t.Fatalf("unexpected error in setup: %v", err)
		}

		_, err = processor("Hello123")
		if err == nil {
			t.Error("expected error for non-matching input 'Hello123', got nil")
		}
	})

	t.Run("no pattern tag", func(t *testing.T) {
		tags := reflect.StructTag(``)
		processor, err := WrapProcessUsingPatternTag(tags, baseProcessor)
		if err != nil {
			t.Fatalf("unexpected error in setup: %v", err)
		}

		val, err := processor("any input")
		if err != nil {
			t.Errorf("unexpected error without pattern tag: %v", err)
		}
		if val != "any input" {
			t.Errorf("expected 'any input', got %v", val)
		}
	})
}
