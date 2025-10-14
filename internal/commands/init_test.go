package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/your-org/sow/internal/config"
)

// TestNewInitCmd verifies init command can be created
func TestNewInitCmd(t *testing.T) {
	cmd := NewInitCmd()

	if cmd == nil {
		t.Fatal("NewInitCmd() returned nil")
	}

	if cmd.Use != "init" {
		t.Errorf("Init command Use = %q, want %q", cmd.Use, "init")
	}

	if cmd.Run == nil && cmd.RunE == nil {
		t.Error("Init command has no Run or RunE function")
	}
}

// TestInitCmdHasForceFlag verifies --force flag exists
func TestInitCmdHasForceFlag(t *testing.T) {
	cmd := NewInitCmd()

	flag := cmd.Flags().Lookup("force")
	if flag == nil {
		t.Error("Init command missing --force flag")
	}
}

// TestInitCmdHasHelpText verifies help text is present
func TestInitCmdHasHelpText(t *testing.T) {
	cmd := NewInitCmd()

	if cmd.Short == "" {
		t.Error("Init command has no Short description")
	}

	if cmd.Long == "" {
		t.Error("Init command has no Long description")
	}

	if cmd.Example == "" {
		t.Error("Init command has no Example text")
	}
}

// TestInitCmdCreatesStructure verifies directory creation
func TestInitCmdCreatesStructure(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Run init command
	cmd := NewInitCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--dir", tmpDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify .sow/ directory exists
	sowDir := filepath.Join(tmpDir, ".sow")
	if _, err := os.Stat(sowDir); os.IsNotExist(err) {
		t.Error(".sow/ directory was not created")
	}

	// Verify .sow/knowledge/ directory exists
	knowledgeDir := filepath.Join(sowDir, "knowledge")
	if _, err := os.Stat(knowledgeDir); os.IsNotExist(err) {
		t.Error(".sow/knowledge/ directory was not created")
	}
}

// TestInitCmdCreatesVersionFile verifies .version file creation
func TestInitCmdCreatesVersionFile(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewInitCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--dir", tmpDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify .sow/.version exists
	versionFile := filepath.Join(tmpDir, ".sow", ".version")
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		t.Error(".sow/.version file was not created")
	}

	// Verify content contains version
	content, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("Failed to read .version file: %v", err)
	}

	if !strings.Contains(string(content), config.Version) {
		t.Errorf(".version file content missing version %q: %s", config.Version, string(content))
	}
}

// TestInitCmdOutputConfirmsCreation verifies confirmation message
func TestInitCmdOutputConfirmsCreation(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewInitCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--dir", tmpDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	output := buf.String()

	// Verify output mentions creation
	if !strings.Contains(output, "created") && !strings.Contains(output, "Created") && !strings.Contains(output, "initialized") {
		t.Errorf("Output missing creation confirmation: %s", output)
	}

	// Verify output mentions directories
	if !strings.Contains(output, ".sow") {
		t.Errorf("Output missing .sow directory mention: %s", output)
	}
}

// TestInitCmdSkipsIfExists verifies skip behavior when .sow/ exists
func TestInitCmdSkipsIfExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .sow/ directory first
	sowDir := filepath.Join(tmpDir, ".sow")
	if err := os.Mkdir(sowDir, 0755); err != nil {
		t.Fatalf("Failed to create .sow/: %v", err)
	}

	cmd := NewInitCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--dir", tmpDir})

	err := cmd.Execute()
	// Should not error, just skip
	if err != nil {
		t.Fatalf("init command errored on existing .sow/: %v", err)
	}

	output := buf.String()

	// Verify output mentions already exists
	if !strings.Contains(output, "already") && !strings.Contains(output, "exists") {
		t.Errorf("Output missing 'already exists' message: %s", output)
	}
}

// TestInitCmdForceRecreates verifies --force flag recreates structure
func TestInitCmdForceRecreates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .sow/ directory with a marker file
	sowDir := filepath.Join(tmpDir, ".sow")
	if err := os.Mkdir(sowDir, 0755); err != nil {
		t.Fatalf("Failed to create .sow/: %v", err)
	}

	markerFile := filepath.Join(sowDir, "marker.txt")
	if err := os.WriteFile(markerFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create marker file: %v", err)
	}

	// Run init with --force
	cmd := NewInitCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--dir", tmpDir, "--force"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("init command with --force failed: %v", err)
	}

	// Verify marker file is gone (directory was recreated)
	if _, err := os.Stat(markerFile); !os.IsNotExist(err) {
		t.Error("--force flag did not recreate structure, marker file still exists")
	}

	// Verify .sow/ still exists
	if _, err := os.Stat(sowDir); os.IsNotExist(err) {
		t.Error(".sow/ directory missing after --force")
	}

	// Verify .version exists
	versionFile := filepath.Join(sowDir, ".version")
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		t.Error(".version file not created after --force")
	}
}

// TestInitCmdListsCreatedDirectories verifies output lists directories
func TestInitCmdListsCreatedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewInitCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--dir", tmpDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	output := buf.String()

	// Verify output lists directories
	expectedDirs := []string{".sow", "knowledge"}
	for _, dir := range expectedDirs {
		if !strings.Contains(output, dir) {
			t.Errorf("Output missing directory %q: %s", dir, output)
		}
	}
}

// TestInitCmdErrorHandling verifies error handling
func TestInitCmdErrorHandling(t *testing.T) {
	// Try to init in a read-only location (behavior depends on permissions)
	// This test verifies command doesn't panic on errors
	cmd := NewInitCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--dir", "/invalid/read/only/path"})

	// Should handle error gracefully (either return error or log it)
	_ = cmd.Execute()
	// Test passes if we don't panic
}
