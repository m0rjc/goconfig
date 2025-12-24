package goconfig

import (
	"bufio"
	"context"
	"os"
	"strings"
)

// NewEnvFileKeyStore returns a KeyStore that reads values from a list of environment files.
// If no filenames are provided, it defaults to ".env".
// Files are processed in the order they are provided. If multiple files contain the same key,
// the first one encountered wins.
func NewEnvFileKeyStore(filenames ...string) KeyStore {
	if len(filenames) == 0 {
		filenames = []string{".env"}
	}

	// Pre-load all files into a map
	values := make(map[string]string)
	for _, filename := range filenames {
		fileValues, err := readEnvFile(filename)
		if err != nil {
			// If a file doesn't exist or can't be read, we just skip it as per typical .env behavior
			continue
		}
		for k, v := range fileValues {
			if _, exists := values[k]; !exists {
				values[k] = v
			}
		}
	}

	return func(ctx context.Context, key string) (string, bool, error) {
		val, ok := values[key]
		return val, ok, nil
	}
}

func readEnvFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split by first '='
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		values[key] = value
	}

	return values, scanner.Err()
}
