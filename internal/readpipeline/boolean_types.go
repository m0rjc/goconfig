package readpipeline

import (
	"reflect"
	"strconv"
)

func NewBoolHandler(_ reflect.Type) PipelineBuilder {
	return WrapTypedHandler(NewTypedBoolHandler())
}

// NewTypedBoolHandler returns a TypedHandler[bool] that uses standard bool parsing and validation.
func NewTypedBoolHandler() TypedHandler[bool] {
	return typeHandlerImpl[bool]{
		Parser: func(rawValue string) (bool, error) {
			return strconv.ParseBool(rawValue)
		},
		ValidationWrapper: func(tags reflect.StructTag, inputProcess FieldProcessor[bool]) (FieldProcessor[bool], error) {
			return inputProcess, nil
		},
	}
}
