package process

import "reflect"

// FieldProcessor takes the user input string and outputs the final value to be set on the struct field.
// Any parsing or validation errors are returned as an error
// FieldProcessors always return values. The system will decide whether to convert to a pointer at the point
// of assignment.
type FieldProcessor[T any] func(rawValue string) (T, error)

// Validator validates a field value.
// Validators always deal with values, even if the target type is a pointer. The system makes the pointer
// at the last minute (before assignment)
type Validator[T any] func(value T) error

// TypedHandler is the strongly typed version of the PipelineBuilder interface.
type TypedHandler[T any] interface {
	// GetParser returns a FieldProcessor[T] that is used to read the raw value and start the read pipeline.
	// It can return nil if this handler doesn't provide a parser (e.g. it's a modification).
	GetParser() FieldProcessor[T]
	// GetWrapper returns a Wrapper[T] that adds validators to the pipeline based on tags.
	// It can return nil if no validation is needed.
	GetWrapper() Wrapper[T]
	// Build creates the final FieldProcessor[any] for the given tags.
	// This causes any TypedHandler to implement the untyped PipelineBuilder interface.
	Build(tags reflect.StructTag) (FieldProcessor[any], error)
}

// PipelineBuilder is the typeless interface used to build the read pipeline.
type PipelineBuilder interface {
	// Build creates the final FieldProcessor[any] for the given tags.
	Build(tags reflect.StructTag) (FieldProcessor[any], error)
}

// Wrapper is a factory that wraps a FieldProcessor according to tags present on the target field
type Wrapper[T any] func(tags reflect.StructTag, inputProcess FieldProcessor[T]) (FieldProcessor[T], error)
