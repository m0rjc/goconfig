# goconfigtools

A simple Go library for loading configuration from environment variables using struct tags.

## Features

- Load configuration from environment variables using struct tags
- Support for nested structs
- Optional default values
- Type conversion for common types: `string`, `bool`, `int`, `uint`, `float`, `time.Duration`
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

## Supported Types

- `string`
- `bool`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `time.Duration` (uses Go's duration format: "30s", "1m", "1h", etc.)

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

## Running Tests

```bash
go test -v
```

## License

MIT
