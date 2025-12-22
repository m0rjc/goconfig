package process

import (
	"reflect"
	"strconv"
)

func NewBoolHandler(fieldType reflect.Type) Handler {
	return TypeHandler[bool]{
		Parser: func(rawValue string) (bool, error) {
			return strconv.ParseBool(rawValue)
		},
		ValidationWrapper: func(tags reflect.StructTag, inputProcess FieldProcessor[bool]) (FieldProcessor[bool], error) {
			return inputProcess, nil
		},
	}
}
