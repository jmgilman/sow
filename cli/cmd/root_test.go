package cmd

import (
	"testing"

	"github.com/jmgilman/sow/libs/project/state"
)

// TestExplorationProjectTypeRegistered verifies that the exploration project type
// is registered in the global registry when the package is loaded.
// This ensures the blank import in root.go is working correctly.
func TestExplorationProjectTypeRegistered(t *testing.T) {
	// When the cmd package loads, it should import both standard and exploration packages
	// via blank imports, which triggers their init() functions and registers them

	// Verify exploration is registered
	config, exists := state.GetConfig("exploration")
	if !exists {
		t.Fatal("exploration project type not registered - check blank import in root.go")
	}

	if config == nil {
		t.Fatal("exploration project type config is nil")
	}

	// Verify the registered config is the correct type
	// We can't check the exact type due to import cycles, but we can verify it's not nil
	// and that it was registered successfully
}

// TestStandardProjectTypeRegistered verifies that the standard project type
// is still registered (regression test to ensure we didn't break existing functionality).
func TestStandardProjectTypeRegistered(t *testing.T) {
	// Verify standard is still registered
	config, exists := state.GetConfig("standard")
	if !exists {
		t.Fatal("standard project type not registered")
	}

	if config == nil {
		t.Fatal("standard project type config is nil")
	}
}
