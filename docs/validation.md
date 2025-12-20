# Validation

Built-in validation helps catch configuration errors early with clear error messages. This guide covers all validation features available in goconfigtools.

## Table of Contents

- [Min/Max Range Validation](#minmax-range-validation)
- [Pattern Validation](#pattern-validation)
- [Custom Validators](#custom-validators)
- [Nested Field Validation](#nested-field-validation)
- [Combining Validators](#combining-validators)
- [Validation Order](#validation-order)
- [Error Messages](#error-messages)

## Min/Max Range Validation

Use `min` and `max` struct tags to enforce numeric ranges on integers, floats, and durations.

### Integer Validation

```go
type ServerConfig struct {
    Port         int   `key:"PORT" default:"8080" min:"1024" max:"65535"`
    MaxConns     int   `key:"MAX_CONNS" default:"100" min:"1" max:"10000"`
    WorkerCount  int8  `key:"WORKERS" default:"4" min:"1" max:"16"`
    BufferSize   uint  `key:"BUFFER_SIZE" default:"1024" min:"512" max:"8192"`
}

func main() {
    var cfg ServerConfig
    if err := goconfigtools.Load(&cfg); err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
    // Port is guaranteed to be between 1024 and 65535
    // MaxConns is guaranteed to be between 1 and 10000
}
```

**Supported integer types:** `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`

### Float Validation

```go
type RateLimitConfig struct {
    RequestsPerSecond float64 `key:"RATE_LIMIT_RPS" default:"100" min:"0.1" max:"10000"`
    LoadFactor        float64 `key:"LOAD_FACTOR" default:"0.75" min:"0.0" max:"1.0"`
    Temperature       float32 `key:"TEMPERATURE" default:"0.7" min:"0.0" max:"2.0"`
}
```

**Supported float types:** `float32`, `float64`

### Duration Validation

Durations use Go's standard duration format (`1s`, `30s`, `5m`, `1h`, etc.):

```go
type TimeoutConfig struct {
    ReadTimeout    time.Duration `key:"READ_TIMEOUT" default:"30s" min:"1s" max:"5m"`
    WriteTimeout   time.Duration `key:"WRITE_TIMEOUT" default:"30s" min:"1s" max:"5m"`
    IdleTimeout    time.Duration `key:"IDLE_TIMEOUT" default:"1m" min:"10s" max:"10m"`
    RequestTimeout time.Duration `key:"REQUEST_TIMEOUT" default:"10s" min:"100ms" max:"1m"`
}
```

**Duration format examples:**
- `100ms` - 100 milliseconds
- `1s` - 1 second
- `30s` - 30 seconds
- `5m` - 5 minutes
- `1h` - 1 hour
- `24h` - 24 hours

## Pattern Validation

Use `pattern` to specify a regular expression that string values must match:

```go
type SecurityConfig struct {
    // Username must be alphanumeric with underscores
    Username string `key:"USERNAME" pattern:"^[a-zA-Z0-9_]+$"`

    // Hostname can include dots and hyphens
    Hostname string `key:"HOSTNAME" pattern:"^[a-zA-Z0-9.-]+$"`

    // Email validation
    Email string `key:"ADMIN_EMAIL" pattern:"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`

    // Semantic version (e.g., "1.2.3")
    Version string `key:"VERSION" pattern:"^\\d+\\.\\d+\\.\\d+$"`
}
```

**Pattern syntax:** Uses Go's `regexp` package syntax (RE2). Remember to escape backslashes in struct tags.

## Custom Validators

Use the `WithValidator` option to add custom validation logic beyond what struct tags provide:

```go
type Config struct {
    APIKey   string `key:"API_KEY" required:"true"`
    Host     string `key:"HOST" default:"localhost"`
    Port     int    `key:"PORT" default:"8080" min:"1024" max:"65535"`
}

func main() {
    var cfg Config

    err := goconfigtools.Load(&cfg,
        // Validate API key format
        goconfigtools.WithValidator("APIKey", func(value any) error {
            key := value.(string)
            if !strings.HasPrefix(key, "sk-") {
                return fmt.Errorf("API key must start with 'sk-'")
            }
            if len(key) < 20 {
                return fmt.Errorf("API key must be at least 20 characters long")
            }
            return nil
        }),

        // Validate host is not an IP address
        goconfigtools.WithValidator("Host", func(value any) error {
            host := value.(string)
            if net.ParseIP(host) != nil {
                return fmt.Errorf("host must be a hostname, not an IP address")
            }
            return nil
        }),

        // Additional port validation
        goconfigtools.WithValidator("Port", func(value any) error {
            port := value.(int64)
            if port%10 != 0 {
                return fmt.Errorf("port must be a multiple of 10")
            }
            return nil
        }),
    )

    if err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
}
```

### Type Assertions in Custom Validators

When writing custom validators, use the appropriate type assertion based on the field type:

| Field Type | Validator Type Assertion | Example |
|------------|-------------------------|---------|
| `string` | `value.(string)` | `key := value.(string)` |
| `int`, `int8`, `int16`, `int32`, `int64` | `value.(int64)` | `port := value.(int64)` |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | `value.(uint64)` | `count := value.(uint64)` |
| `float32`, `float64` | `value.(float64)` | `ratio := value.(float64)` |
| `bool` | `value.(bool)` | `enabled := value.(bool)` |
| `time.Duration` | `value.(time.Duration)` | `timeout := value.(time.Duration)` |

## Nested Field Validation

Validators work with nested structs using dot notation:

```go
type Config struct {
    Database struct {
        Host     string `key:"DB_HOST" default:"localhost"`
        Port     int    `key:"DB_PORT" default:"5432" min:"1024" max:"65535"`
        Username string `key:"DB_USER" default:"postgres"`
    }
    API struct {
        Key      string `key:"API_KEY" required:"true"`
        Endpoint string `key:"API_ENDPOINT" default:"https://api.example.com"`
    }
}

func main() {
    var cfg Config

    err := goconfigtools.Load(&cfg,
        // Validate database host
        goconfigtools.WithValidator("Database.Host", func(value any) error {
            host := value.(string)
            if host == "localhost" {
                return fmt.Errorf("production environments must use a remote database")
            }
            return nil
        }),

        // Validate API endpoint uses HTTPS
        goconfigtools.WithValidator("API.Endpoint", func(value any) error {
            endpoint := value.(string)
            if !strings.HasPrefix(endpoint, "https://") {
                return fmt.Errorf("API endpoint must use HTTPS")
            }
            return nil
        }),
    )

    if err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
}
```

## Combining Validators

You can combine tag-based validation with custom validators. All validations must pass:

```go
type Config struct {
    Port int `key:"PORT" default:"8080" min:"1024" max:"65535"`
}

func main() {
    var cfg Config

    err := goconfigtools.Load(&cfg,
        // This runs AFTER min/max validation from the tag
        goconfigtools.WithValidator("Port", func(value any) error {
            port := value.(int64)
            if port%10 != 0 {
                return fmt.Errorf("port must be a multiple of 10")
            }
            return nil
        }),
    )
}
```

## Validation Order

Validations are executed in this order:

1. **Type conversion** - The string value from the environment is converted to the target type
2. **Tag-based validation** - `min`, `max`, and `pattern` tags are checked
3. **Custom validators** - `WithValidator` functions are executed in registration order

If any validation fails, the error is reported and remaining validations are skipped for that field.

## Error Messages

Validation errors provide clear, actionable messages:

### Min/Max Validation Errors

```
invalid value for PORT: value 500 is below minimum 1024
invalid value for PORT: value 70000 is above maximum 65535
invalid value for TIMEOUT: value 100ms is below minimum 1s
invalid value for TIMEOUT: value 10m is above maximum 5m
```

### Pattern Validation Errors

```
invalid value for USERNAME: value "user@host" does not match pattern "^[a-zA-Z0-9_]+$"
```

### Custom Validation Errors

Custom validators can return any error message:

```
invalid value for APIKey: API key must start with 'sk-'
invalid value for Database.Host: production environments must use a remote database
```

### Multiple Errors

When multiple fields have validation errors, they are all collected and reported together:

```
configuration errors:
  - invalid value for PORT: value 500 is below minimum 1024
  - invalid value for USERNAME: value "user@host" does not match pattern "^[a-zA-Z0-9_]+$"
  - invalid value for API.Endpoint: API endpoint must use HTTPS
```

## Complete Example

See the [validation example](../example/validation) for a comprehensive working example demonstrating all validation features.
