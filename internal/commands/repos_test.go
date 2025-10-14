package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/your-org/sow/internal/repos"
)

func TestReposListCmd(t *testing.T) {
	tmpDir := t.TempDir()
	reposDir := filepath.Join(tmpDir, ".sow", "repos")
	os.MkdirAll(reposDir, 0755)

	// Create test index
	branch := "main"
	idx := &repos.Index{
		Repositories: []repos.Repository{
			{
				Name:      "test-repo",
				Path:      "test-repo",
				Source:    "https://github.com/example/test",
				Purpose:   "Test repository",
				Type:      "clone",
				Branch:    &branch,
				UpdatedAt: "2025-10-13T00:00:00Z",
			},
		},
	}
	indexPath := filepath.Join(reposDir, "index.json")
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("failed to create test index: %v", err)
	}

	// Change to temp directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := NewReposCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("repos list failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test-repo") {
		t.Errorf("output should contain repo name, got: %s", output)
	}
	if !strings.Contains(output, "Test repository") {
		t.Errorf("output should contain purpose, got: %s", output)
	}
}

func TestReposListCmd_EmptyIndex(t *testing.T) {
	tmpDir := t.TempDir()
	reposDir := filepath.Join(tmpDir, ".sow", "repos")
	os.MkdirAll(reposDir, 0755)

	// Create empty index
	idx := &repos.Index{Repositories: []repos.Repository{}}
	indexPath := filepath.Join(reposDir, "index.json")
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("failed to create test index: %v", err)
	}

	// Change to temp directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := NewReposCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("repos list failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No repositories linked") {
		t.Errorf("output should indicate no repos, got: %s", output)
	}
}

func TestReposRemoveCmd(t *testing.T) {
	tmpDir := t.TempDir()
	reposDir := filepath.Join(tmpDir, ".sow", "repos")
	os.MkdirAll(reposDir, 0755)

	// Create test repo directory
	repoDir := filepath.Join(reposDir, "test-repo")
	os.MkdirAll(repoDir, 0755)
	os.WriteFile(filepath.Join(repoDir, "test.txt"), []byte("test content"), 0644)

	// Create test index
	branch := "main"
	idx := &repos.Index{
		Repositories: []repos.Repository{
			{
				Name:      "test-repo",
				Path:      "test-repo",
				Source:    "https://github.com/example/test",
				Purpose:   "Test",
				Type:      "clone",
				Branch:    &branch,
				UpdatedAt: "2025-10-13T00:00:00Z",
			},
		},
	}
	indexPath := filepath.Join(reposDir, "index.json")
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("failed to create test index: %v", err)
	}

	// Change to temp directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := NewReposCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"remove", "test-repo"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("repos remove failed: %v", err)
	}

	// Verify repo directory was removed
	if _, err := os.Stat(repoDir); !os.IsNotExist(err) {
		t.Error("repo directory should have been removed")
	}

	// Verify index was updated
	var updatedIdx repos.Index
	data, _ := os.ReadFile(indexPath)
	json.Unmarshal(data, &updatedIdx)
	if len(updatedIdx.Repositories) != 0 {
		t.Errorf("index should be empty, got %d repos", len(updatedIdx.Repositories))
	}
}

func TestReposRemoveCmd_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	reposDir := filepath.Join(tmpDir, ".sow", "repos")
	os.MkdirAll(reposDir, 0755)

	// Create empty index
	idx := &repos.Index{Repositories: []repos.Repository{}}
	indexPath := filepath.Join(reposDir, "index.json")
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("failed to create test index: %v", err)
	}

	// Change to temp directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cmd := NewReposCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"remove", "non-existent"})

	err := cmd.Execute()
	if err == nil {
		t.Error("repos remove should fail for non-existent repo")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention repo not found, got: %v", err)
	}
}

func TestExtractRepoNameFromArgs(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		nameFlag string
		want     string
	}{
		{
			name:     "explicit name",
			source:   "https://github.com/org/repo",
			nameFlag: "my-repo",
			want:     "my-repo",
		},
		{
			name:     "auto extract from url",
			source:   "https://github.com/org/auth-service.git",
			nameFlag: "",
			want:     "auth-service",
		},
		{
			name:     "auto extract from path",
			source:   "/local/path/to/repo",
			nameFlag: "",
			want:     "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractRepoNameFromArgs(tt.source, tt.nameFlag)
			if got != tt.want {
				t.Errorf("extractRepoNameFromArgs(%q, %q) = %q, want %q",
					tt.source, tt.nameFlag, got, tt.want)
			}
		})
	}
}
