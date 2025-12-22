package process

import (
	"reflect"
)

// FieldProcessor takes the user input string and outputs the final value to be set on the struct field.
// Any parsing or validation errors are returned as an error
// FieldProcessors always return values. The system will decide whether to convert to a pointer at the point
// of assignment.
type FieldProcessor[T any] func(rawValue string) (T, error)

// Validator validates a field value.
// Validators always deal with values, even if the target type is a pointer. The system makes the pointer
// at the last minute (before assignment)
type Validator[T any] func(value T) error

// Pipe combines a processor and a Validator, adding validation to the processor
func Pipe[T any](processor FieldProcessor[T], validator Validator[T]) FieldProcessor[T] {
	return func(rawValue string) (T, error) {
		value, err := processor(rawValue)
		if err != nil {
			return value, err
		}

		if err := validator(value); err != nil {
			return value, err
		}

		return value, nil
	}
}

// PipeMultiple combines a processor and a slice of Validators, adding validation to the processor
// This creates a single validator that runs all the other validators to reduce stack depth
func PipeMultiple[T any](processor FieldProcessor[T], validators []Validator[T]) FieldProcessor[T] {
	if len(validators) == 0 {
		return processor
	}
	// Create a single validator that runs all the other validators to reduce stack depth and closure debugging issues
	return Pipe(processor, func(value T) error {
		for _, validator := range validators {
			if err := validator(value); err != nil {
				return err
			}
		}
		return nil
	})
}

// Wrapper is a factory that wraps a FieldProcessor according to tags present on the target field
type Wrapper[T any] func(tags reflect.StructTag, inputProcess FieldProcessor[T]) (FieldProcessor[T], error)

func NewCompositeWrapper[T any](wrappers ...Wrapper[T]) Wrapper[T] {
	return func(tags reflect.StructTag, inputProcess FieldProcessor[T]) (FieldProcessor[T], error) {
		var wrapped FieldProcessor[T] = inputProcess
		for _, wrapper := range wrappers {
			var err error
			wrapped, err = wrapper(tags, wrapped)
			if err != nil {
				return nil, err
			}
		}
		return wrapped, nil
	}
}
