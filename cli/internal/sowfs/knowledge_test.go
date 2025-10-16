package sowfs

import (
	"testing"

	"github.com/jmgilman/go/fs/billy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKnowledgeFS_ReadFile tests reading files from knowledge directory
func TestKnowledgeFS_ReadFile(t *testing.T) {
	// Setup filesystem with test data
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/knowledge", 0755)
	fs.WriteFile(".sow/knowledge/overview.md", []byte("# Overview\nTest content"), 0644)
	fs.MkdirAll(".sow/knowledge/architecture", 0755)
	fs.WriteFile(".sow/knowledge/architecture/design.md", []byte("# Design"), 0644)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	knowledgeFS := sowFS.Knowledge()

	tests := []struct {
		name     string
		path     string
		want     string
		wantErr  bool
	}{
		{
			name:    "read root file",
			path:    "overview.md",
			want:    "# Overview\nTest content",
			wantErr: false,
		},
		{
			name:    "read nested file",
			path:    "architecture/design.md",
			want:    "# Design",
			wantErr: false,
		},
		{
			name:    "file not found",
			path:    "nonexistent.md",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := knowledgeFS.ReadFile(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, string(data))
			}
		})
	}
}

// TestKnowledgeFS_WriteFile tests writing files to knowledge directory
func TestKnowledgeFS_WriteFile(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/knowledge", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	knowledgeFS := sowFS.Knowledge()

	tests := []struct {
		name    string
		path    string
		content string
	}{
		{
			name:    "write root file",
			path:    "overview.md",
			content: "# Overview",
		},
		{
			name:    "write nested file (creates dirs)",
			path:    "architecture/systems/auth.md",
			content: "# Auth System",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := knowledgeFS.WriteFile(tt.path, []byte(tt.content))
			require.NoError(t, err)

			// Verify file was written
			data, err := knowledgeFS.ReadFile(tt.path)
			require.NoError(t, err)
			assert.Equal(t, tt.content, string(data))
		})
	}
}

// TestKnowledgeFS_Exists tests checking if files exist
func TestKnowledgeFS_Exists(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/knowledge", 0755)
	fs.WriteFile(".sow/knowledge/overview.md", []byte("content"), 0644)
	fs.MkdirAll(".sow/knowledge/architecture", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	knowledgeFS := sowFS.Knowledge()

	tests := []struct {
		name   string
		path   string
		exists bool
	}{
		{
			name:   "file exists",
			path:   "overview.md",
			exists: true,
		},
		{
			name:   "directory exists",
			path:   "architecture",
			exists: true,
		},
		{
			name:   "file does not exist",
			path:   "nonexistent.md",
			exists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := knowledgeFS.Exists(tt.path)
			require.NoError(t, err)
			assert.Equal(t, tt.exists, exists)
		})
	}
}

// TestKnowledgeFS_MkdirAll tests creating directories
func TestKnowledgeFS_MkdirAll(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/knowledge", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	knowledgeFS := sowFS.Knowledge()

	// Create nested directory
	err = knowledgeFS.MkdirAll("architecture/systems/auth")
	require.NoError(t, err)

	// Verify directory exists
	exists, err := knowledgeFS.Exists("architecture/systems/auth")
	require.NoError(t, err)
	assert.True(t, exists)
}

// TestKnowledgeFS_ListADRs tests listing ADR files
func TestKnowledgeFS_ListADRs(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*billy.MemoryFS)
		want    []string
		wantErr bool
	}{
		{
			name: "list ADRs",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/knowledge/adrs", 0755)
				fs.WriteFile(".sow/knowledge/adrs/001-first.md", []byte("content"), 0644)
				fs.WriteFile(".sow/knowledge/adrs/002-second.md", []byte("content"), 0644)
			},
			want: []string{"001-first.md", "002-second.md"},
		},
		{
			name: "no adrs directory",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/knowledge", 0755)
				// Don't create adrs directory
			},
			want: []string{},
		},
		{
			name: "empty adrs directory",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/knowledge/adrs", 0755)
			},
			want: []string{},
		},
		{
			name: "adrs with subdirectory (filtered out)",
			setup: func(fs *billy.MemoryFS) {
				fs.MkdirAll(".sow/knowledge/adrs", 0755)
				fs.WriteFile(".sow/knowledge/adrs/001-first.md", []byte("content"), 0644)
				fs.MkdirAll(".sow/knowledge/adrs/archive", 0755)
			},
			want: []string{"001-first.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := billy.NewMemory()
			tt.setup(fs)

			sowFS, err := NewSowFSWithFS(fs, "/test/repo")
			require.NoError(t, err)

			knowledgeFS := sowFS.Knowledge()

			adrs, err := knowledgeFS.ListADRs()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.want, adrs)
			}
		})
	}
}

// TestKnowledgeFS_ReadADR tests reading ADR files
func TestKnowledgeFS_ReadADR(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/knowledge/adrs", 0755)
	fs.WriteFile(".sow/knowledge/adrs/001-decision.md", []byte("# ADR 001"), 0644)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	knowledgeFS := sowFS.Knowledge()

	tests := []struct {
		name     string
		filename string
		want     string
		wantErr  bool
	}{
		{
			name:     "read existing ADR",
			filename: "001-decision.md",
			want:     "# ADR 001",
		},
		{
			name:     "ADR not found",
			filename: "999-missing.md",
			wantErr:  true,
		},
		{
			name:     "sanitize path traversal attempt",
			filename: "../../../etc/passwd",
			// filepath.Base will return just "passwd"
			// which won't exist in adrs/
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := knowledgeFS.ReadADR(tt.filename)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, string(data))
			}
		})
	}
}

// TestKnowledgeFS_WriteADR tests writing ADR files
func TestKnowledgeFS_WriteADR(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/knowledge", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	knowledgeFS := sowFS.Knowledge()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "write new ADR",
			filename: "001-decision.md",
			content:  "# ADR 001\nDecision content",
		},
		{
			name:     "sanitize path traversal attempt",
			filename: "../../../tmp/test.md",
			content:  "should be written to adrs/test.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := knowledgeFS.WriteADR(tt.filename, []byte(tt.content))
			require.NoError(t, err)

			// Verify file was written to correct location
			// filepath.Base sanitizes the filename
			sanitized := tt.filename
			if tt.name == "sanitize path traversal attempt" {
				sanitized = "test.md"
			}

			data, err := knowledgeFS.ReadADR(sanitized)
			require.NoError(t, err)
			assert.Equal(t, tt.content, string(data))
		})
	}
}

// TestKnowledgeFS_Integration tests full workflow
func TestKnowledgeFS_Integration(t *testing.T) {
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/knowledge", 0755)

	sowFS, err := NewSowFSWithFS(fs, "/test/repo")
	require.NoError(t, err)

	knowledgeFS := sowFS.Knowledge()

	// Write overview
	err = knowledgeFS.WriteFile("overview.md", []byte("# Repository Overview"))
	require.NoError(t, err)

	// Create architecture directory
	err = knowledgeFS.MkdirAll("architecture")
	require.NoError(t, err)

	// Write architecture docs
	err = knowledgeFS.WriteFile("architecture/system.md", []byte("# System Architecture"))
	require.NoError(t, err)

	// Write ADRs
	err = knowledgeFS.WriteADR("001-use-go.md", []byte("# Use Go"))
	require.NoError(t, err)
	err = knowledgeFS.WriteADR("002-use-cue.md", []byte("# Use CUE"))
	require.NoError(t, err)

	// Verify all files exist
	exists, err := knowledgeFS.Exists("overview.md")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = knowledgeFS.Exists("architecture/system.md")
	require.NoError(t, err)
	assert.True(t, exists)

	// List ADRs
	adrs, err := knowledgeFS.ListADRs()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"001-use-go.md", "002-use-cue.md"}, adrs)

	// Read ADR
	data, err := knowledgeFS.ReadADR("001-use-go.md")
	require.NoError(t, err)
	assert.Equal(t, "# Use Go", string(data))
}
