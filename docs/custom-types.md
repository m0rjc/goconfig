# Custom Types - Building Block Guide

This guide explains the building block architecture for custom types in goconfig. The system is designed around **composition** - you build complex type handlers by combining simple, reusable building blocks.

## Table of Contents

- [Overview](#overview)
- [Core Building Blocks](#core-building-blocks)
- [Building Block Functions](#building-block-functions)
- [Composition Patterns](#composition-patterns)
- [Complete Examples](#complete-examples)
- [Advanced Topics](#advanced-topics)

## Overview

The custom type system provides building blocks that you compose to create type handlers:

1. **Parser** - Converts string to your type
2. **Validator** - Validates values of your type
3. **Wrapper** - Adds behavior to a processor (like validation) dynamically based on struct tags
4. **Handler** - Combines parser and wrappers into a complete type handler
5. **Transformer** - Converts between compatible types

These building blocks can be combined in different ways to create exactly the behavior you need.

## Core Building Blocks

### FieldProcessor[T]

A function that converts a string to type T:

```go
type FieldProcessor[T any] func(rawValue string) (T, error)
```

This is the fundamental building block - everything ultimately produces a FieldProcessor.

### Validator[T]

A function that validates a value of type T:

```go
type Validator[T any] func(value T) error
```

Validators are pure functions that check if a value is valid.

### Wrapper[T]

A factory that wraps a FieldProcessor to add behavior. The system allows dynamic behavior based on struct tags:

```go
type Wrapper[T any] func(tags reflect.StructTag, inputProcess FieldProcessor[T]) (FieldProcessor[T], error)
```

Wrappers can:
- Add validation
- Read struct tags to customize behavior
- Transform values
- Add logging or other cross-cutting concerns (not intended, but possible)

### TypedHandler[T]

A handler that builds a FieldProcessor for a specific type:

```go
type TypedHandler[T any] interface {
    BuildPipeline(tags reflect.StructTag) (FieldProcessor[T], error)
}
```

This is the final building block - it combines everything and can read struct tags to customize behavior per field.

## Building Block Functions

### 1. NewCustomType - Start with Parser and Validators

The most common way to create a custom type:

```go
func NewCustomType[T any](parser FieldProcessor[T], validators ...Validator[T]) TypedHandler[T]
```

**Example:**
```go
type Email string

emailHandler := goconfig.NewCustomType(
    // Parser building block
    func(rawValue string) (Email, error) {
        return Email(rawValue), nil
    },
    // Validator building blocks
    func(value Email) error {
        if !strings.Contains(string(value), "@") {
            return errors.New("invalid email format")
        }
        return nil
    },
    func(value Email) error {
        if len(value) < 3 {
            return errors.New("email too short")
        }
        return nil
    },
)
```

### 2. AddValidators - Add Validators to Existing Handler

Extend any handler with additional validators:

```go
func AddValidators[T any](handler TypedHandler[T], validators ...Validator[T]) TypedHandler[T]
```

**Example:**
```go
// Start with default int handler (has tag validation)
baseHandler := goconfig.DefaultIntegerType[int]()

// Add custom validator building block
evenHandler := goconfig.AddValidators(baseHandler,
    func(v int) error {
        if v%2 != 0 {
            return errors.New("must be even")
        }
        return nil
    },
)
// Now has BOTH tag validation AND even check
```

### 3. CastCustomType - Transform Between Compatible Types

Transform a handler from type T to type U:

```go
func CastCustomType[T, U any](handler TypedHandler[T]) TypedHandler[U]
```

This is essential for type aliases - it allows you to reuse handlers for similar types.

**Example:**
```go
type Port int

// Cast int handler to Port type
portHandler := goconfig.CastCustomType[int, Port](
    goconfig.DefaultIntegerType[int](),
)

// Port now gets all int validation (min/max from tags)
```

### 4. NewStringEnumType - Specialized Enum Builder

Create an enum validator for string-based types:

```go
func NewStringEnumType[T ~string](validValues ...T) TypedHandler[T]
```

**Example:**
```go
type LogLevel string
const (
    LogDebug LogLevel = "debug"
    LogInfo  LogLevel = "info"
    LogWarn  LogLevel = "warn"
)

handler := goconfig.NewStringEnumType(LogDebug, LogInfo, LogWarn)
```

### 5. AddDynamicValidation - Advanced Wrapper

Add a wrapper that can read struct tags and customize behavior:

```go
func AddDynamicValidation[T any](handler TypedHandler[T], wrapper Wrapper[T]) TypedHandler[T]
```

**Example:**
```go
// Custom wrapper that reads "allowed" tag
customWrapper := func(tags reflect.StructTag, processor FieldProcessor[string]) (FieldProcessor[string], error) {
    allowed := tags.Get("allowed")
    if allowed == "" {
        return processor, nil
    }

    allowedValues := strings.Split(allowed, ",")
    return func(rawValue string) (string, error) {
        value, err := processor(rawValue)
        if err != nil {
            return value, err
        }
        for _, allowed := range allowedValues {
            if value == allowed {
                return value, nil
            }
        }
        return value, fmt.Errorf("not in allowed list: %v", allowedValues)
    }, nil
}

handler := goconfig.AddDynamicValidation(
    goconfig.DefaultStringType(),
    customWrapper,
)
```

### 6. Default Type Handlers - Building Blocks for Built-in Types

Get handlers for built-in types that already include tag validation:

```go
goconfig.DefaultStringType() // string with pattern validation
goconfig.DefaultIntegerType[int]() // int with min/max validation
goconfig.DefaultUnsignedIntegerType[uint]() // uint with min/max validation
goconfig.DefaultFloatIntegerType[float64]() // float64 with min/max validation
goconfig.DefaultDurationType() // time.Duration with min/max validation
```

These are perfect starting points for composition.

## Composition Patterns

The power of building blocks is in **composition**. Here are common patterns:

### Pattern 1: Start Simple, Add Validators

```go
// 1. Start with parser
handler := goconfig.NewCustomType(parser)

// 2. Add validators later
handler = goconfig.AddValidators(handler, validator1, validator2)
```

### Pattern 2: Extend Built-in Types

```go
// 1. Get default handler
base := goconfig.DefaultIntegerType[int]()

// 2. Add custom validation
custom := goconfig.AddValidators(base, customValidator)
```

### Pattern 3: Cast Then Validate

```go
type Port int

// 1. Cast int handler to Port
handler := goconfig.CastCustomType[int, Port](
    goconfig.DefaultIntegerType[int](),
)

// 2. Add Port-specific validation
handler = goconfig.AddValidators(handler, portValidator)
```

### Pattern 4: Chain Multiple Extensions

```go
// Start with base
handler := goconfig.DefaultStringType()

// Add validator
handler = goconfig.AddValidators(handler, validator1)

// Add dynamic validation
handler = goconfig.AddDynamicValidation(handler, wrapper1)

// Add more validators
handler = goconfig.AddValidators(handler, validator2)
```

## Complete Examples

### Example 1: API Key with Multiple Validations

```go
type APIKey string

apiKeyHandler := goconfig.NewCustomType(
    // Parser
    func(rawValue string) (APIKey, error) {
        return APIKey(rawValue), nil
    },
    // Validator 1: Check prefix
    func(key APIKey) error {
        if !strings.HasPrefix(string(key), "sk-") {
            return errors.New("API key must start with 'sk-'")
        }
        return nil
    },
    // Validator 2: Check length
    func(key APIKey) error {
        if len(key) < 20 {
            return errors.New("API key must be at least 20 characters")
        }
        return nil
    },
    // Validator 3: Check format
    func(key APIKey) error {
        if !regexp.MustCompile(`^sk-[a-zA-Z0-9]+$`).MatchString(string(key)) {
            return errors.New("API key contains invalid characters")
        }
        return nil
    },
)

type Config struct {
    APIKey APIKey `key:"API_KEY" required:"true"`
}

err := goconfig.Load(ctx, &cfg,
    goconfig.WithCustomType[APIKey](apiKeyHandler),
)
```

### Example 2: Port with Range and Multiple-of-10 Validation

```go
type Port int

// Compose building blocks
portHandler := goconfig.AddValidators(
    // Start with int handler (gives us min/max from tags)
    goconfig.CastCustomType[int, Port](
        goconfig.DefaultIntegerType[int](),
    ),
    // Add custom validation
    func(port Port) error {
        if int(port)%10 != 0 {
            return errors.New("port must be a multiple of 10")
        }
        return nil
    },
)

type Config struct {
    HTTP  Port `key:"HTTP_PORT" default:"8080" min:"1024" max:"65535"`
    HTTPS Port `key:"HTTPS_PORT" default:"8443" min:"1024" max:"65535"`
}

err := goconfig.Load(ctx, &cfg,
    goconfig.WithCustomType[Port](portHandler),
)
// Both ports get min/max validation AND multiple-of-10 check
```

### Example 3: URL with Protocol Validation

```go
type HTTPSURL string

urlHandler := goconfig.NewCustomType(
    // Parser with validation
    func(rawValue string) (HTTPSURL, error) {
        u, err := url.Parse(rawValue)
        if err != nil {
            return "", fmt.Errorf("invalid URL: %w", err)
        }
        return HTTPSURL(rawValue), nil
    },
    // Validator 1: Must use HTTPS
    func(urlStr HTTPSURL) error {
        u, _ := url.Parse(string(urlStr))
        if u.Scheme != "https" {
            return errors.New("URL must use HTTPS")
        }
        return nil
    },
    // Validator 2: Must have host
    func(urlStr HTTPSURL) error {
        u, _ := url.Parse(string(urlStr))
        if u.Host == "" {
            return errors.New("URL must have a host")
        }
        return nil
    },
)

type Config struct {
    APIEndpoint HTTPSURL `key:"API_ENDPOINT" required:"true"`
}
```

This sample is also achievable using the `pattern` tag validation.

### Example 4: Complex Type - Server Address

```go
type ServerAddress struct {
    Host string
    Port int
}

addressHandler := goconfig.NewCustomType(
    // Parser - converts "host:port" string to struct
    func(rawValue string) (ServerAddress, error) {
        parts := strings.Split(rawValue, ":")
        if len(parts) != 2 {
            return ServerAddress{}, errors.New("invalid format: expected host:port")
        }
        port, err := strconv.Atoi(parts[1])
        if err != nil {
            return ServerAddress{}, fmt.Errorf("invalid port: %w", err)
        }
        return ServerAddress{Host: parts[0], Port: port}, nil
    },
    // Validator 1: Port range
    func(addr ServerAddress) error {
        if addr.Port < 1024 || addr.Port > 65535 {
            return errors.New("port must be in range 1024-65535")
        }
        return nil
    },
    // Validator 2: Host not empty
    func(addr ServerAddress) error {
        if addr.Host == "" {
            return errors.New("host cannot be empty")
        }
        return nil
    },
    // Validator 3: Not localhost
    func(addr ServerAddress) error {
        if addr.Host == "localhost" || addr.Host == "127.0.0.1" {
            return errors.New("production must use remote address")
        }
        return nil
    },
)

type Config struct {
    Database ServerAddress `key:"DB_ADDR" default:"db.example.com:5432"`
    Cache    ServerAddress `key:"CACHE_ADDR" default:"cache.example.com:6379"`
}
```

### Example 5: Enum with Validation

```go
type Environment string
const (
    EnvDev  Environment = "development"
    EnvStage Environment = "staging"
    EnvProd  Environment = "production"
)

// Use specialized enum building block
envHandler := goconfig.NewStringEnumType(EnvDev, EnvStage, EnvProd)

type Config struct {
    Env Environment `key:"ENVIRONMENT" default:"development"`
}

err := goconfig.Load(ctx, &cfg,
    goconfig.WithCustomType[Environment](envHandler),
)
```

## Advanced Topics

### Global Registration

Register a type handler globally so it applies to all future Load calls:

```go
// Register once at package init or main
func init() {
    goconfig.RegisterCustomType[APIKey](apiKeyHandler)
    goconfig.RegisterCustomType[Port](portHandler)
}

// Now Load automatically uses registered handlers
var cfg Config
err := goconfig.Load(ctx, &cfg)
// No need for WithCustomType options
```

### Per-Load Override

Override global registration for a specific Load:

```go
// Globally registered
goconfig.RegisterCustomType[Port](strictPortHandler)

// Override for this load only
err := goconfig.Load(ctx, &cfg,
    goconfig.WithCustomType[Port](lenientPortHandler),
)
```

### Reusing Handlers

Handlers are composable - build complex handlers from simpler ones:

```go
// Base validators
var (
    notEmpty = func(s string) error {
        if s == "" {
            return errors.New("cannot be empty")
        }
        return nil
    }

    noSpaces = func(s string) error {
        if strings.Contains(s, " ") {
            return errors.New("cannot contain spaces")
        }
        return nil
    }
)

// Compose into handlers
usernameHandler := goconfig.AddValidators(
    goconfig.DefaultStringType(),
    notEmpty,
    noSpaces,
)

passwordHandler := goconfig.AddValidators(
    goconfig.DefaultStringType(),
    notEmpty,
    func(s string) error {
        if len(s) < 8 {
            return errors.New("must be at least 8 characters")
        }
        return nil
    },
)
```

### Testing Custom Types

Test your type handlers independently:

```go
func TestAPIKeyHandler(t *testing.T) {
    handler := createAPIKeyHandler()

    // Build pipeline with empty tags
    processor, err := handler.BuildPipeline("")
    require.NoError(t, err)

    // Test valid key
    key, err := processor("sk-1234567890123456789")
    assert.NoError(t, err)
    assert.Equal(t, APIKey("sk-1234567890123456789"), key)

    // Test invalid prefix
    _, err = processor("invalid-key")
    assert.Error(t, err)

    // Test too short
    _, err = processor("sk-123")
    assert.Error(t, err)
}
```

## Summary

The building block architecture provides:

1. **Composability** - Build complex handlers from simple parts
2. **Reusability** - Share validators and handlers across types
3. **Type Safety** - Generic functions ensure compile-time safety
4. **Flexibility** - Mix and match building blocks as needed
5. **Testability** - Test each building block independently

Start with the simple building blocks (`NewCustomType`, `AddValidators`, `CastCustomType`) and compose them to create exactly the behavior you need.
