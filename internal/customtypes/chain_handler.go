package customtypes

import (
	"reflect"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

func AddWrapper[T any](prior readpipeline.TypedHandler[T], wrapper readpipeline.Wrapper[T]) readpipeline.TypedHandler[T] {
	return &chainHandler[T]{Prior: prior, Wrapper: wrapper}
}

// chainHandler creates a chained pipeline by appending wrappers
type chainHandler[T any] struct {
	Prior   readpipeline.TypedHandler[T]
	Wrapper readpipeline.Wrapper[T]
}

func (c *chainHandler[T]) BuildPipeline(tags reflect.StructTag) (readpipeline.FieldProcessor[T], error) {
	pipeline, err := c.Prior.BuildPipeline(tags)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		return nil, nil
	}

	if c.Wrapper != nil {
		return c.Wrapper(tags, pipeline)
	}
	return pipeline, nil
}
