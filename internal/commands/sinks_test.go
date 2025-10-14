package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/your-org/sow/internal/sinks"
)

func TestSinksListCmd(t *testing.T) {
	tmpDir := t.TempDir()
	sinksDir := filepath.Join(tmpDir, ".sow", "sinks")
	os.MkdirAll(sinksDir, 0755)

	// Create test index
	idx := &sinks.Index{
		Sinks: []sinks.Sink{
			{
				Name:        "test-sink",
				Path:        "test-sink",
				Description: "Test sink description",
				Topics:      []string{"test", "example"},
				WhenToUse:   "For testing",
				Version:     "1.0.0",
				Source:      "https://github.com/example/test",
				UpdatedAt:   "2025-10-13T00:00:00Z",
			},
		},
	}
	indexPath := filepath.Join(sinksDir, "index.json")
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("failed to create test index: %v", err)
	}

	// Change to temp directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := NewSinksCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("sinks list failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test-sink") {
		t.Errorf("output should contain sink name, got: %s", output)
	}
	if !strings.Contains(output, "Test sink description") {
		t.Errorf("output should contain description, got: %s", output)
	}
}

func TestSinksListCmd_EmptyIndex(t *testing.T) {
	tmpDir := t.TempDir()
	sinksDir := filepath.Join(tmpDir, ".sow", "sinks")
	os.MkdirAll(sinksDir, 0755)

	// Create empty index
	idx := &sinks.Index{Sinks: []sinks.Sink{}}
	indexPath := filepath.Join(sinksDir, "index.json")
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("failed to create test index: %v", err)
	}

	// Change to temp directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := NewSinksCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("sinks list failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No sinks installed") {
		t.Errorf("output should indicate no sinks, got: %s", output)
	}
}

func TestSinksRemoveCmd(t *testing.T) {
	tmpDir := t.TempDir()
	sinksDir := filepath.Join(tmpDir, ".sow", "sinks")
	os.MkdirAll(sinksDir, 0755)

	// Create test sink directory
	sinkDir := filepath.Join(sinksDir, "test-sink")
	os.MkdirAll(sinkDir, 0755)
	os.WriteFile(filepath.Join(sinkDir, "test.md"), []byte("test content"), 0644)

	// Create test index
	idx := &sinks.Index{
		Sinks: []sinks.Sink{
			{
				Name:        "test-sink",
				Path:        "test-sink",
				Description: "Test sink",
				Topics:      []string{"test"},
				WhenToUse:   "For testing",
				Version:     "1.0.0",
				Source:      "https://github.com/example/test",
				UpdatedAt:   "2025-10-13T00:00:00Z",
			},
		},
	}
	indexPath := filepath.Join(sinksDir, "index.json")
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("failed to create test index: %v", err)
	}

	// Change to temp directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := NewSinksCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"remove", "test-sink"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("sinks remove failed: %v", err)
	}

	// Verify sink directory was removed
	if _, err := os.Stat(sinkDir); !os.IsNotExist(err) {
		t.Error("sink directory should have been removed")
	}

	// Verify index was updated
	var updatedIdx sinks.Index
	data, _ := os.ReadFile(indexPath)
	json.Unmarshal(data, &updatedIdx)
	if len(updatedIdx.Sinks) != 0 {
		t.Errorf("index should be empty, got %d sinks", len(updatedIdx.Sinks))
	}
}

func TestSinksRemoveCmd_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	sinksDir := filepath.Join(tmpDir, ".sow", "sinks")
	os.MkdirAll(sinksDir, 0755)

	// Create empty index
	idx := &sinks.Index{Sinks: []sinks.Sink{}}
	indexPath := filepath.Join(sinksDir, "index.json")
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("failed to create test index: %v", err)
	}

	// Change to temp directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := NewSinksCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"remove", "non-existent"})

	err := cmd.Execute()
	if err == nil {
		t.Error("sinks remove should fail for non-existent sink")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention sink not found, got: %v", err)
	}
}

func TestExtractNameFromArgs(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		nameFlag string
		want     string
	}{
		{
			name:     "explicit name",
			url:      "https://github.com/org/repo",
			nameFlag: "my-sink",
			want:     "my-sink",
		},
		{
			name:     "auto extract from url",
			url:      "https://github.com/org/python-style.git",
			nameFlag: "",
			want:     "python-style",
		},
		{
			name:     "auto extract from path",
			url:      "/local/path/to/sink",
			nameFlag: "",
			want:     "sink",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractNameFromArgs(tt.url, tt.nameFlag)
			if got != tt.want {
				t.Errorf("extractNameFromArgs(%q, %q) = %q, want %q",
					tt.url, tt.nameFlag, got, tt.want)
			}
		})
	}
}
