package agent

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestNewListCmd_Structure verifies the list command has correct structure.
func TestNewListCmd_Structure(t *testing.T) {
	cmd := newListCmd()

	if cmd.Use != "list" {
		t.Errorf("expected Use='list', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

// TestNewListCmd_ShortDescription verifies the short description is accurate.
func TestNewListCmd_ShortDescription(t *testing.T) {
	cmd := newListCmd()

	if cmd.Short != "List available agents" {
		t.Errorf("expected Short='List available agents', got '%s'", cmd.Short)
	}
}

// TestRunList_OutputsAllAgents verifies all 6 standard agents are in output.
func TestRunList_OutputsAllAgents(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runList(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	expectedAgents := []string{"architect", "decomposer", "implementer", "planner", "researcher", "reviewer"}
	for _, agent := range expectedAgents {
		if !strings.Contains(output, agent) {
			t.Errorf("expected output to contain agent %q", agent)
		}
	}
}

// TestRunList_SortedAlphabetically verifies output is sorted by agent name.
func TestRunList_SortedAlphabetically(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runList(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	expectedOrder := []string{"architect", "decomposer", "implementer", "planner", "researcher", "reviewer"}

	// Find positions of each agent name in the output
	positions := make([]int, len(expectedOrder))
	for i, agent := range expectedOrder {
		pos := strings.Index(output, agent)
		if pos == -1 {
			t.Fatalf("agent %q not found in output", agent)
		}
		positions[i] = pos
	}

	// Verify they appear in order (each position should be greater than previous)
	for i := 1; i < len(positions); i++ {
		if positions[i] <= positions[i-1] {
			t.Errorf("agents not in alphabetical order: %q appears before %q",
				expectedOrder[i], expectedOrder[i-1])
		}
	}
}

// TestRunList_IncludesDescriptions verifies agent descriptions are shown.
func TestRunList_IncludesDescriptions(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runList(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check for key description phrases from agents.go
	expectedDescriptions := []string{
		"System design and architecture decisions",
		"Code implementation using Test-Driven Development",
		"Code review and quality assessment",
	}

	for _, desc := range expectedDescriptions {
		if !strings.Contains(output, desc) {
			t.Errorf("expected output to contain description %q", desc)
		}
	}
}

// TestRunList_HasHeader verifies the output includes a header line.
func TestRunList_HasHeader(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runList(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Available agents:") {
		t.Error("expected output to contain header 'Available agents:'")
	}
}

// TestRunList_FormatsWithAlignment verifies output uses consistent alignment.
func TestRunList_FormatsWithAlignment(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runList(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(output, "\n")

	// Skip header line ("Available agents:")
	// Check that agent lines have consistent indentation
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		// Each agent line should start with two spaces (indentation)
		if !strings.HasPrefix(line, "  ") {
			t.Errorf("expected line to be indented: %q", line)
		}
	}
}
