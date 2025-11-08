package design

import (
	"strings"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// TestConfigurePrompts verifies that configurePrompts properly registers all prompt generators.
func TestConfigurePrompts(t *testing.T) {
	t.Run("returns non-nil builder for chaining", func(t *testing.T) {
		builder := project.NewProjectTypeConfigBuilder("design")
		result := configurePrompts(builder)

		if result == nil {
			t.Fatal("configurePrompts returned nil, expected builder")
		}
	})

	t.Run("registers orchestrator prompt generator", func(t *testing.T) {
		builder := project.NewProjectTypeConfigBuilder("design")
		configurePrompts(builder)
		config := builder.Build()

		// Verify it returns a string by calling it
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
			},
		}
		prompt := config.OrchestratorPrompt(p)
		if prompt == "" {
			t.Error("orchestrator prompt generator returned empty string")
		}
	})

	t.Run("registers Active state prompt generator", func(t *testing.T) {
		builder := project.NewProjectTypeConfigBuilder("design")
		configurePrompts(builder)
		config := builder.Build()

		// Verify it returns a string using GetStatePrompt
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"design": {},
				},
			},
		}
		activeState := state.State(Active)
		prompt := config.GetStatePrompt(activeState, p)
		if prompt == "" {
			t.Error("Active state prompt generator returned empty string")
		}
	})

	t.Run("registers Finalizing state prompt generator", func(t *testing.T) {
		builder := project.NewProjectTypeConfigBuilder("design")
		configurePrompts(builder)
		config := builder.Build()

		// Verify it returns a string using GetStatePrompt
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"finalization": {},
				},
			},
		}
		finalizingState := state.State(Finalizing)
		prompt := config.GetStatePrompt(finalizingState, p)
		if prompt == "" {
			t.Error("Finalizing state prompt generator returned empty string")
		}
	})
}

// TestGenerateOrchestratorPrompt verifies the orchestrator-level prompt generation.
func TestGenerateOrchestratorPrompt(t *testing.T) {
	t.Run("returns non-empty string", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "auth-design",
				Branch: "design/auth",
			},
		}

		prompt := generateOrchestratorPrompt(p)

		if prompt == "" {
			t.Error("expected non-empty prompt, got empty string")
		}
	})

	t.Run("contains key workflow concepts", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "auth-design",
				Branch: "design/auth",
			},
		}

		prompt := generateOrchestratorPrompt(p)

		// Should mention design workflow
		if !strings.Contains(prompt, "design") && !strings.Contains(prompt, "Design") {
			t.Error("prompt should mention design workflow")
		}

		// Should mention project name and branch
		if !strings.Contains(prompt, "auth-design") {
			t.Error("prompt should contain project name")
		}
		if !strings.Contains(prompt, "design/auth") {
			t.Error("prompt should contain branch name")
		}
	})

	t.Run("handles nil project gracefully", func(t *testing.T) {
		prompt := generateOrchestratorPrompt(nil)

		// Should return a string (possibly error message) rather than panic
		if prompt == "" {
			t.Error("expected error message for nil project, got empty string")
		}
	})
}

// TestGenerateActivePrompt verifies the Active state prompt generation.
//
//nolint:funlen // Test contains multiple subtests for comprehensive prompt validation
func TestGenerateActivePrompt(t *testing.T) {
	t.Run("shows project name and branch", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "auth-design",
				Branch: "design/auth",
				Phases: map[string]projschema.PhaseState{
					"design": {},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "auth-design") {
			t.Error("prompt should contain project name")
		}
		if !strings.Contains(prompt, "design/auth") {
			t.Error("prompt should contain branch name")
		}
	})

	t.Run("shows Active Design state", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"design": {},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "Active Design") {
			t.Error("prompt should indicate Active Design state")
		}
	})

	t.Run("lists all tasks with correct status icons", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"design": {
						Tasks: []projschema.TaskState{
							{Id: "010", Name: "Overview Doc", Status: "pending", Created_at: now},
							{Id: "020", Name: "Details Doc", Status: "in_progress", Created_at: now},
							{Id: "030", Name: "Review Doc", Status: "needs_review", Created_at: now},
							{Id: "040", Name: "Approved Doc", Status: "completed", Created_at: now},
							{Id: "050", Name: "Unused Doc", Status: "abandoned", Created_at: now},
						},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		// Check status icons
		if !strings.Contains(prompt, "[ ] 010") {
			t.Error("prompt should show [ ] for pending task")
		}
		if !strings.Contains(prompt, "[~] 020") {
			t.Error("prompt should show [~] for in_progress task")
		}
		if !strings.Contains(prompt, "[?] 030") {
			t.Error("prompt should show [?] for needs_review task")
		}
		if !strings.Contains(prompt, "[✓] 040") {
			t.Error("prompt should show [✓] for completed task")
		}
		if !strings.Contains(prompt, "[✗] 050") {
			t.Error("prompt should show [✗] for abandoned task")
		}
	})

	t.Run("shows task counts accurately", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"design": {
						Tasks: []projschema.TaskState{
							{Id: "010", Status: "pending", Created_at: now},
							{Id: "020", Status: "pending", Created_at: now},
							{Id: "030", Status: "in_progress", Created_at: now},
							{Id: "040", Status: "needs_review", Created_at: now},
							{Id: "050", Status: "completed", Created_at: now},
							{Id: "060", Status: "abandoned", Created_at: now},
						},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "Total: 6 documents") {
			t.Error("prompt should show total count of 6")
		}
		if !strings.Contains(prompt, "Pending: 2") {
			t.Error("prompt should show 2 pending")
		}
		if !strings.Contains(prompt, "In Progress: 1") {
			t.Error("prompt should show 1 in progress")
		}
		if !strings.Contains(prompt, "Needs Review: 1") {
			t.Error("prompt should show 1 needs review")
		}
		if !strings.Contains(prompt, "Completed: 1") {
			t.Error("prompt should show 1 completed")
		}
		if !strings.Contains(prompt, "Abandoned: 1") {
			t.Error("prompt should show 1 abandoned")
		}
	})

	t.Run("shows advancement readiness when guard passes", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"design": {
						Tasks: []projschema.TaskState{
							{Id: "010", Status: "completed", Created_at: now},
							{Id: "020", Status: "completed", Created_at: now},
						},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "All documents approved") || !strings.Contains(prompt, "✓") {
			t.Error("prompt should show advancement readiness when all documents approved")
		}
		if !strings.Contains(prompt, "sow project advance") {
			t.Error("prompt should suggest sow project advance command")
		}
	})

	t.Run("shows unresolved count when guard fails", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"design": {
						Tasks: []projschema.TaskState{
							{Id: "010", Status: "completed", Created_at: now},
							{Id: "020", Status: "in_progress", Created_at: now},
							{Id: "030", Status: "pending", Created_at: now},
						},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "2 documents remaining") && !strings.Contains(prompt, "(2 documents remaining)") {
			t.Error("prompt should show 2 unresolved documents")
		}
	})

	t.Run("handles empty task list", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"design": {
						Tasks: []projschema.TaskState{},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "No documents planned") {
			t.Error("prompt should indicate no documents planned")
		}
	})

	t.Run("handles missing design phase", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "Error") && !strings.Contains(prompt, "error") {
			t.Error("prompt should indicate error when design phase missing")
		}
	})

	t.Run("shows input artifacts when present", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"design": {
						Inputs: []projschema.ArtifactState{
							{
								Path:       "context/exploration-findings.md",
								Created_at: now,
								Metadata: map[string]interface{}{
									"description": "Exploration findings on auth",
								},
							},
						},
						Tasks: []projschema.TaskState{},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "Design Inputs") {
			t.Error("prompt should show Design Inputs section")
		}
		if !strings.Contains(prompt, "context/exploration-findings.md") {
			t.Error("prompt should show input artifact path")
		}
		if !strings.Contains(prompt, "Exploration findings on auth") {
			t.Error("prompt should show input artifact description")
		}
	})

	t.Run("includes static guidance from template", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"design": {},
				},
			},
		}

		prompt := generateActivePrompt(p)

		// Should include template guidance section
		if !strings.Contains(prompt, "---") {
			t.Error("prompt should include separator from template")
		}
	})

	t.Run("shows artifact path and document type in task listing", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"design": {
						Tasks: []projschema.TaskState{
							{
								Id:         "010",
								Name:       "Architecture Doc",
								Status:     "in_progress",
								Created_at: now,
								Metadata: map[string]interface{}{
									"artifact_path": "project/architecture.md",
									"document_type": "architecture",
								},
							},
						},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "project/architecture.md") {
			t.Error("prompt should show artifact path")
		}
		if !strings.Contains(prompt, "architecture") {
			t.Error("prompt should show document type")
		}
	})
}

// TestGenerateFinalizingPrompt verifies the Finalizing state prompt generation.
//
//nolint:funlen // Test contains multiple subtests for comprehensive prompt validation
func TestGenerateFinalizingPrompt(t *testing.T) {
	t.Run("shows project name and branch", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "auth-design",
				Branch: "design/auth",
				Phases: map[string]projschema.PhaseState{
					"finalization": {},
				},
			},
		}

		prompt := generateFinalizingPrompt(p)

		if !strings.Contains(prompt, "auth-design") {
			t.Error("prompt should contain project name")
		}
		if !strings.Contains(prompt, "design/auth") {
			t.Error("prompt should contain branch name")
		}
	})

	t.Run("shows Finalizing state", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"finalization": {},
				},
			},
		}

		prompt := generateFinalizingPrompt(p)

		if !strings.Contains(prompt, "Finalizing") {
			t.Error("prompt should indicate Finalizing state")
		}
	})

	t.Run("lists finalization tasks with checkboxes", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"finalization": {
						Tasks: []projschema.TaskState{
							{Id: "move-docs", Name: "Move approved documents to targets", Status: "completed", Created_at: now},
							{Id: "create-pr", Name: "Create PR with design artifacts", Status: "pending", Created_at: now},
							{Id: "cleanup", Name: "Delete .sow/project/ directory", Status: "pending", Created_at: now},
						},
					},
				},
			},
		}

		prompt := generateFinalizingPrompt(p)

		if !strings.Contains(prompt, "[✓] Move approved documents to targets") {
			t.Error("prompt should show completed task with [✓]")
		}
		if !strings.Contains(prompt, "[ ] Create PR with design artifacts") {
			t.Error("prompt should show pending task with [ ]")
		}
		if !strings.Contains(prompt, "[ ] Delete .sow/project/ directory") {
			t.Error("prompt should show pending task with [ ]")
		}
	})

	t.Run("shows advancement readiness when all tasks complete", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"finalization": {
						Tasks: []projschema.TaskState{
							{Id: "move-docs", Status: "completed", Created_at: now},
							{Id: "create-pr", Status: "completed", Created_at: now},
							{Id: "cleanup", Status: "completed", Created_at: now},
						},
					},
				},
			},
		}

		prompt := generateFinalizingPrompt(p)

		if !strings.Contains(prompt, "All finalization tasks complete") || !strings.Contains(prompt, "✓") {
			t.Error("prompt should show advancement readiness when all tasks complete")
		}
		if !strings.Contains(prompt, "sow project advance") {
			t.Error("prompt should suggest sow project advance command")
		}
	})

	t.Run("handles missing finalization phase", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{},
			},
		}

		prompt := generateFinalizingPrompt(p)

		if !strings.Contains(prompt, "Error") && !strings.Contains(prompt, "error") {
			t.Error("prompt should indicate error when finalization phase missing")
		}
	})

	t.Run("includes static guidance from template", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"finalization": {},
				},
			},
		}

		prompt := generateFinalizingPrompt(p)

		// Should include template guidance section
		if !strings.Contains(prompt, "---") {
			t.Error("prompt should include separator from template")
		}
	})
}
