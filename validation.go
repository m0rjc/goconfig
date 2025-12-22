package goconfig

import (
	"reflect"

	"github.com/m0rjc/goconfig/process"
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
//   - Other types, such as Struct and Map, receive the value as a value not a pointer.
type Validator = process.Validator[any]
