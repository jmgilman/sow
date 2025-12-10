package breakdown

import (
	"strings"
	"testing"
	"time"

	"github.com/jmgilman/sow/libs/project"
	"github.com/jmgilman/sow/libs/project/state"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
)

// TestConfigurePrompts verifies that configurePrompts properly registers all prompt generators.
func TestConfigurePrompts(t *testing.T) {
	t.Run("returns non-nil builder for chaining", func(t *testing.T) {
		builder := project.NewProjectTypeConfigBuilder("breakdown")
		result := configurePrompts(builder)

		if result == nil {
			t.Fatal("configurePrompts returned nil, expected builder")
		}
	})

	t.Run("registers orchestrator prompt generator", func(t *testing.T) {
		builder := project.NewProjectTypeConfigBuilder("breakdown")
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
		builder := project.NewProjectTypeConfigBuilder("breakdown")
		configurePrompts(builder)
		config := builder.Build()

		// Verify it returns a string using GetStatePrompt
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {},
				},
			},
		}
		activeState := project.State(Active)
		prompt := config.GetStatePrompt(string(activeState), p)
		if prompt == "" {
			t.Error("Active state prompt generator returned empty string")
		}
	})

	t.Run("registers Publishing state prompt generator", func(t *testing.T) {
		builder := project.NewProjectTypeConfigBuilder("breakdown")
		configurePrompts(builder)
		config := builder.Build()

		// Verify it returns a string using GetStatePrompt
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {},
				},
			},
		}
		publishingState := project.State(Publishing)
		prompt := config.GetStatePrompt(string(publishingState), p)
		if prompt == "" {
			t.Error("Publishing state prompt generator returned empty string")
		}
	})
}

// TestGenerateOrchestratorPrompt verifies the orchestrator-level prompt generation.
func TestGenerateOrchestratorPrompt(t *testing.T) {
	t.Run("returns non-empty string", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "auth-breakdown",
				Branch: "breakdown/auth",
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
				Name:   "auth-breakdown",
				Branch: "breakdown/auth",
			},
		}

		prompt := generateOrchestratorPrompt(p)

		// Should mention breakdown workflow
		if !strings.Contains(prompt, "breakdown") && !strings.Contains(prompt, "Breakdown") {
			t.Error("prompt should mention breakdown workflow")
		}

		// Should mention project name and branch
		if !strings.Contains(prompt, "auth-breakdown") {
			t.Error("prompt should contain project name")
		}
		if !strings.Contains(prompt, "breakdown/auth") {
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
				Name:   "auth-breakdown",
				Branch: "breakdown/auth",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "auth-breakdown") {
			t.Error("prompt should contain project name")
		}
		if !strings.Contains(prompt, "breakdown/auth") {
			t.Error("prompt should contain branch name")
		}
	})

	t.Run("shows Active Breakdown state", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "Active Breakdown") {
			t.Error("prompt should indicate Active Breakdown state")
		}
	})

	t.Run("lists all tasks with correct status icons", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {
						Tasks: []projschema.TaskState{
							{Id: "001", Name: "Pending Work Unit", Status: "pending", Created_at: now},
							{Id: "002", Name: "In Progress Unit", Status: "in_progress", Created_at: now},
							{Id: "003", Name: "Review Unit", Status: "needs_review", Created_at: now},
							{Id: "004", Name: "Approved Unit", Status: "completed", Created_at: now},
							{Id: "005", Name: "Unused Unit", Status: "abandoned", Created_at: now},
						},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		// Check status icons
		if !strings.Contains(prompt, "[ ] 001") {
			t.Error("prompt should show [ ] for pending task")
		}
		if !strings.Contains(prompt, "[~] 002") {
			t.Error("prompt should show [~] for in_progress task")
		}
		if !strings.Contains(prompt, "[?] 003") {
			t.Error("prompt should show [?] for needs_review task")
		}
		if !strings.Contains(prompt, "[✓] 004") {
			t.Error("prompt should show [✓] for completed task")
		}
		if !strings.Contains(prompt, "[✗] 005") {
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
					"breakdown": {
						Tasks: []projschema.TaskState{
							{Id: "001", Status: "pending", Created_at: now},
							{Id: "002", Status: "pending", Created_at: now},
							{Id: "003", Status: "in_progress", Created_at: now},
							{Id: "004", Status: "needs_review", Created_at: now},
							{Id: "005", Status: "completed", Created_at: now},
							{Id: "006", Status: "abandoned", Created_at: now},
						},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "Total: 6 work units") {
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

	t.Run("shows dependencies in task listing", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {
						Tasks: []projschema.TaskState{
							{
								Id:         "001",
								Name:       "Base Unit",
								Status:     "completed",
								Created_at: now,
							},
							{
								Id:         "002",
								Name:       "Dependent Unit",
								Status:     "in_progress",
								Created_at: now,
								Metadata: map[string]interface{}{
									"dependencies": []interface{}{"001"},
								},
							},
							{
								Id:         "003",
								Name:       "Multi-Dep Unit",
								Status:     "pending",
								Created_at: now,
								Metadata: map[string]interface{}{
									"dependencies": []interface{}{"001", "002"},
								},
							},
						},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		// Should show dependencies
		if !strings.Contains(prompt, "Depends on: [001]") {
			t.Error("prompt should show task 002 depends on [001]")
		}
		if !strings.Contains(prompt, "Depends on: [001 002]") {
			t.Error("prompt should show task 003 depends on [001 002]")
		}
	})

	t.Run("shows artifact path in task listing", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {
						Tasks: []projschema.TaskState{
							{
								Id:         "001",
								Name:       "JWT Middleware",
								Status:     "completed",
								Created_at: now,
								Metadata: map[string]interface{}{
									"artifact_path": "project/work-units/001-jwt.md",
								},
							},
						},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "Spec: project/work-units/001-jwt.md") {
			t.Error("prompt should show artifact path")
		}
	})

	t.Run("shows advancement readiness when guards pass", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {
						Tasks: []projschema.TaskState{
							{Id: "001", Status: "completed", Created_at: now},
							{Id: "002", Status: "completed", Created_at: now},
						},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "All work units approved") || !strings.Contains(prompt, "✓") {
			t.Error("prompt should show advancement readiness when all work units approved")
		}
		if !strings.Contains(prompt, "sow project advance") {
			t.Error("prompt should suggest sow project advance command")
		}
	})

	t.Run("shows advancement blocked when dependencies invalid", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {
						Tasks: []projschema.TaskState{
							{
								Id:         "001",
								Status:     "completed",
								Created_at: now,
								Metadata: map[string]interface{}{
									"dependencies": []interface{}{"002"}, // Cycle
								},
							},
							{
								Id:         "002",
								Status:     "completed",
								Created_at: now,
								Metadata: map[string]interface{}{
									"dependencies": []interface{}{"001"}, // Cycle
								},
							},
						},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "Dependency validation failed") && !strings.Contains(prompt, "cycles") {
			t.Error("prompt should indicate dependency validation failure")
		}
	})

	t.Run("handles empty task list", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {
						Tasks: []projschema.TaskState{},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "No work units") {
			t.Error("prompt should indicate no work units")
		}
	})

	t.Run("handles missing breakdown phase", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "Error") && !strings.Contains(prompt, "error") {
			t.Error("prompt should indicate error when breakdown phase missing")
		}
	})

	t.Run("shows input artifacts when present", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {
						Inputs: []projschema.ArtifactState{
							{
								Path:       "designs/auth-design.md",
								Created_at: now,
								Metadata: map[string]interface{}{
									"description": "Authentication system design",
								},
							},
						},
						Tasks: []projschema.TaskState{},
					},
				},
			},
		}

		prompt := generateActivePrompt(p)

		if !strings.Contains(prompt, "Being Broken Down") {
			t.Error("prompt should show 'Being Broken Down' section")
		}
		if !strings.Contains(prompt, "designs/auth-design.md") {
			t.Error("prompt should show input artifact path")
		}
		if !strings.Contains(prompt, "Authentication system design") {
			t.Error("prompt should show input artifact description")
		}
	})

	t.Run("includes static guidance from template", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {},
				},
			},
		}

		prompt := generateActivePrompt(p)

		// Should include template guidance section
		if !strings.Contains(prompt, "---") {
			t.Error("prompt should include separator from template")
		}
	})
}

// TestGeneratePublishingPrompt verifies the Publishing state prompt generation.
//
//nolint:funlen // Test contains multiple subtests for comprehensive prompt validation
func TestGeneratePublishingPrompt(t *testing.T) {
	t.Run("shows project name and branch", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "auth-breakdown",
				Branch: "breakdown/auth",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {},
				},
			},
		}

		prompt := generatePublishingPrompt(p)

		if !strings.Contains(prompt, "auth-breakdown") {
			t.Error("prompt should contain project name")
		}
		if !strings.Contains(prompt, "breakdown/auth") {
			t.Error("prompt should contain branch name")
		}
	})

	t.Run("shows Publishing state", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {},
				},
			},
		}

		prompt := generatePublishingPrompt(p)

		if !strings.Contains(prompt, "Publishing") {
			t.Error("prompt should indicate Publishing state")
		}
	})

	t.Run("shows publishing counts", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {
						Tasks: []projschema.TaskState{
							{
								Id:         "001",
								Status:     "completed",
								Created_at: now,
								Metadata: map[string]interface{}{
									"published":           true,
									"github_issue_number": 123,
									"github_issue_url":    "https://github.com/org/repo/issues/123",
								},
							},
							{
								Id:         "002",
								Status:     "completed",
								Created_at: now,
								Metadata: map[string]interface{}{
									"published": false,
								},
							},
							{
								Id:         "003",
								Status:     "completed",
								Created_at: now,
							},
						},
					},
				},
			},
		}

		prompt := generatePublishingPrompt(p)

		if !strings.Contains(prompt, "Total work units: 3") {
			t.Error("prompt should show total count of 3")
		}
		if !strings.Contains(prompt, "Published: 1") {
			t.Error("prompt should show 1 published")
		}
		if !strings.Contains(prompt, "Unpublished: 2") {
			t.Error("prompt should show 2 unpublished")
		}
	})

	t.Run("shows published work units with URLs", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {
						Tasks: []projschema.TaskState{
							{
								Id:         "001",
								Name:       "JWT Middleware",
								Status:     "completed",
								Created_at: now,
								Metadata: map[string]interface{}{
									"published":           true,
									"github_issue_number": 123,
									"github_issue_url":    "https://github.com/org/repo/issues/123",
								},
							},
						},
					},
				},
			},
		}

		prompt := generatePublishingPrompt(p)

		if !strings.Contains(prompt, "[✓] Published: https://github.com/org/repo/issues/123") {
			t.Error("prompt should show published task with URL")
		}
		if !strings.Contains(prompt, "001 - JWT Middleware") {
			t.Error("prompt should show task ID and name")
		}
	})

	t.Run("shows unpublished work units", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {
						Tasks: []projschema.TaskState{
							{
								Id:         "002",
								Name:       "Auth Endpoint",
								Status:     "completed",
								Created_at: now,
								Metadata:   map[string]interface{}{},
							},
						},
					},
				},
			},
		}

		prompt := generatePublishingPrompt(p)

		if !strings.Contains(prompt, "[ ] Pending 002 - Auth Endpoint") {
			t.Error("prompt should show unpublished task with [ ] Pending")
		}
	})

	t.Run("shows advancement readiness when all published", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {
						Tasks: []projschema.TaskState{
							{
								Id:         "001",
								Status:     "completed",
								Created_at: now,
								Metadata: map[string]interface{}{
									"published": true,
								},
							},
							{
								Id:         "002",
								Status:     "completed",
								Created_at: now,
								Metadata: map[string]interface{}{
									"published": true,
								},
							},
						},
					},
				},
			},
		}

		prompt := generatePublishingPrompt(p)

		if !strings.Contains(prompt, "All work units published") || !strings.Contains(prompt, "✓") {
			t.Error("prompt should show advancement readiness when all published")
		}
		if !strings.Contains(prompt, "sow project advance") {
			t.Error("prompt should suggest sow project advance command")
		}
	})

	t.Run("handles missing breakdown phase", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{},
			},
		}

		prompt := generatePublishingPrompt(p)

		if !strings.Contains(prompt, "Error") && !strings.Contains(prompt, "error") {
			t.Error("prompt should indicate error when breakdown phase missing")
		}
	})

	t.Run("includes static guidance from template", func(t *testing.T) {
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {},
				},
			},
		}

		prompt := generatePublishingPrompt(p)

		// Should include template guidance section
		if !strings.Contains(prompt, "---") {
			t.Error("prompt should include separator from template")
		}
	})

	t.Run("ignores abandoned tasks in publishing counts", func(t *testing.T) {
		now := time.Now()
		p := &state.Project{
			ProjectState: projschema.ProjectState{
				Name:   "test",
				Branch: "test-branch",
				Phases: map[string]projschema.PhaseState{
					"breakdown": {
						Tasks: []projschema.TaskState{
							{
								Id:         "001",
								Status:     "completed",
								Created_at: now,
								Metadata: map[string]interface{}{
									"published": true,
								},
							},
							{
								Id:         "002",
								Status:     "abandoned",
								Created_at: now,
							},
						},
					},
				},
			},
		}

		prompt := generatePublishingPrompt(p)

		// Should only count completed tasks
		if !strings.Contains(prompt, "Total work units: 1") {
			t.Error("prompt should only count completed tasks, not abandoned")
		}
	})
}
