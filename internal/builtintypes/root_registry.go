package builtintypes

import (
	"net/url"
	"reflect"
	"time"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

// RootTypeRegistry is a registry of Handlers factories for specific types.
// Handlers can be registered for specific types or for a category of types keyed on Kind.
// If a handler is registered for a specific type, it will be used instead of the category handler.
// If a handler is registered for a category, a factory method is called to instantiate the handler given the type.
type RootTypeRegistry struct {
	specialTypeHandlers map[reflect.Type]readpipeline.PipelineBuilder
	kindHandlers        map[reflect.Kind]readpipeline.HandlerFactory
}

// RegisterType registers a custom readpipeline.PipelineBuilder for a given type.
func (r *RootTypeRegistry) RegisterType(t reflect.Type, handler readpipeline.PipelineBuilder) {
	r.specialTypeHandlers[t] = handler
}

// RegisterKind registers a factory method for a given Kind.
func (r *RootTypeRegistry) RegisterKind(kind reflect.Kind, factory readpipeline.HandlerFactory) {
	r.kindHandlers[kind] = factory
}

// HandlerFor returns the readpipeline.PipelineBuilder for the given type, or nil if none is registered.
func (r *RootTypeRegistry) HandlerFor(t reflect.Type) readpipeline.PipelineBuilder {
	// 1. Check for specific type overrides (The "Duration" check)
	if p, ok := r.specialTypeHandlers[t]; ok {
		return p
	}

	// 2. Fall back to category-based logic
	if factory, ok := r.kindHandlers[t.Kind()]; ok {
		return factory(t)
	}

	return nil
}

// NewTypeRegistry creates a new TypeRegistry with the default handlers.
// Types registered here will override the default handlers for this registry instance only
func NewTypeRegistry() readpipeline.TypeRegistry {
	return &LocalTypeRegistry{
		Parent:              rootRegistry,
		SpecialTypeHandlers: map[reflect.Type]readpipeline.PipelineBuilder{},
	}
}

// RegisterType registers a custom PipelineBuilder for a given type in the root registry.
func RegisterType[T any](handler readpipeline.TypedHandler[T]) {
	handlerType := reflect.TypeOf((*T)(nil)).Elem()
	wrapper := readpipeline.WrapTypedHandler(handler)
	rootRegistry.RegisterType(handlerType, wrapper)
}

type LocalTypeRegistry struct {
	Parent              readpipeline.TypeRegistry
	SpecialTypeHandlers map[reflect.Type]readpipeline.PipelineBuilder
}

func (r *LocalTypeRegistry) RegisterType(t reflect.Type, handler readpipeline.PipelineBuilder) {
	r.SpecialTypeHandlers[t] = handler
}

func (r *LocalTypeRegistry) HandlerFor(t reflect.Type) readpipeline.PipelineBuilder {
	if p, ok := r.SpecialTypeHandlers[t]; ok {
		return p
	}
	return r.Parent.HandlerFor(t)
}

var rootRegistry = &RootTypeRegistry{
	specialTypeHandlers: map[reflect.Type]readpipeline.PipelineBuilder{
		reflect.TypeOf(time.Duration(0)): readpipeline.WrapTypedHandler(durationTypeHandler),
		reflect.TypeOf((*url.URL)(nil)):  readpipeline.WrapTypedHandler(NewUrlTypedHandler()),
	},
	kindHandlers: map[reflect.Kind]readpipeline.HandlerFactory{
		reflect.Int:     readpipeline.WrapKindHandler(NewIntHandler),
		reflect.Int8:    readpipeline.WrapKindHandler(NewIntHandler),
		reflect.Int16:   readpipeline.WrapKindHandler(NewIntHandler),
		reflect.Int32:   readpipeline.WrapKindHandler(NewIntHandler),
		reflect.Int64:   readpipeline.WrapKindHandler(NewIntHandler),
		reflect.Uint:    readpipeline.WrapKindHandler(NewUintHandler),
		reflect.Uint8:   readpipeline.WrapKindHandler(NewUintHandler),
		reflect.Uint16:  readpipeline.WrapKindHandler(NewUintHandler),
		reflect.Uint32:  readpipeline.WrapKindHandler(NewUintHandler),
		reflect.Uint64:  readpipeline.WrapKindHandler(NewUintHandler),
		reflect.Struct:  readpipeline.WrapKindHandler(NewJsonPipelineBuilder),
		reflect.Map:     readpipeline.WrapKindHandler(NewJsonPipelineBuilder),
		reflect.String:  readpipeline.WrapKindHandler(NewStringHandler),
		reflect.Bool:    readpipeline.WrapKindHandler(NewBoolHandler),
		reflect.Float32: readpipeline.WrapKindHandler(NewFloatHandler),
		reflect.Float64: readpipeline.WrapKindHandler(NewFloatHandler),
	},
}
