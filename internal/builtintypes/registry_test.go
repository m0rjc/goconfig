package builtintypes

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

// registryMockPipelineBuilder is a simple implementation of PipelineBuilder for testing
type registryMockPipelineBuilder struct {
	buildFunc func(tags reflect.StructTag) (readpipeline.FieldProcessor[any], error)
}

func (m *registryMockPipelineBuilder) Build(tags reflect.StructTag) (readpipeline.FieldProcessor[any], error) {
	if m.buildFunc != nil {
		return m.buildFunc(tags)
	}
	return nil, nil
}

func TestLocalTypeRegistry(t *testing.T) {
	parent := &LocalTypeRegistry{
		SpecialTypeHandlers: map[reflect.Type]readpipeline.PipelineBuilder{},
	}
	registry := &LocalTypeRegistry{
		Parent:              parent,
		SpecialTypeHandlers: map[reflect.Type]readpipeline.PipelineBuilder{},
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
		root := &RootTypeRegistry{
			specialTypeHandlers: map[reflect.Type]readpipeline.PipelineBuilder{},
			kindHandlers:        map[reflect.Kind]readpipeline.HandlerFactory{},
		}
		regWithRoot := &LocalTypeRegistry{
			Parent:              root,
			SpecialTypeHandlers: map[reflect.Type]readpipeline.PipelineBuilder{},
		}
		got := regWithRoot.HandlerFor(boolType)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})
}

func TestRootTypeRegistry(t *testing.T) {
	registry := &RootTypeRegistry{
		specialTypeHandlers: map[reflect.Type]readpipeline.PipelineBuilder{},
		kindHandlers:        map[reflect.Kind]readpipeline.HandlerFactory{},
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
		registry.kindHandlers[reflect.String] = func(t reflect.Type) readpipeline.PipelineBuilder {
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

func TestNewTypeRegistry(t *testing.T) {
	registry := NewTypeRegistry()
	if registry == nil {
		t.Fatal("NewTypeRegistry returned nil")
	}

	local, ok := registry.(*LocalTypeRegistry)
	if !ok {
		t.Fatalf("expected *LocalTypeRegistry, got %T", registry)
	}

	if local.Parent != rootRegistry {
		t.Error("expected parent to be rootRegistry")
	}
}

func TestRegisterTypeGlobal(t *testing.T) {
	type CustomType struct {
		Value string
	}

	// mockTypedHandler is a simple implementation of TypedHandler for testing
	handler := &typeHandlerImpl[CustomType]{
		Parser: func(rawValue string) (CustomType, error) {
			return CustomType{Value: rawValue}, nil
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
		{"URL", (*url.URL)(nil)},
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
