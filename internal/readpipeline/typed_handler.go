package readpipeline

import "reflect"

// typeHandlerImpl is the strongly typed handler for the given pipeline.
// It implements the typeless PipelineBuilder interface for the pipeline by boxing and unboxing the value as required.
type typeHandlerImpl[T any] struct {
	// Parser is the strongly typed version of the FieldProcessor that acts as input for this readpipeline
	Parser FieldProcessor[T]
	// ValidationWrapper is a factory that wraps the FieldProcessor with validation stages
	ValidationWrapper Wrapper[T]
}

func (h typeHandlerImpl[T]) GetParser() FieldProcessor[T] {
	return h.Parser
}

func (h typeHandlerImpl[T]) GetWrapper() Wrapper[T] {
	return h.ValidationWrapper
}

// typedHandlerAdapter adapts a TypedHandler[T] to a PipelineBuilder.
type typedHandlerAdapter[T any] struct {
	Handler TypedHandler[T]
}

func (a typedHandlerAdapter[T]) Build(tags reflect.StructTag) (FieldProcessor[any], error) {
	pipeline := a.Handler.GetParser()
	if pipeline == nil {
		return nil, nil // Return nil if no parser is provided (modification handler)
	}

	wrapper := a.Handler.GetWrapper()
	if wrapper != nil {
		var err error
		pipeline, err = wrapper(tags, pipeline)
		if err != nil {
			return nil, err
		}
	}
	return func(rawValue string) (any, error) {
		return pipeline(rawValue)
	}, nil
}

// WrapTypedHandler wraps a TypedHandler[T] as a PipelineBuilder.
func WrapTypedHandler[T any](handler TypedHandler[T]) PipelineBuilder {
	return typedHandlerAdapter[T]{Handler: handler}
}

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

// NewCompositeWrapper creates a Wrapper that applies a sequence of wrappers to a FieldProcessor
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
