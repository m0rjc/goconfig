package readpipeline

import (
	"reflect"
)

// HandlerFactory is a function that returns a PipelineBuilder for a given type.
type HandlerFactory func(t reflect.Type) PipelineBuilder

// TypedHandlerFactory is a function that returns a TypedHandler for a given type.
type TypedHandlerFactory[T any] func(t reflect.Type) TypedHandler[T]

type TypeRegistry interface {
	RegisterType(t reflect.Type, handler PipelineBuilder)
	HandlerFor(t reflect.Type) PipelineBuilder
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

// WrapTypedHandler wraps a TypedHandler[T] as a PipelineBuilder for use in the typeless registry.
func WrapTypedHandler[T any](handler TypedHandler[T]) PipelineBuilder {
	return typedHandlerAdapter[T]{Handler: handler}
}

// WrapKindHandler wraps a TypedHandlerFactory[T] as a HandlerFactory for use in the typeless registry.
func WrapKindHandler[T any](handler TypedHandlerFactory[T]) HandlerFactory {
	return func(t reflect.Type) PipelineBuilder {
		return WrapTypedHandler(handler(t))
	}
}
