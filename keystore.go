package goconfigtools

import "os"

// KeyStore reads string values given keys.
type KeyStore func(key string) (string, error)

// EnvironmentKeyStore is a key store that reads values from environment variables
func EnvironmentKeyStore(key string) (string, error) {
	return os.Getenv(key), nil
}
