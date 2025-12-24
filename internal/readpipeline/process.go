package readpipeline

import (
	"fmt"
	"reflect"
)

// New creates a FieldProcessor for the given type. It reads struct tags to instantiate required
// validators.
// If the target type is a pointer, it will be unboxed before processing. The output of the readpipeline chain is the value.
// The caller is responsible for assigning the value to the struct field, dealing with pointers as needed.
func New(fieldType reflect.Type, tags reflect.StructTag, registry TypeRegistry) (FieldProcessor[any], error) {
	targetType := fieldType
	handler := registry.HandlerFor(targetType)

	// If the target type is a pointer, try to find a handler for the underlying type
	// Assignment back to the pointer will be handled by the caller's write pipeline
	if handler == nil && fieldType.Kind() == reflect.Ptr {
		targetType = fieldType.Elem()
		handler = registry.HandlerFor(targetType)
	}

	if handler == nil {
		return nil, fmt.Errorf("no handler for type %s", targetType)
	}

	pipeline, err := handler.Build(tags)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		return nil, fmt.Errorf("no parser for type %s", targetType)
	}
	return pipeline, nil
}
