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

func AddDynamicValidation[T any](baseHandler TypedHandler[T], wrapper Wrapper[T]) TypedHandler[T] {
	return customtypes.AddWrapper(baseHandler, wrapper)
}

func CastCustomType[T, U any](baseHandler TypedHandler[T]) TypedHandler[U] {
	return customtypes.NewTransformer[T, U](baseHandler)
}

func DefaultStringType() TypedHandler[string] {
	return readpipeline.NewTypedStringHandler()
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
