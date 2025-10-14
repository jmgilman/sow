package repos

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIndex_Load(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int // number of repos
		wantErr bool
	}{
		{
			name: "empty index",
			content: `{
				"repositories": []
			}`,
			want:    0,
			wantErr: false,
		},
		{
			name: "index with repos",
			content: `{
				"repositories": [
					{
						"name": "auth-service",
						"path": "auth-service",
						"source": "https://github.com/example/auth-service",
						"purpose": "Authentication implementation patterns",
						"type": "clone",
						"branch": "main",
						"updated_at": "2025-10-13T00:00:00Z"
					}
				]
			}`,
			want:    1,
			wantErr: false,
		},
		{
			name:    "invalid json",
			content: `{invalid}`,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			indexPath := filepath.Join(tmpDir, "index.json")

			if err := os.WriteFile(indexPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			idx := &Index{}
			err := idx.Load(indexPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(idx.Repositories) != tt.want {
				t.Errorf("Load() got %d repos, want %d", len(idx.Repositories), tt.want)
			}
		})
	}
}

func TestIndex_Save(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "index.json")

	idx := &Index{
		Repositories: []Repository{
			{
				Name:      "test-repo",
				Path:      "test-repo",
				Source:    "https://github.com/example/test",
				Purpose:   "Testing",
				Type:      "clone",
				Branch:    StringPtr("main"),
				UpdatedAt: time.Now().Format(time.RFC3339),
			},
		},
	}

	err := idx.Save(indexPath)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	var loaded Index
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}

	if len(loaded.Repositories) != 1 {
		t.Errorf("saved index has %d repos, want 1", len(loaded.Repositories))
	}

	if loaded.Repositories[0].Name != "test-repo" {
		t.Errorf("saved repo name = %s, want test-repo", loaded.Repositories[0].Name)
	}
}

func TestIndex_AddRepository(t *testing.T) {
	idx := &Index{Repositories: []Repository{}}

	repo := Repository{
		Name:      "new-repo",
		Path:      "new-repo",
		Source:    "https://github.com/example/new",
		Purpose:   "New repo",
		Type:      "clone",
		Branch:    StringPtr("main"),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	idx.AddRepository(repo)

	if len(idx.Repositories) != 1 {
		t.Errorf("AddRepository() resulted in %d repos, want 1", len(idx.Repositories))
	}

	if idx.Repositories[0].Name != "new-repo" {
		t.Errorf("added repo name = %s, want new-repo", idx.Repositories[0].Name)
	}
}

func TestIndex_RemoveRepository(t *testing.T) {
	idx := &Index{
		Repositories: []Repository{
			{Name: "repo-1", Path: "repo-1"},
			{Name: "repo-2", Path: "repo-2"},
		},
	}

	removed := idx.RemoveRepository("repo-1")
	if !removed {
		t.Error("RemoveRepository() returned false for existing repo")
	}

	if len(idx.Repositories) != 1 {
		t.Errorf("RemoveRepository() resulted in %d repos, want 1", len(idx.Repositories))
	}

	if idx.Repositories[0].Name != "repo-2" {
		t.Errorf("remaining repo name = %s, want repo-2", idx.Repositories[0].Name)
	}

	// Try removing non-existent repo
	removed = idx.RemoveRepository("non-existent")
	if removed {
		t.Error("RemoveRepository() returned true for non-existent repo")
	}
}

func TestExtractNameFromSource(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "https url with .git",
			source: "https://github.com/org/repo.git",
			want:   "repo",
		},
		{
			name:   "https url without .git",
			source: "https://github.com/org/repo",
			want:   "repo",
		},
		{
			name:   "git url",
			source: "git@github.com:org/repo.git",
			want:   "repo",
		},
		{
			name:   "local path",
			source: "/path/to/local/repo",
			want:   "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractNameFromSource(tt.source)
			if got != tt.want {
				t.Errorf("ExtractNameFromSource(%q) = %q, want %q", tt.source, got, tt.want)
			}
		})
	}
}

func TestIsGitSource(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{
			name:   "https git url",
			source: "https://github.com/org/repo",
			want:   true,
		},
		{
			name:   "git ssh url",
			source: "git@github.com:org/repo",
			want:   true,
		},
		{
			name:   "http url",
			source: "http://github.com/org/repo",
			want:   true,
		},
		{
			name:   "local path",
			source: "/path/to/local/repo",
			want:   false,
		},
		{
			name:   "relative path",
			source: "../relative/path",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsGitSource(tt.source)
			if got != tt.want {
				t.Errorf("IsGitSource(%q) = %v, want %v", tt.source, got, tt.want)
			}
		})
	}
}

// Helper function to create string pointers
func StringPtr(s string) *string {
	return &s
}
