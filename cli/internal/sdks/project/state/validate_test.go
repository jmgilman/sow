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

// TestValidateArtifactTypesAllowed tests that artifacts pass when type is in allowed list.
func TestValidateArtifactTypesAllowed(t *testing.T) {
	// Given: Artifacts with types in allowed list
	artifacts := []project.ArtifactState{
		{Type: "task_list"},
		{Type: "design"},
	}
	allowed := []string{"task_list", "design", "review"}

	// When: validateArtifactTypes() called
	err := validateArtifactTypes(artifacts, allowed, "planning", "output")

	// Then: No error returned
	if err != nil {
		t.Errorf("Expected no error for allowed types, got: %v", err)
	}
}

// TestValidateArtifactTypesRejects tests that artifacts fail when type not in allowed list.
func TestValidateArtifactTypesRejects(t *testing.T) {
	// Given: Artifact with type not in allowed list
	artifacts := []project.ArtifactState{
		{Type: "invalid_type"},
	}
	allowed := []string{"task_list", "design"}

	// When: validateArtifactTypes() called
	err := validateArtifactTypes(artifacts, allowed, "planning", "output")

	// Then: Error returned mentioning the invalid type
	if err == nil {
		t.Error("Expected error for invalid artifact type")
	}
	if err != nil && !contains(err.Error(), "invalid_type") {
		t.Errorf("Error should mention invalid type, got: %v", err)
	}
	if err != nil && !contains(err.Error(), "planning") {
		t.Errorf("Error should mention phase name, got: %v", err)
	}
}

// TestValidateArtifactTypesEmptyAllowed tests that empty allowed list allows all types.
func TestValidateArtifactTypesEmptyAllowed(t *testing.T) {
	// Given: Artifacts and empty allowed list
	artifacts := []project.ArtifactState{
		{Type: "any_type"},
		{Type: "another_type"},
	}
	allowed := []string{} // Empty = allow all

	// When: validateArtifactTypes() called
	err := validateArtifactTypes(artifacts, allowed, "planning", "input")

	// Then: No error (all types allowed)
	if err != nil {
		t.Errorf("Expected no error for empty allowed list, got: %v", err)
	}
}

// TestValidateArtifactTypesMultiple tests validation of multiple artifacts.
func TestValidateArtifactTypesMultiple(t *testing.T) {
	// Given: Multiple artifacts, all with allowed types
	artifacts := []project.ArtifactState{
		{Type: "design"},
		{Type: "task_list"},
		{Type: "review"},
	}
	allowed := []string{"design", "task_list", "review", "adr"}

	// When: validateArtifactTypes() called
	err := validateArtifactTypes(artifacts, allowed, "implementation", "output")

	// Then: No error
	if err != nil {
		t.Errorf("Expected no error for multiple valid artifacts, got: %v", err)
	}
}

// TestValidateArtifactTypesErrorMessage tests error message includes all details.
func TestValidateArtifactTypesErrorMessage(t *testing.T) {
	// Given: Invalid artifact type
	artifacts := []project.ArtifactState{
		{Type: "bad_type"},
	}
	allowed := []string{"good_type1", "good_type2"}

	// When: validateArtifactTypes() called
	err := validateArtifactTypes(artifacts, allowed, "review", "input")

	// Then: Error includes phase, category, invalid type, and allowed types
	if err == nil {
		t.Fatal("Expected error for invalid type")
	}
	errMsg := err.Error()
	if !contains(errMsg, "review") {
		t.Errorf("Error should mention phase name 'review', got: %v", errMsg)
	}
	if !contains(errMsg, "input") {
		t.Errorf("Error should mention category 'input', got: %v", errMsg)
	}
	if !contains(errMsg, "bad_type") {
		t.Errorf("Error should mention invalid type 'bad_type', got: %v", errMsg)
	}
}

// TestProjectTypeConfigValidateAllPhases tests that Validate() checks all phases.
func TestProjectTypeConfigValidateAllPhases(t *testing.T) {
	// Given: Config with multiple phases and matching project
	now := time.Now()
	config := &ProjectTypeConfig{
		phaseConfigs: map[string]*PhaseConfig{
			"planning": {
				allowedOutputTypes: []string{"design"},
			},
			"implementation": {
				allowedInputTypes:  []string{"design"},
				allowedOutputTypes: []string{"code"},
			},
		},
	}

	projectState := &Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"planning": {
					Outputs: []project.ArtifactState{
						{Type: "design", Created_at: now},
					},
				},
				"implementation": {
					Inputs: []project.ArtifactState{
						{Type: "design", Created_at: now},
					},
					Outputs: []project.ArtifactState{
						{Type: "code", Created_at: now},
					},
				},
			},
		},
	}

	// When: Validate() called
	err := config.Validate(projectState)

	// Then: No error (all phases valid)
	if err != nil {
		t.Errorf("Expected no error for valid project, got: %v", err)
	}
}

// TestProjectTypeConfigValidateSkipsMissingPhases tests that phases not in state are skipped.
func TestProjectTypeConfigValidateSkipsMissingPhases(t *testing.T) {
	// Given: Config with phase not present in project state
	config := &ProjectTypeConfig{
		phaseConfigs: map[string]*PhaseConfig{
			"planning": {
				allowedOutputTypes: []string{"design"},
			},
			"review": {
				allowedInputTypes: []string{"code"},
			},
		},
	}

	projectState := &Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"planning": {
					Outputs: []project.ArtifactState{
						{Type: "design", Created_at: time.Now()},
					},
				},
				// "review" phase not present
			},
		},
	}

	// When: Validate() called
	err := config.Validate(projectState)

	// Then: No error (missing phase skipped)
	if err != nil {
		t.Errorf("Expected no error when phase missing from state, got: %v", err)
	}
}

// TestProjectTypeConfigValidateInputTypes tests input artifact type validation.
func TestProjectTypeConfigValidateInputTypes(t *testing.T) {
	// Given: Config with restricted input types and invalid input
	config := &ProjectTypeConfig{
		phaseConfigs: map[string]*PhaseConfig{
			"implementation": {
				allowedInputTypes: []string{"design", "task_list"},
			},
		},
	}

	projectState := &Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"implementation": {
					Inputs: []project.ArtifactState{
						{Type: "invalid_input", Created_at: time.Now()},
					},
				},
			},
		},
	}

	// When: Validate() called
	err := config.Validate(projectState)

	// Then: Error returned for invalid input type
	if err == nil {
		t.Error("Expected error for invalid input type")
	}
	if err != nil && !contains(err.Error(), "implementation") {
		t.Errorf("Error should mention phase name, got: %v", err)
	}
	if err != nil && !contains(err.Error(), "input") {
		t.Errorf("Error should mention input category, got: %v", err)
	}
}

// TestProjectTypeConfigValidateOutputTypes tests output artifact type validation.
func TestProjectTypeConfigValidateOutputTypes(t *testing.T) {
	// Given: Config with restricted output types and invalid output
	config := &ProjectTypeConfig{
		phaseConfigs: map[string]*PhaseConfig{
			"planning": {
				allowedOutputTypes: []string{"design", "task_list"},
			},
		},
	}

	projectState := &Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"planning": {
					Outputs: []project.ArtifactState{
						{Type: "code", Created_at: time.Now()}, // Not allowed
					},
				},
			},
		},
	}

	// When: Validate() called
	err := config.Validate(projectState)

	// Then: Error returned for invalid output type
	if err == nil {
		t.Error("Expected error for invalid output type")
	}
	if err != nil && !contains(err.Error(), "planning") {
		t.Errorf("Error should mention phase name, got: %v", err)
	}
	if err != nil && !contains(err.Error(), "output") {
		t.Errorf("Error should mention output category, got: %v", err)
	}
}

// TestProjectTypeConfigValidateMetadataWithSchema tests metadata validation with schema.
func TestProjectTypeConfigValidateMetadataWithSchema(t *testing.T) {
	// Given: Config with metadata schema and valid metadata
	config := &ProjectTypeConfig{
		phaseConfigs: map[string]*PhaseConfig{
			"planning": {
				metadataSchema: `{
					complexity: "low" | "medium" | "high"
				}`,
			},
		},
	}

	projectState := &Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"planning": {
					Metadata: map[string]interface{}{
						"complexity": "medium",
					},
				},
			},
		},
	}

	// When: Validate() called
	err := config.Validate(projectState)

	// Then: No error (metadata valid)
	if err != nil {
		t.Errorf("Expected no error for valid metadata, got: %v", err)
	}
}

// TestProjectTypeConfigValidateMetadataInvalid tests metadata validation failure.
func TestProjectTypeConfigValidateMetadataInvalid(t *testing.T) {
	// Given: Config with metadata schema and invalid metadata
	config := &ProjectTypeConfig{
		phaseConfigs: map[string]*PhaseConfig{
			"planning": {
				metadataSchema: `{
					complexity: "low" | "medium" | "high"
				}`,
			},
		},
	}

	projectState := &Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"planning": {
					Metadata: map[string]interface{}{
						"complexity": "invalid", // Not in allowed values
					},
				},
			},
		},
	}

	// When: Validate() called
	err := config.Validate(projectState)

	// Then: Error returned mentioning phase
	if err == nil {
		t.Error("Expected error for invalid metadata")
	}
	if err != nil && !contains(err.Error(), "planning") {
		t.Errorf("Error should mention phase name, got: %v", err)
	}
}

// TestProjectTypeConfigValidateRejectsUnexpectedMetadata tests rejection of metadata when no schema.
func TestProjectTypeConfigValidateRejectsUnexpectedMetadata(t *testing.T) {
	// Given: Config without metadata schema but project has metadata
	config := &ProjectTypeConfig{
		phaseConfigs: map[string]*PhaseConfig{
			"planning": {
				// No metadataSchema defined
			},
		},
	}

	projectState := &Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"planning": {
					Metadata: map[string]interface{}{
						"unexpected": "data",
					},
				},
			},
		},
	}

	// When: Validate() called
	err := config.Validate(projectState)

	// Then: Error returned
	if err == nil {
		t.Error("Expected error for unexpected metadata")
	}
	if err != nil && !contains(err.Error(), "does not support metadata") {
		t.Errorf("Error should mention metadata not supported, got: %v", err)
	}
}

// TestProjectTypeConfigValidateAllowsMissingMetadata tests that missing metadata is OK when no schema.
func TestProjectTypeConfigValidateAllowsMissingMetadata(t *testing.T) {
	// Given: Config without metadata schema and no metadata in project
	config := &ProjectTypeConfig{
		phaseConfigs: map[string]*PhaseConfig{
			"planning": {
				// No metadataSchema defined
			},
		},
	}

	projectState := &Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"planning": {
					Metadata: map[string]interface{}{}, // Empty metadata
				},
			},
		},
	}

	// When: Validate() called
	err := config.Validate(projectState)

	// Then: No error
	if err != nil {
		t.Errorf("Expected no error for missing metadata when no schema, got: %v", err)
	}
}

// TestProjectTypeConfigValidateFullIntegration tests complete validation scenario.
func TestProjectTypeConfigValidateFullIntegration(t *testing.T) {
	// Given: Full config with artifact types and metadata
	now := time.Now()
	config := &ProjectTypeConfig{
		phaseConfigs: map[string]*PhaseConfig{
			"planning": {
				allowedOutputTypes: []string{"task_list"},
				metadataSchema: `{
					complexity?: "low" | "medium" | "high"
				}`,
			},
		},
	}

	projectState := &Project{
		ProjectState: project.ProjectState{
			Phases: map[string]project.PhaseState{
				"planning": {
					Outputs: []project.ArtifactState{
						{Type: "task_list", Created_at: now},
					},
					Metadata: map[string]interface{}{
						"complexity": "medium",
					},
				},
			},
		},
	}

	// When: Validate() called
	err := config.Validate(projectState)

	// Then: No error (everything valid)
	if err != nil {
		t.Errorf("Expected valid project to pass validation, got: %v", err)
	}
}
