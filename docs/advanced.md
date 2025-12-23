# Advanced Features

This guide covers advanced features for extending goconfig with custom behavior.

## Table of Contents

- [Custom Types](#custom-types)
- [Custom Key Stores](#custom-key-stores)
- [Composite Key Stores](#composite-key-stores)
- [Error Handling and Structured Logging](#error-handling)

## Custom Types

Custom types allow you to define parsing and validation logic for specific types that need special handling beyond the built-in type conversions. The type system uses Go generics for type safety.

### Basic Custom Type

```go
type Config struct {
    // Standard field
    Port int `key:"PORT" default:"8080"`

    // Custom parsed field
    SpecialValue CustomType `key:"SPECIAL_VALUE"`
}

type CustomType struct {
    Value string
    Metadata map[string]string
}

func main() {
    var cfg Config

    customHandler := goconfig.NewCustomHandler(
        func(rawValue string) (CustomType, error) {
            // Parse the value however you need
            parts := strings.Split(rawValue, ":")
            if len(parts) != 2 {
                return CustomType{}, fmt.Errorf("invalid format, expected key:value")
            }

            return CustomType{
                Value: parts[1],
                Metadata: map[string]string{
                    "key": parts[0],
                },
            }, nil
        },
    )

    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithCustomType[CustomType](customHandler),
    )

    if err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
}
```

### Example Use Cases for Custom Types

#### Parsing URLs

```go
type DatabaseURL url.URL

type Config struct {
    DatabaseURL DatabaseURL `key:"DATABASE_URL"`
}

func main() {
    var cfg Config

    urlHandler := goconfig.NewCustomHandler(
        func(rawValue string) (DatabaseURL, error) {
            parsedURL, err := url.Parse(rawValue)
            if err != nil {
                return DatabaseURL{}, fmt.Errorf("invalid URL: %w", err)
            }
            if parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql" {
                return DatabaseURL{}, fmt.Errorf("unsupported database scheme: %s", parsedURL.Scheme)
            }
            return DatabaseURL(*parsedURL), nil
        },
    )

    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithCustomType[DatabaseURL](urlHandler),
    )
}
```

#### Parsing Lists with Custom Delimiters

```go
type HostList []string

type Config struct {
    AllowedHosts HostList `key:"ALLOWED_HOSTS"`
}

func main() {
    var cfg Config

    hostListHandler := goconfig.NewCustomHandler(
        func(rawValue string) (HostList, error) {
            // Split on semicolons instead of commas
            hosts := strings.Split(rawValue, ";")

            // Trim whitespace
            for i, host := range hosts {
                hosts[i] = strings.TrimSpace(host)
            }

            return HostList(hosts), nil
        },
    )

    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithCustomType[HostList](hostListHandler),
    )

    // export ALLOWED_HOSTS="example.com; api.example.com; www.example.com"
}
```


### Parser Error Handling

Custom type parsers should return descriptive errors. For security best practice you should avoid returning the input value.
This avoids the like of `Failed to parse value 'Top Secret' for field 'API_KEY` appearing in logs.

```go
customHandler := goconfig.NewCustomHandler(
    func(rawValue string) (MyType, error) {
        // Good: Descriptive error
        return MyType{}, fmt.Errorf("invalid format: expected 'key=value'")

        // Bad: Generic error
        return MyType{}, fmt.Errorf("parse error")

        // Bad: Potentially leaving secrets in logs
        return MyType{}, fmt.Errorf("parse error: expected 'key=value', got '%s'", rawValue)
    },
)
```

The error will be wrapped with field context automatically:
```
invalid value for FIELD: invalid format: expected 'key=value'
```

## Custom Key Stores

By default, goconfig reads from environment variables using `os.LookupEnv`. You can provide custom key stores to read from other sources.

### Key Store Function Signature

```go
type KeyStore func(ctx context.Context, key string) (value string, found bool, error error)
```

- **ctx**: Context for cancellation and timeouts
- **key**: The key to look up (e.g., "DATABASE_URL")
- **value**: The value if found
- **found**: Whether the key was found (distinguishes "not found" from "found but empty")
- **error**: Any error that occurred during lookup

### Composite Key Stores

`CompositeStore` chains multiple key stores together, trying each in order until one returns a value.

## Error Handling

### ConfigErrors Type

When multiple configuration errors occur, they are collected into a `ConfigErrors` type:

```go
err := goconfig.Load(context.Background(), &config)
if err != nil {
    var configErrs *goconfig.ConfigErrors
    if errors.As(err, &configErrs) {
        // Multiple errors occurred
        for _, e := range configErrs.Unwrap() {
            fmt.Printf("Error: %v\n", e)
        }
    } else {
        // Single error
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Structured Logging

goconfig provides helper functions for structured logging with `slog`:

```go
logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

err := goconfig.Load(context.Background(), &config)
if err != nil {
    // Log all the errors with a customized log message
    goconfig.LogError(logger, err, goconfig.WithLogMessage("config_validation_failed"))
}
```

Alternatively, use `ConfigErrors.LogAll()` for more control:

```go
if configErrs, ok := err.(*goconfig.ConfigErrors); ok {
    configErrs.LogAll(logger, goconfig.WithLogMessage("configuration_error"))
}
```

### Checking Specific Error Types

The goconfig.ConfigErrors type provides the `Unwrap()` method, so implementing the `errors.Is`, `errors.As` contract.

```go
err := goconfig.Load(context.Background(), &config)
if err != nil {
    // Check for missing key
    if errors.Is(err, goconfig.ErrMissingConfigKey) {
        log.Println("Required environment variable not set")
    }

    // Check for missing value
    if errors.Is(err, goconfig.ErrMissingValue) {
        log.Println("Required environment variable is empty")
    }

    // Check for field-specific errors
    var fieldErr *goconfig.FieldError
    if errors.As(err, &fieldErr) {
        log.Printf("Error in field %s (key %s): %v",
            fieldErr.Field, fieldErr.Key, fieldErr.Err)
    }
}
```

## Combining Advanced Features

You can combine multiple advanced features:

```go
func main() {
    var cfg Config

    err := goconfig.Load(context.Background(), &cfg,
        // Custom key store
        goconfig.WithKeyStore(goconfig.CompositeStore(
            goconfig.EnvironmentKeyStore,
            vaultKeyStore(vaultClient, "secret/myapp"),
        )),

        // Custom types
        goconfig.WithCustomType[DatabaseURL](urlHandler),
        goconfig.WithCustomType[EncryptionKey](keyHandler),
        goconfig.WithCustomType[APIKey](apiKeyHandler),
        goconfig.WithCustomType[DatabaseHost](hostHandler),
    )

    if err != nil {
        goconfig.LogError(logger, err)
        os.Exit(1)
    }
}
```

