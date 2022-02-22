package config

import (
	"os"
	"testing"
)

// TestGetEnv Validates the correct behavior of GetEnv function.
func TestGetEnv(t *testing.T) {

	t.Run("key - exist", func(t *testing.T) {
		// Given
		os.Setenv("FOO", "1")

		// When
		value := GetEnv("FOO", "2")

		// Then
		if value != "1" {
			t.Errorf("expected %v, got %v", "1", value)
		}
	})

	t.Run("key - don't exist", func(t *testing.T) {
		// Given/When
		value := GetEnv("BAR", "2")

		// Then
		if value != "2" {
			t.Errorf("expected %v, got %v", "2", value)
		}
	})

}
