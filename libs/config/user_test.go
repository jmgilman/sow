package config

import (
	"errors"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/go/fs/core"
	"github.com/jmgilman/sow/libs/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// strPtr is a helper to create string pointers for test assertions.
func strPtr(s string) *string {
	return &s
}

func TestGetUserConfigPath(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T)
		wantContain string
		wantErr     bool
	}{
		{
			name: "uses XDG_CONFIG_HOME when set",
			setup: func(t *testing.T) {
				t.Setenv("XDG_CONFIG_HOME", "/custom/config")
			},
			wantContain: filepath.Join("/custom/config", "sow", "config.yaml"),
		},
		{
			name: "returns ~/.config/sow/config.yaml on Unix when XDG_CONFIG_HOME not set",
			setup: func(t *testing.T) {
				t.Setenv("XDG_CONFIG_HOME", "")
			},
			wantContain: filepath.Join("sow", "config.yaml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			got, err := GetUserConfigPath()

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Contains(t, got, tt.wantContain)
		})
	}
}

func TestGetUserConfigPath_PlatformSpecific(t *testing.T) {
	// Skip platform-specific tests when not on the target platform
	t.Run("Unix path without XDG_CONFIG_HOME", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping Unix-specific test on Windows")
		}

		t.Setenv("XDG_CONFIG_HOME", "")
		got, err := GetUserConfigPath()

		require.NoError(t, err)
		// Should contain .config/sow/config.yaml
		assert.Contains(t, got, ".config")
		assert.Contains(t, got, filepath.Join("sow", "config.yaml"))
	})

	t.Run("Windows path without XDG_CONFIG_HOME", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Skipping Windows-specific test on non-Windows")
		}

		t.Setenv("XDG_CONFIG_HOME", "")
		got, err := GetUserConfigPath()

		require.NoError(t, err)
		// Should use APPDATA on Windows
		assert.Contains(t, got, filepath.Join("sow", "config.yaml"))
	})
}

//nolint:funlen // Table-driven test with complex inline struct literals
func TestLoadUserConfigFromPath(t *testing.T) {
	tests := []struct {
		name      string
		setupFS   func() (core.FS, string) // returns filesystem and path to config file
		envSetup  func(t *testing.T)
		want      func() *schemas.UserConfig
		wantErr   error
		checkFunc func(t *testing.T, got *schemas.UserConfig)
	}{
		{
			name: "valid config with custom executor",
			setupFS: func() (core.FS, string) {
				memfs := billy.NewMemory()
				path := "home/.config/sow/config.yaml"
				_ = memfs.MkdirAll("home/.config/sow", 0755)
				content := `agents:
  executors:
    my-cursor:
      type: cursor
  bindings:
    implementer: my-cursor
`
				_ = memfs.WriteFile(path, []byte(content), 0644)
				return memfs, path
			},
			checkFunc: func(t *testing.T, got *schemas.UserConfig) {
				require.NotNil(t, got.Agents)
				require.NotNil(t, got.Agents.Executors)
				require.Contains(t, got.Agents.Executors, "my-cursor")
				assert.Equal(t, "cursor", got.Agents.Executors["my-cursor"].Type)
				require.NotNil(t, got.Agents.Bindings)
				require.NotNil(t, got.Agents.Bindings.Implementer)
				assert.Equal(t, "my-cursor", *got.Agents.Bindings.Implementer)
			},
		},
		{
			name: "file not found returns defaults",
			setupFS: func() (core.FS, string) {
				memfs := billy.NewMemory()
				return memfs, "nonexistent.yaml"
			},
			checkFunc: func(t *testing.T, got *schemas.UserConfig) {
				// Should have defaults applied
				require.NotNil(t, got.Agents)
				require.NotNil(t, got.Agents.Executors)
				require.Contains(t, got.Agents.Executors, DefaultExecutorName)
				require.NotNil(t, got.Agents.Bindings)
				require.NotNil(t, got.Agents.Bindings.Orchestrator)
				assert.Equal(t, DefaultExecutorName, *got.Agents.Bindings.Orchestrator)
			},
		},
		{
			name: "empty file returns defaults",
			setupFS: func() (core.FS, string) {
				memfs := billy.NewMemory()
				path := "config.yaml"
				_ = memfs.WriteFile(path, []byte(""), 0644)
				return memfs, path
			},
			checkFunc: func(t *testing.T, got *schemas.UserConfig) {
				require.NotNil(t, got.Agents)
				require.NotNil(t, got.Agents.Executors)
				require.Contains(t, got.Agents.Executors, DefaultExecutorName)
			},
		},
		{
			name: "partial config applies defaults for missing",
			setupFS: func() (core.FS, string) {
				memfs := billy.NewMemory()
				path := "config.yaml"
				content := `agents:
  bindings:
    implementer: claude-code
`
				_ = memfs.WriteFile(path, []byte(content), 0644)
				return memfs, path
			},
			checkFunc: func(t *testing.T, got *schemas.UserConfig) {
				require.NotNil(t, got.Agents)
				require.NotNil(t, got.Agents.Bindings)
				// Implementer from file
				require.NotNil(t, got.Agents.Bindings.Implementer)
				assert.Equal(t, DefaultExecutorName, *got.Agents.Bindings.Implementer)
				// Other bindings get defaults
				require.NotNil(t, got.Agents.Bindings.Orchestrator)
				assert.Equal(t, DefaultExecutorName, *got.Agents.Bindings.Orchestrator)
				// Default executor should be added
				require.Contains(t, got.Agents.Executors, DefaultExecutorName)
			},
		},
		{
			name: "invalid YAML returns error",
			setupFS: func() (core.FS, string) {
				memfs := billy.NewMemory()
				path := "config.yaml"
				content := `invalid: [yaml: without: closing`
				_ = memfs.WriteFile(path, []byte(content), 0644)
				return memfs, path
			},
			wantErr: ErrInvalidYAML,
		},
		{
			name: "invalid executor type returns error",
			setupFS: func() (core.FS, string) {
				memfs := billy.NewMemory()
				path := "config.yaml"
				content := `agents:
  executors:
    bad-executor:
      type: invalid-type
`
				_ = memfs.WriteFile(path, []byte(content), 0644)
				return memfs, path
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "binding references undefined executor returns error",
			setupFS: func() (core.FS, string) {
				memfs := billy.NewMemory()
				path := "config.yaml"
				content := `agents:
  bindings:
    implementer: nonexistent-executor
`
				_ = memfs.WriteFile(path, []byte(content), 0644)
				return memfs, path
			},
			wantErr: ErrInvalidConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars that might interfere
			for _, env := range []string{
				"SOW_AGENTS_ORCHESTRATOR",
				"SOW_AGENTS_IMPLEMENTER",
				"SOW_AGENTS_ARCHITECT",
				"SOW_AGENTS_REVIEWER",
				"SOW_AGENTS_PLANNER",
				"SOW_AGENTS_RESEARCHER",
				"SOW_AGENTS_DECOMPOSER",
			} {
				t.Setenv(env, "")
			}

			if tt.envSetup != nil {
				tt.envSetup(t)
			}

			fsys, path := tt.setupFS()
			got, err := LoadUserConfigFromPath(fsys, path)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr),
					"expected error wrapping %v, got %v", tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			if tt.checkFunc != nil {
				tt.checkFunc(t, got)
			}
		})
	}
}

//nolint:funlen // Table-driven test with complex inline struct literals
func TestValidateUserConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *schemas.UserConfig
		wantErr error
	}{
		{
			name:   "nil config returns nil",
			config: nil,
		},
		{
			name:   "config with nil agents returns nil",
			config: &schemas.UserConfig{},
		},
		{
			name: "valid config with defined executors",
			//nolint:revive // Field names must match generated schemas.UserConfig structure
			config: &schemas.UserConfig{
				Agents: &struct {
					Executors map[string]struct {
						Type     string `json:"type"`
						Settings *struct {
							Yolo_mode *bool   `json:"yolo_mode,omitempty"`
							Model     *string `json:"model,omitempty"`
						} `json:"settings,omitempty"`
						Custom_args []string `json:"custom_args,omitempty"`
					} `json:"executors,omitempty"`
					Bindings *struct {
						Orchestrator *string `json:"orchestrator,omitempty"`
						Implementer  *string `json:"implementer,omitempty"`
						Architect    *string `json:"architect,omitempty"`
						Reviewer     *string `json:"reviewer,omitempty"`
						Planner      *string `json:"planner,omitempty"`
						Researcher   *string `json:"researcher,omitempty"`
						Decomposer   *string `json:"decomposer,omitempty"`
					} `json:"bindings,omitempty"`
				}{
					Executors: map[string]struct {
						Type     string `json:"type"`
						Settings *struct {
							Yolo_mode *bool   `json:"yolo_mode,omitempty"`
							Model     *string `json:"model,omitempty"`
						} `json:"settings,omitempty"`
						Custom_args []string `json:"custom_args,omitempty"`
					}{
						"my-cursor": {Type: "cursor"},
					},
					Bindings: &struct {
						Orchestrator *string `json:"orchestrator,omitempty"`
						Implementer  *string `json:"implementer,omitempty"`
						Architect    *string `json:"architect,omitempty"`
						Reviewer     *string `json:"reviewer,omitempty"`
						Planner      *string `json:"planner,omitempty"`
						Researcher   *string `json:"researcher,omitempty"`
						Decomposer   *string `json:"decomposer,omitempty"`
					}{
						Implementer: strPtr("my-cursor"),
					},
				},
			},
		},
		{
			name: "unknown executor type returns error",
			//nolint:revive // Field names must match generated schemas.UserConfig structure
			config: &schemas.UserConfig{
				Agents: &struct {
					Executors map[string]struct {
						Type     string `json:"type"`
						Settings *struct {
							Yolo_mode *bool   `json:"yolo_mode,omitempty"`
							Model     *string `json:"model,omitempty"`
						} `json:"settings,omitempty"`
						Custom_args []string `json:"custom_args,omitempty"`
					} `json:"executors,omitempty"`
					Bindings *struct {
						Orchestrator *string `json:"orchestrator,omitempty"`
						Implementer  *string `json:"implementer,omitempty"`
						Architect    *string `json:"architect,omitempty"`
						Reviewer     *string `json:"reviewer,omitempty"`
						Planner      *string `json:"planner,omitempty"`
						Researcher   *string `json:"researcher,omitempty"`
						Decomposer   *string `json:"decomposer,omitempty"`
					} `json:"bindings,omitempty"`
				}{
					Executors: map[string]struct {
						Type     string `json:"type"`
						Settings *struct {
							Yolo_mode *bool   `json:"yolo_mode,omitempty"`
							Model     *string `json:"model,omitempty"`
						} `json:"settings,omitempty"`
						Custom_args []string `json:"custom_args,omitempty"`
					}{
						"bad": {Type: "unknown-type"},
					},
				},
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "binding references undefined executor returns error",
			//nolint:revive // Field names must match generated schemas.UserConfig structure
			config: &schemas.UserConfig{
				Agents: &struct {
					Executors map[string]struct {
						Type     string `json:"type"`
						Settings *struct {
							Yolo_mode *bool   `json:"yolo_mode,omitempty"`
							Model     *string `json:"model,omitempty"`
						} `json:"settings,omitempty"`
						Custom_args []string `json:"custom_args,omitempty"`
					} `json:"executors,omitempty"`
					Bindings *struct {
						Orchestrator *string `json:"orchestrator,omitempty"`
						Implementer  *string `json:"implementer,omitempty"`
						Architect    *string `json:"architect,omitempty"`
						Reviewer     *string `json:"reviewer,omitempty"`
						Planner      *string `json:"planner,omitempty"`
						Researcher   *string `json:"researcher,omitempty"`
						Decomposer   *string `json:"decomposer,omitempty"`
					} `json:"bindings,omitempty"`
				}{
					Bindings: &struct {
						Orchestrator *string `json:"orchestrator,omitempty"`
						Implementer  *string `json:"implementer,omitempty"`
						Architect    *string `json:"architect,omitempty"`
						Reviewer     *string `json:"reviewer,omitempty"`
						Planner      *string `json:"planner,omitempty"`
						Researcher   *string `json:"researcher,omitempty"`
						Decomposer   *string `json:"decomposer,omitempty"`
					}{
						Implementer: strPtr("nonexistent"),
					},
				},
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "claude-code binding is always valid (implicit default)",
			//nolint:revive // Field names must match generated schemas.UserConfig structure
			config: &schemas.UserConfig{
				Agents: &struct {
					Executors map[string]struct {
						Type     string `json:"type"`
						Settings *struct {
							Yolo_mode *bool   `json:"yolo_mode,omitempty"`
							Model     *string `json:"model,omitempty"`
						} `json:"settings,omitempty"`
						Custom_args []string `json:"custom_args,omitempty"`
					} `json:"executors,omitempty"`
					Bindings *struct {
						Orchestrator *string `json:"orchestrator,omitempty"`
						Implementer  *string `json:"implementer,omitempty"`
						Architect    *string `json:"architect,omitempty"`
						Reviewer     *string `json:"reviewer,omitempty"`
						Planner      *string `json:"planner,omitempty"`
						Researcher   *string `json:"researcher,omitempty"`
						Decomposer   *string `json:"decomposer,omitempty"`
					} `json:"bindings,omitempty"`
				}{
					// No executors defined
					Bindings: &struct {
						Orchestrator *string `json:"orchestrator,omitempty"`
						Implementer  *string `json:"implementer,omitempty"`
						Architect    *string `json:"architect,omitempty"`
						Reviewer     *string `json:"reviewer,omitempty"`
						Planner      *string `json:"planner,omitempty"`
						Researcher   *string `json:"researcher,omitempty"`
						Decomposer   *string `json:"decomposer,omitempty"`
					}{
						Implementer: strPtr(DefaultExecutorName), // claude-code is always valid
					},
				},
			},
		},
		{
			name: "all valid executor types",
			//nolint:revive // Field names must match generated schemas.UserConfig structure
			config: &schemas.UserConfig{
				Agents: &struct {
					Executors map[string]struct {
						Type     string `json:"type"`
						Settings *struct {
							Yolo_mode *bool   `json:"yolo_mode,omitempty"`
							Model     *string `json:"model,omitempty"`
						} `json:"settings,omitempty"`
						Custom_args []string `json:"custom_args,omitempty"`
					} `json:"executors,omitempty"`
					Bindings *struct {
						Orchestrator *string `json:"orchestrator,omitempty"`
						Implementer  *string `json:"implementer,omitempty"`
						Architect    *string `json:"architect,omitempty"`
						Reviewer     *string `json:"reviewer,omitempty"`
						Planner      *string `json:"planner,omitempty"`
						Researcher   *string `json:"researcher,omitempty"`
						Decomposer   *string `json:"decomposer,omitempty"`
					} `json:"bindings,omitempty"`
				}{
					Executors: map[string]struct {
						Type     string `json:"type"`
						Settings *struct {
							Yolo_mode *bool   `json:"yolo_mode,omitempty"`
							Model     *string `json:"model,omitempty"`
						} `json:"settings,omitempty"`
						Custom_args []string `json:"custom_args,omitempty"`
					}{
						"claude-exec":   {Type: "claude"},
						"cursor-exec":   {Type: "cursor"},
						"windsurf-exec": {Type: "windsurf"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserConfig(tt.config)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr),
					"expected error wrapping %v, got %v", tt.wantErr, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

//nolint:funlen // Table-driven test with multiple test cases
func TestEnvironmentOverrides(t *testing.T) {
	tests := []struct {
		name      string
		setupFS   func() (core.FS, string)
		envSetup  func(t *testing.T)
		checkFunc func(t *testing.T, got *schemas.UserConfig)
	}{
		{
			name: "SOW_AGENTS_IMPLEMENTER overrides binding",
			setupFS: func() (core.FS, string) {
				memfs := billy.NewMemory()
				path := "config.yaml"
				content := `agents:
  bindings:
    implementer: claude-code
`
				_ = memfs.WriteFile(path, []byte(content), 0644)
				return memfs, path
			},
			envSetup: func(t *testing.T) {
				t.Setenv("SOW_AGENTS_IMPLEMENTER", "custom-executor")
			},
			checkFunc: func(t *testing.T, got *schemas.UserConfig) {
				require.NotNil(t, got.Agents)
				require.NotNil(t, got.Agents.Bindings)
				require.NotNil(t, got.Agents.Bindings.Implementer)
				assert.Equal(t, "custom-executor", *got.Agents.Bindings.Implementer)
			},
		},
		{
			name: "multiple env vars can be set",
			setupFS: func() (core.FS, string) {
				memfs := billy.NewMemory()
				return memfs, "nonexistent.yaml"
			},
			envSetup: func(t *testing.T) {
				t.Setenv("SOW_AGENTS_ORCHESTRATOR", "orch-executor")
				t.Setenv("SOW_AGENTS_IMPLEMENTER", "impl-executor")
				t.Setenv("SOW_AGENTS_ARCHITECT", "arch-executor")
			},
			checkFunc: func(t *testing.T, got *schemas.UserConfig) {
				require.NotNil(t, got.Agents)
				require.NotNil(t, got.Agents.Bindings)
				assert.Equal(t, "orch-executor", *got.Agents.Bindings.Orchestrator)
				assert.Equal(t, "impl-executor", *got.Agents.Bindings.Implementer)
				assert.Equal(t, "arch-executor", *got.Agents.Bindings.Architect)
			},
		},
		{
			name: "env vars take precedence over file config",
			setupFS: func() (core.FS, string) {
				memfs := billy.NewMemory()
				path := "config.yaml"
				content := `agents:
  executors:
    my-cursor:
      type: cursor
  bindings:
    implementer: my-cursor
    orchestrator: my-cursor
`
				_ = memfs.WriteFile(path, []byte(content), 0644)
				return memfs, path
			},
			envSetup: func(t *testing.T) {
				t.Setenv("SOW_AGENTS_IMPLEMENTER", "env-override")
			},
			checkFunc: func(t *testing.T, got *schemas.UserConfig) {
				require.NotNil(t, got.Agents)
				require.NotNil(t, got.Agents.Bindings)
				// Implementer overridden by env
				assert.Equal(t, "env-override", *got.Agents.Bindings.Implementer)
				// Orchestrator kept from file
				assert.Equal(t, "my-cursor", *got.Agents.Bindings.Orchestrator)
			},
		},
		{
			name: "all env var overrides work",
			setupFS: func() (core.FS, string) {
				memfs := billy.NewMemory()
				return memfs, "nonexistent.yaml"
			},
			envSetup: func(t *testing.T) {
				t.Setenv("SOW_AGENTS_ORCHESTRATOR", "env-orch")
				t.Setenv("SOW_AGENTS_IMPLEMENTER", "env-impl")
				t.Setenv("SOW_AGENTS_ARCHITECT", "env-arch")
				t.Setenv("SOW_AGENTS_REVIEWER", "env-rev")
				t.Setenv("SOW_AGENTS_PLANNER", "env-plan")
				t.Setenv("SOW_AGENTS_RESEARCHER", "env-res")
				t.Setenv("SOW_AGENTS_DECOMPOSER", "env-dec")
			},
			checkFunc: func(t *testing.T, got *schemas.UserConfig) {
				require.NotNil(t, got.Agents)
				require.NotNil(t, got.Agents.Bindings)
				assert.Equal(t, "env-orch", *got.Agents.Bindings.Orchestrator)
				assert.Equal(t, "env-impl", *got.Agents.Bindings.Implementer)
				assert.Equal(t, "env-arch", *got.Agents.Bindings.Architect)
				assert.Equal(t, "env-rev", *got.Agents.Bindings.Reviewer)
				assert.Equal(t, "env-plan", *got.Agents.Bindings.Planner)
				assert.Equal(t, "env-res", *got.Agents.Bindings.Researcher)
				assert.Equal(t, "env-dec", *got.Agents.Bindings.Decomposer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all SOW_AGENTS_* env vars first
			for _, env := range []string{
				"SOW_AGENTS_ORCHESTRATOR",
				"SOW_AGENTS_IMPLEMENTER",
				"SOW_AGENTS_ARCHITECT",
				"SOW_AGENTS_REVIEWER",
				"SOW_AGENTS_PLANNER",
				"SOW_AGENTS_RESEARCHER",
				"SOW_AGENTS_DECOMPOSER",
			} {
				t.Setenv(env, "")
			}

			if tt.envSetup != nil {
				tt.envSetup(t)
			}

			fsys, path := tt.setupFS()
			got, err := LoadUserConfigFromPath(fsys, path)

			require.NoError(t, err)
			tt.checkFunc(t, got)
		})
	}
}

func TestDefaultValues(t *testing.T) {
	t.Run("default executor is claude-code with type claude", func(t *testing.T) {
		config := getDefaultUserConfig()

		require.NotNil(t, config.Agents)
		require.NotNil(t, config.Agents.Executors)
		require.Contains(t, config.Agents.Executors, DefaultExecutorName)
		assert.Equal(t, "claude", config.Agents.Executors[DefaultExecutorName].Type)
	})

	t.Run("all role bindings default to claude-code", func(t *testing.T) {
		config := getDefaultUserConfig()

		require.NotNil(t, config.Agents)
		require.NotNil(t, config.Agents.Bindings)

		assert.Equal(t, DefaultExecutorName, *config.Agents.Bindings.Orchestrator)
		assert.Equal(t, DefaultExecutorName, *config.Agents.Bindings.Implementer)
		assert.Equal(t, DefaultExecutorName, *config.Agents.Bindings.Architect)
		assert.Equal(t, DefaultExecutorName, *config.Agents.Bindings.Reviewer)
		assert.Equal(t, DefaultExecutorName, *config.Agents.Bindings.Planner)
		assert.Equal(t, DefaultExecutorName, *config.Agents.Bindings.Researcher)
		assert.Equal(t, DefaultExecutorName, *config.Agents.Bindings.Decomposer)
	})

	t.Run("default yolo_mode is false", func(t *testing.T) {
		config := getDefaultUserConfig()

		require.NotNil(t, config.Agents)
		require.NotNil(t, config.Agents.Executors)
		require.Contains(t, config.Agents.Executors, DefaultExecutorName)

		executor := config.Agents.Executors[DefaultExecutorName]
		require.NotNil(t, executor.Settings)
		require.NotNil(t, executor.Settings.Yolo_mode)
		assert.False(t, *executor.Settings.Yolo_mode)
	})
}

func TestLoadUserConfig(t *testing.T) {
	// This test uses an in-memory filesystem
	t.Run("returns config without error", func(t *testing.T) {
		// Clear all SOW_AGENTS_* env vars
		for _, env := range []string{
			"SOW_AGENTS_ORCHESTRATOR",
			"SOW_AGENTS_IMPLEMENTER",
			"SOW_AGENTS_ARCHITECT",
			"SOW_AGENTS_REVIEWER",
			"SOW_AGENTS_PLANNER",
			"SOW_AGENTS_RESEARCHER",
			"SOW_AGENTS_DECOMPOSER",
		} {
			t.Setenv(env, "")
		}

		// Use an in-memory filesystem with no config file
		memfs := billy.NewMemory()

		config, err := LoadUserConfig(memfs)

		require.NoError(t, err)
		require.NotNil(t, config)
		// Should have defaults
		require.NotNil(t, config.Agents)
		require.NotNil(t, config.Agents.Bindings)
	})
}

func TestValidExecutorTypes(t *testing.T) {
	t.Run("contains expected types", func(t *testing.T) {
		assert.True(t, ValidExecutorTypes["claude"])
		assert.True(t, ValidExecutorTypes["cursor"])
		assert.True(t, ValidExecutorTypes["windsurf"])
	})

	t.Run("rejects invalid types", func(t *testing.T) {
		assert.False(t, ValidExecutorTypes["invalid"])
		assert.False(t, ValidExecutorTypes[""])
		assert.False(t, ValidExecutorTypes["vscode"])
	})
}

func TestLoadingPipelineOrder(t *testing.T) {
	t.Run("validates before applying defaults", func(t *testing.T) {
		// If an invalid executor type is in the config, it should fail validation
		// even if defaults would make it work
		memfs := billy.NewMemory()
		path := "config.yaml"
		content := `agents:
  executors:
    bad-exec:
      type: not-valid
`
		_ = memfs.WriteFile(path, []byte(content), 0644)

		// Clear env vars
		for _, env := range []string{
			"SOW_AGENTS_ORCHESTRATOR",
			"SOW_AGENTS_IMPLEMENTER",
			"SOW_AGENTS_ARCHITECT",
			"SOW_AGENTS_REVIEWER",
			"SOW_AGENTS_PLANNER",
			"SOW_AGENTS_RESEARCHER",
			"SOW_AGENTS_DECOMPOSER",
		} {
			t.Setenv(env, "")
		}

		_, err := LoadUserConfigFromPath(memfs, path)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidConfig))
	})

	t.Run("env overrides applied last", func(t *testing.T) {
		memfs := billy.NewMemory()
		path := "config.yaml"
		content := `agents:
  bindings:
    implementer: claude-code
`
		_ = memfs.WriteFile(path, []byte(content), 0644)

		// Clear other env vars
		for _, env := range []string{
			"SOW_AGENTS_ORCHESTRATOR",
			"SOW_AGENTS_ARCHITECT",
			"SOW_AGENTS_REVIEWER",
			"SOW_AGENTS_PLANNER",
			"SOW_AGENTS_RESEARCHER",
			"SOW_AGENTS_DECOMPOSER",
		} {
			t.Setenv(env, "")
		}
		// Set the override
		t.Setenv("SOW_AGENTS_IMPLEMENTER", "env-override")

		config, err := LoadUserConfigFromPath(memfs, path)

		require.NoError(t, err)
		// Env override should win
		assert.Equal(t, "env-override", *config.Agents.Bindings.Implementer)
	})
}
