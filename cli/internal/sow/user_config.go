package sow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/sow/cli/schemas"
	"gopkg.in/yaml.v3"
)

// DefaultExecutorName is the default executor name for all agent bindings.
const DefaultExecutorName = "claude-code"

// GetUserConfigPath returns the path to the user configuration file.
// Uses os.UserConfigDir() for cross-platform compatibility:
// - Linux/Mac: ~/.config/sow/config.yaml
// - Windows: %APPDATA%\sow\config.yaml.
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
	return LoadUserConfigFromPath(path)
}

// LoadUserConfigFromPath loads user configuration from a specific path.
// This is used internally and for testing.
// The full loading pipeline is:
// 1. Read and parse YAML
// 2. Validate (before applying defaults)
// 3. Apply defaults for missing values
// 4. Apply environment overrides (highest priority).
func LoadUserConfigFromPath(path string) (*schemas.UserConfig, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Config doesn't exist, return defaults with env overrides
		config := getDefaultUserConfig()
		applyEnvOverrides(config)
		return config, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", path, err)
	}

	// Handle empty file
	if len(data) == 0 {
		config := &schemas.UserConfig{}
		applyUserConfigDefaults(config)
		applyEnvOverrides(config)
		return config, nil
	}

	// Parse YAML
	var config schemas.UserConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
	}

	// Validate before applying defaults
	if err := ValidateUserConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", path, err)
	}

	// Apply defaults for missing values
	applyUserConfigDefaults(&config)

	// Apply environment overrides (highest priority)
	applyEnvOverrides(&config)

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

// ValidExecutorTypes defines the allowed executor types.
var ValidExecutorTypes = map[string]bool{
	"claude":   true,
	"cursor":   true,
	"windsurf": true,
}

// ValidateUserConfig validates the user configuration.
// Checks:
// - Executor types are valid ("claude", "cursor", "windsurf")
// - Bindings reference defined executors (or default "claude-code")
// Returns nil if valid, error with details if invalid.
func ValidateUserConfig(config *schemas.UserConfig) error {
	if config == nil || config.Agents == nil {
		return nil
	}

	// Validate executor types
	for name, exec := range config.Agents.Executors {
		if !ValidExecutorTypes[exec.Type] {
			return fmt.Errorf("unknown executor type %q for executor %q; must be one of: claude, cursor, windsurf", exec.Type, name)
		}
	}

	// Validate bindings reference defined executors
	if config.Agents.Bindings != nil {
		// Collect all bindings in a map for iteration
		bindings := map[string]*string{
			"orchestrator": config.Agents.Bindings.Orchestrator,
			"implementer":  config.Agents.Bindings.Implementer,
			"architect":    config.Agents.Bindings.Architect,
			"reviewer":     config.Agents.Bindings.Reviewer,
			"planner":      config.Agents.Bindings.Planner,
			"researcher":   config.Agents.Bindings.Researcher,
			"decomposer":   config.Agents.Bindings.Decomposer,
		}

		for agentRole, executorPtr := range bindings {
			if executorPtr == nil {
				continue // Nil bindings are ok, will get defaults
			}
			executorName := *executorPtr

			// "claude-code" is always valid (implicit default executor)
			if executorName == DefaultExecutorName {
				continue
			}

			// Check if executor is defined
			if config.Agents.Executors == nil {
				return fmt.Errorf("binding %q references undefined executor %q", agentRole, executorName)
			}
			if _, ok := config.Agents.Executors[executorName]; !ok {
				return fmt.Errorf("binding %q references undefined executor %q", agentRole, executorName)
			}
		}
	}

	return nil
}

// applyEnvOverrides applies environment variable overrides to the configuration.
// Environment variables take precedence over file configuration.
// Format: SOW_AGENTS_{ROLE}={executor_name}
// Example: SOW_AGENTS_IMPLEMENTER=cursor.
func applyEnvOverrides(config *schemas.UserConfig) {
	// Ensure config.Agents exists
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

	// Ensure config.Agents.Bindings exists
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

	// Map of environment variables to binding fields
	envVars := []struct {
		envVar string
		field  **string
	}{
		{"SOW_AGENTS_ORCHESTRATOR", &config.Agents.Bindings.Orchestrator},
		{"SOW_AGENTS_IMPLEMENTER", &config.Agents.Bindings.Implementer},
		{"SOW_AGENTS_ARCHITECT", &config.Agents.Bindings.Architect},
		{"SOW_AGENTS_REVIEWER", &config.Agents.Bindings.Reviewer},
		{"SOW_AGENTS_PLANNER", &config.Agents.Bindings.Planner},
		{"SOW_AGENTS_RESEARCHER", &config.Agents.Bindings.Researcher},
		{"SOW_AGENTS_DECOMPOSER", &config.Agents.Bindings.Decomposer},
	}

	for _, ev := range envVars {
		if value := os.Getenv(ev.envVar); value != "" {
			// Create a copy of the value for the pointer
			valueCopy := value
			*ev.field = &valueCopy
		}
	}
}
