package goconfig

import (
	"reflect"
	"time"

	"github.com/m0rjc/goconfig/internal/customtypes"
	"github.com/m0rjc/goconfig/internal/readpipeline"
)

type FieldProcessor[T any] = readpipeline.FieldProcessor[T]

type Validator[T any] = readpipeline.Validator[T]

type Wrapper[T any] = readpipeline.Wrapper[T]

type TypedHandler[T any] = readpipeline.TypedHandler[T]

type Transform[T, U any] = customtypes.Transform[T, U]

func RegisterCustomType[T any](handler TypedHandler[T]) {
	readpipeline.RegisterType[T](handler)
}

func NewCustomType[T any](customParser FieldProcessor[T], customValidators ...Validator[T]) TypedHandler[T] {
	handler := customtypes.NewParser(customParser)
	if customValidators != nil && len(customValidators) > 0 {
		handler = customtypes.AddWrapper(handler, customtypes.NewValidatorWrapper(customValidators...))
	}
	return handler
}

func NewStringEnumType[T ~string](validValues ...T) TypedHandler[T] {
	return customtypes.NewStringEnum(validValues...)
}

func AddValidators[T any](baseHandler TypedHandler[T], customValidators ...Validator[T]) TypedHandler[T] {
	if customValidators != nil && len(customValidators) > 0 {
		return customtypes.AddWrapper(baseHandler, customtypes.NewValidatorWrapper(customValidators...))
	}
	return baseHandler
}

// AddDynamicValidation allows a TypedHandler to add validation (or other logic) to the process pipeline dependent
// on struct tags present on the target field.
// See the AddValidatorToPipeline function and the example/custom_tags example for more details.
func AddDynamicValidation[T any](baseHandler TypedHandler[T], wrapper Wrapper[T]) TypedHandler[T] {
	return customtypes.AddWrapper(baseHandler, wrapper)
}

// AddValidatorToPipeline adds a validator to a pipeline. This is used as part of pipeline building in the TypedHandler.
func AddValidatorToPipeline[T any](pipeline FieldProcessor[T], validator Validator[T]) FieldProcessor[T] {
	return func(rawValue string) (T, error) {
		value, err := pipeline(rawValue)
		if err != nil {
			return value, err
		}
		if err = validator(value); err != nil {
			return value, err
		}
		return value, nil
	}
}

func CastCustomType[T, U any](baseHandler TypedHandler[T]) TypedHandler[U] {
	return customtypes.NewCastingTransformer[T, U](baseHandler)
}

// TransformCustomType creates a TypedHandler that applies a Transform function to process data from a base handler.
func TransformCustomType[T, U any](baseHandler TypedHandler[T], transform Transform[T, U]) TypedHandler[U] {
	return customtypes.NewTransformer(baseHandler, transform)
}

func DefaultStringType[T ~string]() TypedHandler[T] {
	pipeline := readpipeline.NewTypedStringHandler()
	return CastCustomType[string, T](pipeline)
}

func DefaultIntegerType[T ~int | ~int8 | ~int16 | ~int32 | ~int64]() TypedHandler[T] {
	t := reflect.TypeOf(T(0))
	return CastCustomType[int64, T](readpipeline.NewTypedIntHandler(t.Bits()))
}

func DefaultUnsignedIntegerType[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64]() TypedHandler[T] {
	t := reflect.TypeOf(T(0))
	return CastCustomType[uint64, T](readpipeline.NewTypedUintHandler(t.Bits()))
}

func DefaultFloatIntegerType[T ~float32 | ~float64]() TypedHandler[T] {
	t := reflect.TypeOf(T(0))
	return CastCustomType[float64, T](readpipeline.NewTypedFloatHandler(t.Bits()))
}

func DefaultDurationType() TypedHandler[time.Duration] {
	return readpipeline.NewTypedDurationHandler()
}
