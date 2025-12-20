package goconfigtools

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Option is a functional option for configuring the Load function.
type Option func(*loadOptions)

// loadOptions holds the configuration options for Load.
type loadOptions struct {
	// keyStore reads the values. Default to os.GetEnv()
	keyStore KeyStore
	// validatorFactories provide validators for a field
	validatorFactories []ValidatorFactory
	// validators maps field paths to their validator functions
	validators map[string][]Validator
}

// WithValidatorFactory registers a custom validator factory.
// Validator factories inspect struct fields and automatically register validators
// based on field metadata (type, tags, name, etc.).
//
// Factories are called for each field during Load, allowing you to:
//   - Add validation based on custom struct tags
//   - Apply type-specific validation rules
//   - Implement domain-specific validation patterns
//
// Example: Adding email validation via custom tag:
//
//	factory := func(fieldType reflect.StructField, registry ValidatorRegistry) error {
//	    if fieldType.Tag.Get("email") == "true" {
//	        registry(func(value any) error {
//	            email := value.(string)
//	            if !strings.Contains(email, "@") {
//	                return fmt.Errorf("invalid email format")
//	            }
//	            return nil
//	        })
//	    }
//	    return nil
//	}
//	Load(&cfg, WithValidatorFactory(factory))
//
// Multiple factories can be registered and will be called in registration order.
// The builtin factory (for min, max, and pattern tags) is always registered first.
func WithValidatorFactory(factory ValidatorFactory) Option {
	return func(opts *loadOptions) {
		opts.addValidatorFactory(factory)
	}
}

// withoutBuiltinValidators removes the builtin validators.
// This option must be supplied first.
// This option is intended for unit testing to allow the test code to isolate the
// validation component.
func withoutBuiltinValidators() Option {
	return func(opts *loadOptions) {
		opts.validatorFactories = make([]ValidatorFactory, 0)
	}
}

// WithValidator registers a custom validator for a specific field path.
// Path uses dot notation for nested fields (e.g., "AI.APIKey", "WebHook.Timeout").
// Multiple validators can be registered for the same field; all will be executed in order.
//
// The validator receives the converted value (after type conversion from the environment
// variable string) and should return an error if validation fails.
//
// Example: Validating a port is a multiple of 10:
//
//	Load(&cfg, WithValidator("Port", func(value any) error {
//	    port := value.(int64)
//	    if port%10 != 0 {
//	        return fmt.Errorf("port must be multiple of 10")
//	    }
//	    return nil
//	}))
//
// Use WithValidatorFactory instead if you want to apply validation based on
// field metadata (tags, type, name) rather than explicit field paths.
func WithValidator(path string, validator Validator) Option {
	return func(opts *loadOptions) {
		opts.addValidator(path, validator)
	}
}

// WithKeyStore replaces the environment variable keystore with an alternative.
// Use this to read from other sources such as a database or properties file.
func WithKeyStore(keyStore KeyStore) Option {
	return func(opts *loadOptions) {
		opts.keyStore = keyStore
	}
}

// newLoadOptions creates default load options.
func newLoadOptions() *loadOptions {
	return &loadOptions{
		keyStore:           EnvironmentKeyStore,
		validatorFactories: []ValidatorFactory{builtinValidatorFactory},
		validators:         make(map[string][]Validator),
	}
}

// applyOptions applies the given options to the load options.
func (opts *loadOptions) applyOptions(options []Option) {
	for _, opt := range options {
		opt(opts)
	}
}

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

	errors := &ConfigErrors{errors: make([]configError, 0)}
	if err := loadStruct(v, "", opts, errors); err != nil {
		return err // configuration error, fail-fast
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

// loadStruct recursively loads configuration values into a struct.
// fieldPath tracks the current position in the struct hierarchy for validators.
func loadStruct(v reflect.Value, fieldPath string, opts *loadOptions, errors *ConfigErrors) error {
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
			if err := loadStruct(field, currentPath, opts, errors); err != nil {
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
		envValue, err := opts.keyStore(key)
		if err != nil {
			// Consider a store error to be fatal
			return err
		}

		// Determine which value to use
		value := envValue
		if value == "" && defaultValue != "" {
			value = defaultValue
		}

		// If still empty, check if it's required
		if value == "" {
			if required {
				errors.Add(key, fmt.Errorf("required environment variable %s is not set", key))
			}
			// Not required and no value, or required but we're collecting errors - leave field unchanged
			continue
		}

		// Parse min/max tags and register validators
		if err := registerValidators(fieldType, currentPath, opts); err != nil {
			return err
		}

		// Set the field value based on its type (with validation)
		setField(field, value, key, currentPath, opts, errors)
	}

	return nil
}

// setField sets a field value based on its type and runs validators.
// Errors are collected in the errors parameter instead of being returned.
func setField(field reflect.Value, value string, key string, fieldPath string, opts *loadOptions, errors *ConfigErrors) {
	var typedValue any

	// Parse and convert the value based on field type
	switch field.Kind() {
	case reflect.String:
		typedValue = value

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			errors.Add(key, fmt.Errorf("invalid bool value %q for %s: %w", value, key, err))
			return
		}
		typedValue = boolVal

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Special handling for time.Duration
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				errors.Add(key, fmt.Errorf("invalid duration value %q for %s: %w", value, key, err))
				return
			}
			typedValue = duration
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				errors.Add(key, fmt.Errorf("invalid int value %q for %s: %w", value, key, err))
				return
			}
			typedValue = intVal
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			errors.Add(key, fmt.Errorf("invalid uint value %q for %s: %w", value, key, err))
			return
		}
		typedValue = uintVal

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			errors.Add(key, fmt.Errorf("invalid float value %q for %s: %w", value, key, err))
			return
		}
		typedValue = floatVal

	default:
		errors.Add(key, fmt.Errorf("unsupported field type: %s", field.Kind()))
		return
	}

	// Run validators before setting the field
	if validators, exists := opts.validators[fieldPath]; exists {
		for _, validator := range validators {
			if err := validator(typedValue); err != nil {
				errors.Add(key, fmt.Errorf("invalid value for %s: %w", key, err))
				return
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
		// This should never happen as we already handled unsupported types above
		panic(fmt.Sprintf("unsupported field type during set: %s", field.Kind()))
	}
}

func (opts *loadOptions) addValidator(fieldPath string, validator Validator) {
	if opts.validators == nil {
		opts.validators = make(map[string][]Validator)
	}
	opts.validators[fieldPath] = append(opts.validators[fieldPath], validator)
}

func (opts *loadOptions) addValidatorFactory(factory ValidatorFactory) {
	if opts.validatorFactories == nil {
		opts.validatorFactories = make([]ValidatorFactory, 0, 1)
	}
	opts.validatorFactories = append(opts.validatorFactories, factory)
}
