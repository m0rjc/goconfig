package goconfig

import (
	"context"
	"os"
)

// KeyStore reads string values given keys.
// Return the value if present (it may be empty), an indication of whether it is present or an error if there was
// an error accessing the store.
type KeyStore func(ctx context.Context, key string) (string, bool, error)

// EnvironmentKeyStore is a key store that reads values from environment variables
func EnvironmentKeyStore(_ context.Context, key string) (string, bool, error) {
	value, present := os.LookupEnv(key)
	return value, present, nil
}

// CompositeStore tries each store in turn until one returns a value or an error.
func CompositeStore(stores ...KeyStore) KeyStore {
	return func(ctx context.Context, key string) (string, bool, error) {
		for _, store := range stores {
			value, present, err := store(ctx, key)
			if present || err != nil {
				return value, present, err
			}
		}
		return "", false, nil
	}
}
