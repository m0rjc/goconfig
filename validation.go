package goconfig

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

// ValidatorRegistry is the callback to add a validator to the current field.
// Validator factories call this function to register validators for a field.
// The registry handles associating the validator with the current field path.
type ValidatorRegistry func(validator Validator)

// ValidatorFactory inspects a struct field and registers appropriate validators.
// Factories can examine the field's type, tags, and name to determine which validators to add.
// The registry parameter is used to register validators for the current field.
//
// Example: A factory that validates email fields based on a custom tag:
//
//	func emailValidatorFactory(fieldType reflect.StructField, registry ValidatorRegistry) error {
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
//
// Factories are called during configuration loading for each field that has a key tag.
// They are called before values are loaded, so they only have access to field metadata.
type ValidatorFactory func(fieldType reflect.StructField, registry ValidatorRegistry) error

// Validator validates a field's value after type conversion.
// The validator receives the converted value and returns an error if validation fails.
// Validators are called after the environment variable or default value is converted
// to the field's type but before it is assigned to the struct field.
//
// The value parameter type depends on the field type:
//   - int types receive int64
//   - uint types receive uint64
//   - float types receive float64
//   - string types receive string
//   - bool types receive bool
//   - time.Duration types receive time.Duration
type Validator func(value any) error

func validate(value any, key, fieldPath string, field reflect.StructField, opts *loadOptions, errors *ConfigErrors) (bool, error) {
	// TODO: No need to register just to fetch them back again.

	// Parse min/max tags and register validators
	if err := registerValidators(field, fieldPath, opts); err != nil {
		return false, err
	}

	// Run validators before setting the field
	ok := true
	if validators, exists := opts.validators[fieldPath]; exists {
		for _, validator := range validators {
			if err := validator(value); err != nil {
				ok = false
				errors.Add(key, fmt.Errorf("invalid value for %s: %w", key, err))
			}
		}
	}

	return ok, nil
}

// registerValidators registers any validators for the given field
func registerValidators(fieldType reflect.StructField, fieldPath string, opts *loadOptions) error {
	registry := func(validator Validator) {
		opts.addValidator(fieldPath, validator)
	}

	for _, factory := range opts.validatorFactories {
		if err := factory(fieldType, registry); err != nil {
			return err
		}
	}
	return nil
}

func builtinValidatorFactory(fieldType reflect.StructField, registry ValidatorRegistry) error {
	minTag := fieldType.Tag.Get("min")
	maxTag := fieldType.Tag.Get("max")
	patternTag := fieldType.Tag.Get("pattern")

	// Check for time.Duration type before checking kind
	// This is necessary because time.Duration is an alias for int64, but we want to
	// parse the min/max tags as duration strings (e.g., "30s", "5m") rather than integers
	if fieldType.Type == reflect.TypeOf(time.Duration(0)) {
		// Register min validator for duration
		if minTag != "" {
			validator, err := createMinDurationValidator(minTag)
			if err != nil {
				return fmt.Errorf("invalid min tag value %q for field %s: %w", minTag, fieldType.Name, err)
			}
			registry(validator)
		}

		// Register max validator for duration
		if maxTag != "" {
			validator, err := createMaxDurationValidator(maxTag)
			if err != nil {
				return fmt.Errorf("invalid max tag value %q for field %s: %w", maxTag, fieldType.Name, err)
			}
			registry(validator)
		}

		return nil
	}

	kind := fieldType.Type.Kind()

	// Register min validator
	if minTag != "" {
		validator, err := createMinValidator(kind, minTag)
		if err != nil {
			return fmt.Errorf("invalid min tag value %q for field %s: %w", minTag, fieldType.Name, err)
		}
		registry(validator)
	}

	// Register max validator
	if maxTag != "" {
		validator, err := createMaxValidator(kind, maxTag)
		if err != nil {
			return fmt.Errorf("invalid max tag value %q for field %s: %w", maxTag, fieldType.Name, err)
		}
		registry(validator)
	}

	// Register pattern validator for strings
	if patternTag != "" {
		validator, err := createPatternValidator(kind, patternTag)
		if err != nil {
			return fmt.Errorf("invalid pattern tag value %q for field %s: %w", patternTag, fieldType.Name, err)
		}
		registry(validator)
	}

	return nil
}

// createMinValidator creates a minimum value validator for the given type.
func createMinValidator(kind reflect.Kind, minStr string) (Validator, error) {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		min, err := strconv.ParseInt(minStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return func(value any) error {
			v := value.(int64)
			if v < min {
				return fmt.Errorf("value %d is below minimum %d", v, min)
			}
			return nil
		}, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		min, err := strconv.ParseUint(minStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return func(value any) error {
			v := value.(uint64)
			if v < min {
				return fmt.Errorf("value %d is below minimum %d", v, min)
			}
			return nil
		}, nil

	case reflect.Float32, reflect.Float64:
		min, err := strconv.ParseFloat(minStr, 64)
		if err != nil {
			return nil, err
		}
		return func(value any) error {
			v := value.(float64)
			if v < min {
				return fmt.Errorf("value %f is below minimum %f", v, min)
			}
			return nil
		}, nil

	default:
		return nil, fmt.Errorf("min tag not supported for type %s", kind)
	}
}

// createMaxValidator creates a maximum value validator for the given type.
func createMaxValidator(kind reflect.Kind, maxStr string) (Validator, error) {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		max, err := strconv.ParseInt(maxStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return func(value any) error {
			v := value.(int64)
			if v > max {
				return fmt.Errorf("value %d exceeds maximum %d", v, max)
			}
			return nil
		}, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		max, err := strconv.ParseUint(maxStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return func(value any) error {
			v := value.(uint64)
			if v > max {
				return fmt.Errorf("value %d exceeds maximum %d", v, max)
			}
			return nil
		}, nil

	case reflect.Float32, reflect.Float64:
		max, err := strconv.ParseFloat(maxStr, 64)
		if err != nil {
			return nil, err
		}
		return func(value any) error {
			v := value.(float64)
			if v > max {
				return fmt.Errorf("value %f exceeds maximum %f", v, max)
			}
			return nil
		}, nil

	default:
		return nil, fmt.Errorf("max tag not supported for type %s", kind)
	}
}

// createMinDurationValidator creates a minimum duration validator.
func createMinDurationValidator(minStr string) (Validator, error) {
	min, err := time.ParseDuration(minStr)
	if err != nil {
		return nil, err
	}
	return func(value any) error {
		v := value.(time.Duration)
		if v < min {
			return fmt.Errorf("value %s is below minimum %s", v, min)
		}
		return nil
	}, nil
}

// createMaxDurationValidator creates a maximum duration validator.
func createMaxDurationValidator(maxStr string) (Validator, error) {
	max, err := time.ParseDuration(maxStr)
	if err != nil {
		return nil, err
	}
	return func(value any) error {
		v := value.(time.Duration)
		if v > max {
			return fmt.Errorf("value %s exceeds maximum %s", v, max)
		}
		return nil
	}, nil
}

// createPatternValidator creates a pattern validator for strings.
func createPatternValidator(kind reflect.Kind, patternStr string) (Validator, error) {
	switch kind {
	case reflect.String:
		pattern, err := regexp.Compile(patternStr)
		if err != nil {
			return nil, fmt.Errorf("Invalid pattern %s: %w", pattern, err)
		}
		return func(value any) error {
			v := value.(string)
			if !pattern.MatchString(v) {
				return fmt.Errorf("value %s does not match pattern %s", v, patternStr)
			}
			return nil
		}, nil
	default:
		return nil, fmt.Errorf("pattern tag not supported for type %s", kind)
	}
}
