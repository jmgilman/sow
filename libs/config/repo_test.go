package config

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/jmgilman/sow/libs/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ptr is a helper to create string pointers for test assertions.
func ptr(s string) *string {
	return &s
}

func TestLoadRepoConfigFromBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    *schemas.Config
		wantErr error
	}{
		{
			name:  "valid complete config",
			input: []byte("artifacts:\n  adrs: custom-adrs\n  design_docs: docs"),
			//nolint:revive // Field names must match generated schemas.Config structure
			want: &schemas.Config{
				Artifacts: &struct {
					Adrs        *string `json:"adrs,omitempty"`
					Design_docs *string `json:"design_docs,omitempty"`
				}{
					Adrs:        ptr("custom-adrs"),
					Design_docs: ptr("docs"),
				},
			},
		},
		{
			name:  "valid partial config - only adrs",
			input: []byte("artifacts:\n  adrs: my-adrs"),
			//nolint:revive // Field names must match generated schemas.Config structure
			want: &schemas.Config{
				Artifacts: &struct {
					Adrs        *string `json:"adrs,omitempty"`
					Design_docs *string `json:"design_docs,omitempty"`
				}{
					Adrs:        ptr("my-adrs"),
					Design_docs: ptr(DefaultDesignDocsPath), // default applied
				},
			},
		},
		{
			name:  "valid partial config - only design_docs",
			input: []byte("artifacts:\n  design_docs: my-docs"),
			//nolint:revive // Field names must match generated schemas.Config structure
			want: &schemas.Config{
				Artifacts: &struct {
					Adrs        *string `json:"adrs,omitempty"`
					Design_docs *string `json:"design_docs,omitempty"`
				}{
					Adrs:        ptr(DefaultADRsPath), // default applied
					Design_docs: ptr("my-docs"),
				},
			},
		},
		{
			name:  "empty bytes returns default config",
			input: []byte{},
			want:  DefaultConfig(),
		},
		{
			name:  "nil bytes returns default config",
			input: nil,
			want:  DefaultConfig(),
		},
		{
			name:  "empty artifacts struct returns defaults",
			input: []byte("artifacts: {}"),
			want:  DefaultConfig(),
		},
		{
			name:  "only whitespace returns default config",
			input: []byte("   \n\t  "),
			want:  DefaultConfig(),
		},
		{
			name:    "invalid yaml - unclosed bracket",
			input:   []byte("invalid: [yaml: without: closing"),
			wantErr: ErrInvalidYAML,
		},
		{
			name:    "invalid yaml - duplicate keys",
			input:   []byte(":\n\t- :"),
			wantErr: ErrInvalidYAML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadRepoConfigFromBytes(tt.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr),
					"expected error wrapping %v, got %v", tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// mockFS implements core.FS for testing.
type mockFS struct {
	files map[string][]byte
	err   error
}

func (m *mockFS) ReadFile(name string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	data, ok := m.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return data, nil
}

func TestLoadRepoConfig(t *testing.T) {
	tests := []struct {
		name    string
		fs      *mockFS
		want    *schemas.Config
		wantErr error
	}{
		{
			name: "config file exists with all fields",
			fs: &mockFS{
				files: map[string][]byte{
					"config.yaml": []byte("artifacts:\n  adrs: custom-adrs\n  design_docs: custom-docs"),
				},
			},
			//nolint:revive // Field names must match generated schemas.Config structure
			want: &schemas.Config{
				Artifacts: &struct {
					Adrs        *string `json:"adrs,omitempty"`
					Design_docs *string `json:"design_docs,omitempty"`
				}{
					Adrs:        ptr("custom-adrs"),
					Design_docs: ptr("custom-docs"),
				},
			},
		},
		{
			name: "config file exists with partial fields",
			fs: &mockFS{
				files: map[string][]byte{
					"config.yaml": []byte("artifacts:\n  adrs: partial-adrs"),
				},
			},
			//nolint:revive // Field names must match generated schemas.Config structure
			want: &schemas.Config{
				Artifacts: &struct {
					Adrs        *string `json:"adrs,omitempty"`
					Design_docs *string `json:"design_docs,omitempty"`
				}{
					Adrs:        ptr("partial-adrs"),
					Design_docs: ptr(DefaultDesignDocsPath),
				},
			},
		},
		{
			name: "config file doesn't exist returns default config",
			fs: &mockFS{
				files: map[string][]byte{}, // no config.yaml
			},
			want: DefaultConfig(),
		},
		{
			name: "empty config file returns default config",
			fs: &mockFS{
				files: map[string][]byte{
					"config.yaml": []byte(""),
				},
			},
			want: DefaultConfig(),
		},
		{
			name: "invalid YAML returns error",
			fs: &mockFS{
				files: map[string][]byte{
					"config.yaml": []byte("invalid: [yaml: without: closing"),
				},
			},
			wantErr: ErrInvalidYAML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadRepoConfig(tt.fs)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr),
					"expected error wrapping %v, got %v", tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
