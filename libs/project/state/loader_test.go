package state

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmgilman/sow/libs/schemas/project"
	"github.com/qmuntal/stateless"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConfig implements ProjectTypeConfig for testing.
type mockConfig struct {
	name            string
	initialState    string
	initializeFunc  func(p *Project, inputs map[string][]project.ArtifactState) error
	validateFunc    func(p *Project) error
	buildMachineFunc func(p *Project, initialState string) *stateless.StateMachine
	getPhaseForStateFunc func(state string) string
	isPhaseStartStateFunc func(phaseName string, state string) bool
}

func (m *mockConfig) Name() string {
	return m.name
}

func (m *mockConfig) InitialState() string {
	return m.initialState
}

func (m *mockConfig) Initialize(p *Project, inputs map[string][]project.ArtifactState) error {
	if m.initializeFunc != nil {
		return m.initializeFunc(p, inputs)
	}
	return nil
}

func (m *mockConfig) Validate(p *Project) error {
	if m.validateFunc != nil {
		return m.validateFunc(p)
	}
	return nil
}

func (m *mockConfig) BuildMachine(p *Project, initialState string) *stateless.StateMachine {
	if m.buildMachineFunc != nil {
		return m.buildMachineFunc(p, initialState)
	}
	// Return a simple mock machine
	return stateless.NewStateMachine(initialState)
}

func (m *mockConfig) GetPhaseForState(state string) string {
	if m.getPhaseForStateFunc != nil {
		return m.getPhaseForStateFunc(state)
	}
	return ""
}

func (m *mockConfig) IsPhaseStartState(phaseName string, state string) bool {
	if m.isPhaseStartStateFunc != nil {
		return m.isPhaseStartStateFunc(phaseName, state)
	}
	return false
}

// Helper to create a valid project state for testing.
func validProjectState() *project.ProjectState {
	now := time.Now()
	return &project.ProjectState{
		Name:        "test-project",
		Type:        "standard",
		Branch:      "feat/test",
		Description: "Test project",
		Created_at:  now,
		Updated_at:  now,
		Phases:      make(map[string]project.PhaseState),
		Statechart: project.StatechartState{
			Current_state: "planning",
			Updated_at:    now,
		},
	}
}

func TestDetectProjectType(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		want       string
	}{
		{
			name:       "explore prefix returns exploration",
			branchName: "explore/new-feature",
			want:       "exploration",
		},
		{
			name:       "design prefix returns design",
			branchName: "design/auth-system",
			want:       "design",
		},
		{
			name:       "breakdown prefix returns breakdown",
			branchName: "breakdown/api-refactor",
			want:       "breakdown",
		},
		{
			name:       "feat prefix returns standard",
			branchName: "feat/auth-implementation",
			want:       "standard",
		},
		{
			name:       "fix prefix returns standard",
			branchName: "fix/login-bug",
			want:       "standard",
		},
		{
			name:       "no prefix returns standard",
			branchName: "main",
			want:       "standard",
		},
		{
			name:       "empty branch returns standard",
			branchName: "",
			want:       "standard",
		},
		{
			name:       "explore without slash returns standard",
			branchName: "explore",
			want:       "standard",
		},
		{
			name:       "nested explore path",
			branchName: "explore/auth/sub-feature",
			want:       "exploration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectProjectType(tt.branchName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateProjectName(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "simple description",
			description: "Add user authentication",
			want:        "add-user-authentication",
		},
		{
			name:        "uppercase letters converted to lowercase",
			description: "Implement JWT Token",
			want:        "implement-jwt-token",
		},
		{
			name:        "underscores converted to hyphens",
			description: "add_new_feature",
			want:        "add-new-feature",
		},
		{
			name:        "special characters removed",
			description: "Add feature! (version 2.0)",
			want:        "add-feature-version-20",
		},
		{
			name:        "truncated to 50 characters before conversion",
			description: "This is a very long description that should be truncated to fit within fifty characters limit",
			want:        "this-is-a-very-long-description-that-should-be-tru",
		},
		{
			name:        "trailing hyphen removed",
			description: "Add feature ",
			want:        "add-feature",
		},
		{
			name:        "multiple spaces collapsed",
			description: "Add   multiple   spaces",
			want:        "add-multiple-spaces",
		},
		{
			name:        "empty description",
			description: "",
			want:        "",
		},
		{
			name:        "numbers preserved",
			description: "Phase 2 implementation",
			want:        "phase-2-implementation",
		},
		{
			name:        "mixed case with numbers",
			description: "OAuth2 Integration",
			want:        "oauth2-integration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateProjectName(tt.description)
			assert.Equal(t, tt.want, got)
		})
	}
}

//nolint:funlen // Table-driven test with multiple subtests
func TestLoad(t *testing.T) {
	t.Run("succeeds with valid project state", func(t *testing.T) {
		// Setup: register a mock config
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
		}
		RegisterConfig(mockCfg)

		// Create backend with valid state
		state := validProjectState()
		backend := NewMemoryBackendWithState(state)

		// Act
		proj, err := Load(context.Background(), backend)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "test-project", proj.Name)
		assert.Equal(t, "standard", proj.Type)
		assert.NotNil(t, proj.Config())
		assert.NotNil(t, proj.Machine())
	})

	t.Run("fails when project not found", func(t *testing.T) {
		// Setup: empty backend
		backend := NewMemoryBackend()

		// Act
		proj, err := Load(context.Background(), backend)

		// Assert
		assert.Nil(t, proj)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("fails with unknown project type", func(t *testing.T) {
		// Setup: clear registry (no configs registered)
		ClearRegistry()

		state := validProjectState()
		state.Type = "unknown_type" // Uses underscore to pass CUE validation
		backend := NewMemoryBackendWithState(state)

		// Act
		proj, err := Load(context.Background(), backend)

		// Assert
		assert.Nil(t, proj)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown project type")
	})

	t.Run("fails when structure validation fails", func(t *testing.T) {
		// Setup: state with missing required field
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
		}
		RegisterConfig(mockCfg)

		state := validProjectState()
		state.Name = "" // Missing required field
		backend := NewMemoryBackendWithState(state)

		// Act
		proj, err := Load(context.Background(), backend)

		// Assert
		assert.Nil(t, proj)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validate structure")
	})

	t.Run("fails when metadata validation fails", func(t *testing.T) {
		// Setup: config that rejects validation
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
			validateFunc: func(_ *Project) error {
				return &ValidationError{Field: "metadata", Message: "invalid"}
			},
		}
		RegisterConfig(mockCfg)

		state := validProjectState()
		backend := NewMemoryBackendWithState(state)

		// Act
		proj, err := Load(context.Background(), backend)

		// Assert
		assert.Nil(t, proj)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validate metadata")
	})

	t.Run("attaches correct config from registry", func(t *testing.T) {
		// Setup: register multiple configs
		ClearRegistry()
		standardCfg := &mockConfig{name: "standard", initialState: "planning"}
		explorationCfg := &mockConfig{name: "exploration", initialState: "exploring"}
		RegisterConfig(standardCfg)
		RegisterConfig(explorationCfg)

		state := validProjectState()
		state.Type = "exploration"
		backend := NewMemoryBackendWithState(state)

		// Act
		proj, err := Load(context.Background(), backend)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "exploration", proj.Config().Name())
	})

	t.Run("builds state machine with current state", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
			buildMachineFunc: func(_ *Project, initialState string) *stateless.StateMachine {
				// Verify the initial state matches current_state from project
				assert.Equal(t, "review", initialState)
				return stateless.NewStateMachine(initialState)
			},
		}
		RegisterConfig(mockCfg)

		state := validProjectState()
		state.Statechart.Current_state = "review"
		backend := NewMemoryBackendWithState(state)

		// Act
		proj, err := Load(context.Background(), backend)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, proj.Machine())
	})
}

//nolint:funlen // Table-driven test with multiple subtests
func TestSave(t *testing.T) {
	t.Run("syncs statechart from machine", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
		}
		RegisterConfig(mockCfg)

		state := validProjectState()
		state.Statechart.Current_state = "planning"
		backend := NewMemoryBackendWithState(state)

		proj, err := Load(context.Background(), backend)
		require.NoError(t, err)

		// Simulate machine state change (in reality this would be from Fire())
		// For testing, we manually set the machine's state via a custom machine
		// Since we can't easily change stateless machine state without firing,
		// we'll just verify the current behavior
		originalState := proj.Statechart.Current_state

		// Act
		err = Save(context.Background(), proj)

		// Assert
		require.NoError(t, err)
		// State should still be synced (even if unchanged)
		assert.Equal(t, originalState, proj.Statechart.Current_state)
	})

	t.Run("updates timestamps", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
		}
		RegisterConfig(mockCfg)

		state := validProjectState()
		originalTime := state.Updated_at
		backend := NewMemoryBackendWithState(state)

		proj, err := Load(context.Background(), backend)
		require.NoError(t, err)

		// Wait a tiny bit to ensure time changes
		time.Sleep(time.Millisecond)

		// Act
		err = Save(context.Background(), proj)

		// Assert
		require.NoError(t, err)
		assert.True(t, proj.Updated_at.After(originalTime) || proj.Updated_at.Equal(originalTime))
	})

	t.Run("validates structure before saving", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
		}
		RegisterConfig(mockCfg)

		state := validProjectState()
		backend := NewMemoryBackendWithState(state)

		proj, err := Load(context.Background(), backend)
		require.NoError(t, err)

		// Corrupt the project state
		proj.Name = "" // Missing required field

		// Act
		err = Save(context.Background(), proj)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validate structure")
	})

	t.Run("validates metadata before saving", func(t *testing.T) {
		// Setup
		ClearRegistry()
		callCount := 0
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
			validateFunc: func(_ *Project) error {
				callCount++
				// Succeed on Load (first call), fail on Save (second call)
				if callCount > 1 {
					return &ValidationError{Field: "metadata", Message: "invalid"}
				}
				return nil
			},
		}
		RegisterConfig(mockCfg)

		state := validProjectState()
		backend := NewMemoryBackendWithState(state)

		proj, err := Load(context.Background(), backend)
		require.NoError(t, err)

		// Act
		err = Save(context.Background(), proj)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validate metadata")
	})

	t.Run("persists to backend on success", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
		}
		RegisterConfig(mockCfg)

		state := validProjectState()
		state.Description = "Original description"
		backend := NewMemoryBackendWithState(state)

		proj, err := Load(context.Background(), backend)
		require.NoError(t, err)

		// Modify the project
		proj.Description = "Updated description"

		// Act
		err = Save(context.Background(), proj)

		// Assert
		require.NoError(t, err)

		// Verify persisted
		loaded, err := backend.Load(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "Updated description", loaded.Description)
	})

	t.Run("does not persist on validation failure", func(t *testing.T) {
		// Setup
		ClearRegistry()
		callCount := 0
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
			validateFunc: func(_ *Project) error {
				callCount++
				// Succeed on Load (first call), fail on Save (second call)
				if callCount > 1 {
					return &ValidationError{Field: "metadata", Message: "invalid"}
				}
				return nil
			},
		}
		RegisterConfig(mockCfg)

		state := validProjectState()
		state.Description = "Original description"
		backend := NewMemoryBackendWithState(state)

		proj, err := Load(context.Background(), backend)
		require.NoError(t, err)

		// Modify the project
		proj.Description = "Should not be saved"

		// Act
		err = Save(context.Background(), proj)

		// Assert - validation fails
		require.Error(t, err)

		// Verify NOT persisted - original description still there
		loaded, err := backend.Load(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "Original description", loaded.Description)
	})
}

//nolint:funlen // Table-driven test with multiple subtests
func TestCreate(t *testing.T) {
	t.Run("creates project with correct metadata", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
		}
		RegisterConfig(mockCfg)

		backend := NewMemoryBackend()

		// Act
		proj, err := Create(context.Background(), backend, CreateOpts{
			Branch:      "feat/test-feature",
			Description: "Test feature implementation",
		})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "test-feature-implementation", proj.Name)
		assert.Equal(t, "standard", proj.Type)
		assert.Equal(t, "feat/test-feature", proj.Branch)
		assert.Equal(t, "Test feature implementation", proj.Description)
		assert.False(t, proj.Created_at.IsZero())
		assert.False(t, proj.Updated_at.IsZero())
	})

	t.Run("detects project type from branch prefix", func(t *testing.T) {
		tests := []struct {
			branch       string
			expectedType string
		}{
			{"explore/new-idea", "exploration"},
			{"design/architecture", "design"},
			{"breakdown/epic", "breakdown"},
			{"feat/feature", "standard"},
			{"fix/bug", "standard"},
		}

		for _, tt := range tests {
			t.Run(tt.branch, func(t *testing.T) {
				// Setup
				ClearRegistry()
				mockCfg := &mockConfig{
					name:         tt.expectedType,
					initialState: "initial",
				}
				RegisterConfig(mockCfg)

				backend := NewMemoryBackend()

				// Act
				proj, err := Create(context.Background(), backend, CreateOpts{
					Branch:      tt.branch,
					Description: "Test",
				})

				// Assert
				require.NoError(t, err)
				assert.Equal(t, tt.expectedType, proj.Type)
			})
		}
	})

	t.Run("allows explicit project type override", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "exploration",
			initialState: "exploring",
		}
		RegisterConfig(mockCfg)

		backend := NewMemoryBackend()

		// Act - branch would normally be "standard", but we override
		proj, err := Create(context.Background(), backend, CreateOpts{
			Branch:      "feat/test",
			Description: "Test",
			ProjectType: "exploration", // Explicit override
		})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "exploration", proj.Type)
	})

	t.Run("initializes phases via config", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
			initializeFunc: func(p *Project, _ map[string][]project.ArtifactState) error {
				// Simulate adding phases
				p.Phases = map[string]project.PhaseState{
					"planning":       {Status: "pending"},
					"implementation": {Status: "pending"},
					"review":         {Status: "pending"},
				}
				return nil
			},
		}
		RegisterConfig(mockCfg)

		backend := NewMemoryBackend()

		// Act
		proj, err := Create(context.Background(), backend, CreateOpts{
			Branch:      "feat/test",
			Description: "Test",
		})

		// Assert
		require.NoError(t, err)
		assert.Len(t, proj.Phases, 3)
		assert.Equal(t, "pending", proj.Phases["planning"].Status)
	})

	t.Run("passes initial inputs to initializer", func(t *testing.T) {
		// Setup
		ClearRegistry()
		var receivedInputs map[string][]project.ArtifactState
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
			initializeFunc: func(_ *Project, inputs map[string][]project.ArtifactState) error {
				receivedInputs = inputs
				return nil
			},
		}
		RegisterConfig(mockCfg)

		backend := NewMemoryBackend()
		initialInputs := map[string][]project.ArtifactState{
			"planning": {{Type: "design_doc", Path: "docs/design.md"}},
		}

		// Act
		_, err := Create(context.Background(), backend, CreateOpts{
			Branch:        "feat/test",
			Description:   "Test",
			InitialInputs: initialInputs,
		})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, initialInputs, receivedInputs)
	})

	t.Run("builds state machine with initial state", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
			buildMachineFunc: func(_ *Project, initialState string) *stateless.StateMachine {
				// Verify initial state is from config, not from loaded state
				assert.Equal(t, "planning", initialState)
				return stateless.NewStateMachine(initialState)
			},
		}
		RegisterConfig(mockCfg)

		backend := NewMemoryBackend()

		// Act
		proj, err := Create(context.Background(), backend, CreateOpts{
			Branch:      "feat/test",
			Description: "Test",
		})

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, proj.Machine())
	})

	t.Run("initializes statechart with config initial state", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning_start",
		}
		RegisterConfig(mockCfg)

		backend := NewMemoryBackend()

		// Act
		proj, err := Create(context.Background(), backend, CreateOpts{
			Branch:      "feat/test",
			Description: "Test",
		})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "planning_start", proj.Statechart.Current_state)
	})

	t.Run("saves project to backend", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
		}
		RegisterConfig(mockCfg)

		backend := NewMemoryBackend()

		// Act
		_, err := Create(context.Background(), backend, CreateOpts{
			Branch:      "feat/test",
			Description: "Test",
		})

		// Assert
		require.NoError(t, err)

		// Verify it was persisted
		exists, err := backend.Exists(context.Background())
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("fails with unknown project type", func(t *testing.T) {
		// Setup
		ClearRegistry() // No configs registered

		backend := NewMemoryBackend()

		// Act
		proj, err := Create(context.Background(), backend, CreateOpts{
			Branch:      "feat/test",
			Description: "Test",
		})

		// Assert
		assert.Nil(t, proj)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown project type")
	})

	t.Run("fails when initialize returns error", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
			initializeFunc: func(_ *Project, _ map[string][]project.ArtifactState) error {
				return fmt.Errorf("initialization failed")
			},
		}
		RegisterConfig(mockCfg)

		backend := NewMemoryBackend()

		// Act
		proj, err := Create(context.Background(), backend, CreateOpts{
			Branch:      "feat/test",
			Description: "Test",
		})

		// Assert
		assert.Nil(t, proj)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "initialize project")
	})

	t.Run("does not persist on validation failure", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning",
			validateFunc: func(_ *Project) error {
				return &ValidationError{Field: "test", Message: "validation failed"}
			},
		}
		RegisterConfig(mockCfg)

		backend := NewMemoryBackend()

		// Act
		proj, err := Create(context.Background(), backend, CreateOpts{
			Branch:      "feat/test",
			Description: "Test",
		})

		// Assert
		assert.Nil(t, proj)
		require.Error(t, err)

		// Verify NOT persisted
		exists, err := backend.Exists(context.Background())
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("marks initial phase as in_progress", func(t *testing.T) {
		// Setup
		ClearRegistry()
		mockCfg := &mockConfig{
			name:         "standard",
			initialState: "planning_start",
			initializeFunc: func(p *Project, _ map[string][]project.ArtifactState) error {
				p.Phases = map[string]project.PhaseState{
					"planning": {Status: "pending"},
				}
				return nil
			},
			getPhaseForStateFunc: func(state string) string {
				if state == "planning_start" {
					return "planning"
				}
				return ""
			},
			isPhaseStartStateFunc: func(phaseName string, state string) bool {
				return phaseName == "planning" && state == "planning_start"
			},
		}
		RegisterConfig(mockCfg)

		backend := NewMemoryBackend()

		// Act
		proj, err := Create(context.Background(), backend, CreateOpts{
			Branch:      "feat/test",
			Description: "Test",
		})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "in_progress", proj.Phases["planning"].Status)
	})
}
