# JSON Deserialization

goconfigtools can deserialize JSON strings from environment variables into Go structs and maps.

## Table of Contents

- [Map Deserialization](#map-deserialization)
- [Struct Deserialization](#struct-deserialization)
- [Pointer Types](#pointer-types)
- [Nested JSON](#nested-json)
- [Error Handling](#error-handling)

## Map Deserialization

Convert JSON strings to `map[string]interface{}`:

```go
type Config struct {
    ModelParams map[string]interface{} `key:"OPENAI_MODEL_PARAMS"`
}

func main() {
    var config Config

    // Set in environment:
    // export OPENAI_MODEL_PARAMS='{"temperature":0.7,"max_tokens":1000}'

    if err := goconfigtools.Load(&config); err != nil {
        log.Fatalf("Configuration error: %v", err)
    }

    // Access map values
    if temp, ok := config.ModelParams["temperature"].(float64); ok {
        fmt.Printf("Temperature: %.1f\n", temp)
    }
}
```

### Accessing Map Values

Since `map[string]interface{}` values are of type `interface{}`, you need type assertions:

```go
params := config.ModelParams

// Number (always float64 in JSON)
temperature := params["temperature"].(float64)

// String
model := params["model"].(string)

// Boolean
stream := params["stream"].(bool)

// Nested object
if metadata, ok := params["metadata"].(map[string]interface{}); ok {
    version := metadata["version"].(string)
}

// Array
if tags, ok := params["tags"].([]interface{}); ok {
    for _, tag := range tags {
        fmt.Println(tag.(string))
    }
}
```

## Struct Deserialization

Convert JSON strings to typed structs for better type safety:

```go
type ModelParameters struct {
    Temperature float64 `json:"temperature"`
    MaxTokens   int     `json:"max_tokens"`
    TopP        float64 `json:"top_p"`
    Model       string  `json:"model"`
}

type Config struct {
    ModelParams ModelParameters `key:"OPENAI_MODEL_PARAMS"`
}

func main() {
    var config Config

    // Set in environment:
    // export OPENAI_MODEL_PARAMS='{"temperature":0.7,"max_tokens":1000,"top_p":0.9,"model":"gpt-4"}'

    if err := goconfigtools.Load(&config); err != nil {
        log.Fatalf("Configuration error: %v", err)
    }

    // Type-safe access
    fmt.Printf("Temperature: %.1f\n", config.ModelParams.Temperature)
    fmt.Printf("Max Tokens: %d\n", config.ModelParams.MaxTokens)
}
```

### Struct Tag Options

Use standard `json` tags to control deserialization:

```go
type ModelParameters struct {
    // Rename field
    Temperature float64 `json:"temperature"`

    // Optional field (omitempty on marshal)
    MaxTokens int `json:"max_tokens,omitempty"`

    // Ignore field
    Internal string `json:"-"`

    // Default JSON field name (uses field name "TopP")
    TopP float64
}
```

## Pointer Types

Both maps and structs work with pointer types:

### Pointer to Map

```go
type Config struct {
    ModelParams *map[string]interface{} `key:"OPENAI_MODEL_PARAMS"`
}

func main() {
    var config Config

    if err := goconfigtools.Load(&config); err != nil {
        log.Fatalf("Configuration error: %v", err)
    }

    // Check if set
    if config.ModelParams != nil {
        temp := (*config.ModelParams)["temperature"].(float64)
        fmt.Printf("Temperature: %.1f\n", temp)
    }
}
```

### Pointer to Struct

```go
type ModelParameters struct {
    Temperature float64 `json:"temperature"`
    MaxTokens   int     `json:"max_tokens"`
}

type Config struct {
    ModelParams *ModelParameters `key:"OPENAI_MODEL_PARAMS"`
}

func main() {
    var config Config

    if err := goconfigtools.Load(&config); err != nil {
        log.Fatalf("Configuration error: %v", err)
    }

    // Check if set
    if config.ModelParams != nil {
        fmt.Printf("Temperature: %.1f\n", config.ModelParams.Temperature)
    }
}
```

## Nested JSON

JSON can contain nested objects and arrays:

```go
type DatabaseConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Username string `json:"username"`
    Password string `json:"password"`
}

type RedisConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

type ServiceConfig struct {
    Database DatabaseConfig `json:"database"`
    Redis    RedisConfig    `json:"redis"`
    Replicas []string       `json:"replicas"`
}

type Config struct {
    Services ServiceConfig `key:"SERVICES_CONFIG"`
}

func main() {
    var config Config

    // Set in environment:
    // export SERVICES_CONFIG='{
    //   "database": {"host": "db.example.com", "port": 5432, "username": "app", "password": "secret"},
    //   "redis": {"host": "redis.example.com", "port": 6379},
    //   "replicas": ["replica1.example.com", "replica2.example.com"]
    // }'

    if err := goconfigtools.Load(&config); err != nil {
        log.Fatalf("Configuration error: %v", err)
    }

    fmt.Printf("Database: %s:%d\n",
        config.Services.Database.Host,
        config.Services.Database.Port)

    for _, replica := range config.Services.Replicas {
        fmt.Printf("Replica: %s\n", replica)
    }
}
```

## Default Values with JSON

You can provide default JSON values:

```go
type Config struct {
    ModelParams ModelParameters `key:"OPENAI_MODEL_PARAMS" default:"{\"temperature\":0.7,\"max_tokens\":1000}"`
}
```

**Important:** Remember to escape quotes in the default tag value.

## Error Handling

JSON deserialization errors are reported with context:

```go
err := goconfigtools.Load(&config)
if err != nil {
    // Error will indicate which field had invalid JSON
    // Example: "invalid value for OPENAI_MODEL_PARAMS: invalid character '}' looking for beginning of value"
    log.Fatalf("Configuration error: %v", err)
}
```

Common JSON errors:
- **Syntax errors** - Invalid JSON format
- **Type mismatches** - JSON field type doesn't match struct field type
- **Missing required fields** - JSON doesn't include fields without `omitempty`

## Best Practices

1. **Use structs for known structure** - Better type safety and IDE support
2. **Use maps for dynamic configuration** - When structure isn't known at compile time
3. **Validate JSON types** - Ensure JSON field types match struct field types
4. **Use pointers for optional JSON** - Distinguish between "not set" and "set to zero value"
5. **Escape quotes in defaults** - Remember to escape quotes in `default` tags
6. **Consider external files for complex JSON** - For very large JSON, consider loading from files instead

## Example: Feature Flags

JSON is great for feature flag configuration:

```go
type FeatureFlags struct {
    EnableBetaFeatures bool     `json:"enable_beta_features"`
    AllowedUsers       []string `json:"allowed_users"`
    RolloutPercentage  int      `json:"rollout_percentage"`
}

type Config struct {
    Features FeatureFlags `key:"FEATURE_FLAGS" default:"{\"enable_beta_features\":false,\"allowed_users\":[],\"rollout_percentage\":0}"`
}

func main() {
    var config Config

    // export FEATURE_FLAGS='{"enable_beta_features":true,"allowed_users":["alice","bob"],"rollout_percentage":25}'

    if err := goconfigtools.Load(&config); err != nil {
        log.Fatalf("Configuration error: %v", err)
    }

    if config.Features.EnableBetaFeatures {
        fmt.Println("Beta features enabled")
        fmt.Printf("Allowed users: %v\n", config.Features.AllowedUsers)
        fmt.Printf("Rollout: %d%%\n", config.Features.RolloutPercentage)
    }
}
```

## Combining with Other Features

JSON fields work with all other goconfigtools features:

```go
type Config struct {
    // Required JSON field
    ModelParams ModelParameters `key:"OPENAI_MODEL_PARAMS" required:"true"`

    // JSON with default value
    Features FeatureFlags `key:"FEATURE_FLAGS" default:"{\"enabled\":false}"`

    // Optional JSON (pointer)
    Advanced *AdvancedConfig `key:"ADVANCED_CONFIG"`
}
```

Note: Validation with `min`, `max`, and `pattern` tags applies to the JSON string itself, not the deserialized values. Use custom validators for validating deserialized JSON content.
