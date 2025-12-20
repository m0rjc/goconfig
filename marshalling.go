package goconfigtools

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type Parser func(value string) (any, error)

func parseValue(rawValue string, fieldPath string, field reflect.Value, opts *loadOptions) (any, error) {
	var typedValue any
	var err error

	// 1. DEREFERENCE LOGIC
	// We want to know what the target 'final' type is
	targetType := field.Type()
	targetKind := field.Kind()

	// If it's a pointer, we look at the element it points to
	if targetKind == reflect.Pointer {
		targetType = targetType.Elem()
		targetKind = targetType.Kind()
	}

	// 2. STRATEGY SELECTION
	if customParser, exists := opts.parsers[fieldPath]; exists {
		typedValue, err = customParser(rawValue)
	} else if targetType == reflect.TypeOf(time.Duration(0)) {
		// SPECIAL CASE: Duration (even if it's a pointer to a duration)
		typedValue, err = time.ParseDuration(rawValue)
	} else if targetKind == reflect.Struct || targetKind == reflect.Map {
		// Handle as JSON
		// We create a new instance of targetType (e.g. TypedJsonStruct)
		// reflect.New returns a pointer to that type
		ptr := reflect.New(targetType).Interface()
		err = json.Unmarshal([]byte(rawValue), ptr)
		if err == nil {
			// If the original field was a pointer, we keep 'ptr' as is.
			// If it was a value, we need to Elem() it.
			if field.Kind() == reflect.Pointer {
				typedValue = ptr
			} else {
				typedValue = reflect.ValueOf(ptr).Elem().Interface()
			}
		}
	} else {
		// Standard primitives
		typedValue, err = parseDefault(targetKind, rawValue)
	}

	if err != nil {
		return nil, err
	}

	return typedValue, nil
}

func parseDefault(kind reflect.Kind, rawValue string) (any, error) {
	switch kind {
	case reflect.String:
		return rawValue, nil
	case reflect.Bool:
		return strconv.ParseBool(rawValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.ParseInt(rawValue, 10, 64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.ParseUint(rawValue, 10, 64)
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(rawValue, 64)
	default:
		return nil, fmt.Errorf("unsupported type: %s", kind)
	}
}
