package customtypes

import (
	"errors"
	"testing"
)

type Source string
type Target string

func TestNewTransformer(t *testing.T) {
	sourceHandler := NewParser[Source](func(rawValue string) (Source, error) {
		return Source(rawValue), nil
	})

	t.Run("ValidConversion", func(t *testing.T) {
		handler := NewTransformer[Source, Target](sourceHandler)
		pipeline, err := handler.BuildPipeline("")
		if err != nil {
			t.Fatalf("BuildPipeline failed: %v", err)
		}

		val, err := pipeline("hello")
		if err != nil {
			t.Fatalf("pipeline failed: %v", err)
		}
		if val != Target("hello") {
			t.Errorf("expected Target(hello), got %v", val)
		}
	})

	t.Run("IncompatibleTypes", func(t *testing.T) {
		// int and string are not convertible directly via reflect.Value.Convert if they are not underlying types
		// Actually, int to string conversion is allowed in Go but let's use something that is definitely not convertible
		type Unrelated struct{ X int }

		handler := NewTransformer[Source, Unrelated](sourceHandler)
		_, err := handler.BuildPipeline("")
		if err == nil {
			t.Error("expected error for incompatible types, got nil")
		}
	})

	t.Run("UpstreamError", func(t *testing.T) {
		errHandler := NewParser[Source](func(rawValue string) (Source, error) {
			return "", errors.New("upstream error")
		})
		handler := NewTransformer[Source, Target](errHandler)
		pipeline, err := handler.BuildPipeline("")
		if err != nil {
			t.Fatalf("BuildPipeline failed: %v", err)
		}

		_, err = pipeline("any")
		if err == nil {
			t.Error("expected upstream error, got nil")
		} else if err.Error() != "upstream error" {
			t.Errorf("expected 'upstream error', got %v", err)
		}
	})
}
