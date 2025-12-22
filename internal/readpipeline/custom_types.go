package readpipeline

import (
	"fmt"
	"reflect"
)

// NewCustomHandler creates a new handler that uses the custom parser and validators.
func NewCustomHandler[T any](customParser FieldProcessor[T], customValidators ...Validator[T]) TypedHandler[T] {
	return typeHandlerImpl[T]{
		Parser:            customParser,
		ValidationWrapper: newCustomValidatorWrapper(customValidators),
	}
}

func NewEnumHandler[T ~string](validValues ...T) TypedHandler[T] {
	return NewCustomHandler[T](func(rawValue string) (T, error) {
		for _, validValue := range validValues {
			if rawValue == string(validValue) {
				return validValue, nil
			}
		}
		return "", fmt.Errorf("invalid value: %s", rawValue)
	})
}

func newCustomValidatorWrapper[T any](customValidators []Validator[T]) Wrapper[T] {
	return func(tags reflect.StructTag, inputProcess FieldProcessor[T]) (FieldProcessor[T], error) {
		if customValidators != nil && len(customValidators) > 0 {
			inputProcess = PipeMultiple(inputProcess, customValidators)
		}
		return inputProcess, nil
	}
}

func ReplaceParser[B, T any](baseHandler TypedHandler[B], customParser FieldProcessor[T]) (TypedHandler[T], error) {
	adaptedWrapper := castWrapper[B, T](baseHandler.GetWrapper())

	return typeHandlerImpl[T]{
		Parser:            customParser,
		ValidationWrapper: adaptedWrapper,
	}, nil
}

func PrependValidators[B, T any](baseHandler TypedHandler[B], customValidators ...Validator[T]) (TypedHandler[T], error) {
	parser, err := castPipeline[B, T](baseHandler.GetParser())
	if err != nil {
		return nil, err
	}

	adaptedWrapper := castWrapper[B, T](baseHandler.GetWrapper())

	return typeHandlerImpl[T]{
		Parser:            parser,
		ValidationWrapper: NewCompositeWrapper[T](adaptedWrapper, newCustomValidatorWrapper(customValidators)),
	}, nil
}

func CastHandler[B, T any](handler TypedHandler[B]) (TypedHandler[T], error) {
	parser, err := castPipeline[B, T](handler.GetParser())
	if err != nil {
		return nil, err
	}

	wrapper := castWrapper[B, T](handler.GetWrapper())

	return typeHandlerImpl[T]{
		Parser:            parser,
		ValidationWrapper: wrapper,
	}, nil
}

func castPipeline[B, T any](parser FieldProcessor[B]) (FieldProcessor[T], error) {
	if parser == nil {
		return nil, nil
	}

	baseType := reflect.TypeOf((*B)(nil)).Elem()
	newType := reflect.TypeOf((*T)(nil)).Elem()
	if !baseType.ConvertibleTo(newType) {
		return nil, fmt.Errorf("incompatible type conversion: %s -> %s", baseType, newType)
	}

	return func(rawValue string) (T, error) {
		val, err := parser(rawValue)
		if err != nil {
			var zero T
			return zero, err
		}
		// Convert B to T (e.g., string to Foo)
		return reflect.ValueOf(val).Convert(newType).Interface().(T), nil
	}, nil
}

func castWrapper[B, T any](wrapper Wrapper[B]) Wrapper[T] {
	if wrapper == nil {
		return nil
	}

	return func(tags reflect.StructTag, inputProcess FieldProcessor[T]) (FieldProcessor[T], error) {
		// 1. Down-convert the inputProcess (T -> B) so the base wrapper can use it
		inputPipeline, err := castPipeline[T, B](inputProcess)
		if err != nil {
			return nil, fmt.Errorf("input conversion for validators: %w", err)
		}

		// 2. Run the base wrapper logic
		wrappedBase, err := wrapper(tags, inputPipeline)
		if err != nil {
			return nil, err
		}

		// 3. Up-convert the result (B -> T) for the final pipeline
		outputPipeline, err := castPipeline[B, T](wrappedBase)
		if err != nil {
			return nil, fmt.Errorf("output conversion for validators: %w", err)
		}

		return outputPipeline, nil
	}
}
