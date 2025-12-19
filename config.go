package goconfigtools

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

// Load populates the given configuration struct from environment variables
// using the `key`, `default`, and `required` struct tags.
//
// Value resolution follows this precedence (highest to lowest):
//  1. Environment variable (if set)
//  2. Tag default (if specified with default:"value")
//  3. Pre-initialized struct value (allows coded defaults)
//
// Fields without any value source are left unchanged, allowing you to
// set defaults by initializing the struct before calling Load.
//
// Use required:"true" to enforce that a field must be set via environment
// variable or default tag.
func Load(config interface{}) error {
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	return loadStruct(v)
}

// loadStruct recursively loads configuration values into a struct
func loadStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Handle nested structs
		if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Duration(0)) {
			if err := loadStruct(field); err != nil {
				return fmt.Errorf("error loading nested struct %s: %w", fieldType.Name, err)
			}
			continue
		}

		// Get the key tag
		key := fieldType.Tag.Get("key")
		if key == "" {
			// No key tag, skip this field
			continue
		}

		// Get the default value
		defaultValue := fieldType.Tag.Get("default")

		// Check if field is required
		required := fieldType.Tag.Get("required") == "true"

		// Get the environment variable value
		envValue := os.Getenv(key)

		// Determine which value to use
		value := envValue
		if value == "" && defaultValue != "" {
			value = defaultValue
		}

		// If still empty, check if it's required
		if value == "" {
			if required {
				return fmt.Errorf("required environment variable %s is not set", key)
			}
			// Not required and no value, leave field unchanged (preserves pre-initialized values)
			continue
		}

		// Set the field value based on its type
		if err := setField(field, value); err != nil {
			return fmt.Errorf("error setting field %s from key %s: %w", fieldType.Name, key, err)
		}
	}

	return nil
}

// setField sets a field value based on its type
func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value %q: %w", value, err)
		}
		field.SetBool(boolVal)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Special handling for time.Duration
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration value %q: %w", value, err)
			}
			field.SetInt(int64(duration))
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int value %q: %w", value, err)
			}
			field.SetInt(intVal)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint value %q: %w", value, err)
		}
		field.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value %q: %w", value, err)
		}
		field.SetFloat(floatVal)

	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}
