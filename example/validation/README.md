# Validation Example

This example demonstrates comprehensive validation features in goconfig, including:

- **Min/Max validation** for integers, floats, and durations
- **Pattern validation** for strings using regular expressions
- **Custom validators** for complex business logic
- **Nested struct validation** using dot notation

## Running the Example

### With Valid Configuration

```bash
export API_KEY="sk-1234567890abcdefghij"
export SERVER_PORT="8080"
export DB_USER="appuser"
go run .
```

### Testing Validation Errors

#### Port out of range (below minimum)
```bash
export SERVER_PORT="80"  # Error: below minimum 1024
export API_KEY="sk-1234567890abcdefghij"
go run .
```

#### Port out of range (above maximum)
```bash
export SERVER_PORT="70000"  # Error: above maximum 65535
export API_KEY="sk-1234567890abcdefghij"
go run .
```

#### Invalid duration (below minimum)
```bash
export READ_TIMEOUT="100ms"  # Error: below minimum 1s
export API_KEY="sk-1234567890abcdefghij"
go run .
```

#### Invalid pattern (username with special characters)
```bash
export DB_USER="user@host"  # Error: doesn't match pattern ^[a-zA-Z0-9_]+$
export API_KEY="sk-1234567890abcdefghij"
go run .
```

#### Custom validation (API key too short)
```bash
export API_KEY="sk-short"  # Error: must be at least 20 characters
go run .
```

#### Custom validation (API key wrong prefix)
```bash
export API_KEY="pk-1234567890abcdefghij"  # Error: must start with 'sk-'
go run .
```

## Features Demonstrated

### Integer Range Validation
```go
Port int `key:"SERVER_PORT" default:"8080" min:"1024" max:"65535"`
```

### Float Range Validation
```go
LoadFactor float64 `key:"LOAD_FACTOR" default:"0.75" min:"0.0" max:"1.0"`
```

### Duration Range Validation
```go
ReadTimeout time.Duration `key:"READ_TIMEOUT" default:"30s" min:"1s" max:"5m"`
```

### String Pattern Validation
```go
Hostname string `key:"HOSTNAME" default:"localhost" pattern:"^[a-zA-Z0-9.-]+$"`
```

### Custom Validators
```go
goconfig.WithValidator("API.APIKey", func(value any) error {
    key := value.(string)
    if !strings.HasPrefix(key, "sk-") {
        return fmt.Errorf("API key must start with 'sk-'")
    }
    return nil
})
```

### Nested Field Validation
```go
// Validates the Host field within the Database struct
goconfig.WithValidator("Database.Host", func(value any) error {
    // validation logic
    return nil
})
```

## Validation Order

Validations are applied in this order:
1. Tag-based validation (`min`, `max`, `pattern`)
2. Custom validators (in the order they are registered)

All validations must pass for the configuration to be considered valid.
