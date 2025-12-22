package process

import (
	"fmt"
	"reflect"
)

// New creates a FieldProcessor for the given type. It reads struct tags to instantiate required
// validators. It allows override using custom parsing and validation if supplied. Default validation is
// applied before custom validators.
// If the target type is a pointer, it will be unboxed before processing. The output of the process chain is the value.
// The caller is responsible for assigning the value to the struct field, dealing with pointers as needed.
func New(fieldType reflect.Type, tags reflect.StructTag, customParser FieldProcessor[any], customValidators []Validator[any]) (FieldProcessor[any], error) {
	var err error
	targetType := fieldType
	isPointer := fieldType.Kind() == reflect.Ptr

	if isPointer {
		// Pointer writing is handled by the setFieldValue side of the process
		// in config.go
		targetType = targetType.Elem()
	}

	handler := handlerFor(targetType)
	if customParser != nil {
		handler = NewCustomHandler(customParser, handler)
	}
	if handler == nil {
		return nil, fmt.Errorf("no handler for type %s", fieldType)
	}

	pipeline := handler.GetParser()

	if customValidators != nil {
		pipeline = PipeMultiple(pipeline, customValidators)
	}

	pipeline, err = handler.AddValidatorsToPipeline(tags, pipeline)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}
