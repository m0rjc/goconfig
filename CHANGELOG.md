# Changelog

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

