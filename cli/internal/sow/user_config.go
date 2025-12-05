package sow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/sow/cli/schemas"
	"gopkg.in/yaml.v3"
)

// Default executor name for all agent bindings.
const DefaultExecutorName = "claude-code"

// GetUserConfigPath returns the path to the user configuration file.
// Uses os.UserConfigDir() for cross-platform compatibility:
// - Linux/Mac: ~/.config/sow/config.yaml
// - Windows: %APPDATA%\sow\config.yaml
func GetUserConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, "sow", "config.yaml"), nil
}

// LoadUserConfig loads the user configuration from the standard location.
// Returns default configuration if file doesn't exist (zero-config experience).
// Returns error only for actual failures (parse errors, permission issues).
func LoadUserConfig() (*schemas.UserConfig, error) {
	path, err := GetUserConfigPath()
	if err != nil {
		// If we can't determine the path, return defaults
		return getDefaultUserConfig(), nil
	}
	return loadUserConfigFromPath(path)
}

// loadUserConfigFromPath loads user configuration from a specific path.
// This is used internally and for testing.
func loadUserConfigFromPath(path string) (*schemas.UserConfig, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Config doesn't exist, return defaults (zero-config experience)
		return getDefaultUserConfig(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", path, err)
	}

	// Handle empty file
	if len(data) == 0 {
		config := &schemas.UserConfig{}
		applyUserConfigDefaults(config)
		return config, nil
	}

	// Parse YAML
	var config schemas.UserConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
	}

	// Apply defaults for missing values
	applyUserConfigDefaults(&config)

	return &config, nil
}

// getDefaultUserConfig returns a UserConfig with all default values.
// Default: All agents use claude-code executor with safe settings.
func getDefaultUserConfig() *schemas.UserConfig {
	yoloMode := false
	claudeCode := DefaultExecutorName

	return &schemas.UserConfig{
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
				DefaultExecutorName: {
					Type: "claude",
					Settings: &struct {
						Yolo_mode *bool   `json:"yolo_mode,omitempty"`
						Model     *string `json:"model,omitempty"`
					}{
						Yolo_mode: &yoloMode,
					},
				},
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
				Orchestrator: &claudeCode,
				Implementer:  &claudeCode,
				Architect:    &claudeCode,
				Reviewer:     &claudeCode,
				Planner:      &claudeCode,
				Researcher:   &claudeCode,
				Decomposer:   &claudeCode,
			},
		},
	}
}

// applyUserConfigDefaults fills in missing configuration values with defaults.
// This allows partial configuration - user only specifies what they want to change.
func applyUserConfigDefaults(config *schemas.UserConfig) {
	claudeCode := DefaultExecutorName
	yoloMode := false

	// Initialize Agents if nil
	if config.Agents == nil {
		config.Agents = &struct {
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
		}{}
	}

	// Initialize Executors if nil
	if config.Agents.Executors == nil {
		config.Agents.Executors = make(map[string]struct {
			Type     string `json:"type"`
			Settings *struct {
				Yolo_mode *bool   `json:"yolo_mode,omitempty"`
				Model     *string `json:"model,omitempty"`
			} `json:"settings,omitempty"`
			Custom_args []string `json:"custom_args,omitempty"`
		})
	}

	// Add default claude-code executor if not present
	if _, ok := config.Agents.Executors[DefaultExecutorName]; !ok {
		config.Agents.Executors[DefaultExecutorName] = struct {
			Type     string `json:"type"`
			Settings *struct {
				Yolo_mode *bool   `json:"yolo_mode,omitempty"`
				Model     *string `json:"model,omitempty"`
			} `json:"settings,omitempty"`
			Custom_args []string `json:"custom_args,omitempty"`
		}{
			Type: "claude",
			Settings: &struct {
				Yolo_mode *bool   `json:"yolo_mode,omitempty"`
				Model     *string `json:"model,omitempty"`
			}{
				Yolo_mode: &yoloMode,
			},
		}
	}

	// Initialize Bindings if nil
	if config.Agents.Bindings == nil {
		config.Agents.Bindings = &struct {
			Orchestrator *string `json:"orchestrator,omitempty"`
			Implementer  *string `json:"implementer,omitempty"`
			Architect    *string `json:"architect,omitempty"`
			Reviewer     *string `json:"reviewer,omitempty"`
			Planner      *string `json:"planner,omitempty"`
			Researcher   *string `json:"researcher,omitempty"`
			Decomposer   *string `json:"decomposer,omitempty"`
		}{}
	}

	// Apply defaults for each nil binding
	if config.Agents.Bindings.Orchestrator == nil {
		config.Agents.Bindings.Orchestrator = &claudeCode
	}
	if config.Agents.Bindings.Implementer == nil {
		config.Agents.Bindings.Implementer = &claudeCode
	}
	if config.Agents.Bindings.Architect == nil {
		config.Agents.Bindings.Architect = &claudeCode
	}
	if config.Agents.Bindings.Reviewer == nil {
		config.Agents.Bindings.Reviewer = &claudeCode
	}
	if config.Agents.Bindings.Planner == nil {
		config.Agents.Bindings.Planner = &claudeCode
	}
	if config.Agents.Bindings.Researcher == nil {
		config.Agents.Bindings.Researcher = &claudeCode
	}
	if config.Agents.Bindings.Decomposer == nil {
		config.Agents.Bindings.Decomposer = &claudeCode
	}
}
