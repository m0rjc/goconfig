### Development Guidelines for goconfig

#### 1. Build and Configuration
The project is a standard Go library and follows standard Go build procedures.

- **Go Version**: The project requires Go 1.25.4 or later as specified in `go.mod`.
- **Dependencies**: Uses minimal external dependencies. Standard library packages like `reflect`, `context`, and `encoding/json` are heavily used.
- **Build**: Run `go build ./...` to ensure everything compiles. Note that during large refactors, some internal packages may be in flux.

#### 2. Testing Information
Testing is a critical part of this library, given its reliance on reflection and type conversion.

- **Running Tests**: 
  - Run all tests: `go test ./...`
  - Run with verbose output: `go test -v ./...`
- **Adding New Tests**:
  - Tests should be added to `*_test.go` files in the root directory for public API testing, or within the respective package for internal testing.
  - When testing `Load`, it is recommended to use a mock `KeyStore` to avoid dependency on the environment.
- **Demonstration Process**:
  To test a new configuration struct:
  1. Define your struct with `key` tags.
  2. Create a mock `KeyStore` that returns the desired values for testing.
  3. Call `goconfig.Load` with the mock `KeyStore` using `WithKeyStore` option.

**Example Test**:
```go
func TestExampleLoad(t *testing.T) {
    type Config struct {
        Port int `key:"PORT" default:"8080"`
    }
    
    ctx := context.Background()
    var cfg Config
    
    // Mock KeyStore
    mockStore := func(ctx context.Context, key string) (string, bool, error) {
        if key == "PORT" {
            return "9000", true, nil
        }
        return "", false, nil
    }

    err := Load(ctx, &cfg, WithKeyStore(mockStore))
    if err != nil {
        t.Fatalf("Failed to load: %v", err)
    }
    
    if cfg.Port != 9000 {
        t.Errorf("Expected 9000, got %d", cfg.Port)
    }
}
```

#### 3. Additional Development Information
- **Internal Pipeline**: The library uses a pipeline approach located in `internal/process`. Each field is processed by a `FieldProcessor` which is built based on the field type and tags.
- **Type Handlers**: Type-specific logic is registered in `internal/process/typeregistry.go`. To add support for a new type, implement a `Handler` and register it in `kindParsers` or `specialTypeParsers`.
- **Validation**: Validation is integrated into the processing pipeline. The `Pipe` function in `internal/process/types.go` is used to chain validators to processors.
- **Struct Tags**:
  - `key`: The name of the environment variable/key in the store.
  - `default`: Default value if the key is missing.
  - `required`: If "true", the key must be present and non-empty.
  - `keyRequired`: If "true", the key must be present (can be empty).
  - `min`, `max`: Range validation for numbers and durations.
  - `pattern`: Regex validation for strings.
- **Reflection**: The core logic in `config.go` uses reflection to traverse structs. Ensure that fields are exported (start with an upper-case letter) so that they can be set by the library.
