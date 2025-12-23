# goconfig

A simple, type-safe Go library for loading configuration from environment variables using struct tags.

## Features

- 🏷️ **Struct-based configuration** - Define config with Go structs and tags
- ✅ **Built-in validation** - `min`, `max`, and `pattern` tags plus custom type validators ([docs](docs/validation.md) | [example](example/validation))
- 🎯 **Type-safe** - Automatic conversion for primitives, durations, and JSON with generic type handlers
- 🧱 **Building block architecture** - Compose custom types from simple, reusable components ([docs](docs/custom-types.md))
- 🔄 **Flexible defaults** - Struct tags or pre-initialized values ([docs](docs/defaulting.md))
- 🌳 **Nested structs** - Organize configuration hierarchically
- 🔧 **Extensible** - Custom types and key stores ([docs](docs/advanced.md))
- 💬 **Clear errors** - Descriptive validation and missing field errors

## Installation

```bash
go get github.com/m0rjc/goconfig
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/m0rjc/goconfig"
)

type Config struct {
    // Basic fields
    APIKey string `key:"API_KEY" required:"true"`
    Host   string `key:"HOST" default:"localhost"`

    // With validation
    Port    int           `key:"PORT" default:"8080" min:"1024" max:"65535"`
    Timeout time.Duration `key:"TIMEOUT" default:"30s" min:"1s" max:"5m"`

    // Nested configuration
    Database struct {
        Host     string `key:"DB_HOST" default:"localhost"`
        Port     int    `key:"DB_PORT" default:"5432"`
        Username string `key:"DB_USER" required:"true"`
        Password string `key:"DB_PASSWORD" required:"true"`
    }
}

func main() {
    var config Config

    if err := goconfig.Load(context.Background(), &config); err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    fmt.Printf("Server: %s:%d\n", config.Host, config.Port)
    fmt.Printf("Database: %s:%d\n", config.Database.Host, config.Database.Port)
}
```

Set environment variables and run:

```bash
export API_KEY="sk-your-api-key"
export DB_USER="appuser"
export DB_PASSWORD="secret"
export PORT="8080"
go run main.go
```

## Struct Tags

| Tag | Purpose | Example |
|-----|---------|---------|
| `key` | Environment variable name (required) | `key:"PORT"` |
| `default` | Default value if not set | `default:"8080"` |
| `min` | Minimum value (numbers, durations) | `min:"1024"` |
| `max` | Maximum value (numbers, durations) | `max:"65535"` |
| `pattern` | Regex pattern for strings | `pattern:"^[a-z]+$"` |
| `required` | Must be present and non-empty | `required:"true"` |
| `keyRequired` | Must be present (can be empty) | `keyRequired:"true"` |

## Supported Types

- **Primitives:** `string`, `bool`
- **Integers:** `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- **Floats:** `float32`, `float64`
- **Duration:** `time.Duration` - uses Go format: "30s", "5m", "1h"
- **JSON:** `map[string]interface{}` or structs with `json` tags
- **Pointers:** All above types as pointers
- **Nested structs:** Organize configuration hierarchically

## Validation

### Built-in Validators

```go
type ServerConfig struct {
    Port       int           `key:"PORT" default:"8080" min:"1024" max:"65535"`
    MaxConns   int           `key:"MAX_CONNS" default:"100" min:"1" max:"10000"`
    LoadFactor float64       `key:"LOAD_FACTOR" default:"0.75" min:"0.0" max:"1.0"`
    Timeout    time.Duration `key:"TIMEOUT" default:"30s" min:"1s" max:"5m"`
    Username   string        `key:"USERNAME" pattern:"^[a-zA-Z0-9_]+$"`
}
```

Validation errors provide clear messages:
```
invalid value for PORT: below minimum 1024
```

### Custom Types

Define custom types with validation using the **building block architecture** - compose simple, reusable components:

```go
type APIKey string

type Config struct {
    APIKey APIKey `key:"API_KEY" required:"true"`
}

// Building block approach: parser + validators
apiKeyHandler := goconfig.NewCustomType(
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

err := goconfig.Load(context.Background(), &cfg,
    goconfig.WithCustomType[APIKey](apiKeyHandler),
)
```

The building block system lets you compose handlers:
- `NewCustomType` - Start with parser and validators
- `AddValidators` - Add validators to existing handlers
- `CastCustomType` - Transform handlers for type aliases
- `NewStringEnumType` - Specialized enum builder

📚 **[Custom Types Guide](docs/custom-types.md)** | **[Validation Guide](docs/validation.md)** | **[Example](example/validation)**

## JSON Configuration

Load complex JSON structures from environment variables:

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

📚 **[JSON Guide](docs/json.md)**

## Documentation

- 📖 **[Documentation Index](docs/)** - Complete guides and reference
- 🧱 **[Custom Types Guide](docs/custom-types.md)** - Building block architecture for custom types
- 📋 **[Validation](docs/validation.md)** - Min/max, pattern, and custom type validators
- ⚙️ **[Defaulting & Required Fields](docs/defaulting.md)** - How defaults and required work
- 🔄 **[JSON Deserialization](docs/json.md)** - Working with JSON config
- 🔧 **[Advanced Features](docs/advanced.md)** - Custom key stores and advanced patterns
- 💡 **[Examples](example/)** - Working code examples

## Examples

- **[Simple Example](example/simple)** - Basic usage with defaults and nested structs
- **[Validation Example](example/validation)** - Comprehensive validation demonstration

## Advanced Usage

### Custom Key Stores

Read from sources other than environment variables:

```go
// Composite store: try environment, then fall back to file
store := goconfig.CompositeStore(
    goconfig.EnvironmentKeyStore,
    fileKeyStore("/etc/myapp/config"),
)

err := goconfig.Load(context.Background(), &cfg,
    goconfig.WithKeyStore(store),
)
```

Supports AWS Secrets Manager, HashiCorp Vault, config files, and more.

📚 **[Advanced Guide](docs/advanced.md)**

## Error Handling

```go
err := goconfig.Load(context.Background(), &config)
if err != nil {
    // Check for specific errors
    if errors.Is(err, goconfig.ErrMissingConfigKey) {
        log.Fatal("Missing required environment variable")
    }

    log.Fatalf("Configuration error: %v", err)
}
```

Multiple errors are collected and reported together for easier debugging.

## Testing

```bash
go test -v
```

## Contributing

Contributions welcome! Please open an issue or pull request on [GitHub](https://github.com/m0rjc/goconfig).

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Links

- 📚 [Documentation](docs/)
- 🐙 [GitHub Repository](https://github.com/m0rjc/goconfig)
- 📦 [pkg.go.dev](https://pkg.go.dev/github.com/m0rjc/goconfig)
