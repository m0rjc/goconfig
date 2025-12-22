package readpipeline

import (
	"encoding/json"
	"reflect"
)

func NewJsonHandler(targetType reflect.Type) PipelineBuilder {
	return typeHandlerImpl[any]{
		Parser: func(rawValue string) (any, error) {
			ptr := reflect.New(targetType).Interface()
			err := json.Unmarshal([]byte(rawValue), ptr)

			if err != nil {
				return nil, err
			}

			// Dereference the value to maintain consistency with the maxim "Pipelines always readpipeline values"
			return reflect.ValueOf(ptr).Elem().Interface(), nil
		},

		ValidationWrapper: func(tags reflect.StructTag, inputProcess FieldProcessor[any]) (FieldProcessor[any], error) {
			return inputProcess, nil
		},
	}
}
