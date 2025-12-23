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

### 3. Custom Type System - Building Block Approach

The custom type system uses a **building block architecture** where you compose simple, reusable components to create type handlers. This approach replaced the old parser/validator system and provides maximum flexibility through composition.

#### Core Building Blocks

**TypedHandler[T]** - A handler that knows how to parse and validate values of type T. Handlers are built by combining building blocks.

**FieldProcessor[T]** - A function that converts a string to type T:
```go
type FieldProcessor[T any] func(rawValue string) (T, error)
```

**Validator[T]** - A function that validates a value of type T:
```go
type Validator[T any] func(value T) error
```

**Wrapper[T]** - A factory that wraps a FieldProcessor to add behavior (like validation):
```go
type Wrapper[T any] func(tags reflect.StructTag, inputProcess FieldProcessor[T]) (FieldProcessor[T], error)
```

#### Building Block Functions

These are the core building blocks you compose to create custom types:

**NewCustomType[T](parser, validators...)** - Start with a parser and add validators
- Creates a complete handler from a parser function and optional validators
- This is the most common way to create custom types

**AddValidators[T](handler, validators...)** - Add validators to an existing handler
- Takes any handler and wraps it with additional validators
- Useful for extending built-in type handlers

**AddDynamicValidation[T](handler, wrapper)** - Add dynamic validation that reads struct tags
- Adds a wrapper that can read struct tags and customize behavior per field
- For advanced scenarios where validation depends on struct tags

**CastCustomType[T, U](handler)** - Transform a handler from type T to type U
- Uses Go's type conversion rules to transform between compatible types
- Useful for creating handlers for type aliases (e.g., `type Port int`)

**NewStringEnumType[T](values...)** - Create an enum validator for string-based types
- Specialized builder for string enum types
- Automatically validates that values match one of the provided options

#### Default Type Handlers

These provide base handlers for built-in types that you can extend:
- `DefaultStringType()` - Returns `TypedHandler[string]`
- `DefaultIntegerType[T]()` - Returns `TypedHandler[T]` for int types (int, int8, int16, int32, int64)
- `DefaultUnsignedIntegerType[T]()` - Returns `TypedHandler[T]` for uint types (uint, uint8, uint16, uint32, uint64)
- `DefaultFloatIntegerType[T]()` - Returns `TypedHandler[T]` for float types (float32, float64)
- `DefaultDurationType()` - Returns `TypedHandler[time.Duration]`

These handlers already include tag-based validation (min, max, pattern). Use them as building blocks to add custom validation.

#### Registering Custom Types

Use `WithCustomType[T]()` to register a handler when loading config:
```go
err := goconfig.Load(ctx, &config,
    goconfig.WithCustomType[APIKey](apiKeyHandler),
    goconfig.WithCustomType[Status](statusHandler),
)
```

Or use `RegisterCustomType[T]()` to register globally (before Load):
```go
goconfig.RegisterCustomType[APIKey](apiKeyHandler)
// Now all APIKey fields will use this handler
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

### New System (Building Block Approach)
```go
// New: Type-based handlers with building blocks
type APIKey string

apiKeyHandler := goconfig.NewCustomType(
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
4. **Composability**: Building block architecture - compose handlers using `AddValidators()`, `CastCustomType()`, `AddDynamicValidation()`, etc.

## Common Patterns

### Custom String Types with Validation

```go
type Email string

// Building block approach: start with parser, add validators
emailHandler := goconfig.NewCustomType(
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

// Use the specialized enum building block
err := goconfig.Load(ctx, &config,
    goconfig.WithCustomType[LogLevel](
        goconfig.NewStringEnumType(LogDebug, LogInfo, LogWarn, LogError),
    ),
)
```

### Adding Validation to Built-in Types (Composition)

```go
// Building block approach: Take default int handler and add validators
baseHandler := goconfig.DefaultIntegerType[int]()
evenHandler := goconfig.AddValidators(baseHandler, func(v int) error {
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
// Now all int fields get tag validation (min/max) AND the even check
```

### Type Aliases with CastCustomType

```go
type Port int

// Building block: Cast the int64 handler to Port type
portHandler := goconfig.CastCustomType[int64, Port](
    goconfig.DefaultIntegerType[int64](),
)

// Or add validators after casting
portWithValidation := goconfig.AddValidators(portHandler, func(p Port) error {
    if int(p)%10 != 0 {
        return errors.New("port must be multiple of 10")
    }
    return nil
})

type Config struct {
    Port Port `key:"PORT" default:"8080" min:"1024" max:"65535"`
}

err := goconfig.Load(ctx, &config,
    goconfig.WithCustomType[Port](portWithValidation),
)
```

### Complex Custom Types

```go
type ServerAddress struct {
    Host string
    Port int
}

// Building block: Custom parser for complex type
addressHandler := goconfig.NewCustomType(
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
    // Add validators as additional building blocks
    func(addr ServerAddress) error {
        if addr.Port < 1024 || addr.Port > 65535 {
            return errors.New("port must be in range 1024-65535")
        }
        return nil
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

## Building Block Composition Patterns

The power of the building block system is in **composition**. Here are common patterns:

### Pattern 1: Start with Parser, Add Validators
```go
handler := goconfig.NewCustomType(parser, validator1, validator2, ...)
```

### Pattern 2: Extend Default Type with Validators
```go
handler := goconfig.AddValidators(goconfig.DefaultIntegerType[int](), myValidator)
```

### Pattern 3: Cast Then Validate
```go
handler := goconfig.AddValidators(
    goconfig.CastCustomType[int64, Port](goconfig.DefaultIntegerType[int64]()),
    portValidator,
)
```

### Pattern 4: Chain Multiple Wrappers
```go
handler := goconfig.AddDynamicValidation(
    goconfig.AddValidators(baseHandler, validator1),
    customWrapper,
)
```

## Key Files

- `custom_types.go` - Public API for building blocks
- `internal/customtypes/parser.go` - Parser building block
- `internal/customtypes/validation_wrapper.go` - Validation wrapper building block
- `internal/customtypes/chain_handler.go` - Composition via chaining
- `internal/customtypes/transformer.go` - Type transformation building block
- `internal/customtypes/enum.go` - Enum building block
- `internal/readpipeline/typed_handler.go` - TypedHandler interface
- `internal/readpipeline/typeregistry.go` - Type registry
- `loadoptions.go` - WithCustomType() option
- `example/validation/main.go` - Example using custom types
