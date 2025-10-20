package refs

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmgilman/sow/cli/schemas"
)

func TestGitType_Name(t *testing.T) {
	g := &GitType{}
	if g.Name() != "git" {
		t.Errorf("GitType.Name() = %q, want %q", g.Name(), "git")
	}
}

func TestGitType_IsEnabled(t *testing.T) {
	g := &GitType{}
	ctx := context.Background()

	enabled, err := g.IsEnabled(ctx)
	if err != nil {
		t.Fatalf("GitType.IsEnabled() error = %v", err)
	}

	// We can't assert the value since it depends on the system
	// But we can verify it returns without error
	t.Logf("GitType.IsEnabled() = %v (depends on git binary in PATH)", enabled)
}

func TestGitType_ValidateConfig(t *testing.T) {
	g := &GitType{}

	tests := []struct {
		name      string
		config    schemas.RefConfig
		wantError bool
	}{
		{
			name: "valid empty config",
			config: schemas.RefConfig{
				Branch: "",
				Path:   "",
			},
			wantError: false,
		},
		{
			name: "valid with branch",
			config: schemas.RefConfig{
				Branch: "main",
				Path:   "",
			},
			wantError: false,
		},
		{
			name: "valid with path",
			config: schemas.RefConfig{
				Branch: "",
				Path:   "docs",
			},
			wantError: false,
		},
		{
			name: "valid with nested path",
			config: schemas.RefConfig{
				Branch: "",
				Path:   "docs/guides",
			},
			wantError: false,
		},
		{
			name: "valid with both",
			config: schemas.RefConfig{
				Branch: "main",
				Path:   "docs",
			},
			wantError: false,
		},
		{
			name: "invalid absolute path",
			config: schemas.RefConfig{
				Branch: "",
				Path:   "/absolute/path",
			},
			wantError: true,
		},
		{
			name: "invalid path with traversal",
			config: schemas.RefConfig{
				Branch: "",
				Path:   "../../../etc/passwd",
			},
			wantError: true,
		},
		{
			name: "invalid path with dot-dot",
			config: schemas.RefConfig{
				Branch: "",
				Path:   "docs/../../../etc",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := g.ValidateConfig(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("GitType.ValidateConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGitType_CachePath(t *testing.T) {
	g := &GitType{}

	ref := &schemas.Ref{
		Id:     "test-ref",
		Source: "git+https://github.com/org/repo",
	}

	cacheDir := "/cache"
	path := g.CachePath(cacheDir, ref)

	expectedPrefix := filepath.Join("/cache", "git", "checkouts")
	if !strings.HasPrefix(path, expectedPrefix) {
		t.Errorf("GitType.CachePath() = %q, want prefix %q", path, expectedPrefix)
	}

	// Should include ref.Id somewhere in the path
	if !strings.Contains(path, ref.Id) {
		t.Errorf("GitType.CachePath() = %q, should contain ref.Id %q", path, ref.Id)
	}
}

func TestGitType_Interface(_ *testing.T) {
	// Verify GitType implements RefType interface
	var _ RefType = (*GitType)(nil)
}
