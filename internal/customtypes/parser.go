package customtypes

import (
	"reflect"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

type customType[T any] struct {
	Parser readpipeline.FieldProcessor[T]
}

func NewParser[T any](parser readpipeline.FieldProcessor[T]) readpipeline.TypedHandler[T] {
	return &customType[T]{Parser: parser}
}

func (c *customType[T]) BuildPipeline(tags reflect.StructTag) (readpipeline.FieldProcessor[T], error) {
	return c.Parser, nil
}
