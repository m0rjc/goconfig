package goconfig

import (
	"errors"
	"log/slog"
	"strings"
)

var (
	ErrMissingConfigKey = errors.New("no configuration found for this key")
	ErrMissingValue     = errors.New("missing or blank value for this key")
)

// ConfigErrors collects multiple runtime configuration errors.
// It maintains the order errors were encountered and provides formatted output.
type ConfigErrors struct {
	// Errors contains the collected errors
	Errors []ConfigError
}

// ConfigError represents a single configuration error for a specific environment variable.
type ConfigError struct {
	Key string // Environment variable name (e.g., "DB_PORT", "API_KEY")
	Err error  // The underlying error
}

// Error implements the error interface.
// It formats all collected errors as: "KEY1: error1; KEY2: error2"
func (ce *ConfigErrors) Error() string {
	if len(ce.Errors) == 0 {
		return ""
	}

	var parts []string
	for _, e := range ce.Errors {
		msg := e.Err.Error()
		// Strip "invalid value for KEY: " prefix to avoid duplication
		prefix := "invalid value for " + e.Key + ": "
		msg = strings.TrimPrefix(msg, prefix)
		parts = append(parts, e.Key+": "+msg)
	}
	return strings.Join(parts, "\n")
}

// Add adds a new error for the given environment variable.
func (ce *ConfigErrors) Add(key string, err error) {
	ce.Errors = append(ce.Errors, ConfigError{Key: key, Err: err})
}

// HasErrors returns true if any errors were collected.
func (ce *ConfigErrors) HasErrors() bool {
	return len(ce.Errors) > 0
}

// Len returns the number of errors collected.
func (ce *ConfigErrors) Len() int {
	return len(ce.Errors)
}

// Unwrap returns all underlying errors for Go 1.20+ error inspection.
func (ce *ConfigErrors) Unwrap() []error {
	result := make([]error, len(ce.Errors))
	for i, e := range ce.Errors {
		result[i] = e.Err
	}
	return result
}

// ErrorLogOption provides options for the ConfigErrors.LogAll method
type ErrorLogOption func(*logSettings)

type logSettings struct {
	message string
}

// WithLogMessage sets the log message to be passed to the Logger in the ConfigErrors.LogAll method
func WithLogMessage(message string) ErrorLogOption {
	return func(s *logSettings) { s.message = message }
}

// LogError is a convenience function to log either a single error from the configuration load
// or the collection of validation errors. If a collection of validation errors is returned then
// they will be logged individually using ConfigErrors.LogAll().
func LogError(logger *slog.Logger, err error, opts ...ErrorLogOption) {
	if configErrs, ok := err.(*ConfigErrors); ok {
		configErrs.LogAll(logger, opts...)
	} else {
		settings := getLogSettings(opts...)
		logger.Error(settings.message, "error", err)
	}
}

// LogAll logs each configuration error using structured logging.
//
// Example usage:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
//	if err := LoadConfig(&cfg); err != nil {
//	    if configErrs, ok := err.(*ConfigErrors); ok {
//	        configErrs.LogAll(logger, WithLogMessage("configuration_error"))
//	    }
//	}
func (ce *ConfigErrors) LogAll(logger *slog.Logger, opts ...ErrorLogOption) {
	settings := getLogSettings(opts...)

	for _, e := range ce.Errors {
		logger.Error(settings.message,
			"key", e.Key,
			"error", e.Err,
		)
	}
}

// getLogSettings resolves the logging settings given the options.
func getLogSettings(opts ...ErrorLogOption) *logSettings {
	settings := &logSettings{
		message: "configuration error",
	}

	for _, opt := range opts {
		opt(settings)
	}

	return settings
}
