package goconfigtools

import (
	"fmt"
	"reflect"
)

// Option is a functional option for configuring the Load function.
type Option func(*loadOptions)

// loadOptions holds the configuration options for Load.
type loadOptions struct {
	// parsers allows the parser for a given key to be overridden
	parsers map[string]Parser // key is fieldPath
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

// WithParser registers a custom parser at a given path.
func WithParser(path string, parser Parser) Option {
	return func(opts *loadOptions) {
		opts.addParser(path, parser)
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
		parsers:            make(map[string]Parser),
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

		// Get the key tag
		key := fieldType.Tag.Get("key")
		if key == "" {
			// If it's a struct then recurse into it
			if field.Kind() == reflect.Struct {
				if err := loadStruct(field, currentPath, opts, errors); err != nil {
					return err
				}
			}
			// No key tag, skip this field
			continue
		}

		configuredValue, err := getConfiguredValue(fieldType.Tag, key, opts)
		if err != nil {
			errors.Add(key, fmt.Errorf("error reading key %s: %w", key, err))
			continue
		}

		// If empty, check if it's required
		if configuredValue == "" {
			required := fieldType.Tag.Get("required") == "true"
			if required {
				errors.Add(key, fmt.Errorf("required environment variable %s is not set", key))
			}
			// Either blank for no value, or a missing required value, and we're collecting errors.
			continue
		}

		// Parse the configured value to produce a raw value
		rawValue, err := parseValue(configuredValue, currentPath, field, opts)
		if err != nil {
			errors.Add(key, fmt.Errorf("error parsing value %s: %w", configuredValue, err))
			continue
		}

		// Validate the raw value
		ok, err := validate(rawValue, key, currentPath, fieldType, opts, errors)
		if err != nil {
			return err
		}

		// Set the field value based on its type
		if ok {
			setField(field, rawValue, key, errors)
		}
	}

	return nil
}

// getConfiguredValue reads the string value to use for the field. This is read from the keystore or
// any default provided in the tag.
func getConfiguredValue(tag reflect.StructTag, key string, opts *loadOptions) (string, error) {
	// Get the default value
	defaultValue := tag.Get("default")

	// Get the environment variable value
	envValue, err := opts.keyStore(key)
	if err != nil {
		// Consider a store error to be fatal
		return "", err
	}

	// Determine which value to use
	value := envValue
	if value == "" && defaultValue != "" {
		value = defaultValue
	}
	return value, nil
}

// setField sets a field value based on its type
func setField(field reflect.Value, value any, key string, errors *ConfigErrors) {
	// Set the field after validation passes
	val := reflect.ValueOf(value)
	fieldType := field.Type()

	if val.Type().ConvertibleTo(fieldType) {
		field.Set(val.Convert(fieldType))
	} else {
		errors.Add(key, fmt.Errorf("value of type %s cannot be converted to %s", val.Type(), fieldType))
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

func (opts *loadOptions) addParser(path string, parser Parser) {
	if opts.parsers == nil {
		opts.parsers = make(map[string]Parser)
	}
	opts.parsers[path] = parser
}
