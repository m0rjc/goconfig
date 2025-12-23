# goconfig Documentation

Comprehensive guides for using goconfig.

## Quick Links

- [Main README](../README.md) - Quick start and installation
- [Examples](../example) - Working code examples
- [API Documentation](https://pkg.go.dev/github.com/m0rjc/goconfig) - Generated from code

## Documentation Guides

### 🧱 [Custom Types - Building Block Guide](custom-types.md)
Learn the building block architecture for creating custom types.

**Topics covered:**
- Core building blocks (parser, validator, wrapper, handler, transformer)
- Building block functions (NewCustomType, AddValidators, CastCustomType, etc.)
- Composition patterns
- Complete examples (API keys, ports, URLs, complex types, enums)
- Advanced topics (global registration, reusable handlers, testing)

**When to read:** Essential when you need custom types beyond built-in validation. Start here to understand the composable building block approach.

---

### 📋 [Validation](validation.md)
Learn how to validate configuration values using built-in and custom validators.

**Topics covered:**
- Min/max range validation for integers, floats, and durations
- Pattern validation using regular expressions
- Custom type validation with building blocks
- Nested field validation
- Combining multiple validators
- Error messages and debugging

**When to read:** Essential for production applications that need to ensure configuration values are within acceptable ranges.

---

### ⚙️ [Defaulting and Required Fields](defaulting.md)
Understand how default values and required fields work.

**Topics covered:**
- Setting default values with the `default` tag
- Understanding `required` vs `keyRequired` tags
- Defaulting behavior with environment variables
- Pre-initialized struct defaults
- Sentinel errors (`ErrMissingConfigKey`, `ErrMissingValue`)

**When to read:** Read this when you need to understand the interaction between defaults, required fields, and environment variables.

---

### 🔄 [JSON Deserialization](json.md)
Work with JSON configuration values from environment variables.

**Topics covered:**
- Deserializing to `map[string]interface{}`
- Deserializing to typed structs
- Using pointer types for optional JSON
- Nested JSON structures
- Default JSON values
- Error handling

**When to read:** When you need to pass complex structured data through environment variables.

---

### 🔧 [Advanced Features](advanced.md)
Extend goconfig with custom behavior.

**Topics covered:**
- Custom parsers for specialized types
- Custom key stores (files, AWS Secrets Manager, Vault, etc.)
- Composite key stores for multi-source configuration
- Error handling and structured logging
- Testing with in-memory stores

**When to read:** When you need to read from sources other than environment variables, or parse custom data formats.

---

## Quick Reference

### Struct Tags

| Tag | Purpose | Example |
|-----|---------|---------|
| `key` | Environment variable name (required) | `key:"PORT"` |
| `default` | Default value if not set | `default:"8080"` |
| `min` | Minimum value (numbers, durations) | `min:"1024"` |
| `max` | Maximum value (numbers, durations) | `max:"65535"` |
| `pattern` | Regex pattern (strings) | `pattern:"^[a-z]+$"` |
| `required` | Must be present and non-empty | `required:"true"` |
| `keyRequired` | Must be present (can be empty) | `keyRequired:"true"` |

### Supported Types

- **Primitives:** `string`, `bool`
- **Integers:** `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- **Floats:** `float32`, `float64`
- **Duration:** `time.Duration` (e.g., "30s", "5m", "1h")
- **JSON:** `map[string]interface{}`, structs with `json` tags
- **Pointers:** All above types as pointers

### Load Options

```go
// Load with options
err := goconfig.Load(context.Background(), &config,
    goconfig.WithKeyStore(customStore),
    goconfig.WithCustomType[APIKey](apiKeyHandler),
)
```

## Examples

### Basic Configuration

```go
type Config struct {
    Port int    `key:"PORT" default:"8080"`
    Host string `key:"HOST" default:"localhost"`
}

var cfg Config
err := goconfig.Load(context.Background(), &cfg)
```

See: [simple example](../example/simple)

### With Validation

```go
type Config struct {
    Port    int           `key:"PORT" default:"8080" min:"1024" max:"65535"`
    Timeout time.Duration `key:"TIMEOUT" default:"30s" min:"1s" max:"5m"`
}

var cfg Config
err := goconfig.Load(context.Background(), &cfg)
```

See: [validation example](../example/validation), [validation.md](validation.md)

### With Custom Types (Building Blocks)

```go
type APIKey string

// Building block: parser + validator
apiKeyHandler := goconfig.NewCustomType(
    func(rawValue string) (APIKey, error) {
        return APIKey(rawValue), nil
    },
    func(key APIKey) error {
        if !strings.HasPrefix(string(key), "sk-") {
            return fmt.Errorf("API key must start with 'sk-'")
        }
        return nil
    },
)

type Config struct {
    APIKey APIKey `key:"API_KEY" required:"true"`
}

err := goconfig.Load(context.Background(), &cfg,
    goconfig.WithCustomType[APIKey](apiKeyHandler),
)
```

See: [custom-types.md](custom-types.md), [validation.md](validation.md#custom-validators)

### JSON Configuration

```go
type ModelParams struct {
    Temperature float64 `json:"temperature"`
    MaxTokens   int     `json:"max_tokens"`
}

type Config struct {
    Params ModelParams `key:"MODEL_PARAMS"`
}
```

```bash
export MODEL_PARAMS='{"temperature":0.7,"max_tokens":1000}'
```

See: [json.md](json.md)

### Custom Key Store

```go
store := goconfig.CompositeStore(
    goconfig.EnvironmentKeyStore,
    fileKeyStore("/etc/myapp/config"),
)

err := goconfig.Load(context.Background(), &cfg,
    goconfig.WithKeyStore(store),
)
```

See: [advanced.md](advanced.md#custom-key-stores)

## Common Patterns

### Production Configuration

```go
type Config struct {
    // Required fields
    APIKey      string `key:"API_KEY" required:"true"`
    DatabaseURL string `key:"DATABASE_URL" required:"true"`

    // With validation
    Port        int           `key:"PORT" default:"8080" min:"1024" max:"65535"`
    MaxConns    int           `key:"MAX_CONNS" default:"100" min:"1" max:"1000"`
    Timeout     time.Duration `key:"TIMEOUT" default:"30s" min:"1s" max:"5m"`

    // Feature flags
    DebugMode   bool   `key:"DEBUG" default:"false"`
    Environment string `key:"ENV" default:"production" pattern:"^(development|staging|production)$"`
}
```

### Testing Configuration

```go
func TestMyApp(t *testing.T) {
    testConfig := map[string]string{
        "PORT":         "8080",
        "API_KEY":      "sk-test-key",
        "DATABASE_URL": "postgres://localhost/test",
    }

    var cfg Config
    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithKeyStore(mapKeyStore(testConfig)),
    )

    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }
}
```

See: [advanced.md](advanced.md#in-memory-key-store-for-testing)

## Error Handling

```go
err := goconfig.Load(context.Background(), &config)
if err != nil {
    // Check for specific errors
    if errors.Is(err, goconfig.ErrMissingConfigKey) {
        log.Fatal("Missing required environment variable")
    }

    // Handle multiple errors
    var configErrs *goconfig.ConfigErrors
    if errors.As(err, &configErrs) {
        for _, e := range configErrs.Unwrap() {
            log.Printf("Config error: %v", e)
        }
    }

    log.Fatalf("Configuration error: %v", err)
}
```

See: [defaulting.md](defaulting.md#sentinel-errors), [advanced.md](advanced.md#error-handling)

## Getting Help

- **Questions?** Open an issue on [GitHub](https://github.com/m0rjc/goconfig/issues)
- **Found a bug?** Report it on [GitHub Issues](https://github.com/m0rjc/goconfig/issues)
- **Want a feature?** Request it on [GitHub Discussions](https://github.com/m0rjc/goconfig/discussions)

## Contributing

Contributions are welcome! Please see the main [README](../README.md) for contribution guidelines.
