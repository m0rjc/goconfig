package goconfig

import (
	"context"
	"fmt"
	"reflect"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

// Load populates the given configuration struct from environment variables
// using the `key`, `default`, `required`, `min`, `max`, and `pattern` struct tags.
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
//
// Builtin Validation Tags:
//   - min:"value" and max:"value": Numeric range validation (int, uint, float types)
//   - pattern:"regex": Regular expression validation (string types only)
//
// Custom Validation:
//   - WithValidator(path, validator): Add a validator for a specific field path
//   - WithValidatorFactory(factory): Register a factory to auto-add validators based on field metadata
//   - Validators run after type conversion but before field assignment
//
// Options:
//   - WithValidator(path, validator): Register custom validator for a specific field
//   - WithValidatorFactory(factory): Register a custom validator factory
//
// Example:
//
//	type Config struct {
//	    Port     int    `key:"PORT" default:"8080" min:"1024" max:"65535"`
//	    Username string `key:"USERNAME" pattern:"^[a-zA-Z0-9_]+$"`
//	    Email    string `key:"EMAIL"`
//	}
//
//	cfg := Config{}
//	err := Load(&cfg, WithValidator("Email", emailValidator))
func Load(ctx context.Context, config interface{}, options ...Option) error {
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	opts := newLoadOptions()
	opts.applyOptions(options)

	errors := &ConfigErrors{Errors: make([]ConfigError, 0)}
	if err := loadStruct(ctx, v, "", opts, errors); err != nil {
		return err // configuration error, fail-fast
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

// loadStruct recursively loads configuration values into a struct.
// fieldPath tracks the current position in the struct hierarchy for validators.
func loadStruct(ctx context.Context, v reflect.Value, fieldPath string, opts *loadOptions, errors *ConfigErrors) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Get the key tag
		key := fieldType.Tag.Get("key")

		// Skip unexported fields, but error if they have a key tag
		if !field.CanSet() {
			if key != "" {
				return fmt.Errorf("field %s is unexported but has a key tag", fieldType.Name)
			}
			continue
		}

		// Build the current field path
		currentPath := fieldType.Name
		if fieldPath != "" {
			currentPath = fieldPath + "." + fieldType.Name
		}

		if key == "" {
			// If it's a struct or pointer to struct then recurse into it
			effectiveField := field
			if field.Kind() == reflect.Ptr {
				if field.IsNil() && field.Type().Elem().Kind() == reflect.Struct {
					field.Set(reflect.New(field.Type().Elem()))
				}
				effectiveField = field.Elem()
			}

			if effectiveField.Kind() == reflect.Struct {
				if err := loadStruct(ctx, effectiveField, currentPath, opts, errors); err != nil {
					return err
				}
			}
			// No key tag, skip this field
			continue
		}

		configuredValue, present, err := getConfiguredValue(ctx, fieldType.Tag, key, opts)
		if err != nil {
			return err
		}

		isKeyRequired := fieldType.Tag.Get("keyRequired") == "true"
		isValueRequired := fieldType.Tag.Get("required") == "true"
		if !present {
			if isKeyRequired || isValueRequired {
				errors.Add(key, ErrMissingConfigKey)
			}
			continue
		}

		// If empty, check if it's required
		if configuredValue == "" && isValueRequired {
			errors.Add(key, ErrMissingValue)
			continue
		}

		// Configure the processor, then run it
		processor, err := readpipeline.New(fieldType.Type, fieldType.Tag, opts.typeRegistry)
		if err != nil {
			return fmt.Errorf("setting up field readpipeline %s: %v", currentPath, err)
		}

		// Parse the configured value to produce a raw value
		rawValue, err := processor(configuredValue)
		if err != nil {
			errors.Add(key, err)
			continue
		}

		setField(field, rawValue, key, errors)
	}

	return nil
}

// getConfiguredValue reads the string value to use for the field. This is read from the keystore or
// any default provided in the tag.
func getConfiguredValue(ctx context.Context, tag reflect.StructTag, key string, opts *loadOptions) (string, bool, error) {
	// Get the environment variable value
	envValue, present, err := opts.keyStore(ctx, key)
	if present || err != nil {
		return envValue, present, err
	}

	// Get the default value
	defaultValue, defaultPresent := tag.Lookup("default")
	if defaultPresent {
		return defaultValue, true, nil
	}

	return "", false, nil
}

// setField sets a field value based on its type. It automatically handles pointer fields
func setField(field reflect.Value, value any, key string, errors *ConfigErrors) {
	// Set the field after validation passes
	val := reflect.ValueOf(value)
	fieldType := field.Type()

	if val.Type().ConvertibleTo(fieldType) {
		field.Set(val.Convert(fieldType))
	} else if fieldType.Kind() == reflect.Ptr && val.Type().ConvertibleTo(fieldType.Elem()) {
		// Create a new pointer of the required type
		ptr := reflect.New(fieldType.Elem())
		// Dereference the pointer and set the converted value
		ptr.Elem().Set(val.Convert(fieldType.Elem()))
		// Assign the pointer to the field
		field.Set(ptr)
	} else {
		// This is unexpected because our pipeline setup system should always ensure that we have a pipeline
		// that is compatible with the target field.
		errors.Add(key, fmt.Errorf("value of type %s cannot be converted to %s", val.Type(), fieldType))
	}
}
