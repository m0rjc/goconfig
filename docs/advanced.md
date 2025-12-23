# Advanced Features

This guide covers advanced features for extending goconfig with custom behavior.

## Table of Contents

- [Custom Types](#custom-types)
- [Custom Key Stores](#custom-key-stores)
- [Composite Key Stores](#composite-key-stores)
- [Error Handling](#error-handling)

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

### Use Cases for Custom Types

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

#### Parsing Custom Time Formats

```go
type Timestamp time.Time

type Config struct {
    Timestamp Timestamp `key:"TIMESTAMP"`
}

func main() {
    var cfg Config

    timestampHandler := goconfig.NewCustomHandler(
        func(rawValue string) (Timestamp, error) {
            // Parse RFC3339 format
            t, err := time.Parse(time.RFC3339, rawValue)
            if err != nil {
                return Timestamp{}, fmt.Errorf("invalid timestamp format: %w", err)
            }
            return Timestamp(t), nil
        },
    )

    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithCustomType[Timestamp](timestampHandler),
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

#### Parsing Binary Data

```go
type EncryptionKey []byte

type Config struct {
    EncryptionKey EncryptionKey `key:"ENCRYPTION_KEY"`
}

func main() {
    var cfg Config

    keyHandler := goconfig.NewCustomHandler(
        func(rawValue string) (EncryptionKey, error) {
            // Decode base64-encoded key
            key, err := base64.StdEncoding.DecodeString(rawValue)
            if err != nil {
                return nil, fmt.Errorf("invalid base64: %w", err)
            }
            if len(key) != 32 {
                return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
            }
            return EncryptionKey(key), nil
        },
    )

    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithCustomType[EncryptionKey](keyHandler),
    )
}
```

### Parser Error Handling

Custom type parsers should return descriptive errors:

```go
customHandler := goconfig.NewCustomHandler(
    func(rawValue string) (MyType, error) {
        // Good: Descriptive error
        return MyType{}, fmt.Errorf("invalid format: expected 'key=value', got '%s'", rawValue)

        // Bad: Generic error
        return MyType{}, fmt.Errorf("parse error")
    },
)
```

The error will be wrapped with field context automatically:
```
invalid value for FIELD: invalid format: expected 'key=value', got 'invalid-input'
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

### Reading from a File

```go
func fileKeyStore(filename string) goconfig.KeyStore {
    return func(ctx context.Context, key string) (string, bool, error) {
        data, err := os.ReadFile(filename)
        if err != nil {
            return "", false, fmt.Errorf("failed to read config file: %w", err)
        }

        // Parse simple KEY=VALUE format
        lines := strings.Split(string(data), "\n")
        for _, line := range lines {
            line = strings.TrimSpace(line)
            if line == "" || strings.HasPrefix(line, "#") {
                continue
            }

            parts := strings.SplitN(line, "=", 2)
            if len(parts) == 2 && parts[0] == key {
                return parts[1], true, nil
            }
        }

        return "", false, nil
    }
}

func main() {
    var cfg Config

    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithKeyStore(fileKeyStore("/etc/myapp/config")),
    )
}
```

### Reading from AWS Secrets Manager

```go
func awsSecretsKeyStore(secretsClient *secretsmanager.Client, secretName string) goconfig.KeyStore {
    return func(ctx context.Context, key string) (string, bool, error) {
        result, err := secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
            SecretId: aws.String(secretName),
        })
        if err != nil {
            return "", false, fmt.Errorf("failed to get secret: %w", err)
        }

        // Parse JSON secret
        var secrets map[string]string
        if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
            return "", false, fmt.Errorf("failed to parse secret: %w", err)
        }

        value, found := secrets[key]
        return value, found, nil
    }
}
```

### Reading from HashiCorp Vault

```go
func vaultKeyStore(client *vault.Client, path string) goconfig.KeyStore {
    return func(ctx context.Context, key string) (string, bool, error) {
        secret, err := client.KVv2("secret").Get(ctx, path)
        if err != nil {
            return "", false, fmt.Errorf("failed to read from vault: %w", err)
        }

        value, found := secret.Data[key].(string)
        if !found {
            return "", false, nil
        }

        return value, true, nil
    }
}
```

### In-Memory Key Store (for testing)

```go
func mapKeyStore(data map[string]string) goconfig.KeyStore {
    return func(ctx context.Context, key string) (string, bool, error) {
        value, found := data[key]
        return value, found, nil
    }
}

func TestConfig(t *testing.T) {
    var cfg Config

    testData := map[string]string{
        "PORT":    "8080",
        "HOST":    "localhost",
        "API_KEY": "sk-test-key-12345678901234",
    }

    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithKeyStore(mapKeyStore(testData)),
    )

    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }
}
```

## Composite Key Stores

`CompositeStore` chains multiple key stores together, trying each in order until one returns a value.

### Environment Variables with File Fallback

```go
func main() {
    var cfg Config

    // Try environment first, then fall back to config file
    store := goconfig.CompositeStore(
        goconfig.EnvironmentKeyStore,
        fileKeyStore("/etc/myapp/config"),
    )

    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithKeyStore(store),
    )
}
```

### Secrets Manager with Environment Variable Override

```go
func main() {
    var cfg Config

    // Environment variables can override secrets manager
    store := goconfig.CompositeStore(
        goconfig.EnvironmentKeyStore,
        awsSecretsKeyStore(secretsClient, "prod/myapp"),
    )

    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithKeyStore(store),
    )
}
```

### Multi-Stage Fallback

```go
func main() {
    var cfg Config

    // Try multiple sources in order:
    // 1. Environment variables (highest priority)
    // 2. Vault secrets
    // 3. Local config file
    // 4. Default values in struct tags (automatic fallback)
    store := goconfig.CompositeStore(
        goconfig.EnvironmentKeyStore,
        vaultKeyStore(vaultClient, "secret/myapp"),
        fileKeyStore("/etc/myapp/config"),
    )

    err := goconfig.Load(context.Background(), &cfg,
        goconfig.WithKeyStore(store),
    )
}
```

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

See `LogError` function in `errors.go` for an example of logging configuration errors to structured logs:

```go
func LogError(logger *slog.Logger, err error) {
    var configErrs *ConfigErrors
    if errors.As(err, &configErrs) {
        for _, e := range configErrs.Unwrap() {
            var fieldErr *FieldError
            if errors.As(e, &fieldErr) {
                logger.Error("configuration error",
                    "field", fieldErr.Field,
                    "key", fieldErr.Key,
                    "error", fieldErr.Err,
                )
            }
        }
    }
}
```

### Checking Specific Error Types

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

## Best Practices

1. **Use custom types for domain validation** - Create custom types (e.g., `APIKey`, `Email`) instead of using raw strings for fields that need validation
2. **Leverage type safety** - Use the type system's generics for compile-time type safety
3. **Reuse type handlers** - Define handlers once and use them across multiple fields of the same type
4. **Use CompositeStore for flexibility** - Allow environment overrides in production
5. **Cache key store results** - Avoid repeated API calls for the same key
6. **Handle context cancellation** - Respect context timeouts in custom key stores
7. **Return descriptive errors** - Help users understand what went wrong
8. **Test with in-memory stores** - Use `mapKeyStore` for unit tests
9. **Fail fast on key store errors** - Don't silently ignore lookup failures
10. **Use `NewEnumHandler` for enums** - Automatic validation for string-based enum types
