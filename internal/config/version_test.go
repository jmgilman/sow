package config

import (
	"testing"
)

// TestVersionVariablesExist verifies version variables are defined
func TestVersionVariablesExist(t *testing.T) {
	// Version should have a default value
	if Version == "" {
		t.Error("Version is empty string")
	}

	// BuildDate should have a default value
	if BuildDate == "" {
		t.Error("BuildDate is empty string")
	}

	// Commit should have a default value
	if Commit == "" {
		t.Error("Commit is empty string")
	}
}

// TestDefaultVersionValues verifies default values before ldflags injection
func TestDefaultVersionValues(t *testing.T) {
	// When not built with ldflags, should have sensible defaults
	// This test validates that the code works before build-time injection

	if Version != "dev" {
		t.Logf("Version is '%s' (expected 'dev' unless built with ldflags)", Version)
	}

	if BuildDate != "unknown" {
		t.Logf("BuildDate is '%s' (expected 'unknown' unless built with ldflags)", BuildDate)
	}

	if Commit != "none" {
		t.Logf("Commit is '%s' (expected 'none' unless built with ldflags)", Commit)
	}
}
