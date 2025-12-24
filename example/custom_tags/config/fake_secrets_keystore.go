package config

import (
	"context"
	"os"
)

// fakeSecretsKeyStore is a fake keystore that reads secrets from environment variables.
// The environment variables must be prefixed with SECRET_.
// A production keystore could read secrets from a secure store, though I've just had Kubernetes inject them into
// the environment and used the normal keystore.
func fakeSecretsKeyStore(ctx context.Context, key string) (string, bool, error) {
	value, ok := os.LookupEnv("SECRET_" + key)
	return value, ok, nil
}
