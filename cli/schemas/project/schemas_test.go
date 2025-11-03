package project

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

// Helper function to load and validate data against a schema definition.
func validateSchema(t *testing.T, schemaPath string, data map[string]any) error {
	t.Helper()

	ctx := cuecontext.New()

	// Load the project package - test_helper.cue ensures all schemas are referenced
	instances := load.Instances([]string{"."}, &load.Config{
		Dir: ".",
	})
	if len(instances) == 0 {
		t.Fatal("failed to load schema instances")
	}
	if instances[0].Err != nil {
		t.Fatalf("failed to load schema: %v", instances[0].Err)
	}

	schema := ctx.BuildInstance(instances[0])
	if schema.Err() != nil {
		t.Fatalf("failed to build schema: %v", schema.Err())
	}

	// Lookup the specific schema
	schemaType := schema.LookupPath(cue.ParsePath(schemaPath))
	if schemaType.Err() != nil {
		t.Fatalf("failed to lookup %s: %v", schemaPath, schemaType.Err())
	}

	// Convert test data to CUE value
	dataJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	dataValue := ctx.CompileBytes(dataJSON)
	if dataValue.Err() != nil {
		t.Fatalf("failed to compile test data: %v", dataValue.Err())
	}

	// Validate
	unified := schemaType.Unify(dataValue)
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}
	return nil
}

// ============================================================
// ProjectState Tests
// ============================================================

func TestValidProjectState(t *testing.T) {
	data := map[string]any{
		"name":       "test-project",
		"type":       "standard",
		"branch":     "main",
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"phases":     map[string]any{},
		"statechart": map[string]any{
			"current_state": "NoProject",
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err != nil {
		t.Errorf("valid project should pass validation: %v", err)
	}
}

func TestValidProjectState_WithDescription(t *testing.T) {
	data := map[string]any{
		"name":        "test-project",
		"type":        "standard",
		"branch":      "feat/test",
		"description": "A test project description",
		"created_at":  time.Now().Format(time.RFC3339),
		"updated_at":  time.Now().Format(time.RFC3339),
		"phases":      map[string]any{},
		"statechart": map[string]any{
			"current_state": "NoProject",
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err != nil {
		t.Errorf("project with description should pass validation: %v", err)
	}
}

func TestValidProjectNames(t *testing.T) {
	validNames := []string{
		"test",
		"my-project",
		"proj123",
		"a1",
		"test-123-project",
		"123project456",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			data := map[string]any{
				"name":       name,
				"type":       "standard",
				"branch":     "main",
				"created_at": time.Now().Format(time.RFC3339),
				"updated_at": time.Now().Format(time.RFC3339),
				"phases":     map[string]any{},
				"statechart": map[string]any{
					"current_state": "NoProject",
					"updated_at":    time.Now().Format(time.RFC3339),
				},
			}

			err := validateSchema(t, "#ProjectState", data)
			if err != nil {
				t.Errorf("valid project name '%s' should pass validation: %v", name, err)
			}
		})
	}
}

func TestInvalidProjectName_StartsWithHyphen(t *testing.T) {
	data := map[string]any{
		"name":       "-project",
		"type":       "standard",
		"branch":     "main",
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"phases":     map[string]any{},
		"statechart": map[string]any{
			"current_state": "NoProject",
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err == nil {
		t.Error("project name starting with hyphen should fail validation")
	}
}

func TestInvalidProjectName_EndsWithHyphen(t *testing.T) {
	data := map[string]any{
		"name":       "project-",
		"type":       "standard",
		"branch":     "main",
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"phases":     map[string]any{},
		"statechart": map[string]any{
			"current_state": "NoProject",
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err == nil {
		t.Error("project name ending with hyphen should fail validation")
	}
}

func TestInvalidProjectName_ContainsUppercase(t *testing.T) {
	data := map[string]any{
		"name":       "MyProject",
		"type":       "standard",
		"branch":     "main",
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"phases":     map[string]any{},
		"statechart": map[string]any{
			"current_state": "NoProject",
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err == nil {
		t.Error("project name with uppercase should fail validation")
	}
}

func TestInvalidProjectName_Empty(t *testing.T) {
	data := map[string]any{
		"name":       "",
		"type":       "standard",
		"branch":     "main",
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"phases":     map[string]any{},
		"statechart": map[string]any{
			"current_state": "NoProject",
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err == nil {
		t.Error("empty project name should fail validation")
	}
}

func TestInvalidProjectName_SingleCharWithHyphen(t *testing.T) {
	data := map[string]any{
		"name":       "a-",
		"type":       "standard",
		"branch":     "main",
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"phases":     map[string]any{},
		"statechart": map[string]any{
			"current_state": "NoProject",
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err == nil {
		t.Error("project name ending with hyphen should fail validation")
	}
}

func TestValidProjectTypes(t *testing.T) {
	validTypes := []string{
		"standard",
		"exploration",
		"design",
		"breakdown",
		"custom_type",
		"type123",
	}

	for _, typ := range validTypes {
		t.Run(typ, func(t *testing.T) {
			data := map[string]any{
				"name":       "test",
				"type":       typ,
				"branch":     "main",
				"created_at": time.Now().Format(time.RFC3339),
				"updated_at": time.Now().Format(time.RFC3339),
				"phases":     map[string]any{},
				"statechart": map[string]any{
					"current_state": "NoProject",
					"updated_at":    time.Now().Format(time.RFC3339),
				},
			}

			err := validateSchema(t, "#ProjectState", data)
			if err != nil {
				t.Errorf("valid project type '%s' should pass validation: %v", typ, err)
			}
		})
	}
}

func TestInvalidProjectType_ContainsHyphen(t *testing.T) {
	data := map[string]any{
		"name":       "test",
		"type":       "my-type",
		"branch":     "main",
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"phases":     map[string]any{},
		"statechart": map[string]any{
			"current_state": "NoProject",
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err == nil {
		t.Error("project type with hyphen should fail validation")
	}
}

func TestInvalidProjectType_Empty(t *testing.T) {
	data := map[string]any{
		"name":       "test",
		"type":       "",
		"branch":     "main",
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"phases":     map[string]any{},
		"statechart": map[string]any{
			"current_state": "NoProject",
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err == nil {
		t.Error("empty project type should fail validation")
	}
}

func TestInvalidProjectState_EmptyBranch(t *testing.T) {
	data := map[string]any{
		"name":       "test",
		"type":       "standard",
		"branch":     "",
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"phases":     map[string]any{},
		"statechart": map[string]any{
			"current_state": "NoProject",
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err == nil {
		t.Error("empty branch should fail validation")
	}
}

func TestInvalidProjectState_MissingCreatedAt(t *testing.T) {
	data := map[string]any{
		"name":       "test",
		"type":       "standard",
		"branch":     "main",
		"updated_at": time.Now().Format(time.RFC3339),
		"phases":     map[string]any{},
		"statechart": map[string]any{
			"current_state": "NoProject",
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err == nil {
		t.Error("missing created_at should fail validation")
	}
}

func TestInvalidProjectState_MissingUpdatedAt(t *testing.T) {
	data := map[string]any{
		"name":       "test",
		"type":       "standard",
		"branch":     "main",
		"created_at": time.Now().Format(time.RFC3339),
		"phases":     map[string]any{},
		"statechart": map[string]any{
			"current_state": "NoProject",
			"updated_at":    time.Now().Format(time.RFC3339),
		},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err == nil {
		t.Error("missing updated_at should fail validation")
	}
}

// Note: phases has a default value ({}) in CUE, so it can be omitted
// and will be filled in with an empty map. This is correct behavior.

func TestInvalidProjectState_MissingStatechart(t *testing.T) {
	data := map[string]any{
		"name":       "test",
		"type":       "standard",
		"branch":     "main",
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"phases":     map[string]any{},
	}

	err := validateSchema(t, "#ProjectState", data)
	if err == nil {
		t.Error("missing statechart should fail validation")
	}
}

// ============================================================
// StatechartState Tests
// ============================================================

func TestValidStatechartState(t *testing.T) {
	data := map[string]any{
		"current_state": "NoProject",
		"updated_at":    time.Now().Format(time.RFC3339),
	}

	err := validateSchema(t, "#StatechartState", data)
	if err != nil {
		t.Errorf("valid statechart should pass validation: %v", err)
	}
}

func TestValidStatechartState_VariousStates(t *testing.T) {
	states := []string{
		"NoProject",
		"Planning",
		"Implementation",
		"Review",
		"Completed",
	}

	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			data := map[string]any{
				"current_state": state,
				"updated_at":    time.Now().Format(time.RFC3339),
			}

			err := validateSchema(t, "#StatechartState", data)
			if err != nil {
				t.Errorf("valid state '%s' should pass validation: %v", state, err)
			}
		})
	}
}

func TestInvalidStatechart_EmptyCurrentState(t *testing.T) {
	data := map[string]any{
		"current_state": "",
		"updated_at":    time.Now().Format(time.RFC3339),
	}

	err := validateSchema(t, "#StatechartState", data)
	if err == nil {
		t.Error("empty current_state should fail validation")
	}
}

func TestInvalidStatechart_MissingUpdatedAt(t *testing.T) {
	data := map[string]any{
		"current_state": "NoProject",
	}

	err := validateSchema(t, "#StatechartState", data)
	if err == nil {
		t.Error("missing updated_at should fail validation")
	}
}

// ============================================================
// PhaseState Tests
// ============================================================

func TestValidPhaseState(t *testing.T) {
	data := map[string]any{
		"status":     "pending",
		"enabled":    true,
		"created_at": time.Now().Format(time.RFC3339),
		"inputs":     []any{},
		"outputs":    []any{},
		"tasks":      []any{},
	}

	err := validateSchema(t, "#PhaseState", data)
	if err != nil {
		t.Errorf("valid phase should pass validation: %v", err)
	}
}

func TestValidPhaseState_WithOptionalFields(t *testing.T) {
	data := map[string]any{
		"status":       "in_progress",
		"enabled":      true,
		"created_at":   time.Now().Format(time.RFC3339),
		"started_at":   time.Now().Format(time.RFC3339),
		"completed_at": time.Now().Format(time.RFC3339),
		"metadata": map[string]any{
			"key1": "value1",
			"key2": 123,
			"nested": map[string]any{
				"inner": "data",
			},
		},
		"inputs":  []any{},
		"outputs": []any{},
		"tasks":   []any{},
	}

	err := validateSchema(t, "#PhaseState", data)
	if err != nil {
		t.Errorf("phase with optional fields should pass validation: %v", err)
	}
}

func TestValidPhaseState_WithPopulatedCollections(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"status":     "in_progress",
		"enabled":    true,
		"created_at": now,
		"inputs": []any{
			map[string]any{
				"type":       "task_list",
				"path":       "input.md",
				"approved":   true,
				"created_at": now,
			},
		},
		"outputs": []any{
			map[string]any{
				"type":       "review",
				"path":       "output.md",
				"approved":   false,
				"created_at": now,
			},
		},
		"tasks": []any{
			map[string]any{
				"id":             "001",
				"name":           "Test Task",
				"phase":          "implementation",
				"status":         "pending",
				"created_at":     now,
				"updated_at":     now,
				"iteration":      1,
				"assigned_agent": "implementer",
				"inputs":         []any{},
				"outputs":        []any{},
			},
		},
	}

	err := validateSchema(t, "#PhaseState", data)
	if err != nil {
		t.Errorf("phase with populated collections should pass validation: %v", err)
	}
}

func TestInvalidPhaseState_EmptyStatus(t *testing.T) {
	data := map[string]any{
		"status":     "",
		"enabled":    true,
		"created_at": time.Now().Format(time.RFC3339),
		"inputs":     []any{},
		"outputs":    []any{},
		"tasks":      []any{},
	}

	err := validateSchema(t, "#PhaseState", data)
	if err == nil {
		t.Error("empty status should fail validation")
	}
}

func TestInvalidPhaseState_MissingEnabled(t *testing.T) {
	data := map[string]any{
		"status":     "pending",
		"created_at": time.Now().Format(time.RFC3339),
		"inputs":     []any{},
		"outputs":    []any{},
		"tasks":      []any{},
	}

	err := validateSchema(t, "#PhaseState", data)
	if err == nil {
		t.Error("missing enabled field should fail validation")
	}
}

func TestInvalidPhaseState_MissingCreatedAt(t *testing.T) {
	data := map[string]any{
		"status":  "pending",
		"enabled": true,
		"inputs":  []any{},
		"outputs": []any{},
		"tasks":   []any{},
	}

	err := validateSchema(t, "#PhaseState", data)
	if err == nil {
		t.Error("missing created_at should fail validation")
	}
}

// Note: inputs, outputs, and tasks arrays have default values ([])
// in CUE, so they can be omitted and will be filled in with empty arrays.
// This is correct behavior per the schema definition.

// ============================================================
// ArtifactState Tests
// ============================================================

func TestValidArtifactState(t *testing.T) {
	data := map[string]any{
		"type":       "task_list",
		"path":       "tasks.md",
		"approved":   true,
		"created_at": time.Now().Format(time.RFC3339),
	}

	err := validateSchema(t, "#ArtifactState", data)
	if err != nil {
		t.Errorf("valid artifact should pass validation: %v", err)
	}
}

func TestValidArtifactState_WithMetadata(t *testing.T) {
	data := map[string]any{
		"type":       "review",
		"path":       "review.md",
		"approved":   false,
		"created_at": time.Now().Format(time.RFC3339),
		"metadata": map[string]any{
			"assessment": "needs_revision",
			"score":      7,
		},
	}

	err := validateSchema(t, "#ArtifactState", data)
	if err != nil {
		t.Errorf("artifact with metadata should pass validation: %v", err)
	}
}

func TestValidArtifactState_VariousTypes(t *testing.T) {
	types := []string{
		"task_list",
		"review",
		"design_doc",
		"adr",
		"test_report",
	}

	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			data := map[string]any{
				"type":       typ,
				"path":       "test.md",
				"approved":   true,
				"created_at": time.Now().Format(time.RFC3339),
			}

			err := validateSchema(t, "#ArtifactState", data)
			if err != nil {
				t.Errorf("artifact type '%s' should pass validation: %v", typ, err)
			}
		})
	}
}

func TestValidArtifactState_ApprovedFalse(t *testing.T) {
	data := map[string]any{
		"type":       "review",
		"path":       "test.md",
		"approved":   false,
		"created_at": time.Now().Format(time.RFC3339),
	}

	err := validateSchema(t, "#ArtifactState", data)
	if err != nil {
		t.Errorf("artifact with approved=false should pass validation: %v", err)
	}
}

func TestInvalidArtifactState_EmptyType(t *testing.T) {
	data := map[string]any{
		"type":       "",
		"path":       "test.md",
		"approved":   true,
		"created_at": time.Now().Format(time.RFC3339),
	}

	err := validateSchema(t, "#ArtifactState", data)
	if err == nil {
		t.Error("empty type should fail validation")
	}
}

func TestInvalidArtifactState_EmptyPath(t *testing.T) {
	data := map[string]any{
		"type":       "task_list",
		"path":       "",
		"approved":   true,
		"created_at": time.Now().Format(time.RFC3339),
	}

	err := validateSchema(t, "#ArtifactState", data)
	if err == nil {
		t.Error("empty path should fail validation")
	}
}

func TestInvalidArtifactState_MissingApproved(t *testing.T) {
	data := map[string]any{
		"type":       "task_list",
		"path":       "test.md",
		"created_at": time.Now().Format(time.RFC3339),
	}

	err := validateSchema(t, "#ArtifactState", data)
	if err == nil {
		t.Error("missing approved field should fail validation")
	}
}

func TestInvalidArtifactState_MissingCreatedAt(t *testing.T) {
	data := map[string]any{
		"type":     "task_list",
		"path":     "test.md",
		"approved": true,
	}

	err := validateSchema(t, "#ArtifactState", data)
	if err == nil {
		t.Error("missing created_at should fail validation")
	}
}

// ============================================================
// TaskState Tests
// ============================================================

func TestValidTaskState(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "001",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "pending",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err != nil {
		t.Errorf("valid task should pass validation: %v", err)
	}
}

func TestValidTaskState_AllStatuses(t *testing.T) {
	statuses := []string{
		"pending",
		"in_progress",
		"completed",
		"abandoned",
	}

	now := time.Now().Format(time.RFC3339)
	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			data := map[string]any{
				"id":             "001",
				"name":           "Test Task",
				"phase":          "implementation",
				"status":         status,
				"created_at":     now,
				"updated_at":     now,
				"iteration":      1,
				"assigned_agent": "implementer",
				"inputs":         []any{},
				"outputs":        []any{},
			}

			err := validateSchema(t, "#TaskState", data)
			if err != nil {
				t.Errorf("task with status '%s' should pass validation: %v", status, err)
			}
		})
	}
}

func TestValidTaskState_WithOptionalFields(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "042",
		"name":           "Complex Task",
		"phase":          "design",
		"status":         "completed",
		"created_at":     now,
		"started_at":     now,
		"updated_at":     now,
		"completed_at":   now,
		"iteration":      3,
		"assigned_agent": "architect",
		"inputs":         []any{},
		"outputs":        []any{},
		"metadata": map[string]any{
			"github_issue_url":    "https://github.com/org/repo/issues/42",
			"github_issue_number": 42,
			"custom_field":        "value",
		},
	}

	err := validateSchema(t, "#TaskState", data)
	if err != nil {
		t.Errorf("task with optional fields should pass validation: %v", err)
	}
}

func TestValidTaskState_VariousIDs(t *testing.T) {
	ids := []string{
		"001",
		"042",
		"123",
		"999",
		"000",
	}

	now := time.Now().Format(time.RFC3339)
	for _, id := range ids {
		t.Run(id, func(t *testing.T) {
			data := map[string]any{
				"id":             id,
				"name":           "Test Task",
				"phase":          "implementation",
				"status":         "pending",
				"created_at":     now,
				"updated_at":     now,
				"iteration":      1,
				"assigned_agent": "implementer",
				"inputs":         []any{},
				"outputs":        []any{},
			}

			err := validateSchema(t, "#TaskState", data)
			if err != nil {
				t.Errorf("task ID '%s' should pass validation: %v", id, err)
			}
		})
	}
}

func TestValidTaskState_VariousIterations(t *testing.T) {
	iterations := []int{1, 2, 5, 10, 100}

	now := time.Now().Format(time.RFC3339)
	for _, iteration := range iterations {
		t.Run(string(rune(iteration)), func(t *testing.T) {
			data := map[string]any{
				"id":             "001",
				"name":           "Test Task",
				"phase":          "implementation",
				"status":         "pending",
				"created_at":     now,
				"updated_at":     now,
				"iteration":      iteration,
				"assigned_agent": "implementer",
				"inputs":         []any{},
				"outputs":        []any{},
			}

			err := validateSchema(t, "#TaskState", data)
			if err != nil {
				t.Errorf("task with iteration %d should pass validation: %v", iteration, err)
			}
		})
	}
}

func TestValidTaskState_WithArtifacts(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "001",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "completed",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs": []any{
			map[string]any{
				"type":       "design_doc",
				"path":       "design.md",
				"approved":   true,
				"created_at": now,
			},
		},
		"outputs": []any{
			map[string]any{
				"type":       "test_report",
				"path":       "tests.md",
				"approved":   false,
				"created_at": now,
			},
		},
	}

	err := validateSchema(t, "#TaskState", data)
	if err != nil {
		t.Errorf("task with artifacts should pass validation: %v", err)
	}
}

func TestInvalidTask_TwoDigitID(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "01",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "pending",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("two-digit task ID should fail validation")
	}
}

func TestInvalidTask_FourDigitID(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "0001",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "pending",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("four-digit task ID should fail validation")
	}
}

func TestInvalidTask_NonNumericID(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "abc",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "pending",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("non-numeric task ID should fail validation")
	}
}

func TestInvalidTask_EmptyID(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "pending",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("empty task ID should fail validation")
	}
}

func TestInvalidTask_InvalidStatus_Skipped(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "001",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "skipped",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("invalid status 'skipped' should fail validation")
	}
}

func TestInvalidTask_InvalidStatus_Blocked(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "001",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "blocked",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("invalid status 'blocked' should fail validation")
	}
}

func TestInvalidTask_EmptyStatus(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "001",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("empty status should fail validation")
	}
}

func TestInvalidTask_EmptyName(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "001",
		"name":           "",
		"phase":          "implementation",
		"status":         "pending",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("empty name should fail validation")
	}
}

func TestInvalidTask_EmptyPhase(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "001",
		"name":           "Test Task",
		"phase":          "",
		"status":         "pending",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("empty phase should fail validation")
	}
}

func TestInvalidTask_IterationZero(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "001",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "pending",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      0,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("iteration 0 should fail validation")
	}
}

func TestInvalidTask_IterationNegative(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "001",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "pending",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      -1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("negative iteration should fail validation")
	}
}

func TestInvalidTask_EmptyAssignedAgent(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "001",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "pending",
		"created_at":     now,
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("empty assigned_agent should fail validation")
	}
}

func TestInvalidTask_MissingCreatedAt(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "001",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "pending",
		"updated_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("missing created_at should fail validation")
	}
}

func TestInvalidTask_MissingUpdatedAt(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	data := map[string]any{
		"id":             "001",
		"name":           "Test Task",
		"phase":          "implementation",
		"status":         "pending",
		"created_at":     now,
		"iteration":      1,
		"assigned_agent": "implementer",
		"inputs":         []any{},
		"outputs":        []any{},
	}

	err := validateSchema(t, "#TaskState", data)
	if err == nil {
		t.Error("missing updated_at should fail validation")
	}
}

// Note: inputs and outputs arrays have default values ([]) in CUE,
// so they can be omitted and will be filled in with empty arrays.
// This is correct behavior per the schema definition.
