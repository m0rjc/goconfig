# Changelog

## [v0.4.0] - 2025-12-24

This release feeds back some things I've found using the package in a larger project.
This is an interim release. I'll add some issues to GitHub for things I'd like to add in future releases.

### Added

Improvements to goconfig:
* Support for `*url.URL` type
  - Support for pointer types in the read pipeline
  - Separate the read pipeline code and the built-in types code for easier navigation

Improvements to samples:
* Example of using custom type handlers
* Example of using custom validation tags
  - Improvements to the helper methods in goconfig's `custom_types.go` to make this easier.
* Example of combining keystores using CompositeStore.
  - A simple store that reads values from a Properties file.

## [v0.3.0] - 2025-12-23

This is a return to 0.x versions due to the significance of the breaking changes. It's a real about turn in the
mechanism for providing custom validation. If you've not been using custom validators, then nothing changes for you
apart from the module rename.

The AI-generated validation code was becoming messy, with switch statements all over the place. A pipeline mechanism
was made which used a typed pipeline to convert from raw values to typed values. A custom types system was built on top
of this, providing building blocks for custom validators. The raw building blocks are currently an internal package
with much of their functionality exposed through functions and type aliases in the root `goconfig` package.

### Added

- **Range Validator**: If both `min` and `max` are present then use a Range Validator. This improves error messages.
  For example `PORT: must be between 1024 and 65536`

### Fixed

- **Do not log input**: Error messages were changed to not mention input values. This is for security, to prevent secrets
  leakage into logs. [#15](https://github.com/m0rjc/goconfig/issues/15)

### Breaking Changes

- **Module Rename**: The module was renamed from `goconfigtools` to `goconfig`. Its github repository was also renamed.
- **Custom Type System**: The old parser and validation mechanism was replaced by a custom type system.

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
- Initial release of goconfig
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

