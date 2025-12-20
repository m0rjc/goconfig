package goconfigtools

import "strings"

// ConfigErrors collects multiple runtime configuration errors.
// It maintains the order errors were encountered and provides formatted output.
type ConfigErrors struct {
	errors []configError
}

// configError represents a single configuration error for a specific environment variable.
type configError struct {
	key string // Environment variable name (e.g., "DB_PORT", "API_KEY")
	err error  // The underlying error
}

// Error implements the error interface.
// It formats all collected errors as: "KEY1: error1; KEY2: error2"
func (ce *ConfigErrors) Error() string {
	if len(ce.errors) == 0 {
		return ""
	}

	var parts []string
	for _, e := range ce.errors {
		msg := e.err.Error()
		// Strip "invalid value for KEY: " prefix to avoid duplication
		prefix := "invalid value for " + e.key + ": "
		msg = strings.TrimPrefix(msg, prefix)
		parts = append(parts, e.key+": "+msg)
	}
	return strings.Join(parts, "; ")
}

// Add adds a new error for the given environment variable.
func (ce *ConfigErrors) Add(key string, err error) {
	ce.errors = append(ce.errors, configError{key: key, err: err})
}

// HasErrors returns true if any errors were collected.
func (ce *ConfigErrors) HasErrors() bool {
	return len(ce.errors) > 0
}

// Len returns the number of errors collected.
func (ce *ConfigErrors) Len() int {
	return len(ce.errors)
}

// Unwrap returns all underlying errors for Go 1.20+ error inspection.
func (ce *ConfigErrors) Unwrap() []error {
	result := make([]error, len(ce.errors))
	for i, e := range ce.errors {
		result[i] = e.err
	}
	return result
}
