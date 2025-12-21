# Defaulting and Required Fields

This guide explains how goconfig handles default values and required fields.

## Table of Contents

- [Default Values](#default-values)
- [Required Fields](#required-fields)
- [Defaulting Behavior](#defaulting-behavior)
- [Sentinel Errors](#sentinel-errors)

## Default Values

Use the `default` struct tag to specify a fallback value when an environment variable is not set:

```go
type Config struct {
    Port     int    `key:"PORT" default:"8080"`
    Host     string `key:"HOST" default:"localhost"`
    Debug    bool   `key:"DEBUG" default:"false"`
    Timeout  time.Duration `key:"TIMEOUT" default:"30s"`
}
```

If `PORT` is not set in the environment, it will default to `8080`.

### Default Values with Validation

Defaults are validated just like user-provided values:

```go
type Config struct {
    // Default value must satisfy min/max constraints
    Port int `key:"PORT" default:"8080" min:"1024" max:"65535"`
}
```

If you specify a default that violates validation rules, you'll get an error during `Load()`.

## Required Fields

There are three ways to mark fields as required:

### 1. No Tags (Optional)

Without `required` or `keyRequired`, the field is optional:

```go
type Config struct {
    OptionalField string `key:"OPTIONAL_FIELD"`
}
```

If `OPTIONAL_FIELD` is not set, the field retains its zero value (empty string for strings).

### 2. `required:"true"` - Non-Empty Value Required

The field must be present AND have a non-empty value:

```go
type Config struct {
    APIKey string `key:"API_KEY" required:"true"`
}
```

This will error if:
- `API_KEY` is not set (unset)
- `API_KEY` is set to empty string (`export API_KEY=`)

### 3. `keyRequired:"true"` - Key Must Be Present

The field must be present in the environment, but can be empty:

```go
type Config struct {
    Password string `key:"PASSWORD" keyRequired:"true"`
}
```

This will error if:
- `PASSWORD` is not set (unset)

But allows:
- `PASSWORD` set to empty string (`export PASSWORD=`)

This is useful when you want to explicitly distinguish between "not configured" and "configured as empty".

## Defaulting Behavior

The interaction between defaults, required tags, and environment variables follows this logic:

| Environment State | No Required Tags | `keyRequired:"true"` | `required:"true"` |
|-------------------|------------------|----------------------|-------------------|
| `export FOO=bar` | Value becomes "bar" | Value becomes "bar" | Value becomes "bar" |
| `export FOO=` | Value becomes "" (empty) | Value becomes "" (empty) | **Error**: ErrMissingValue |
| `unset FOO` | Value remains unchanged | **Error**: ErrMissingConfigKey | **Error**: ErrMissingConfigKey |

### With Default Tags

When a `default` tag is present:

| Environment State | Behavior |
|-------------------|----------|
| `export FOO=bar` | Value becomes "bar" (environment overrides default) |
| `export FOO=` | Value becomes "" (empty environment overrides default) |
| `unset FOO` | Value becomes the default value |

**Important notes about `default` tags:**

* When a key is **unset**, the `default` value is used. This satisfies both `keyRequired="true"` and `required="true"`.
* When a key is **set to empty** (e.g., `export FOO=`), it overrides the default and the empty value is used. This will cause an error if `required="true"`.
* Defaults are only applied when the key is completely absent from the key store.

### Examples

#### Example 1: Optional Field with Default

```go
type Config struct {
    Port int `key:"PORT" default:"8080"`
}
```

- `PORT` not set → `Port = 8080`
- `PORT=9000` → `Port = 9000`
- `PORT=` → Error (invalid integer)

#### Example 2: Required Field with Default

```go
type Config struct {
    APIKey string `key:"API_KEY" default:"dev-key" required:"true"`
}
```

- `API_KEY` not set → `APIKey = "dev-key"` ✓ (default satisfies required)
- `API_KEY=sk-prod` → `APIKey = "sk-prod"` ✓
- `API_KEY=` → Error: ErrMissingValue (empty value not allowed)

#### Example 3: keyRequired vs required

```go
type Config struct {
    Password    string `key:"PASSWORD" keyRequired:"true"`
    Username    string `key:"USERNAME" required:"true"`
}
```

```bash
# This is valid
export PASSWORD=
export USERNAME=admin

# This is invalid - PASSWORD must be present
unset PASSWORD
export USERNAME=admin

# This is invalid - USERNAME cannot be empty
export PASSWORD=
export USERNAME=
```

## Sentinel Errors

goconfig provides two sentinel errors for missing configuration:

```go
import "errors"

// ErrMissingConfigKey - key not found in key store
var ErrMissingConfigKey = errors.New("missing configuration key")

// ErrMissingValue - key found but value is empty when required="true"
var ErrMissingValue = errors.New("missing value")
```

You can check for these errors using `errors.Is()`:

```go
err := goconfig.Load(&config)
if err != nil {
    if errors.Is(err, goconfig.ErrMissingConfigKey) {
        // Handle missing key
        log.Println("Required environment variable not set")
    }
    if errors.Is(err, goconfig.ErrMissingValue) {
        // Handle empty value
        log.Println("Required environment variable is empty")
    }
}
```

## Pre-Initialized Defaults

You can pre-initialize struct fields instead of using `default` tags:

```go
type Config struct {
    Port int    `key:"PORT"`
    Host string `key:"HOST"`
}

func main() {
    // Initialize with defaults
    cfg := Config{
        Port: 8080,
        Host: "localhost",
    }

    // Load will only override fields that are present in environment
    if err := goconfig.Load(&cfg); err != nil {
        log.Fatalf("Configuration error: %v", err)
    }

    // If PORT and HOST are not set, they remain 8080 and "localhost"
}
```

This approach gives you more flexibility for complex default values that can't be expressed as strings.

## Best Practices

1. **Use `required:"true"` for critical configuration** - API keys, database URLs, etc.
2. **Use `default` for sensible defaults** - Port numbers, timeouts, feature flags
3. **Use `keyRequired:"true"` sparingly** - Only when you need to distinguish "not set" from "empty"
4. **Pre-initialize for complex defaults** - When default values can't be expressed as strings
5. **Document your defaults** - Make it clear what the default behavior is

## Examples

See these examples for practical demonstrations:
- [Simple example](../example/simple) - Basic usage with defaults
- [Validation example](../example/validation) - Defaults with validation
