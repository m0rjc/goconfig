package goconfigtools

import (
	"reflect"
	"testing"
)

// mockRegistry is a mock implementation for testing that collects validators
type mockRegistry struct {
	validators []Validator
}

func newMockRegistry() (*mockRegistry, ValidatorRegistry) {
	mock := &mockRegistry{
		validators: make([]Validator, 0),
	}
	registry := ValidatorRegistry(func(validator Validator) {
		mock.validators = append(mock.validators, validator)
	})
	return mock, registry
}

// TestCreateMinValidator_Int tests min validation for integer types
func TestCreateMinValidator_Int(t *testing.T) {
	validator, err := createMinValidator(reflect.Int, "1024")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name      string
		value     int64
		shouldErr bool
		errMsg    string
	}{
		{"above minimum", int64(2000), false, ""},
		{"at minimum", int64(1024), false, ""},
		{"below minimum", int64(500), true, "value 500 is below minimum 1024"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error %q, got nil", tt.errMsg)
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestCreateMinValidator_Uint tests min validation for unsigned integer types
func TestCreateMinValidator_Uint(t *testing.T) {
	validator, err := createMinValidator(reflect.Uint, "512")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name      string
		value     uint64
		shouldErr bool
		errMsg    string
	}{
		{"above minimum", uint64(1000), false, ""},
		{"at minimum", uint64(512), false, ""},
		{"below minimum", uint64(100), true, "value 100 is below minimum 512"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error %q, got nil", tt.errMsg)
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestCreateMinValidator_Float tests min validation for float types
func TestCreateMinValidator_Float(t *testing.T) {
	validator, err := createMinValidator(reflect.Float64, "0.0")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name      string
		value     float64
		shouldErr bool
		errMsg    string
	}{
		{"above minimum", 0.75, false, ""},
		{"at minimum", 0.0, false, ""},
		{"below minimum", -0.5, true, "value -0.500000 is below minimum 0.000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error %q, got nil", tt.errMsg)
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestCreateMinValidator_InvalidValue tests error handling for invalid min values
func TestCreateMinValidator_InvalidValue(t *testing.T) {
	tests := []struct {
		name  string
		kind  reflect.Kind
		value string
	}{
		{"invalid int", reflect.Int, "not-a-number"},
		{"invalid uint", reflect.Uint, "not-a-number"},
		{"invalid float", reflect.Float64, "not-a-number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := createMinValidator(tt.kind, tt.value)
			if err == nil {
				t.Error("Expected error for invalid min value")
			}
		})
	}
}

// TestCreateMinValidator_UnsupportedType tests error for unsupported types
func TestCreateMinValidator_UnsupportedType(t *testing.T) {
	_, err := createMinValidator(reflect.String, "10")
	if err == nil {
		t.Fatal("Expected error for unsupported type")
	}
	if err.Error() != "min tag not supported for type string" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestCreateMaxValidator_Int tests max validation for integer types
func TestCreateMaxValidator_Int(t *testing.T) {
	validator, err := createMaxValidator(reflect.Int, "65535")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name      string
		value     int64
		shouldErr bool
		errMsg    string
	}{
		{"below maximum", int64(8080), false, ""},
		{"at maximum", int64(65535), false, ""},
		{"above maximum", int64(70000), true, "value 70000 exceeds maximum 65535"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error %q, got nil", tt.errMsg)
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestCreateMaxValidator_Uint tests max validation for unsigned integer types
func TestCreateMaxValidator_Uint(t *testing.T) {
	validator, err := createMaxValidator(reflect.Uint, "4096")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name      string
		value     uint64
		shouldErr bool
		errMsg    string
	}{
		{"below maximum", uint64(1024), false, ""},
		{"at maximum", uint64(4096), false, ""},
		{"above maximum", uint64(5000), true, "value 5000 exceeds maximum 4096"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error %q, got nil", tt.errMsg)
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestCreateMaxValidator_Float tests max validation for float types
func TestCreateMaxValidator_Float(t *testing.T) {
	validator, err := createMaxValidator(reflect.Float64, "1.0")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name      string
		value     float64
		shouldErr bool
		errMsg    string
	}{
		{"below maximum", 0.75, false, ""},
		{"at maximum", 1.0, false, ""},
		{"above maximum", 1.5, true, "value 1.500000 exceeds maximum 1.000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error %q, got nil", tt.errMsg)
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestCreateMaxValidator_InvalidValue tests error handling for invalid max values
func TestCreateMaxValidator_InvalidValue(t *testing.T) {
	tests := []struct {
		name  string
		kind  reflect.Kind
		value string
	}{
		{"invalid int", reflect.Int, "not-a-number"},
		{"invalid uint", reflect.Uint, "not-a-number"},
		{"invalid float", reflect.Float64, "not-a-number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := createMaxValidator(tt.kind, tt.value)
			if err == nil {
				t.Error("Expected error for invalid max value")
			}
		})
	}
}

// TestCreateMaxValidator_UnsupportedType tests error for unsupported types
func TestCreateMaxValidator_UnsupportedType(t *testing.T) {
	_, err := createMaxValidator(reflect.String, "10")
	if err == nil {
		t.Fatal("Expected error for unsupported type")
	}
	if err.Error() != "max tag not supported for type string" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestCreatePatternValidator_SimplePattern tests pattern validation with simple patterns
func TestCreatePatternValidator_SimplePattern(t *testing.T) {
	validator, err := createPatternValidator(reflect.String, "^[a-zA-Z0-9_]+$")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
		errMsg    string
	}{
		{"valid alphanumeric", "user123", false, ""},
		{"valid with underscore", "user_name", false, ""},
		{"invalid with space", "user name", true, "value user name does not match pattern ^[a-zA-Z0-9_]+$"},
		{"invalid with dash", "user-name", true, "value user-name does not match pattern ^[a-zA-Z0-9_]+$"},
		{"invalid with special char", "user@name", true, "value user@name does not match pattern ^[a-zA-Z0-9_]+$"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error %q, got nil", tt.errMsg)
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestCreatePatternValidator_EmailPattern tests pattern validation with email regex
func TestCreatePatternValidator_EmailPattern(t *testing.T) {
	validator, err := createPatternValidator(reflect.String, `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"valid simple email", "user@example.com", false},
		{"valid email with dots", "user.name@example.com", false},
		{"valid email with plus", "user+tag@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"invalid no at sign", "userexample.com", true},
		{"invalid no domain", "user@", true},
		{"invalid no tld", "user@example", true},
		{"invalid spaces", "user @example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error for value %q, got nil", tt.value)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for value %q, got %v", tt.value, err)
				}
			}
		})
	}
}

// TestCreatePatternValidator_URLPattern tests pattern validation with URL regex
func TestCreatePatternValidator_URLPattern(t *testing.T) {
	validator, err := createPatternValidator(reflect.String, `^https?://[a-zA-Z0-9.-]+.*$`)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com", false},
		{"valid with path", "https://example.com/webhook", false},
		{"valid with query", "https://example.com/webhook?token=123", false},
		{"invalid ftp", "ftp://example.com", true},
		{"invalid no protocol", "example.com", true},
		{"invalid ws protocol", "ws://example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error for value %q, got nil", tt.value)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for value %q, got %v", tt.value, err)
				}
			}
		})
	}
}

// TestCreatePatternValidator_CaseSensitive tests that pattern matching is case-sensitive
func TestCreatePatternValidator_CaseSensitive(t *testing.T) {
	validator, err := createPatternValidator(reflect.String, "^[a-z]{3}$")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"lowercase matches", "abc", false},
		{"uppercase doesn't match", "ABC", true},
		{"mixed case doesn't match", "Abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error for value %q, got nil", tt.value)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for value %q, got %v", tt.value, err)
				}
			}
		})
	}
}

// TestCreatePatternValidator_SpecialCharacters tests pattern with special characters
func TestCreatePatternValidator_SpecialCharacters(t *testing.T) {
	validator, err := createPatternValidator(reflect.String, "^[a-zA-Z0-9_-]{8,}$")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"valid token with letters and numbers", "abcd1234", false},
		{"valid token with underscore", "test_token_123", false},
		{"valid token with hyphen", "test-token-123", false},
		{"valid mixed case", "TestToken123", false},
		{"invalid with spaces", "test token", true},
		{"invalid with special char", "test@token", true},
		{"invalid too short", "test", true},
		{"invalid with dot", "test.token", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.value)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error for value %q, got nil", tt.value)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for value %q, got %v", tt.value, err)
				}
			}
		})
	}
}

// TestCreatePatternValidator_InvalidRegex tests error handling for invalid regex
func TestCreatePatternValidator_InvalidRegex(t *testing.T) {
	_, err := createPatternValidator(reflect.String, "[invalid(regex")
	if err == nil {
		t.Fatal("Expected error for invalid regex pattern")
	}
}

// TestCreatePatternValidator_UnsupportedType tests error for non-string types
func TestCreatePatternValidator_UnsupportedType(t *testing.T) {
	_, err := createPatternValidator(reflect.Int, "^[0-9]+$")
	if err == nil {
		t.Fatal("Expected error for pattern on non-string type")
	}
	if err.Error() != "pattern tag not supported for type int" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestBuiltinValidatorFactory_MinTag tests that min tags are processed correctly
func TestBuiltinValidatorFactory_MinTag(t *testing.T) {
	mock, registry := newMockRegistry()
	fieldType := reflect.StructField{
		Name: "Port",
		Type: reflect.TypeOf(int(0)),
		Tag:  `min:"1024"`,
	}

	err := builtinValidatorFactory(fieldType, registry)
	if err != nil {
		t.Fatalf("Failed to register validators: %v", err)
	}

	if len(mock.validators) != 1 {
		t.Errorf("Expected 1 validator, got %d", len(mock.validators))
	}

	// Test the validator works
	validator := mock.validators[0]
	if err := validator(int64(2000)); err != nil {
		t.Errorf("Validator should pass for value 2000: %v", err)
	}
	if err := validator(int64(500)); err == nil {
		t.Error("Validator should fail for value 500")
	}
}

// TestBuiltinValidatorFactory_MaxTag tests that max tags are processed correctly
func TestBuiltinValidatorFactory_MaxTag(t *testing.T) {
	mock, registry := newMockRegistry()
	fieldType := reflect.StructField{
		Name: "Port",
		Type: reflect.TypeOf(int(0)),
		Tag:  `max:"65535"`,
	}

	err := builtinValidatorFactory(fieldType, registry)
	if err != nil {
		t.Fatalf("Failed to register validators: %v", err)
	}

	if len(mock.validators) != 1 {
		t.Errorf("Expected 1 validator, got %d", len(mock.validators))
	}

	// Test the validator works
	validator := mock.validators[0]
	if err := validator(int64(8080)); err != nil {
		t.Errorf("Validator should pass for value 8080: %v", err)
	}
	if err := validator(int64(70000)); err == nil {
		t.Error("Validator should fail for value 70000")
	}
}

// TestBuiltinValidatorFactory_PatternTag tests that pattern tags are processed correctly
func TestBuiltinValidatorFactory_PatternTag(t *testing.T) {
	mock, registry := newMockRegistry()
	fieldType := reflect.StructField{
		Name: "Username",
		Type: reflect.TypeOf(""),
		Tag:  `pattern:"^[a-zA-Z0-9_]+$"`,
	}

	err := builtinValidatorFactory(fieldType, registry)
	if err != nil {
		t.Fatalf("Failed to register validators: %v", err)
	}

	if len(mock.validators) != 1 {
		t.Errorf("Expected 1 validator, got %d", len(mock.validators))
	}

	// Test the validator works
	validator := mock.validators[0]
	if err := validator("user123"); err != nil {
		t.Errorf("Validator should pass for value 'user123': %v", err)
	}
	if err := validator("user-name"); err == nil {
		t.Error("Validator should fail for value 'user-name'")
	}
}

// TestBuiltinValidatorFactory_MultipleTags tests that multiple validation tags work together
func TestBuiltinValidatorFactory_MultipleTags(t *testing.T) {
	mock, registry := newMockRegistry()
	fieldType := reflect.StructField{
		Name: "Port",
		Type: reflect.TypeOf(int(0)),
		Tag:  `min:"1024" max:"65535"`,
	}

	err := builtinValidatorFactory(fieldType, registry)
	if err != nil {
		t.Fatalf("Failed to register validators: %v", err)
	}

	if len(mock.validators) != 2 {
		t.Errorf("Expected 2 validators, got %d", len(mock.validators))
	}

	// Test both validators work
	minValidator := mock.validators[0]
	maxValidator := mock.validators[1]

	if err := minValidator(int64(8080)); err != nil {
		t.Errorf("Min validator should pass for value 8080: %v", err)
	}
	if err := maxValidator(int64(8080)); err != nil {
		t.Errorf("Max validator should pass for value 8080: %v", err)
	}

	if err := minValidator(int64(500)); err == nil {
		t.Error("Min validator should fail for value 500")
	}
	if err := maxValidator(int64(70000)); err == nil {
		t.Error("Max validator should fail for value 70000")
	}
}

// TestBuiltinValidatorFactory_InvalidMinTag tests error handling for invalid min tags
func TestBuiltinValidatorFactory_InvalidMinTag(t *testing.T) {
	_, registry := newMockRegistry()
	fieldType := reflect.StructField{
		Name: "Port",
		Type: reflect.TypeOf(int(0)),
		Tag:  `min:"not-a-number"`,
	}

	err := builtinValidatorFactory(fieldType, registry)
	if err == nil {
		t.Fatal("Expected error for invalid min tag")
	}
}

// TestBuiltinValidatorFactory_InvalidMaxTag tests error handling for invalid max tags
func TestBuiltinValidatorFactory_InvalidMaxTag(t *testing.T) {
	_, registry := newMockRegistry()
	fieldType := reflect.StructField{
		Name: "Port",
		Type: reflect.TypeOf(int(0)),
		Tag:  `max:"not-a-number"`,
	}

	err := builtinValidatorFactory(fieldType, registry)
	if err == nil {
		t.Fatal("Expected error for invalid max tag")
	}
}

// TestBuiltinValidatorFactory_InvalidPatternTag tests error handling for invalid pattern tags
func TestBuiltinValidatorFactory_InvalidPatternTag(t *testing.T) {
	_, registry := newMockRegistry()
	fieldType := reflect.StructField{
		Name: "Field",
		Type: reflect.TypeOf(""),
		Tag:  `pattern:"[invalid(regex"`,
	}

	err := builtinValidatorFactory(fieldType, registry)
	if err == nil {
		t.Fatal("Expected error for invalid pattern tag")
	}
}

// TestBuiltinValidatorFactory_PatternOnNonStringType tests error for pattern on non-string type
func TestBuiltinValidatorFactory_PatternOnNonStringType(t *testing.T) {
	_, registry := newMockRegistry()
	fieldType := reflect.StructField{
		Name: "Port",
		Type: reflect.TypeOf(int(0)),
		Tag:  `pattern:"^[0-9]+$"`,
	}

	err := builtinValidatorFactory(fieldType, registry)
	if err == nil {
		t.Fatal("Expected error for pattern tag on non-string type")
	}

	expectedMsg := "pattern tag not supported for type int"
	if err.Error() != "invalid pattern tag value \"^[0-9]+$\" for field Port: "+expectedMsg {
		t.Errorf("Expected error %q, got %q", "invalid pattern tag value \"^[0-9]+$\" for field Port: "+expectedMsg, err.Error())
	}
}
