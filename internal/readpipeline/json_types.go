package readpipeline

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func NewJsonPipelineBuilder(targetType reflect.Type) TypedHandler[any] {
	return &typeHandlerImpl[any]{
		Parser: func(rawValue string) (any, error) {
			ptr := reflect.New(targetType).Interface()
			err := json.Unmarshal([]byte(rawValue), ptr)

			if err != nil {
				// We arrive here quite often if the system has not recognized the type.
				// Many types are structs under the covers.
				return nil, fmt.Errorf("error parsing json: %w", err)
			}

			// Dereference the value to maintain consistency with the maxim "Pipelines always readpipeline values"
			return reflect.ValueOf(ptr).Elem().Interface(), nil
		},

		ValidationWrapper: func(tags reflect.StructTag, inputProcess FieldProcessor[any]) (FieldProcessor[any], error) {
			return inputProcess, nil
		},
	}
}
