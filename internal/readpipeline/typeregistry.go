package readpipeline

import (
	"reflect"
	"time"
)

// HandlerFactory is a function that returns a PipelineBuilder for a given type.
type HandlerFactory func(t reflect.Type) PipelineBuilder

// TypeRegistry is a registry of Handlers factories for specific types.
// Handlers can be registered for specific types or for a category of types keyed on Kind.
// If a handler is registered for a specific type, it will be used instead of the category handler.
// If a handler is registered for a category, a factory method is called to instantiate the handler given the type.
type TypeRegistry struct {
	specialTypeHandlers map[reflect.Type]PipelineBuilder
	kindHandlers        map[reflect.Kind]HandlerFactory
}

// NewDefaultTypeRegistry creates a new TypeRegistry with the default handlers.
func NewDefaultTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		specialTypeHandlers: map[reflect.Type]PipelineBuilder{
			reflect.TypeOf(time.Duration(0)): durationTypeHandler,
		},
		kindHandlers: map[reflect.Kind]HandlerFactory{
			reflect.Int:     NewIntHandler,
			reflect.Int8:    NewIntHandler,
			reflect.Int16:   NewIntHandler,
			reflect.Int32:   NewIntHandler,
			reflect.Int64:   NewIntHandler,
			reflect.Uint:    NewUintHandler,
			reflect.Uint8:   NewUintHandler,
			reflect.Uint16:  NewUintHandler,
			reflect.Uint32:  NewUintHandler,
			reflect.Uint64:  NewUintHandler,
			reflect.Struct:  NewJsonHandler,
			reflect.Map:     NewJsonHandler,
			reflect.String:  NewStringHandler,
			reflect.Bool:    NewBoolHandler,
			reflect.Float32: NewFloatHandler,
			reflect.Float64: NewFloatHandler,
		},
	}
}

// RegisterKind registers a factory function for a given kind.
func (r *TypeRegistry) RegisterKind(kind reflect.Kind, factory func(t reflect.Type) PipelineBuilder) {
	r.kindHandlers[kind] = factory
}

// RegisterType registers a custom PipelineBuilder for a given type.
func (r *TypeRegistry) RegisterType(t reflect.Type, handler PipelineBuilder) {
	r.specialTypeHandlers[t] = handler
}

// HandlerFor returns the PipelineBuilder for the given type, or nil if none is registered.
func (r *TypeRegistry) HandlerFor(t reflect.Type) PipelineBuilder {
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
