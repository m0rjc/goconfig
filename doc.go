// Package goconfig provides a simple way to load configuration from environment
// variables using struct tags.
//
// # Features
//
//   - Load configuration from environment variables using struct tags
//   - Support for nested structs
//   - Optional default values
//   - Type conversion for common types: string, bool, int, uint, float, time.Duration
//   - Ability to convert JSON strings into maps or JSON annotated structs
//   - Built-in min, max, pattern validators plus support for custom validation
//   - Support for custom field parsing
//   - Support for a custom key-value store as an alternative to environment variables
//   - Clear error messages for missing required fields or invalid values
//
// # Basic Usage
//
// Define your configuration struct with 'key' and optional 'default' tags:
//
//	type Config struct {
//	    APIKey  string        `key:"API_KEY"`                                      // Required
//	    Model   string        `key:"MODEL" default:"gpt-4"`                        // Optional with default
//	    Port    int           `key:"PORT" default:"8080" min:"1024" max:"65535"`  // With min/max validation
//	    Timeout time.Duration `key:"TIMEOUT" default:"30s" min:"1s" max:"5m"`     // Duration with validation
//	}
//
//	func main() {
//	    var config Config
//	    if err := goconfig.Load(context.Background(), &config); err != nil {
//	        log.Fatalf("Failed to load configuration: %v", err)
//	    }
//	    // Port is guaranteed to be between 1024 and 65535
//	    // Timeout is guaranteed to be between 1s and 5m
//	}
//
// # Struct Tags
//
//   - key: The environment variable name to read from (required)
//   - default: The default value to use if the environment variable is not set (optional)
//   - min: Minimum value for numeric types (optional)
//   - max: Maximum value for numeric types (optional)
//   - pattern: Regular expression for string types (optional)
//   - required: Set to "true" to require the field to not be empty (optional)
//   - keyRequired: Set to "true" to require the field to be present, though it can be explicitly blank
//
// # Supported Types
//
//   - string
//   - bool
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64
//   - time.Duration (uses Go's duration format: "30s", "1m", "1h", etc.)
//   - map[string]interface{} using JSON deserialisation
//   - struct using JSON deserialisation
//   - pointers to the above
//
// # Custom Validation
//
// Use the WithValidator option to add custom validation logic:
//
//	err := goconfig.Load(ctx, &cfg,
//	    goconfig.WithValidator("APIKey", func(value any) error {
//	        key := value.(string)
//	        if !strings.HasPrefix(key, "sk-") {
//	            return fmt.Errorf("API key must start with 'sk-'")
//	        }
//	        return nil
//	    }),
//	)
//
// # Custom Parsers
//
// Use the WithParser option to provide custom parsing logic for specific fields:
//
//	err := goconfig.Load(ctx, &cfg,
//	    goconfig.WithParser("SpecialField", func(value string) (any, error) {
//	        // Custom parsing logic
//	        return customParse(value)
//	    }),
//	)
//
// # Custom Key Stores
//
// By default, the package reads from environment variables, but you can provide
// a custom key-value store:
//
//	myStore := func(ctx context.Context, key string) (string, bool, error) {
//	    // Custom logic to retrieve configuration values
//	    return value, found, nil
//	}
//
//	err := goconfig.Load(ctx, &config, goconfig.WithKeyStore(myStore))
//
// You can also chain multiple key stores using CompositeStore, which tries each
// store in order until one returns a value:
//
//	store := goconfig.CompositeStore(
//	    customStore,
//	    goconfig.EnvironmentKeyStore,
//	)
//	err := goconfig.Load(ctx, &config, goconfig.WithKeyStore(store))
//
// # Error Handling
//
// The package provides two sentinel errors for common cases:
//
//   - ErrMissingConfigKey: returned when a required key is not found in the key store
//   - ErrMissingValue: returned when a key is found but has a blank value when required="true"
//
// When multiple configuration errors occur, they are collected into a ConfigErrors
// type, which implements error and provides an Unwrap method for Go 1.20+ error inspection:
//
//	err := goconfig.Load(ctx, &config)
//	if err != nil {
//	    var configErrs *goconfig.ConfigErrors
//	    if errors.As(err, &configErrs) {
//	        for _, e := range configErrs.Unwrap() {
//	            if errors.Is(e, goconfig.ErrMissingConfigKey) {
//	                // Handle missing key
//	            }
//	        }
//	    }
//	}
//
// Error logging to structured logs is supported.
//
// # Documentation
//
// For detailed guides and examples, see:
//
//   - https://github.com/m0rjc/goconfig/tree/main/docs - Full documentation
//   - https://github.com/m0rjc/goconfig/tree/main/docs/validation.md - Validation guide
//   - https://github.com/m0rjc/goconfig/tree/main/docs/defaulting.md - Defaulting behavior
//   - https://github.com/m0rjc/goconfig/tree/main/docs/json.md - JSON deserialization
//   - https://github.com/m0rjc/goconfig/tree/main/docs/advanced.md - Advanced features
//   - https://github.com/m0rjc/goconfig/tree/main/example - Working examples
package goconfig
