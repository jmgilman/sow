package sinks

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
		want    int // number of sinks
		wantErr bool
	}{
		{
			name: "empty index",
			content: `{
				"sinks": []
			}`,
			want:    0,
			wantErr: false,
		},
		{
			name: "index with sinks",
			content: `{
				"sinks": [
					{
						"name": "python-style",
						"path": "python-style",
						"description": "Python coding standards",
						"topics": ["python", "style"],
						"when_to_use": "When working with Python code",
						"version": "1.0.0",
						"source": "https://github.com/example/python-style",
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

			if !tt.wantErr && len(idx.Sinks) != tt.want {
				t.Errorf("Load() got %d sinks, want %d", len(idx.Sinks), tt.want)
			}
		})
	}
}

func TestIndex_Save(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "index.json")

	idx := &Index{
		Sinks: []Sink{
			{
				Name:        "test-sink",
				Path:        "test-sink",
				Description: "Test sink",
				Topics:      []string{"test"},
				WhenToUse:   "For testing",
				Version:     "1.0.0",
				Source:      "https://github.com/example/test",
				UpdatedAt:   time.Now().Format(time.RFC3339),
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

	if len(loaded.Sinks) != 1 {
		t.Errorf("saved index has %d sinks, want 1", len(loaded.Sinks))
	}

	if loaded.Sinks[0].Name != "test-sink" {
		t.Errorf("saved sink name = %s, want test-sink", loaded.Sinks[0].Name)
	}
}

func TestIndex_AddSink(t *testing.T) {
	idx := &Index{Sinks: []Sink{}}

	sink := Sink{
		Name:        "new-sink",
		Path:        "new-sink",
		Description: "New sink",
		Topics:      []string{"new"},
		WhenToUse:   "When new",
		Version:     "1.0.0",
		Source:      "https://github.com/example/new",
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	idx.AddSink(sink)

	if len(idx.Sinks) != 1 {
		t.Errorf("AddSink() resulted in %d sinks, want 1", len(idx.Sinks))
	}

	if idx.Sinks[0].Name != "new-sink" {
		t.Errorf("added sink name = %s, want new-sink", idx.Sinks[0].Name)
	}
}

func TestIndex_RemoveSink(t *testing.T) {
	idx := &Index{
		Sinks: []Sink{
			{Name: "sink-1", Path: "sink-1"},
			{Name: "sink-2", Path: "sink-2"},
		},
	}

	removed := idx.RemoveSink("sink-1")
	if !removed {
		t.Error("RemoveSink() returned false for existing sink")
	}

	if len(idx.Sinks) != 1 {
		t.Errorf("RemoveSink() resulted in %d sinks, want 1", len(idx.Sinks))
	}

	if idx.Sinks[0].Name != "sink-2" {
		t.Errorf("remaining sink name = %s, want sink-2", idx.Sinks[0].Name)
	}

	// Try removing non-existent sink
	removed = idx.RemoveSink("non-existent")
	if removed {
		t.Error("RemoveSink() returned true for non-existent sink")
	}
}

func TestExtractNameFromURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "https url with .git",
			url:  "https://github.com/org/repo.git",
			want: "repo",
		},
		{
			name: "https url without .git",
			url:  "https://github.com/org/repo",
			want: "repo",
		},
		{
			name: "git url",
			url:  "git@github.com:org/repo.git",
			want: "repo",
		},
		{
			name: "with path",
			url:  "https://github.com/org/repo/path/to/dir",
			want: "dir",
		},
		{
			name: "local path",
			url:  "/path/to/local/repo",
			want: "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractNameFromURL(tt.url)
			if got != tt.want {
				t.Errorf("ExtractNameFromURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestIsGitURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "https git url",
			url:  "https://github.com/org/repo",
			want: true,
		},
		{
			name: "git ssh url",
			url:  "git@github.com:org/repo",
			want: true,
		},
		{
			name: "http url",
			url:  "http://github.com/org/repo",
			want: true,
		},
		{
			name: "local path",
			url:  "/path/to/local/repo",
			want: false,
		},
		{
			name: "relative path",
			url:  "../relative/path",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsGitURL(tt.url)
			if got != tt.want {
				t.Errorf("IsGitURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
