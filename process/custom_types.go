package process

import "reflect"

// NewCustomHandler creates a new handler that uses the custom parser and validators.
// The custom parser cannot be nil.
func NewCustomHandler[T any](customParser FieldProcessor[T], customValidators ...Validator[T]) Handler {
	return &TypeHandler[T]{
		Parser: customParser,
		ValidationWrapper: func(tags reflect.StructTag, inputProcess FieldProcessor[T]) (FieldProcessor[T], error) {
			if customValidators != nil && len(customValidators) > 0 {
				inputProcess = PipeMultiple(inputProcess, customValidators)
			}
			return inputProcess, nil
		},
	}
}
