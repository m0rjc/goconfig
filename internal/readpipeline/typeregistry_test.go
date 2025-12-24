package readpipeline

import (
	"errors"
	"reflect"
	"testing"
)

// mockTypedHandler is a simple implementation of TypedHandler for testing
type mockTypedHandler[T any] struct {
	buildPipelineFunc func(tags reflect.StructTag) (FieldProcessor[T], error)
}

func (m *mockTypedHandler[T]) BuildPipeline(tags reflect.StructTag) (FieldProcessor[T], error) {
	if m.buildPipelineFunc != nil {
		return m.buildPipelineFunc(tags)
	}
	return nil, nil
}

func TestTypedHandlerAdapter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		inner := &mockTypedHandler[int]{
			buildPipelineFunc: func(tags reflect.StructTag) (FieldProcessor[int], error) {
				return func(rawValue string) (int, error) {
					return 42, nil
				}, nil
			},
		}

		adapter := WrapTypedHandler[int](inner)
		pipeline, err := adapter.Build("")
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}

		val, err := pipeline("anything")
		if err != nil {
			t.Fatalf("pipeline failed: %v", err)
		}
		if val != 42 {
			t.Errorf("expected 42, got %v", val)
		}
	})

	t.Run("NilPipeline", func(t *testing.T) {
		inner := &mockTypedHandler[int]{
			buildPipelineFunc: func(tags reflect.StructTag) (FieldProcessor[int], error) {
				return nil, nil
			},
		}

		adapter := WrapTypedHandler[int](inner)
		pipeline, err := adapter.Build("")
		if err != nil {
			t.Fatalf("Build failed: %v", err)
		}
		if pipeline != nil {
			t.Error("expected nil pipeline")
		}
	})

	t.Run("BuildError", func(t *testing.T) {
		inner := &mockTypedHandler[int]{
			buildPipelineFunc: func(tags reflect.StructTag) (FieldProcessor[int], error) {
				return nil, errors.New("build error")
			},
		}

		adapter := WrapTypedHandler[int](inner)
		_, err := adapter.Build("")
		if err == nil {
			t.Error("expected build error, got nil")
		}
	})
}
