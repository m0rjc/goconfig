package goconfigtools

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

// Option is a functional option for configuring the Load function.
type Option func(*loadOptions)

// loadOptions holds the configuration options for Load.
type loadOptions struct {
	// validators maps field paths to their validator functions
	validators map[string][]Validator
}

// WithValidator registers a custom validator for a specific field path.
// Path uses dot notation for nested fields (e.g., "AI.APIKey", "WebHook.Timeout").
// Multiple validators can be registered for the same field; all will be executed in order.
func WithValidator(path string, validator Validator) Option {
	return func(opts *loadOptions) {
		if opts.validators == nil {
			opts.validators = make(map[string][]Validator)
		}
		opts.validators[path] = append(opts.validators[path], validator)
	}
}

// newLoadOptions creates default load options.
func newLoadOptions() *loadOptions {
	return &loadOptions{
		validators: make(map[string][]Validator),
	}
}

// applyOptions applies the given options to the load options.
func (opts *loadOptions) applyOptions(options []Option) {
	for _, opt := range options {
		opt(opts)
	}
}

// Load populates the given configuration struct from environment variables
// using the `key`, `default`, `required`, `min`, and `max` struct tags.
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
// Validation:
//   - Use min:"value" and max:"value" tags for numeric range validation
//   - Use WithValidator option for custom validation logic
//   - Validators run after type conversion but before field assignment
//
// Options:
//   - WithValidator(path, validator): Register custom validator for a field
func Load(config interface{}, options ...Option) error {
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

	return loadStruct(v, "", opts)
}

// loadStruct recursively loads configuration values into a struct.
// fieldPath tracks the current position in the struct hierarchy for validators.
func loadStruct(v reflect.Value, fieldPath string, opts *loadOptions) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Build the current field path
		currentPath := fieldType.Name
		if fieldPath != "" {
			currentPath = fieldPath + "." + fieldType.Name
		}

		// Handle nested structs
		// TODO: When we implement custom marshallers allow a marshaller to be registered against
		//       a whole struct and take over here.
		if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Duration(0)) {
			if err := loadStruct(field, currentPath, opts); err != nil {
				return err
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

		// Parse min/max tags and register validators
		if err := registerBuiltinValidators(fieldType, currentPath, opts); err != nil {
			return err
		}

		// Set the field value based on its type (with validation)
		if err := setField(field, value, key, currentPath, opts); err != nil {
			return err
		}
	}

	return nil
}

// setField sets a field value based on its type and runs validators.
func setField(field reflect.Value, value string, key string, fieldPath string, opts *loadOptions) error {
	var typedValue any

	// Parse and convert the value based on field type
	switch field.Kind() {
	case reflect.String:
		typedValue = value

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value %q for %s: %w", value, key, err)
		}
		typedValue = boolVal

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Special handling for time.Duration
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration value %q for %s: %w", value, key, err)
			}
			typedValue = duration
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int value %q for %s: %w", value, key, err)
			}
			typedValue = intVal
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid uint value %q for %s: %w", value, key, err)
		}
		typedValue = uintVal

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value %q for %s: %w", value, key, err)
		}
		typedValue = floatVal

	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	// Run validators before setting the field
	if validators, exists := opts.validators[fieldPath]; exists {
		for _, validator := range validators {
			if err := validator(typedValue); err != nil {
				return fmt.Errorf("invalid value for %s: %w", key, err)
			}
		}
	}

	// Set the field after validation passes
	switch field.Kind() {
	case reflect.String:
		field.SetString(typedValue.(string))
	case reflect.Bool:
		field.SetBool(typedValue.(bool))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			field.SetInt(int64(typedValue.(time.Duration)))
		} else {
			field.SetInt(typedValue.(int64))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetUint(typedValue.(uint64))
	case reflect.Float32, reflect.Float64:
		field.SetFloat(typedValue.(float64))
	default:
		// My IDE warns me that this would be iota fields. We'd not expect an iota field here.
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}

func (opts *loadOptions) addValidator(fieldPath string, validator Validator) {
	if opts.validators == nil {
		opts.validators = make(map[string][]Validator)
	}
	opts.validators[fieldPath] = append(opts.validators[fieldPath], validator)
}
