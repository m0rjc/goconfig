package readpipeline

import (
	"errors"
	"reflect"
	"testing"
	"time"
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

// registryMockPipelineBuilder is a simple implementation of PipelineBuilder for testing
type registryMockPipelineBuilder struct {
	buildFunc func(tags reflect.StructTag) (FieldProcessor[any], error)
}

func (m *registryMockPipelineBuilder) Build(tags reflect.StructTag) (FieldProcessor[any], error) {
	if m.buildFunc != nil {
		return m.buildFunc(tags)
	}
	return nil, nil
}

func TestLocalTypeRegistry(t *testing.T) {
	parent := &localTypeRegistry{
		specialTypeHandlers: map[reflect.Type]PipelineBuilder{},
	}
	registry := &localTypeRegistry{
		parent:              parent,
		specialTypeHandlers: map[reflect.Type]PipelineBuilder{},
	}

	intType := reflect.TypeOf(0)
	stringType := reflect.TypeOf("")

	handler1 := &registryMockPipelineBuilder{}
	handler2 := &registryMockPipelineBuilder{}

	t.Run("RegisterAndRetrieve", func(t *testing.T) {
		registry.RegisterType(intType, handler1)
		got := registry.HandlerFor(intType)
		if got != handler1 {
			t.Errorf("expected handler1, got %v", got)
		}
	})

	t.Run("FallbackToParent", func(t *testing.T) {
		parent.RegisterType(stringType, handler2)
		got := registry.HandlerFor(stringType)
		if got != handler2 {
			t.Errorf("expected handler2 from parent, got %v", got)
		}
	})

	t.Run("OverrideParent", func(t *testing.T) {
		registry.RegisterType(stringType, handler1)
		got := registry.HandlerFor(stringType)
		if got != handler1 {
			t.Errorf("expected handler1 (override), got %v", got)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		boolType := reflect.TypeOf(true)
		// We need a parent that returns nil if not found
		root := &rootTypeRegistry{
			specialTypeHandlers: map[reflect.Type]PipelineBuilder{},
			kindHandlers:        map[reflect.Kind]HandlerFactory{},
		}
		regWithRoot := &localTypeRegistry{
			parent:              root,
			specialTypeHandlers: map[reflect.Type]PipelineBuilder{},
		}
		got := regWithRoot.HandlerFor(boolType)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})
}

func TestRootTypeRegistry(t *testing.T) {
	registry := &rootTypeRegistry{
		specialTypeHandlers: map[reflect.Type]PipelineBuilder{},
		kindHandlers:        map[reflect.Kind]HandlerFactory{},
	}

	intType := reflect.TypeOf(0)
	handler1 := &registryMockPipelineBuilder{}

	t.Run("SpecialTypeHandler", func(t *testing.T) {
		registry.RegisterType(intType, handler1)
		got := registry.HandlerFor(intType)
		if got != handler1 {
			t.Errorf("expected handler1, got %v", got)
		}
	})

	t.Run("KindHandler", func(t *testing.T) {
		stringType := reflect.TypeOf("")
		handler2 := &registryMockPipelineBuilder{}
		registry.kindHandlers[reflect.String] = func(t reflect.Type) PipelineBuilder {
			return handler2
		}

		got := registry.HandlerFor(stringType)
		if got != handler2 {
			t.Errorf("expected handler2 from kind factory, got %v", got)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		boolType := reflect.TypeOf(true)
		got := registry.HandlerFor(boolType)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})
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

func TestNewTypeRegistry(t *testing.T) {
	registry := NewTypeRegistry()
	if registry == nil {
		t.Fatal("NewTypeRegistry returned nil")
	}

	local, ok := registry.(*localTypeRegistry)
	if !ok {
		t.Fatalf("expected *localTypeRegistry, got %T", registry)
	}

	if local.parent != rootRegistry {
		t.Error("expected parent to be rootRegistry")
	}
}

func TestRegisterTypeGlobal(t *testing.T) {
	type CustomType struct {
		Value string
	}

	handler := &mockTypedHandler[CustomType]{
		buildPipelineFunc: func(tags reflect.StructTag) (FieldProcessor[CustomType], error) {
			return func(rawValue string) (CustomType, error) {
				return CustomType{Value: rawValue}, nil
			}, nil
		},
	}

	RegisterType[CustomType](handler)

	// Verify it's in rootRegistry
	customType := reflect.TypeOf(CustomType{})
	pb := rootRegistry.HandlerFor(customType)
	if pb == nil {
		t.Fatal("expected handler to be registered in rootRegistry")
	}

	pipeline, err := pb.Build("")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	val, err := pipeline("hello")
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}

	expected := CustomType{Value: "hello"}
	if val != expected {
		t.Errorf("expected %v, got %v", expected, val)
	}
}

func TestDefaultHandlers(t *testing.T) {
	tests := []struct {
		name string
		val  any
	}{
		{"Int", 0},
		{"Int8", int8(0)},
		{"Int16", int16(0)},
		{"Int32", int32(0)},
		{"Int64", int64(0)},
		{"Uint", uint(0)},
		{"Uint8", uint8(0)},
		{"Uint16", uint16(0)},
		{"Uint32", uint32(0)},
		{"Uint64", uint64(0)},
		{"String", ""},
		{"Bool", true},
		{"Float32", float32(0)},
		{"Float64", float64(0)},
		{"Duration", time.Duration(0)},
		{"Struct", struct{ X int }{}},
		{"Map", map[string]int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.val)
			handler := rootRegistry.HandlerFor(typ)
			if handler == nil {
				t.Errorf("no default handler for %v", typ)
			}
		})
	}
}
