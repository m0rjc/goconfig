# goconfigtools

A simple Go library for loading configuration from environment variables using struct tags.

## Features

- Load configuration from environment variables using struct tags
- Support for nested structs
- Optional default values
- Type conversion for common types: `string`, `bool`, `int`, `uint`, `float`, `time.Duration`
- Ability to convert JSON strings into maps or JSON annotated structs
- Built-in min,max,pattern validators plus support for custom validation
- Support for custom field parsing
- Support for a custom key-value store as an alternative to environment variables
- Clear error messages for missing required fields or invalid values

## Installation

```bash
go get github.com/m0rjc/goconfigtools
```

## Usage

Define your configuration struct with `key` and optional `default` tags:

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/m0rjc/goconfigtools"
)

type WebhookConfig struct {
    Path    string        `key:"WEBHOOK_PATH" default:"webhook"`
    Timeout time.Duration `key:"WEBHOOK_TIMEOUT"` // No default - required
}

type AIConfig struct {
    APIKey string `key:"OPENAI_API_KEY"`     // Required
    Model  string `key:"OPENAI_MODEL" default:"gpt-4"` // Optional with default
}

type Config struct {
    AI       AIConfig
    WebHook  WebhookConfig
    EnableAI bool `key:"ENABLE_AI" default:"false"`
}

func main() {
    var config Config

    if err := goconfigtools.Load(&config); err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    fmt.Printf("Configuration loaded: %+v\n", config)
}
```

## Struct Tags

- `key`: The environment variable name to read from (required)
- `default`: The default value to use if the environment variable is not set (optional)
- `min`: Minimum value for numeric types (optional)
- `max`: Maximum value for numeric types (optional)
- `pattern`: Regular expression for string types (optional)
- `required`: Set to "true" to require the field to not be empty (optional)
- `keyRequired`: Set to "true" to require the field to present, though it can be explicitly blank

## Supported Types

- `string`
- `bool`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `time.Duration` (uses Go's duration format: "30s", "1m", "1h", etc.)
- `map[string]interface{}` using JSON deserialisation
- `struct` using JSON deserialisation
- pointers to the above

## Examples

### Setting environment variables

```bash
export OPENAI_API_KEY="sk-..."
export WEBHOOK_TIMEOUT="30s"
export ENABLE_AI="true"
go run example/main.go
```

### Using defaults

If you don't set optional variables, defaults will be used:

```bash
export OPENAI_API_KEY="sk-..."
export WEBHOOK_TIMEOUT="30s"
# WEBHOOK_PATH will default to "webhook"
# OPENAI_MODEL will default to "gpt-4"
# ENABLE_AI will default to "false"
go run example/main.go
```


### Defaulting behaviour and required fields

* The `required="true"` tag checks that the key is both present and has a non-blank value.
* The `keyRequired="true"` tag requires the key to be present, but it may have a blank value.
* Failing the above errors, the value is set on the struct if the key is found in configuration
* If the key is absent then the value is not modified. This allows a struct to be passed with initialised defaults.

When using environment variables (default Key Store)

| Environment      | No Required Tags         | keyRequired="true"     | required="true"     |
|------------------|--------------------------|------------------------|---------------------|
| `export FOO=bar` | Value becomes "bar"      | Value becomes "bar"    | Value becomes "bar" |
| `export FOO=`    | Value becomes ""         | Value becomes ""       | _Error_             |
| `unset FOO`      | Value remains unchanged  | _Error_                | _Error_             |

**Important notes about `default` tags:**

* When a key is **unset**, the `default` value is used. This satisfies both `keyRequired="true"` and `required="true"`.
* When a key is **set to empty** (e.g., `export FOO=`), it overrides the default and the empty value is used. This will cause an error if `required="true"`.
* Defaults are only applied when the key is completely absent from the key store.

The sentinel errors are ErrMissingConfigKey and ErrMissingValue.

## Validation

### Min/Max Range Validation

Use `min` and `max` struct tags to enforce numeric ranges:

```go
type ServerConfig struct {
    Port       int     `key:"PORT" default:"8080" min:"1024" max:"65535"`
    MaxConns   int     `key:"MAX_CONNS" default:"100" min:"1" max:"10000"`
    LoadFactor float64 `key:"LOAD_FACTOR" default:"0.75" min:"0.0" max:"1.0"`
}

func main() {
    var cfg ServerConfig
    if err := goconfigtools.Load(&cfg); err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
    // Port is guaranteed to be between 1024 and 65535
}
```

If a value is outside the specified range, you'll get a clear error message:
```
invalid value for PORT: value 500 is below minimum 1024
```

### Pattern Validation

Use `pattern` to specify a regular expression to test string values.

### Custom Validators

Use the `WithValidator` option to add custom validation logic:

```go
type Config struct {
    APIKey string `key:"API_KEY" required:"true"`
    Host   string `key:"HOST" default:"localhost"`
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
                return fmt.Errorf("API key too short")
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
    )

    if err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
}
```

### Validators on Nested Fields

Validators work with nested structs using dot notation:

```go
type Config struct {
    Database struct {
        Host string `key:"DB_HOST" default:"localhost"`
        Port int    `key:"DB_PORT" default:"5432" min:"1024" max:"65535"`
    }
}

func main() {
    var cfg Config

    err := goconfigtools.Load(&cfg,
        goconfigtools.WithValidator("Database.Host", func(value any) error {
            host := value.(string)
            if host == "localhost" {
                return fmt.Errorf("production environments must use a remote database")
            }
            return nil
        }),
    )
}
```

### Combining Validators

You can combine tag-based min/max validation with custom validators:

```go
type Config struct {
    Port int `key:"PORT" default:"8080" min:"1024" max:"65535"`
}

func main() {
    var cfg Config

    err := goconfigtools.Load(&cfg,
        // Additional validation: port must be a multiple of 10
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

Multiple validators are executed in order, and all must pass for the configuration to be valid.

### JSON Deserialisation

Given `OPENAI_MODEL_PARAMS={"temperature":0.7}`

```go
type Config struct {
	Value map[string]interface{} `key:"OPENAI_MODEL_PARAMS"`
}

config := Config{}
err := Load(&config, WithKeyStore(keyStore))
```

This also works with typed structures

```go
type ModelParameters struct {
	Temperature float32 `json:"temperature"`
}

type Config struct {
	Value ModelParameters `key:"OPENAI_MODEL_PARAMS"`
}
```

It also works with pointers
```go
type ModelParameters struct {
	Temperature float32 `json:"temperature"`
}

type Config struct {
	Value *ModelParameters `key:"OPENAI_MODEL_PARAMS"`
}
```

### Error Handling

See the LogError function in `errors.go` for an example of logging errors from the Load function to a
structured log.

## Running Tests

```bash
go test -v
```

## License

MIT
