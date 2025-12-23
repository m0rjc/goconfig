package customtypes

import (
	"fmt"
	"reflect"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

type transformer[T, U any] struct {
	Prior readpipeline.TypedHandler[T]
	Cast  func(T) (U, error)
}

func (t *transformer[T, U]) BuildPipeline(tags reflect.StructTag) (readpipeline.FieldProcessor[U], error) {
	pipeline, err := t.Prior.BuildPipeline(tags)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		return nil, nil
	}

	return func(rawValue string) (U, error) {
		val, upstreamError := pipeline(rawValue)
		if upstreamError != nil {
			var zero U
			return zero, upstreamError
		}
		return t.Cast(val)
	}, nil
}

// badTransformer holds on to an error that will be returned by BuildPipeline.
type badTransformer[T any] struct {
	Err error
}

func (b *badTransformer[T]) BuildPipeline(tags reflect.StructTag) (readpipeline.FieldProcessor[T], error) {
	return nil, b.Err
}

func NewTransformer[T, U any](handler readpipeline.TypedHandler[T]) readpipeline.TypedHandler[U] {
	sourceType := reflect.TypeOf((*T)(nil)).Elem()
	newType := reflect.TypeOf((*U)(nil)).Elem()
	if !sourceType.ConvertibleTo(newType) {
		return &badTransformer[U]{fmt.Errorf("incompatible type conversion: %s -> %s", sourceType, newType)}
	}

	cast := func(value T) (U, error) {
		reflected := reflect.ValueOf(value)
		if !reflected.IsValid() {
			var zero U
			return zero, fmt.Errorf("invalid value in type conversion")
		}
		if !reflected.Type().ConvertibleTo(newType) {
			var zero U
			return zero, fmt.Errorf("cannot convert from %s to %s", reflected.Type(), newType)
		}
		return reflect.ValueOf(value).Convert(newType).Interface().(U), nil
	}

	return &transformer[T, U]{Prior: handler, Cast: cast}
}
