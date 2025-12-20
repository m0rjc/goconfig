# Changelog

## [v1.0.0-beta.2] - 2025-12-20

### Fixed

- **Min and Max Durations**: The `min` and `max` tags now support durations
- **Restructured documentation**: Asked Claude to restructure the docs so that its future self can do a better job of
  using this code. It's no longer my voice, but I think that's clear enough to anybody reading.

## [v1.0.0-beta.1] - 2025-12-20

### Breaking Changes
- **Context support**: `Load()` function now requires `context.Context` as the first parameter
  ```go
  // Before: Load(&config)
  // After:  Load(context.Background(), &config)
  ```
- **KeyStore signature change**: Custom keystores must now return `(value string, present bool, err error)` instead of `(string, error)`
  - The `present` boolean indicates whether the key exists in the store (value may be empty string)
  - Allows distinguishing between "key not found" vs "key found with empty value"
- **Exported error types**: `ConfigErrors` and `ConfigError` are now exported (previously `configErrors` and `configError`)

### Added
- **Context support throughout the library**: All key lookups now receive context for cancellation and timeout support
- **Enhanced required field semantics**:
  - `keyRequired="true"`: Key must be present in store but can have empty value
  - `required="true"`: Key must be present AND have non-empty value
- **Sentinel errors**: Exported `ErrMissingConfigKey` and `ErrMissingValue` for error inspection via `errors.Is()`
- **Structured logging support**:
  - `ConfigErrors.LogAll()`: Log each error with structured fields
  - `LogError()`: Convenience function that handles both single errors and ConfigErrors
  - `WithLogMessage()`: Option to customize log message
- **CompositeStore**: Chain multiple keystores with automatic fallback
- **JSON deserialization support**: Convert JSON strings from config into `map[string]interface{}` or JSON-annotated structs
- **Custom field parsing**: Support for types implementing custom unmarshaling interfaces
- **Custom keystores**: Use alternative key-value stores instead of environment variables via `WithKeyStore` option
- **Validation system**:
  - Built-in `min`/`max` validators for numeric types
  - `pattern` validator for regular expression matching on strings
  - Custom validators via `WithValidator` option
  - Support for validators on nested fields using dot notation
- **Multi-error reporting**: Collects and reports all configuration errors in a single pass instead of failing on the first error
- **Pointer type support**: All supported types can now be used as pointers

### Changed
- **Defaulting behavior**: Defaults are only applied when key is completely absent from keystore, not when set to empty string
  - Previously: `export FOO=` would use the default
  - Now: `export FOO=` uses the empty string, overriding the default

## [v0.2.0] - 2025-12-19

### Added
- Custom validation mechanism

## [v0.1.0] - 2025-12-19

### Added
- Initial release of goconfigtools
- Load configuration from environment variables using struct tags
- Support for nested structs
- Optional default values via `default` tag
- Required field validation via `required` tag
- Type conversion for common types:
  - `string`
  - `bool`
  - `int`, `int8`, `int16`, `int32`, `int64`
  - `uint`, `uint8`, `uint16`, `uint32`, `uint64`
  - `float32`, `float64`
  - `time.Duration`
- Value resolution precedence: environment variable > tag default > pre-initialized value
- Clear error messages for missing required fields or invalid values
- Comprehensive test suite

