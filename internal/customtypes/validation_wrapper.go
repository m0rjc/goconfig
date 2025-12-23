package customtypes

import (
	"reflect"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

func NewValidatorWrapper[T any](customValidators ...readpipeline.Validator[T]) readpipeline.Wrapper[T] {
	return func(tags reflect.StructTag, inputProcess readpipeline.FieldProcessor[T]) (readpipeline.FieldProcessor[T], error) {
		if customValidators != nil && len(customValidators) > 0 {
			inputProcess = readpipeline.PipeMultiple(inputProcess, customValidators)
		}
		return inputProcess, nil
	}
}
