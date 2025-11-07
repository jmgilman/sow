package internal

import (
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
)

// TestHuhLibraryInstalled verifies that the huh library and spinner subpackage
// are correctly installed and can be imported. This test ensures the dependencies
// are maintained in go.mod.
func TestHuhLibraryInstalled(t *testing.T) {
	// Verify huh.NewForm can be called (basic library functionality)
	form := huh.NewForm()
	if form == nil {
		t.Error("huh.NewForm() returned nil")
	}

	// Verify spinner.New can be called (spinner subpackage functionality)
	spin := spinner.New()
	if spin == nil {
		t.Error("spinner.New() returned nil")
	}
}
