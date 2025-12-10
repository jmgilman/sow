package config

import (
	"path/filepath"
	"testing"

	"github.com/jmgilman/sow/libs/schemas"
	"github.com/stretchr/testify/assert"
)

func TestGetADRsPath(t *testing.T) {
	customPath := "custom-adrs"

	//nolint:revive // Field names must match generated schemas.Config structure
	tests := []struct {
		name     string
		repoRoot string
		config   *schemas.Config
		want     string
	}{
		{
			name:     "custom path configured",
			repoRoot: "/repo",
			config: &schemas.Config{
				Artifacts: &struct {
					Adrs        *string `json:"adrs,omitempty"`
					Design_docs *string `json:"design_docs,omitempty"`
				}{
					Adrs: &customPath,
				},
			},
			want: filepath.Join("/repo", ".sow", "knowledge", "custom-adrs"),
		},
		{
			name:     "nil config uses default",
			repoRoot: "/repo",
			config:   nil,
			want:     filepath.Join("/repo", ".sow", "knowledge", "adrs"),
		},
		{
			name:     "nil Artifacts uses default",
			repoRoot: "/repo",
			config:   &schemas.Config{Artifacts: nil},
			want:     filepath.Join("/repo", ".sow", "knowledge", "adrs"),
		},
		{
			name:     "nil Adrs field uses default",
			repoRoot: "/repo",
			config: &schemas.Config{
				Artifacts: &struct {
					Adrs        *string `json:"adrs,omitempty"`
					Design_docs *string `json:"design_docs,omitempty"`
				}{
					Adrs: nil,
				},
			},
			want: filepath.Join("/repo", ".sow", "knowledge", "adrs"),
		},
		{
			name:     "absolute repoRoot path",
			repoRoot: "/home/user/project",
			config:   nil,
			want:     filepath.Join("/home/user/project", ".sow", "knowledge", "adrs"),
		},
		{
			name:     "relative repoRoot path",
			repoRoot: "relative/path",
			config:   nil,
			want:     filepath.Join("relative/path", ".sow", "knowledge", "adrs"),
		},
		{
			name:     "repoRoot with trailing slash handled by filepath.Join",
			repoRoot: "/repo/",
			config:   nil,
			want:     filepath.Join("/repo/", ".sow", "knowledge", "adrs"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetADRsPath(tt.repoRoot, tt.config)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetDesignDocsPath(t *testing.T) {
	customPath := "custom-design"

	//nolint:revive // Field names must match generated schemas.Config structure
	tests := []struct {
		name     string
		repoRoot string
		config   *schemas.Config
		want     string
	}{
		{
			name:     "custom path configured",
			repoRoot: "/repo",
			config: &schemas.Config{
				Artifacts: &struct {
					Adrs        *string `json:"adrs,omitempty"`
					Design_docs *string `json:"design_docs,omitempty"`
				}{
					Design_docs: &customPath,
				},
			},
			want: filepath.Join("/repo", ".sow", "knowledge", "custom-design"),
		},
		{
			name:     "nil config uses default",
			repoRoot: "/repo",
			config:   nil,
			want:     filepath.Join("/repo", ".sow", "knowledge", "design"),
		},
		{
			name:     "nil Artifacts uses default",
			repoRoot: "/repo",
			config:   &schemas.Config{Artifacts: nil},
			want:     filepath.Join("/repo", ".sow", "knowledge", "design"),
		},
		{
			name:     "nil Design_docs field uses default",
			repoRoot: "/repo",
			config: &schemas.Config{
				Artifacts: &struct {
					Adrs        *string `json:"adrs,omitempty"`
					Design_docs *string `json:"design_docs,omitempty"`
				}{
					Design_docs: nil,
				},
			},
			want: filepath.Join("/repo", ".sow", "knowledge", "design"),
		},
		{
			name:     "absolute repoRoot path",
			repoRoot: "/home/user/project",
			config:   nil,
			want:     filepath.Join("/home/user/project", ".sow", "knowledge", "design"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDesignDocsPath(tt.repoRoot, tt.config)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetExplorationsPath(t *testing.T) {
	tests := []struct {
		name     string
		repoRoot string
		want     string
	}{
		{
			name:     "returns correct path with repoRoot",
			repoRoot: "/repo",
			want:     filepath.Join("/repo", ".sow", "knowledge", "explorations"),
		},
		{
			name:     "absolute repoRoot path",
			repoRoot: "/home/user/project",
			want:     filepath.Join("/home/user/project", ".sow", "knowledge", "explorations"),
		},
		{
			name:     "relative repoRoot path",
			repoRoot: "relative/path",
			want:     filepath.Join("relative/path", ".sow", "knowledge", "explorations"),
		},
		{
			name:     "repoRoot with trailing slash",
			repoRoot: "/repo/",
			want:     filepath.Join("/repo/", ".sow", "knowledge", "explorations"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExplorationsPath(tt.repoRoot)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetKnowledgePath(t *testing.T) {
	tests := []struct {
		name     string
		repoRoot string
		want     string
	}{
		{
			name:     "returns correct path with repoRoot",
			repoRoot: "/repo",
			want:     filepath.Join("/repo", ".sow", "knowledge"),
		},
		{
			name:     "absolute repoRoot path",
			repoRoot: "/home/user/project",
			want:     filepath.Join("/home/user/project", ".sow", "knowledge"),
		},
		{
			name:     "relative repoRoot path",
			repoRoot: "relative/path",
			want:     filepath.Join("relative/path", ".sow", "knowledge"),
		},
		{
			name:     "repoRoot with trailing slash",
			repoRoot: "/repo/",
			want:     filepath.Join("/repo/", ".sow", "knowledge"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetKnowledgePath(tt.repoRoot)
			assert.Equal(t, tt.want, got)
		})
	}
}
