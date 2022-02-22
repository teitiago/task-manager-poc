package config

import "os"

// GetEnv collects a given env variable, key and allows to provide a fallback string
// when the key is not set.
func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
