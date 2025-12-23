# Validation

Built-in validation helps catch configuration errors early with clear error messages. This guide covers all validation features available in goconfig.

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
    if err := goconfig.Load(&cfg); err != nil {
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

Use custom types with the `WithCustomType` option to add custom validation logic beyond what struct tags provide. The new type system uses Go generics for type-safe validation.

### Basic Custom Types

```go
// Define custom types for fields that need special validation
type APIKey string
type Hostname string
type Port int

type Config struct {
    APIKey APIKey   `key:"API_KEY" required:"true"`
    Host   Hostname `key:"HOST" default:"localhost"`
    Port   Port     `key:"PORT" default:"8080" min:"1024" max:"65535"`
}

func main() {
    var cfg Config

    err := goconfig.Load(context.Background(), &cfg,
        // Validate API key format
        goconfig.WithCustomType[APIKey](goconfig.NewCustomHandler(
            func(rawValue string) (APIKey, error) {
                return APIKey(rawValue), nil
            },
            func(value APIKey) error {
                key := string(value)
                if !strings.HasPrefix(key, "sk-") {
                    return fmt.Errorf("API key must start with 'sk-'")
                }
                if len(key) < 20 {
                    return fmt.Errorf("API key must be at least 20 characters long")
                }
                return nil
            },
        )),

        // Validate host is not an IP address
        goconfig.WithCustomType[Hostname](goconfig.NewCustomHandler(
            func(rawValue string) (Hostname, error) {
                return Hostname(rawValue), nil
            },
            func(value Hostname) error {
                host := string(value)
                if net.ParseIP(host) != nil {
                    return fmt.Errorf("host must be a hostname, not an IP address")
                }
                return nil
            },
        )),

        // Additional port validation (even numbers only)
        goconfig.WithCustomType[Port](goconfig.NewCustomHandler(
            func(rawValue string) (Port, error) {
                v, err := strconv.Atoi(rawValue)
                return Port(v), err
            },
            func(value Port) error {
                if int(value)%10 != 0 {
                    return fmt.Errorf("port must be a multiple of 10")
                }
                return nil
            },
        )),
    )

    if err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
}
```

### Type-Safe Validators

Custom validators are type-safe - no type assertions needed:

```go
// String-based custom type
type Email string
emailHandler := goconfig.NewCustomHandler(
    func(rawValue string) (Email, error) {
        return Email(rawValue), nil
    },
    func(value Email) error {  // value is Email, not any
        if !strings.Contains(string(value), "@") {
            return errors.New("invalid email format")
        }
        return nil
    },
)

// Int-based custom type
type EvenPort int
evenPortHandler := goconfig.NewCustomHandler(
    func(rawValue string) (EvenPort, error) {
        v, err := strconv.Atoi(rawValue)
        return EvenPort(v), err
    },
    func(value EvenPort) error {  // value is EvenPort, not any
        if int(value)%2 != 0 {
            return errors.New("port must be even")
        }
        return nil
    },
)

// Duration-based custom type
type RequestTimeout time.Duration
timeoutHandler := goconfig.NewCustomHandler(
    func(rawValue string) (RequestTimeout, error) {
        d, err := time.ParseDuration(rawValue)
        return RequestTimeout(d), err
    },
    func(value RequestTimeout) error {  // value is RequestTimeout, not any
        if value < RequestTimeout(100*time.Millisecond) {
            return errors.New("timeout too short")
        }
        return nil
    },
)

### Enum Types

For string-based enums, use `NewEnumHandler` for automatic validation:

```go
type LogLevel string

const (
    LogDebug LogLevel = "debug"
    LogInfo  LogLevel = "info"
    LogWarn  LogLevel = "warn"
    LogError LogLevel = "error"
)

type Config struct {
    Level LogLevel `key:"LOG_LEVEL" default:"info"`
}

func main() {
    var cfg Config

    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithCustomType[LogLevel](
            goconfig.NewEnumHandler(LogDebug, LogInfo, LogWarn, LogError),
        ),
    )
}
```

The enum handler will automatically validate that the value is one of the provided options and return a clear error if not.

## Nested Field Validation

Custom type validators work with nested structs. Define custom types and they will be validated regardless of where they appear in the config structure:

```go
type DatabaseHost string
type APIEndpoint string

type Config struct {
    Database struct {
        Host     DatabaseHost `key:"DB_HOST" default:"localhost"`
        Port     int          `key:"DB_PORT" default:"5432" min:"1024" max:"65535"`
        Username string       `key:"DB_USER" default:"postgres"`
    }
    API struct {
        Key      string      `key:"API_KEY" required:"true"`
        Endpoint APIEndpoint `key:"API_ENDPOINT" default:"https://api.example.com"`
    }
}

func main() {
    var cfg Config

    err := goconfig.Load(context.Background(), &cfg,
        // Validate database host
        goconfig.WithCustomType[DatabaseHost](goconfig.NewCustomHandler(
            func(rawValue string) (DatabaseHost, error) {
                return DatabaseHost(rawValue), nil
            },
            func(value DatabaseHost) error {
                host := string(value)
                if host == "localhost" {
                    return fmt.Errorf("production environments must use a remote database")
                }
                return nil
            },
        )),

        // Validate API endpoint uses HTTPS
        goconfig.WithCustomType[APIEndpoint](goconfig.NewCustomHandler(
            func(rawValue string) (APIEndpoint, error) {
                return APIEndpoint(rawValue), nil
            },
            func(value APIEndpoint) error {
                endpoint := string(value)
                if !strings.HasPrefix(endpoint, "https://") {
                    return fmt.Errorf("API endpoint must use HTTPS")
                }
                return nil
            },
        )),
    )

    if err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
}
```

**Note:** Custom type handlers are registered by TYPE, not by field path. All fields of the same type will use the same handler, regardless of nesting level.

## Combining Validators

You can combine tag-based validation with custom type validators. All validations must pass:

### Using Custom Types with Tag Validation

```go
type Port int

type Config struct {
    Port Port `key:"PORT" default:"8080" min:"1024" max:"65535"`
}

func main() {
    var cfg Config

    // Create handler with custom validation that runs AFTER min/max from tags
    portHandler := goconfig.NewCustomHandler(
        func(rawValue string) (Port, error) {
            v, err := strconv.Atoi(rawValue)
            return Port(v), err
        },
        func(value Port) error {
            if int(value)%10 != 0 {
                return fmt.Errorf("port must be a multiple of 10")
            }
            return nil
        },
    )

    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithCustomType[Port](portHandler),
    )
}
```

### Adding Validators to Built-in Types

Use `PrependValidators()` to add custom validation to built-in types while keeping their tag validation:

```go
type Config struct {
    Port int `key:"PORT" default:"8080" min:"1024" max:"65535"`
}

func main() {
    var cfg Config

    // Get the standard int handler
    baseHandler := goconfig.NewTypedIntHandler[int]()

    // Add custom validator (port must be even)
    evenIntHandler, err := goconfig.PrependValidators(baseHandler, func(v int) error {
        if v%2 != 0 {
            return errors.New("must be even")
        }
        return nil
    })
    if err != nil {
        log.Fatal(err)
    }

    err = goconfig.Load(context.Background(), &cfg,
        goconfig.WithCustomType[int](evenIntHandler),
    )
}
```

### Multiple Validators

You can pass multiple validators to `NewCustomHandler`:

```go
type APIKey string

apiKeyHandler := goconfig.NewCustomHandler(
    func(rawValue string) (APIKey, error) {
        return APIKey(rawValue), nil
    },
    // Multiple validators
    func(value APIKey) error {
        if !strings.HasPrefix(string(value), "sk-") {
            return errors.New("must start with 'sk-'")
        }
        return nil
    },
    func(value APIKey) error {
        if len(value) < 20 {
            return errors.New("must be at least 20 characters")
        }
        return nil
    },
    func(value APIKey) error {
        if !strings.HasSuffix(string(value), "==") {
            return errors.New("must end with '=='")
        }
        return nil
    },
)
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
invalid value for PORT: below minimum 1024
invalid value for PORT: exceeds maximum 65535
invalid value for TIMEOUT: below minimum 1s
invalid value for TIMEOUT: exceeds maximum 5m
```

### Pattern Validation Errors

```
invalid value for USERNAME: does not match pattern "^[a-zA-Z0-9_]+$"
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
  - invalid value for PORT: below minimum 1024
  - invalid value for USERNAME: does not match pattern "^[a-zA-Z0-9_]+$"
  - invalid value for API.Endpoint: API endpoint must use HTTPS
```

## Complete Example

See the [validation example](../example/validation) for a comprehensive working example demonstrating all validation features.
