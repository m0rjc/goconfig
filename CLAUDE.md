# goconfig - Developer Guide for Claude

This document provides an overview of the goconfig library architecture and type system for AI assistants helping with development.

## Overview

goconfig is a Go library for loading configuration from environment variables (or other sources) into typed structs with validation. It uses a type-based system with struct tags for declaring configuration fields.

## Core Concepts

### 1. Struct Tags

Configuration fields are declared using struct tags:

```go
type Config struct {
    Port int `key:"PORT" default:"8080" min:"1024" max:"65535"`
    Host string `key:"HOST" default:"localhost"`
}
```

**Available tags:**
- `key` - Environment variable name (required)
- `default` - Default value if not set
- `required` - Must be present and non-empty
- `keyRequired` - Must be present (can be empty)
- `min` - Minimum value (numbers, durations)
- `max` - Maximum value (numbers, durations)
- `pattern` - Regex pattern for strings

### 2. Built-in Type Support

The library has built-in support for:
- Primitives: `string`, `bool`
- Integers: `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- Floats: `float32`, `float64`
- Special types: `time.Duration`
- Complex types: JSON (for maps and structs with json tags)
- Pointers: All above types as pointers
- Nested structs

### 3. Custom Type System

The new type system (replaced the old parser/validator system) uses typed handlers registered per type.

#### Key Types

**TypedHandler[T]** - A handler that knows how to parse and validate values of type T
- Has a parser: `FieldProcessor[T]` (converts string to T)
- Has a validation wrapper: `Wrapper[T]` (adds validation stages)

**FieldProcessor[T]** - A function that converts a string to type T:
```go
type FieldProcessor[T any] func(rawValue string) (T, error)
```

**Validator[T]** - A function that validates a value of type T:
```go
type Validator[T any] func(value T) error
```

**Wrapper[T]** - A factory that wraps a FieldProcessor with validation:
```go
type Wrapper[T any] func(tags reflect.StructTag, inputProcess FieldProcessor[T]) (FieldProcessor[T], error)
```

#### Creating Custom Type Handlers

**NewCustomHandler[T]()** - Create a handler with custom parsing and validation:
```go
handler := goconfig.NewCustomHandler(
    func(rawValue string) (APIKey, error) {
        return APIKey(rawValue), nil
    },
    func(value APIKey) error {
        if !strings.HasPrefix(string(value), "sk-") {
            return fmt.Errorf("API key must start with 'sk-'")
        }
        return nil
    },
)
```

**NewEnumHandler[T]()** - Create a handler for enum types:
```go
type Status string
const (
    StatusActive Status = "active"
    StatusInactive Status = "inactive"
)
handler := goconfig.NewEnumHandler(StatusActive, StatusInactive)
```

**ReplaceParser()** - Replace the parser while keeping validators:
```go
baseHandler := goconfig.NewTypedIntHandler[int]()
customHandler, err := goconfig.ReplaceParser(baseHandler, func(rawValue string) (int, error) {
    v, err := strconv.Atoi(rawValue)
    return v * 2, err  // Example: multiply by 2
})
```

**AddValidators()** - Add validators to an existing handler:
```go
baseHandler := goconfig.NewTypedIntHandler[int]()
validatedHandler, err := goconfig.AddValidators(baseHandler, func(v int) error {
    if v%2 != 0 {
        return errors.New("must be even")
    }
    return nil
})
```

#### Standard Type Handlers

For extending built-in types with custom validation:
- `NewTypedStringHandler()` - Returns `TypedHandler[string]`
- `NewTypedIntHandler[T]()` - Returns `TypedHandler[T]` for int types
- `NewTypedUintHandler[T]()` - Returns `TypedHandler[T]` for uint types
- `NewTypedFloatHandler[T]()` - Returns `TypedHandler[T]` for float types
- `NewTypedDurationHandler()` - Returns `TypedHandler[time.Duration]`

#### Registering Custom Types

Use `WithCustomType[T]()` to register a handler:
```go
err := goconfig.Load(ctx, &config,
    goconfig.WithCustomType[APIKey](apiKeyHandler),
    goconfig.WithCustomType[Status](statusHandler),
)
```

**Important:** Handlers are registered by TYPE, not by field name. All fields of type T will use the same handler.

## Architecture

### Load Process

1. User calls `Load(ctx, &config, options...)`
2. Options are applied (custom types, key stores)
3. The config struct is reflected to find all fields
4. For each field:
   - Look up the handler in the type registry (by type)
   - If not found, create a default handler for the type
   - Build the processing pipeline (parser + validators)
   - Read the value from the key store
   - Parse and validate the value
   - Set the field value
5. Collect and return all errors

### Type Registry

The type registry maps `reflect.Type` to `PipelineBuilder`. When you register a custom type with `WithCustomType[T]()`, it:
1. Gets the reflect.Type for T
2. Wraps the TypedHandler[T] in a `PipelineBuilder` adapter
3. Registers it in the type registry

When processing a field, the system:
1. Gets the field's reflect.Type
2. Looks up the handler in the registry
3. Calls `Build(tags)` to create a `FieldProcessor[any]` for that field
4. Uses that processor to parse and validate the value

## Migration from Old System

### Old System (Deprecated)
```go
// Old: Field-based validators
err := goconfig.Load(ctx, &cfg,
    goconfig.WithValidator("APIKey", func(value any) error {
        key := value.(string)
        // validation...
    }),
    goconfig.WithParser("DatabaseURL", func(value string) (any, error) {
        // parsing...
    }),
)
```

### New System
```go
// New: Type-based handlers
type APIKey string

apiKeyHandler := goconfig.NewCustomHandler(
    func(rawValue string) (APIKey, error) {
        return APIKey(rawValue), nil
    },
    func(value APIKey) error {
        // validation...
    },
)

err := goconfig.Load(ctx, &cfg,
    goconfig.WithCustomType[APIKey](apiKeyHandler),
)
```

### Key Differences

1. **Type-based vs Field-based**: Old system used field names ("APIKey"), new system uses types (APIKey)
2. **Type safety**: New system uses generics for compile-time type safety
3. **Reusability**: Custom types can be reused across multiple fields automatically
4. **Composability**: New system provides `ReplaceParser()`, `AddValidators()`, etc. for composing handlers

## Common Patterns

### Custom String Types with Validation

```go
type Email string

emailHandler := goconfig.NewCustomHandler(
    func(rawValue string) (Email, error) {
        return Email(rawValue), nil
    },
    func(value Email) error {
        if !strings.Contains(string(value), "@") {
            return errors.New("invalid email format")
        }
        return nil
    },
)

type Config struct {
    AdminEmail Email `key:"ADMIN_EMAIL" required:"true"`
    UserEmail  Email `key:"USER_EMAIL"`  // Same handler applies
}

err := goconfig.Load(ctx, &config, goconfig.WithCustomType[Email](emailHandler))
```

### Enum Types

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

err := goconfig.Load(ctx, &config,
    goconfig.WithCustomType[LogLevel](
        goconfig.NewEnumHandler(LogDebug, LogInfo, LogWarn, LogError),
    ),
)
```

### Adding Validation to Built-in Types

```go
// Make all ints even
baseHandler := goconfig.NewTypedIntHandler[int]()
evenHandler, _ := goconfig.AddValidators(baseHandler, func(v int) error {
    if v%2 != 0 {
        return errors.New("must be even")
    }
    return nil
})

type Config struct {
    Port int `key:"PORT" default:"8080" min:"1024" max:"65535"`
}

err := goconfig.Load(ctx, &config,
    goconfig.WithCustomType[int](evenHandler),
)
```

### Complex Custom Types

```go
type ServerAddress struct {
    Host string
    Port int
}

addressHandler := goconfig.NewCustomHandler(
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
)

type Config struct {
    Server ServerAddress `key:"SERVER_ADDR" default:"localhost:8080"`
}
```

## Testing

Use in-memory key stores for testing:
```go
func TestConfig(t *testing.T) {
    testStore := func(ctx context.Context, key string) (string, bool, error) {
        data := map[string]string{
            "PORT": "8080",
            "HOST": "localhost",
        }
        value, found := data[key]
        return value, found, nil
    }

    var cfg Config
    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithKeyStore(testStore),
    )
}
```

## Error Handling

Errors are collected and returned together:
```go
err := goconfig.Load(ctx, &config)
if err != nil {
    var configErrs *goconfig.ConfigErrors
    if errors.As(err, &configErrs) {
        // Multiple errors - all will be logged
        goconfig.LogError(logger, err)
    }
}
```

## Key Files

- `custom_types.go` - Public API for custom types
- `internal/readpipeline/custom_types.go` - Custom type implementation
- `internal/readpipeline/typed_handler.go` - TypedHandler interface and implementation
- `internal/readpipeline/typeregistry.go` - Type registry
- `loadoptions.go` - WithCustomType() option
- `example/validation/main.go` - Example using custom types
