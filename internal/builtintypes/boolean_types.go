package builtintypes

import (
	"reflect"
	"strconv"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

func NewBoolHandler(_ reflect.Type) readpipeline.TypedHandler[bool] {
	return NewTypedBoolHandler()
}

// NewTypedBoolHandler returns a TypedHandler[bool] that uses standard bool parsing and validation.
func NewTypedBoolHandler() readpipeline.TypedHandler[bool] {
	return &typeHandlerImpl[bool]{
		Parser: func(rawValue string) (bool, error) {
			return strconv.ParseBool(rawValue)
		},
		ValidationWrapper: func(tags reflect.StructTag, inputProcess readpipeline.FieldProcessor[bool]) (readpipeline.FieldProcessor[bool], error) {
			return inputProcess, nil
		},
	}
}
