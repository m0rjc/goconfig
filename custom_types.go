package goconfig

import (
	"reflect"
	"time"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

type FieldProcessor[T any] = readpipeline.FieldProcessor[T]

type Validator[T any] = readpipeline.Validator[T]

type Wrapper[T any] = readpipeline.Wrapper[T]

type TypedHandler[T any] = readpipeline.TypedHandler[T]

func NewCustomHandler[T any](customParser FieldProcessor[T], customValidators ...Validator[T]) TypedHandler[T] {
	return readpipeline.NewCustomHandler(customParser, customValidators...)
}

func NewEnumHandler[T ~string](validValues ...T) TypedHandler[T] {
	return readpipeline.NewEnumHandler(validValues...)
}

func ReplaceParser[B, T any](baseHandler TypedHandler[B], customParser FieldProcessor[T]) (TypedHandler[T], error) {
	return readpipeline.ReplaceParser(baseHandler, customParser)
}

func PrependValidators[B, T any](baseHandler TypedHandler[B], customValidators ...Validator[T]) (TypedHandler[T], error) {
	return readpipeline.PrependValidators(baseHandler, customValidators...)
}

func NewTypedStringHandler() TypedHandler[string] {
	return readpipeline.NewTypedStringHandler()
}

func NewTypedIntHandler[T ~int | ~int8 | ~int16 | ~int32 | ~int64]() TypedHandler[T] {
	t := reflect.TypeOf(T(0))
	cast, err := readpipeline.CastHandler[int64, T](readpipeline.NewTypedIntHandler(t.Bits()))
	if err != nil {
		// Should never happen
		panic(err)
	}
	return cast
}

func NewTypedUintHandler[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64]() TypedHandler[T] {
	t := reflect.TypeOf(T(0))
	cast, err := readpipeline.CastHandler[uint64, T](readpipeline.NewTypedUintHandler(t.Bits()))
	if err != nil {
		// Should never happen
		panic(err)
	}
	return cast
}

func NewTypedFloatHandler[T ~float32 | ~float64]() TypedHandler[T] {
	t := reflect.TypeOf(T(0))
	cast, err := readpipeline.CastHandler[float64, T](readpipeline.NewTypedFloatHandler(t.Bits()))
	if err != nil {
		// Should never happen
		panic(err)
	}
	return cast
}

func NewTypedDurationHandler() TypedHandler[time.Duration] {
	return readpipeline.NewTypedDurationHandler()
}
