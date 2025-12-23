package customtypes

import (
	"errors"
	"reflect"
	"testing"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

func TestAddWrapper(t *testing.T) {
	parser := NewParser[int](func(rawValue string) (int, error) {
		if rawValue == "1" {
			return 1, nil
		}
		return 0, errors.New("parse error")
	})

	t.Run("Success", func(t *testing.T) {
		wrapper := func(tags reflect.StructTag, input readpipeline.FieldProcessor[int]) (readpipeline.FieldProcessor[int], error) {
			return func(rawValue string) (int, error) {
				val, err := input(rawValue)
				if err != nil {
					return val, err
				}
				return val * 10, nil
			}, nil
		}

		handler := AddWrapper[int](parser, wrapper)
		pipeline, err := handler.BuildPipeline("")
		if err != nil {
			t.Fatalf("BuildPipeline failed: %v", err)
		}

		val, err := pipeline("1")
		if err != nil {
			t.Fatalf("pipeline failed: %v", err)
		}
		if val != 10 {
			t.Errorf("expected 10, got %v", val)
		}
	})

	t.Run("WrapperError", func(t *testing.T) {
		wrapper := func(tags reflect.StructTag, input readpipeline.FieldProcessor[int]) (readpipeline.FieldProcessor[int], error) {
			return nil, errors.New("wrapper build error")
		}

		handler := AddWrapper[int](parser, wrapper)
		_, err := handler.BuildPipeline("")
		if err == nil {
			t.Error("expected error from wrapper build, got nil")
		}
	})

	t.Run("NilWrapper", func(t *testing.T) {
		handler := AddWrapper[int](parser, nil)
		pipeline, err := handler.BuildPipeline("")
		if err != nil {
			t.Fatalf("BuildPipeline failed: %v", err)
		}

		val, err := pipeline("1")
		if err != nil {
			t.Fatalf("pipeline failed: %v", err)
		}
		if val != 1 {
			t.Errorf("expected 1, got %v", val)
		}
	})
}
