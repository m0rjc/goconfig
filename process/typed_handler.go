package process

import "reflect"

// typeHandlerImpl is the strongly typed handler for the given pipeline.
// It implements the typeless PipelineBuilder interface for the pipeline by boxing and unboxing the value as required.
type typeHandlerImpl[T any] struct {
	// Parser is the strongly typed version of the FieldProcessor that acts as input for this process
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

func (h typeHandlerImpl[T]) Build(tags reflect.StructTag) (FieldProcessor[any], error) {
	pipeline := h.GetParser()
	if pipeline == nil {
		return nil, nil // Return nil if no parser is provided (modification handler)
	}

	wrapper := h.GetWrapper()
	if wrapper != nil {
		var err error
		pipeline, err = wrapper(tags, pipeline)
		if err != nil {
			return nil, err
		}
	}
	return typedToUntypedPipeline(pipeline), nil
}

// typedToUntypedPipeline converts from the strongly typed world of the typeHandlerImpl to the weak type world of the process pipeline.
func typedToUntypedPipeline[T any](parser FieldProcessor[T]) FieldProcessor[any] {
	if parser == nil {
		return nil
	}
	return func(rawValue string) (any, error) {
		return parser(rawValue)
	}
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
