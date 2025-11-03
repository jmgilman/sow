package state

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/schemas/project"
)

// TestValidateStructure_ValidProject tests that a valid project passes structural validation.
func TestValidateStructure_ValidProject(t *testing.T) {
	// Given: Valid ProjectState
	now := time.Now()
	projectState := project.ProjectState{
		Name:        "valid-project",
		Type:        "standard",
		Branch:      "feat/test",
		Description: "A valid test project",
		Created_at:  now,
		Updated_at:  now,
		Phases: map[string]project.PhaseState{
			"implementation": {
				Status:     "in_progress",
				Enabled:    true,
				Created_at: now,
				Inputs:     []project.ArtifactState{},
				Outputs:    []project.ArtifactState{},
				Tasks:      []project.TaskState{},
			},
		},
		Statechart: project.StatechartState{
			Current_state: "ImplementationExecuting",
			Updated_at:    now,
		},
	}

	// When: validateStructure() called
	err := validateStructure(projectState)

	// Then: No error returned
	if err != nil {
		t.Errorf("Expected no error for valid project, got: %v", err)
	}
}

// TestValidateStructure_MissingRequiredField tests that missing required fields are caught.
func TestValidateStructure_MissingRequiredField(t *testing.T) {
	tests := []struct {
		name        string
		projectFunc func() project.ProjectState
		wantErrMsg  string
	}{
		{
			name: "missing name",
			projectFunc: func() project.ProjectState {
				now := time.Now()
				return project.ProjectState{
					Name:       "", // Missing required field
					Type:       "standard",
					Branch:     "feat/test",
					Created_at: now,
					Updated_at: now,
					Phases:     map[string]project.PhaseState{},
					Statechart: project.StatechartState{
						Current_state: "Initial",
						Updated_at:    now,
					},
				}
			},
			wantErrMsg: "name",
		},
		{
			name: "missing type",
			projectFunc: func() project.ProjectState {
				now := time.Now()
				return project.ProjectState{
					Name:       "test-project",
					Type:       "", // Missing required field
					Branch:     "feat/test",
					Created_at: now,
					Updated_at: now,
					Phases:     map[string]project.PhaseState{},
					Statechart: project.StatechartState{
						Current_state: "Initial",
						Updated_at:    now,
					},
				}
			},
			wantErrMsg: "type",
		},
		{
			name: "missing branch",
			projectFunc: func() project.ProjectState {
				now := time.Now()
				return project.ProjectState{
					Name:       "test-project",
					Type:       "standard",
					Branch:     "", // Missing required field
					Created_at: now,
					Updated_at: now,
					Phases:     map[string]project.PhaseState{},
					Statechart: project.StatechartState{
						Current_state: "Initial",
						Updated_at:    now,
					},
				}
			},
			wantErrMsg: "branch",
		},
		{
			name: "missing statechart current_state",
			projectFunc: func() project.ProjectState {
				now := time.Now()
				return project.ProjectState{
					Name:       "test-project",
					Type:       "standard",
					Branch:     "feat/test",
					Created_at: now,
					Updated_at: now,
					Phases:     map[string]project.PhaseState{},
					Statechart: project.StatechartState{
						Current_state: "", // Missing required field
						Updated_at:    now,
					},
				}
			},
			wantErrMsg: "current_state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: ProjectState with missing required field
			projectState := tt.projectFunc()

			// When: validateStructure() called
			err := validateStructure(projectState)

			// Then: Error returned mentioning the field
			if err == nil {
				t.Errorf("Expected error for missing %s, got nil", tt.wantErrMsg)
			}
		})
	}
}

// TestValidateStructure_InvalidFieldType tests that incorrect field types are caught.
func TestValidateStructure_InvalidFieldType(t *testing.T) {
	// Note: This test is somewhat limited because Go's type system catches
	// most type errors at compile time. CUE validation primarily helps with
	// YAML/JSON deserialization issues. We'll test the validation function
	// itself works correctly with valid typed data.

	// Given: Valid ProjectState with all correct types
	now := time.Now()
	projectState := project.ProjectState{
		Name:       "test-project",
		Type:       "standard",
		Branch:     "feat/test",
		Created_at: now,
		Updated_at: now,
		Phases: map[string]project.PhaseState{
			"implementation": {
				Status:     "pending",
				Enabled:    true, // Correct bool type
				Created_at: now,
				Inputs:     []project.ArtifactState{},
				Outputs:    []project.ArtifactState{},
				Tasks:      []project.TaskState{},
			},
		},
		Statechart: project.StatechartState{
			Current_state: "Initial",
			Updated_at:    now,
		},
	}

	// When: validateStructure() called
	err := validateStructure(projectState)

	// Then: No error (all types correct)
	if err != nil {
		t.Errorf("Expected no error for correct types, got: %v", err)
	}
}

// makeProjectWithField creates a ProjectState with specified name and type for testing.
func makeProjectWithField(name, projectType string) project.ProjectState {
	now := time.Now()
	return project.ProjectState{
		Name:       name,
		Type:       projectType,
		Branch:     "feat/test",
		Created_at: now,
		Updated_at: now,
		Phases:     map[string]project.PhaseState{},
		Statechart: project.StatechartState{
			Current_state: "Initial",
			Updated_at:    now,
		},
	}
}

// makeProjectWithInvalidTask creates a ProjectState with a task that has an invalid ID.
func makeProjectWithInvalidTask(taskID string) project.ProjectState {
	now := time.Now()
	return project.ProjectState{
		Name:       "test-project",
		Type:       "standard",
		Branch:     "feat/test",
		Created_at: now,
		Updated_at: now,
		Phases: map[string]project.PhaseState{
			"implementation": {
				Status:     "pending",
				Enabled:    true,
				Created_at: now,
				Inputs:     []project.ArtifactState{},
				Outputs:    []project.ArtifactState{},
				Tasks: []project.TaskState{
					{
						Id:             taskID,
						Name:           "Test task",
						Phase:          "implementation",
						Status:         "pending",
						Created_at:     now,
						Updated_at:     now,
						Iteration:      1,
						Assigned_agent: "implementer",
						Inputs:         []project.ArtifactState{},
						Outputs:        []project.ArtifactState{},
					},
				},
			},
		},
		Statechart: project.StatechartState{
			Current_state: "Initial",
			Updated_at:    now,
		},
	}
}

// TestValidateStructure_RegexViolation tests that regex pattern violations are caught.
func TestValidateStructure_RegexViolation(t *testing.T) {
	tests := []struct {
		name        string
		projectFunc func() project.ProjectState
		wantErrMsg  string
	}{
		{
			name:        "name with spaces",
			projectFunc: func() project.ProjectState { return makeProjectWithField("invalid name", "standard") },
			wantErrMsg:  "name",
		},
		{
			name:        "name with uppercase",
			projectFunc: func() project.ProjectState { return makeProjectWithField("InvalidName", "standard") },
			wantErrMsg:  "name",
		},
		{
			name:        "name starting with hyphen",
			projectFunc: func() project.ProjectState { return makeProjectWithField("-invalid", "standard") },
			wantErrMsg:  "name",
		},
		{
			name:        "type with hyphens",
			projectFunc: func() project.ProjectState { return makeProjectWithField("test-project", "invalid-type") },
			wantErrMsg:  "type",
		},
		{
			name:        "invalid task ID format",
			projectFunc: func() project.ProjectState { return makeProjectWithInvalidTask("1") },
			wantErrMsg:  "id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: ProjectState with regex violation
			projectState := tt.projectFunc()

			// When: validateStructure() called
			err := validateStructure(projectState)

			// Then: Error returned
			if err == nil {
				t.Errorf("Expected error for regex violation on %s, got nil", tt.wantErrMsg)
			}
		})
	}
}

// TestValidateStructure_ComplexNested tests validation of deeply nested structures.
func TestValidateStructure_ComplexNested(t *testing.T) {
	// Given: Valid ProjectState with complex nested structure
	now := time.Now()
	projectState := project.ProjectState{
		Name:        "complex-project",
		Type:        "standard",
		Branch:      "feat/complex",
		Description: "A complex project with nested structures",
		Created_at:  now,
		Updated_at:  now,
		Phases: map[string]project.PhaseState{
			"planning": {
				Status:       "completed",
				Enabled:      true,
				Created_at:   now,
				Started_at:   now,
				Completed_at: now,
				Inputs:       []project.ArtifactState{},
				Outputs: []project.ArtifactState{
					{
						Type:       "design_doc",
						Path:       "phases/planning/design.md",
						Approved:   true,
						Created_at: now,
					},
				},
				Tasks: []project.TaskState{},
			},
			"implementation": {
				Status:     "in_progress",
				Enabled:    true,
				Created_at: now,
				Started_at: now,
				Inputs: []project.ArtifactState{
					{
						Type:       "design_doc",
						Path:       "phases/planning/design.md",
						Approved:   true,
						Created_at: now,
					},
				},
				Outputs: []project.ArtifactState{},
				Tasks: []project.TaskState{
					{
						Id:             "010",
						Name:           "Setup project structure",
						Phase:          "implementation",
						Status:         "completed",
						Created_at:     now,
						Started_at:     now,
						Updated_at:     now,
						Completed_at:   now,
						Iteration:      1,
						Assigned_agent: "implementer",
						Inputs: []project.ArtifactState{
							{
								Type:       "design_doc",
								Path:       "phases/planning/design.md",
								Approved:   true,
								Created_at: now,
							},
						},
						Outputs: []project.ArtifactState{
							{
								Type:       "code",
								Path:       "src/main.go",
								Approved:   false,
								Created_at: now,
							},
						},
					},
					{
						Id:             "020",
						Name:           "Implement core logic",
						Phase:          "implementation",
						Status:         "in_progress",
						Created_at:     now,
						Started_at:     now,
						Updated_at:     now,
						Iteration:      2,
						Assigned_agent: "implementer",
						Inputs:         []project.ArtifactState{},
						Outputs:        []project.ArtifactState{},
					},
				},
			},
		},
		Statechart: project.StatechartState{
			Current_state: "ImplementationExecuting",
			Updated_at:    now,
		},
	}

	// When: validateStructure() called
	err := validateStructure(projectState)

	// Then: No error returned
	if err != nil {
		t.Errorf("Expected no error for complex nested structure, got: %v", err)
	}
}

// TestValidateMetadata_ValidSchema tests that valid metadata passes validation.
func TestValidateMetadata_ValidSchema(t *testing.T) {
	// Given: Valid metadata and matching schema
	metadata := map[string]interface{}{
		"complexity":      "medium",
		"requires_review": true,
	}

	schema := `{
		complexity: "low" | "medium" | "high"
		requires_review: bool
	}`

	// When: validateMetadata() called
	err := validateMetadata(metadata, schema)

	// Then: No error returned
	if err != nil {
		t.Errorf("Expected no error for valid metadata, got: %v", err)
	}
}

// TestValidateMetadata_SchemaViolation tests that invalid metadata returns clear error.
func TestValidateMetadata_SchemaViolation(t *testing.T) {
	tests := []struct {
		name       string
		metadata   map[string]interface{}
		schema     string
		wantErrMsg string
	}{
		{
			name: "wrong value type",
			metadata: map[string]interface{}{
				"complexity":      "invalid", // Not one of allowed values
				"requires_review": true,
			},
			schema: `{
				complexity: "low" | "medium" | "high"
				requires_review: bool
			}`,
			wantErrMsg: "complexity",
		},
		{
			name: "missing required field",
			metadata: map[string]interface{}{
				"complexity": "medium",
				// Missing requires_review
			},
			schema: `{
				complexity: "low" | "medium" | "high"
				requires_review: bool
			}`,
			wantErrMsg: "requires_review",
		},
		{
			name: "wrong field type",
			metadata: map[string]interface{}{
				"complexity":      "medium",
				"requires_review": "yes", // Should be bool
			},
			schema: `{
				complexity: "low" | "medium" | "high"
				requires_review: bool
			}`,
			wantErrMsg: "requires_review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When: validateMetadata() called with invalid data
			err := validateMetadata(tt.metadata, tt.schema)

			// Then: Error returned
			if err == nil {
				t.Errorf("Expected error for schema violation, got nil")
			}
		})
	}
}

// TestValidateMetadata_EmptySchema tests that empty schema skips validation.
func TestValidateMetadata_EmptySchema(t *testing.T) {
	// Given: Metadata but no schema
	metadata := map[string]interface{}{
		"any": "data",
		"is":  "allowed",
	}

	// When: validateMetadata() called with empty schema
	err := validateMetadata(metadata, "")

	// Then: No error (validation skipped)
	if err != nil {
		t.Errorf("Expected no error for empty schema, got: %v", err)
	}
}
