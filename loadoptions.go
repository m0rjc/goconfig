package goconfig

import (
	"reflect"

	"github.com/m0rjc/goconfig/internal/readpipeline"
)

// Option is a functional option for configuring the Load function.
type Option func(*loadOptions)

// WithKeyStore replaces the environment variable keystore with an alternative.
// Use this to read from other sources such as a database or properties file.
func WithKeyStore(keyStore KeyStore) Option {
	return func(opts *loadOptions) {
		opts.keyStore = keyStore
	}
}

// WithCustomType registers a custom type handler for a given type.
func WithCustomType[T any](handler TypedHandler[T]) Option {
	var typedNil *T
	t := reflect.TypeOf(typedNil).Elem()

	return func(opts *loadOptions) {
		opts.typeRegistry.RegisterType(t, readpipeline.WrapTypedHandler(handler))
	}
}

// loadOptions holds the configuration options for Load.
type loadOptions struct {
	// keyStore reads the values. Default to os.GetEnv()
	keyStore KeyStore
	// typeRegistry holds the handlers for specific types
	typeRegistry readpipeline.TypeRegistry
}

// newLoadOptions creates default load options.
func newLoadOptions() *loadOptions {
	return &loadOptions{
		keyStore:     EnvironmentKeyStore,
		typeRegistry: readpipeline.NewTypeRegistry(),
	}
}

// applyOptions applies the given options to the load options.
func (opts *loadOptions) applyOptions(options []Option) {
	for _, opt := range options {
		opt(opts)
	}
}
