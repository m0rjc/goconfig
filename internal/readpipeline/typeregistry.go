package readpipeline

import (
	"reflect"
	"time"
)

// HandlerFactory is a function that returns a PipelineBuilder for a given type.
type HandlerFactory func(t reflect.Type) PipelineBuilder

// TypedHandlerFactory is a function that returns a TypedHandler for a given type.
type TypedHandlerFactory[T any] func(t reflect.Type) TypedHandler[any]

type TypeRegistry interface {
	RegisterType(t reflect.Type, handler PipelineBuilder)
	HandlerFor(t reflect.Type) PipelineBuilder
}

// NewTypeRegistry creates a new TypeRegistry with the default handlers.
// Types registered here will override the default handlers for this registry instance only
func NewTypeRegistry() TypeRegistry {
	return &localTypeRegistry{
		parent:              rootRegistry,
		specialTypeHandlers: map[reflect.Type]PipelineBuilder{},
	}
}

// RegisterType registers a custom PipelineBuilder for a given type in the root registry.
func RegisterType[T any](handler TypedHandler[T]) {
	handlerType := reflect.TypeOf((*T)(nil)).Elem()
	wrapper := WrapTypedHandler(handler)
	rootRegistry.RegisterType(handlerType, wrapper)
}

type localTypeRegistry struct {
	parent              TypeRegistry
	specialTypeHandlers map[reflect.Type]PipelineBuilder
}

func (r *localTypeRegistry) RegisterType(t reflect.Type, handler PipelineBuilder) {
	r.specialTypeHandlers[t] = handler
}

func (r *localTypeRegistry) HandlerFor(t reflect.Type) PipelineBuilder {
	if p, ok := r.specialTypeHandlers[t]; ok {
		return p
	}
	return r.parent.HandlerFor(t)
}

// BaseTypeRegistry is a registry of Handlers factories for specific types.
// Handlers can be registered for specific types or for a category of types keyed on Kind.
// If a handler is registered for a specific type, it will be used instead of the category handler.
// If a handler is registered for a category, a factory method is called to instantiate the handler given the type.
type rootTypeRegistry struct {
	specialTypeHandlers map[reflect.Type]PipelineBuilder
	kindHandlers        map[reflect.Kind]HandlerFactory
}

// RegisterType registers a custom PipelineBuilder for a given type.
func (r *rootTypeRegistry) RegisterType(t reflect.Type, handler PipelineBuilder) {
	r.specialTypeHandlers[t] = handler
}

// HandlerFor returns the PipelineBuilder for the given type, or nil if none is registered.
func (r *rootTypeRegistry) HandlerFor(t reflect.Type) PipelineBuilder {
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

// typedHandlerAdapter adapts a TypedHandler[T] to a PipelineBuilder.
type typedHandlerAdapter[T any] struct {
	Handler TypedHandler[T]
}

func (a typedHandlerAdapter[T]) Build(tags reflect.StructTag) (FieldProcessor[any], error) {
	pipeline, err := a.Handler.BuildPipeline(tags)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		return nil, nil // Return nil if no parser is provided (modification handler)
	}
	return func(rawValue string) (any, error) {
		return pipeline(rawValue)
	}, nil
}

// WrapTypedHandler wraps a TypedHandler[T] as a PipelineBuilder.
func WrapTypedHandler[T any](handler TypedHandler[T]) PipelineBuilder {
	return typedHandlerAdapter[T]{Handler: handler}
}

var rootRegistry = &rootTypeRegistry{
	specialTypeHandlers: map[reflect.Type]PipelineBuilder{
		reflect.TypeOf(time.Duration(0)): WrapTypedHandler(durationTypeHandler),
	},
	kindHandlers: map[reflect.Kind]HandlerFactory{
		reflect.Int:     func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewIntHandler(t)) },
		reflect.Int8:    func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewIntHandler(t)) },
		reflect.Int16:   func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewIntHandler(t)) },
		reflect.Int32:   func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewIntHandler(t)) },
		reflect.Int64:   func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewIntHandler(t)) },
		reflect.Uint:    func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewUintHandler(t)) },
		reflect.Uint8:   func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewUintHandler(t)) },
		reflect.Uint16:  func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewUintHandler(t)) },
		reflect.Uint32:  func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewUintHandler(t)) },
		reflect.Uint64:  func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewUintHandler(t)) },
		reflect.Struct:  func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewJsonPipelineBuilder(t)) },
		reflect.Map:     func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewJsonPipelineBuilder(t)) },
		reflect.String:  func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewStringHandler(t)) },
		reflect.Bool:    func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewBoolHandler(t)) },
		reflect.Float32: func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewFloatHandler(t)) },
		reflect.Float64: func(t reflect.Type) PipelineBuilder { return WrapTypedHandler(NewFloatHandler(t)) },
	},
}
