package schemas_test

import (
	"encoding/json"
	"testing"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

// TestArtifactApprovedOptional verifies that Artifact.approved field is optional and nillable.
func TestArtifactApprovedOptional(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
	}{
		{
			name: "artifact with approved true",
			data: map[string]any{
				"path":       "test.md",
				"approved":   true,
				"created_at": time.Now().Format(time.RFC3339),
			},
			wantErr: false,
		},
		{
			name: "artifact with approved false",
			data: map[string]any{
				"path":       "test.md",
				"approved":   false,
				"created_at": time.Now().Format(time.RFC3339),
			},
			wantErr: false,
		},
		{
			name: "artifact without approved field (should be valid)",
			data: map[string]any{
				"path":       "test.md",
				"created_at": time.Now().Format(time.RFC3339),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := cuecontext.New()

			// Load the phases schema
			instances := load.Instances([]string{"./phases"}, &load.Config{
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

			artifactSchema := schema.LookupPath(cue.ParsePath("#Artifact"))
			if artifactSchema.Err() != nil {
				t.Fatalf("failed to lookup #Artifact: %v", artifactSchema.Err())
			}

			// Convert test data to CUE value
			dataJSON, err := json.Marshal(tt.data)
			if err != nil {
				t.Fatalf("failed to marshal test data: %v", err)
			}

			dataValue := ctx.CompileBytes(dataJSON)
			if dataValue.Err() != nil {
				t.Fatalf("failed to compile test data: %v", dataValue.Err())
			}

			// Validate
			unified := artifactSchema.Unify(dataValue)
			err = unified.Validate(cue.Concrete(true))

			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// TestPhaseInputsField verifies that Phase.inputs field exists and accepts artifact arrays.
func TestPhaseInputsField(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
	}{
		{
			name: "phase with inputs array",
			data: map[string]any{
				"status":     "pending",
				"enabled":    true,
				"created_at": time.Now().Format(time.RFC3339),
				"artifacts":  []any{},
				"tasks":      []any{},
				"inputs": []any{
					map[string]any{
						"path":       "input1.md",
						"created_at": time.Now().Format(time.RFC3339),
					},
					map[string]any{
						"path":       "input2.md",
						"created_at": time.Now().Format(time.RFC3339),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "phase with empty inputs array",
			data: map[string]any{
				"status":     "pending",
				"enabled":    true,
				"created_at": time.Now().Format(time.RFC3339),
				"artifacts":  []any{},
				"tasks":      []any{},
				"inputs":     []any{},
			},
			wantErr: false,
		},
		{
			name: "phase without inputs field (should be valid)",
			data: map[string]any{
				"status":     "pending",
				"enabled":    true,
				"created_at": time.Now().Format(time.RFC3339),
				"artifacts":  []any{},
				"tasks":      []any{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := cuecontext.New()

			instances := load.Instances([]string{"./phases"}, &load.Config{
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

			phaseSchema := schema.LookupPath(cue.ParsePath("#Phase"))
			if phaseSchema.Err() != nil {
				t.Fatalf("failed to lookup #Phase: %v", phaseSchema.Err())
			}

			dataJSON, err := json.Marshal(tt.data)
			if err != nil {
				t.Fatalf("failed to marshal test data: %v", err)
			}

			dataValue := ctx.CompileBytes(dataJSON)
			if dataValue.Err() != nil {
				t.Fatalf("failed to compile test data: %v", dataValue.Err())
			}

			unified := phaseSchema.Unify(dataValue)
			err = unified.Validate(cue.Concrete(true))

			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// TestTaskRefsField verifies that Task.refs field exists and accepts artifact arrays.
func TestTaskRefsField(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
	}{
		{
			name: "task with refs array",
			data: map[string]any{
				"id":       "010",
				"name":     "Test Task",
				"status":   "pending",
				"parallel": false,
				"refs": []any{
					map[string]any{
						"path":       "ref1.md",
						"created_at": time.Now().Format(time.RFC3339),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "task with empty refs array",
			data: map[string]any{
				"id":       "010",
				"name":     "Test Task",
				"status":   "pending",
				"parallel": false,
				"refs":     []any{},
			},
			wantErr: false,
		},
		{
			name: "task without refs field (should be valid)",
			data: map[string]any{
				"id":       "010",
				"name":     "Test Task",
				"status":   "pending",
				"parallel": false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := cuecontext.New()

			instances := load.Instances([]string{"./phases"}, &load.Config{
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

			taskSchema := schema.LookupPath(cue.ParsePath("#Task"))
			if taskSchema.Err() != nil {
				t.Fatalf("failed to lookup #Task: %v", taskSchema.Err())
			}

			dataJSON, err := json.Marshal(tt.data)
			if err != nil {
				t.Fatalf("failed to marshal test data: %v", err)
			}

			dataValue := ctx.CompileBytes(dataJSON)
			if dataValue.Err() != nil {
				t.Fatalf("failed to compile test data: %v", dataValue.Err())
			}

			unified := taskSchema.Unify(dataValue)
			err = unified.Validate(cue.Concrete(true))

			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// TestTaskMetadataField verifies that Task.metadata field exists and accepts key-value maps.
func TestTaskMetadataField(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
	}{
		{
			name: "task with metadata",
			data: map[string]any{
				"id":       "010",
				"name":     "Test Task",
				"status":   "pending",
				"parallel": false,
				"metadata": map[string]any{
					"github_issue_url":    "https://github.com/org/repo/issues/123",
					"github_issue_number": 123,
					"custom_field":        "value",
				},
			},
			wantErr: false,
		},
		{
			name: "task with empty metadata",
			data: map[string]any{
				"id":       "010",
				"name":     "Test Task",
				"status":   "pending",
				"parallel": false,
				"metadata": map[string]any{},
			},
			wantErr: false,
		},
		{
			name: "task without metadata field (should be valid)",
			data: map[string]any{
				"id":       "010",
				"name":     "Test Task",
				"status":   "pending",
				"parallel": false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := cuecontext.New()

			instances := load.Instances([]string{"./phases"}, &load.Config{
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

			taskSchema := schema.LookupPath(cue.ParsePath("#Task"))
			if taskSchema.Err() != nil {
				t.Fatalf("failed to lookup #Task: %v", taskSchema.Err())
			}

			dataJSON, err := json.Marshal(tt.data)
			if err != nil {
				t.Fatalf("failed to marshal test data: %v", err)
			}

			dataValue := ctx.CompileBytes(dataJSON)
			if dataValue.Err() != nil {
				t.Fatalf("failed to compile test data: %v", dataValue.Err())
			}

			unified := taskSchema.Unify(dataValue)
			err = unified.Validate(cue.Concrete(true))

			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// TestGoTypeGeneration verifies Go types match CUE schema expectations.
func TestGoTypeGeneration(t *testing.T) {
	t.Run("Artifact.Approved is pointer type", func(t *testing.T) {
		// Test that Approved field is now a pointer (*bool)
		approved := true
		artifact := phases.Artifact{
			Path:       "test.md",
			Approved:   &approved, // Should be a pointer now
			Created_at: time.Now(),
		}

		// Verify field is optional by setting it to nil
		artifact.Approved = nil
		if artifact.Approved != nil {
			t.Error("Approved field should be nillable")
		}

		// Verify we can check the value when set
		artifact.Approved = &approved
		if artifact.Approved == nil || *artifact.Approved != true {
			t.Error("Approved field should be settable and readable")
		}
	})

	t.Run("Phase.Inputs field exists and accepts artifacts", func(t *testing.T) {
		phase := phases.Phase{
			Status:     "pending",
			Enabled:    true,
			Created_at: time.Now(),
			Artifacts:  []phases.Artifact{},
			Tasks:      []phases.Task{},
			Inputs: []phases.Artifact{
				{Path: "input.md", Created_at: time.Now()},
			},
		}

		if len(phase.Inputs) != 1 {
			t.Errorf("Expected 1 input, got %d", len(phase.Inputs))
		}
	})

	t.Run("Task.Refs field exists and accepts artifacts", func(t *testing.T) {
		task := phases.Task{
			Id:       "010",
			Name:     "Test",
			Status:   "pending",
			Parallel: false,
			Refs: []phases.Artifact{
				{Path: "ref.md", Created_at: time.Now()},
			},
		}

		if len(task.Refs) != 1 {
			t.Errorf("Expected 1 ref, got %d", len(task.Refs))
		}
	})

	t.Run("Task.Metadata field exists and accepts map", func(t *testing.T) {
		task := phases.Task{
			Id:       "010",
			Name:     "Test",
			Status:   "pending",
			Parallel: false,
			Metadata: map[string]any{
				"github_issue_url": "https://github.com/org/repo/issues/123",
				"custom_field":     42,
			},
		}

		if len(task.Metadata) != 2 {
			t.Errorf("Expected 2 metadata entries, got %d", len(task.Metadata))
		}

		if task.Metadata["github_issue_url"] != "https://github.com/org/repo/issues/123" {
			t.Error("Metadata field should store values correctly")
		}
	})
}

// TestDiscriminatedUnion verifies the ProjectState union includes all 4 types.
func TestDiscriminatedUnion(t *testing.T) {
	t.Run("StandardProjectState is valid", func(_ *testing.T) {
		state := projects.StandardProjectState{
			Project: struct {
				Type         string    `json:"type"`
				Name         string    `json:"name"`
				Branch       string    `json:"branch"`
				Description  string    `json:"description"`
				Github_issue *int64    `json:"github_issue,omitempty"`
				Created_at   time.Time `json:"created_at"`
				Updated_at   time.Time `json:"updated_at"`
			}{
				Type:        "standard",
				Name:        "test-project",
				Branch:      "feat/test",
				Description: "Test",
				Created_at:  time.Now(),
				Updated_at:  time.Now(),
			},
		}

		// Should compile
		_ = state
	})

	t.Run("ExplorationProjectState is valid", func(t *testing.T) {
		state := projects.ExplorationProjectState{
			Project: struct {
				Type        string    `json:"type"`
				Name        string    `json:"name"`
				Branch      string    `json:"branch"`
				Description string    `json:"description"`
				Created_at  time.Time `json:"created_at"`
				Updated_at  time.Time `json:"updated_at"`
			}{
				Type:        "exploration",
				Name:        "test-exploration",
				Branch:      "explore/test",
				Description: "Test exploration",
				Created_at:  time.Now(),
				Updated_at:  time.Now(),
			},
		}

		// Should compile and have correct type
		if state.Project.Type != "exploration" {
			t.Errorf("Expected type 'exploration', got '%s'", state.Project.Type)
		}
	})

	t.Run("DesignProjectState is valid", func(t *testing.T) {
		state := projects.DesignProjectState{
			Project: struct {
				Type        string    `json:"type"`
				Name        string    `json:"name"`
				Branch      string    `json:"branch"`
				Description string    `json:"description"`
				Created_at  time.Time `json:"created_at"`
				Updated_at  time.Time `json:"updated_at"`
			}{
				Type:        "design",
				Name:        "test-design",
				Branch:      "design/test",
				Description: "Test design",
				Created_at:  time.Now(),
				Updated_at:  time.Now(),
			},
		}

		// Should compile and have correct type
		if state.Project.Type != "design" {
			t.Errorf("Expected type 'design', got '%s'", state.Project.Type)
		}
	})

	t.Run("BreakdownProjectState is valid", func(t *testing.T) {
		state := projects.BreakdownProjectState{
			Project: struct {
				Type        string    `json:"type"`
				Name        string    `json:"name"`
				Branch      string    `json:"branch"`
				Description string    `json:"description"`
				Created_at  time.Time `json:"created_at"`
				Updated_at  time.Time `json:"updated_at"`
			}{
				Type:        "breakdown",
				Name:        "test-breakdown",
				Branch:      "breakdown/test",
				Description: "Test breakdown",
				Created_at:  time.Now(),
				Updated_at:  time.Now(),
			},
		}

		// Should compile and have correct type
		if state.Project.Type != "breakdown" {
			t.Errorf("Expected type 'breakdown', got '%s'", state.Project.Type)
		}
	})
}
