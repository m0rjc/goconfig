package goconfigtools

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// Validator validates a field's value after type conversion.
// The validator returns an error if the field value is not acceptable.
type Validator func(value any) error

// registerBuiltinValidators parses min and max tags and registers appropriate validators.
func registerBuiltinValidators(fieldType reflect.StructField, fieldPath string, opts *loadOptions) error {
	minTag := fieldType.Tag.Get("min")
	maxTag := fieldType.Tag.Get("max")
	patternTag := fieldType.Tag.Get("pattern")

	kind := fieldType.Type.Kind()

	// Register min validator
	if minTag != "" {
		validator, err := createMinValidator(kind, minTag)
		if err != nil {
			return fmt.Errorf("invalid min tag value %q for field %s: %w", minTag, fieldType.Name, err)
		}
		opts.addValidator(fieldPath, validator)
	}

	// Register max validator
	if maxTag != "" {
		validator, err := createMaxValidator(kind, maxTag)
		if err != nil {
			return fmt.Errorf("invalid max tag value %q for field %s: %w", maxTag, fieldType.Name, err)
		}
		opts.addValidator(fieldPath, validator)
	}

	// Register pattern validator for strings
	if patternTag != "" {
		validator, err := createPatternValidator(kind, patternTag)
		if err != nil {
			return fmt.Errorf("invalid pattern tag value %q for field %s: %w", patternTag, fieldType.Name, err)
		}
		opts.addValidator(fieldPath, validator)
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

// createMaxValidator creates a maximum value validator for the given type.
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
