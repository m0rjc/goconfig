package process

import (
	"reflect"
	"time"
)

// TypeHandler is the strongly typed handler for the given pipeline.
// It implements the typeless Handler interface for the pipeline by boxing and unboxing the value as required.
type TypeHandler[T any] struct {
	Parser            FieldProcessor[T]
	ValidationWrapper Wrapper[T]
}

// Handler is the typeless interface used to build the read pipeline.
type Handler interface {
	// GetParser returns a FieldProcessor[any] that is used to read the raw value and start the read pipeline
	GetParser() FieldProcessor[any]
	// AddValidatorsToPipeline adds validators to the pipeline based on tags found in the StructTag for the target field.
	AddValidatorsToPipeline(tags reflect.StructTag, p FieldProcessor[any]) (FieldProcessor[any], error)
}

func (h TypeHandler[T]) GetParser() FieldProcessor[any] {
	// This wrapper function converts from the strongly typed world of the TypeHandler to the weak type world of the process pipeline.
	return func(rawValue string) (any, error) {
		return h.Parser(rawValue)
	}
}

func (h TypeHandler[T]) AddValidatorsToPipeline(tags reflect.StructTag, p FieldProcessor[any]) (FieldProcessor[any], error) {
	// Convert FieldProcessor[any] back to FieldProcessor[T] safely
	typedP := func(s string) (T, error) {
		val, err := p(s)
		if err != nil {
			var zero T
			return zero, err
		}
		return val.(T), nil
	}

	wrapped, err := h.ValidationWrapper(tags, typedP)
	if err != nil {
		return nil, err
	}

	// Erase type again for the pipeline
	return func(s string) (any, error) {
		return wrapped(s)
	}, nil
}

// Specific type overrides (Duration, etc.)
var specialTypeParsers = map[reflect.Type]Handler{
	reflect.TypeOf(time.Duration(0)): durationTypeHandler,
}

// Category-based parsers
var kindParsers = map[reflect.Kind]func(t reflect.Type) Handler{
	reflect.Int:     NewIntHandler,
	reflect.Uint:    NewUintHandler,
	reflect.Struct:  NewJsonHandler,
	reflect.Map:     NewJsonHandler,
	reflect.String:  NewStringHandler,
	reflect.Bool:    NewBoolHandler,
	reflect.Float32: NewFloatHandler,
	reflect.Float64: NewFloatHandler,
}

// handlerFor returns the Handler for the given type, or nil if none is registered.
func handlerFor(t reflect.Type) Handler {
	// 1. Check for specific type overrides (The "Duration" check)
	if p, ok := specialTypeParsers[t]; ok {
		return p
	}

	// 2. Fall back to category-based logic
	if factory, ok := kindParsers[t.Kind()]; ok {
		return factory(t)
	}

	return nil
}
