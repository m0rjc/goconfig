package readpipeline

import (
	"errors"
	"reflect"
	"testing"
)

// mockPipelineBuilder is a mock implementation of PipelineBuilder for testing.
type mockPipelineBuilder struct {
	buildFunc func(tags reflect.StructTag) (FieldProcessor[any], error)
}

func (m *mockPipelineBuilder) Build(tags reflect.StructTag) (FieldProcessor[any], error) {
	return m.buildFunc(tags)
}

func TestNew(t *testing.T) {
	t.Run("BareType", func(t *testing.T) {
		registry := &rootTypeRegistry{
			specialTypeHandlers: make(map[reflect.Type]PipelineBuilder),
			kindHandlers:        make(map[reflect.Kind]HandlerFactory),
		}

		expectedValue := "success"
		mockBuilder := &mockPipelineBuilder{
			buildFunc: func(tags reflect.StructTag) (FieldProcessor[any], error) {
				return func(rawValue string) (any, error) {
					return expectedValue, nil
				}, nil
			},
		}

		stringType := reflect.TypeOf("")
		registry.RegisterType(stringType, mockBuilder)

		processor, err := New(stringType, "", registry)
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}

		val, err := processor("input")
		if err != nil {
			t.Fatalf("Processor failed: %v", err)
		}

		if val != expectedValue {
			t.Errorf("Expected %v, got %v", expectedValue, val)
		}
	})

	t.Run("PointerType", func(t *testing.T) {
		registry := &rootTypeRegistry{
			specialTypeHandlers: make(map[reflect.Type]PipelineBuilder),
			kindHandlers:        make(map[reflect.Kind]HandlerFactory),
		}

		expectedValue := 42
		mockBuilder := &mockPipelineBuilder{
			buildFunc: func(tags reflect.StructTag) (FieldProcessor[any], error) {
				return func(rawValue string) (any, error) {
					return expectedValue, nil
				}, nil
			},
		}

		intType := reflect.TypeOf(0)
		registry.RegisterType(intType, mockBuilder)

		// Pass *int to New
		ptrIntType := reflect.TypeOf((*int)(nil))
		processor, err := New(ptrIntType, "", registry)
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}

		val, err := processor("input")
		if err != nil {
			t.Fatalf("Processor failed: %v", err)
		}

		if val != expectedValue {
			t.Errorf("Expected %v, got %v", expectedValue, val)
		}
	})

	t.Run("NoHandlerError", func(t *testing.T) {
		registry := &rootTypeRegistry{
			specialTypeHandlers: make(map[reflect.Type]PipelineBuilder),
			kindHandlers:        make(map[reflect.Kind]HandlerFactory),
		}

		_, err := New(reflect.TypeOf(0), "", registry)
		if err == nil {
			t.Fatal("Expected error for missing handler, got nil")
		}

		expectedErr := "no handler for type int"
		if err.Error() != expectedErr {
			t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
		}
	})

	t.Run("BuildError", func(t *testing.T) {
		registry := &rootTypeRegistry{
			specialTypeHandlers: make(map[reflect.Type]PipelineBuilder),
			kindHandlers:        make(map[reflect.Kind]HandlerFactory),
		}

		expectedBuildErr := errors.New("build failed")
		mockBuilder := &mockPipelineBuilder{
			buildFunc: func(tags reflect.StructTag) (FieldProcessor[any], error) {
				return nil, expectedBuildErr
			},
		}

		intType := reflect.TypeOf(0)
		registry.RegisterType(intType, mockBuilder)

		_, err := New(intType, "", registry)
		if err == nil {
			t.Fatal("Expected error from Build, got nil")
		}

		if !errors.Is(err, expectedBuildErr) {
			t.Errorf("Expected error %v, got %v", expectedBuildErr, err)
		}
	})

	t.Run("NilPipelineError", func(t *testing.T) {
		registry := &rootTypeRegistry{
			specialTypeHandlers: make(map[reflect.Type]PipelineBuilder),
			kindHandlers:        make(map[reflect.Kind]HandlerFactory),
		}

		mockBuilder := &mockPipelineBuilder{
			buildFunc: func(tags reflect.StructTag) (FieldProcessor[any], error) {
				return nil, nil
			},
		}

		intType := reflect.TypeOf(0)
		registry.RegisterType(intType, mockBuilder)

		_, err := New(intType, "", registry)
		if err == nil {
			t.Fatal("Expected error for nil pipeline, got nil")
		}

		expectedErr := "no parser for type int"
		if err.Error() != expectedErr {
			t.Errorf("Expected error %q, got %q", expectedErr, err.Error())
		}
	})

	t.Run("ProcessorError", func(t *testing.T) {
		registry := &rootTypeRegistry{
			specialTypeHandlers: make(map[reflect.Type]PipelineBuilder),
			kindHandlers:        make(map[reflect.Kind]HandlerFactory),
		}

		expectedProcErr := errors.New("processing failed")
		mockBuilder := &mockPipelineBuilder{
			buildFunc: func(tags reflect.StructTag) (FieldProcessor[any], error) {
				return func(rawValue string) (any, error) {
					return nil, expectedProcErr
				}, nil
			},
		}

		intType := reflect.TypeOf(0)
		registry.RegisterType(intType, mockBuilder)

		processor, err := New(intType, "", registry)
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}

		_, err = processor("input")
		if err == nil {
			t.Fatal("Expected error from processor, got nil")
		}

		if !errors.Is(err, expectedProcErr) {
			t.Errorf("Expected error %v, got %v", expectedProcErr, err)
		}
	})
}
