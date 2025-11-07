package project

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projectschema "github.com/jmgilman/sow/cli/schemas/project"
)

// TestCompleteProjectTypeWorkflow demonstrates the complete SDK workflow from
// project type definition through state machine advancement with guards and actions.
// This is the primary integration test proving the SDK works end-to-end.
func TestCompleteProjectTypeWorkflow(t *testing.T) {
	// 1. DEFINE: Create project type using builder API
	config := NewProjectTypeConfigBuilder("simple").
		WithPhase("work",
			WithOutputs("result"),
			WithTasks(),
			WithMetadataSchema(`{
				complexity: "low" | "medium" | "high"
			}`),
		).
		SetInitialState(sdkstate.State("Idle")).
		AddTransition(
			sdkstate.State("Idle"),
			sdkstate.State("Working"),
			sdkstate.Event("Start"),
		).
		AddTransition(
			sdkstate.State("Working"),
			sdkstate.State("Done"),
			sdkstate.Event("Complete"),
			WithGuard("result approved", func(p *state.Project) bool {
				return p.PhaseOutputApproved("work", "result")
			}),
		).
		Build()

	if config == nil {
		t.Fatal("Builder.Build() returned nil")
	}

	// 2. CREATE: Make project state
	proj := createTestProject(t, "simple")

	// 3. BUILD: Create state machine
	machine := config.BuildMachine(proj, sdkstate.State("Idle"))
	if machine == nil {
		t.Fatal("BuildMachine returned nil")
	}
	if machine.State() != sdkstate.State("Idle") {
		t.Errorf("expected initial state=Idle, got %s", machine.State())
	}

	// 4. ADVANCE: Progress from Idle → Working
	can, err := machine.CanFire(sdkstate.Event("Start"))
	if err != nil {
		t.Fatalf("CanFire(Start) failed: %v", err)
	}
	if !can {
		t.Error("expected Start event to be firable from Idle")
	}

	err = machine.Fire(sdkstate.Event("Start"))
	if err != nil {
		t.Fatalf("Fire(Start) failed: %v", err)
	}
	if machine.State() != sdkstate.State("Working") {
		t.Errorf("expected state=Working, got %s", machine.State())
	}

	// 5. GUARD TEST: Guard blocks without approved output
	can, err = machine.CanFire(sdkstate.Event("Complete"))
	if err != nil {
		t.Fatalf("CanFire(Complete) failed: %v", err)
	}
	if can {
		t.Error("expected Complete event to be blocked by guard (output not approved)")
	}

	// 6. MODIFY: Approve output
	phase := proj.Phases["work"]
	phase.Outputs[0].Approved = true
	proj.Phases["work"] = phase

	// 7. ADVANCE: Now transition should work
	can, err = machine.CanFire(sdkstate.Event("Complete"))
	if err != nil {
		t.Fatalf("CanFire(Complete) after approval failed: %v", err)
	}
	if !can {
		t.Error("expected Complete event to be allowed after approval")
	}

	err = machine.Fire(sdkstate.Event("Complete"))
	if err != nil {
		t.Fatalf("Fire(Complete) failed: %v", err)
	}
	if machine.State() != sdkstate.State("Done") {
		t.Errorf("expected state=Done, got %s", machine.State())
	}
}

// TestBuilderPhaseConfiguration verifies that phase configuration via the builder
// works correctly and is accessible in the built config.
func TestBuilderPhaseConfiguration(t *testing.T) {
	// Define project type with multiple phase configurations
	config := NewProjectTypeConfigBuilder("multi-phase").
		WithPhase("planning",
			WithOutputs("task_list", "design"),
		).
		WithPhase("implementation",
			WithInputs("task_list"),
			WithOutputs("code", "tests"),
			WithTasks(),
			WithMetadataSchema(`{
				complexity: "low" | "medium" | "high"
			}`),
		).
		WithPhase("review",
			WithInputs("code", "tests"),
			WithOutputs("approval"),
		).
		SetInitialState(sdkstate.State("Planning")).
		Build()

	if config == nil {
		t.Fatal("Builder.Build() returned nil")
	}

	// Verify we can build a machine with this config
	proj := &state.Project{
		ProjectState: projectschema.ProjectState{
			Name:   "test-project",
			Type:   "multi-phase",
			Branch: "test-branch",
			Phases: map[string]projectschema.PhaseState{
				"planning":       {Status: "pending", Enabled: true},
				"implementation": {Status: "pending", Enabled: true},
				"review":         {Status: "pending", Enabled: true},
			},
			Statechart: projectschema.StatechartState{
				Current_state: "Planning",
				Updated_at:    time.Now(),
			},
		},
	}

	machine := config.BuildMachine(proj, sdkstate.State("Planning"))
	if machine == nil {
		t.Fatal("BuildMachine returned nil")
	}
	if machine.State() != sdkstate.State("Planning") {
		t.Errorf("expected state=Planning, got %s", machine.State())
	}
}

// TestOnEntryOnExitActionsIntegration verifies that entry and exit actions
// execute correctly during state transitions.
func TestOnEntryOnExitActionsIntegration(t *testing.T) {

	// Track action executions
	var actionsExecuted []string

	// Define project type with entry/exit actions
	config := NewProjectTypeConfigBuilder("with-actions").
		WithPhase("work").
		SetInitialState(sdkstate.State("Start")).
		AddTransition(
			sdkstate.State("Start"),
			sdkstate.State("Middle"),
			sdkstate.Event("Advance"),
			WithOnExit(func(p *state.Project) error {
				actionsExecuted = append(actionsExecuted, "exit-Start")
				// Mutate project state
				phase := p.Phases["work"]
				if phase.Metadata == nil {
					phase.Metadata = make(map[string]interface{})
				}
				phase.Metadata["left_start"] = true
				p.Phases["work"] = phase
				return nil
			}),
			WithOnEntry(func(p *state.Project) error {
				actionsExecuted = append(actionsExecuted, "enter-Middle")
				// Mutate project state
				phase := p.Phases["work"]
				if phase.Metadata == nil {
					phase.Metadata = make(map[string]interface{})
				}
				phase.Metadata["entered_middle"] = true
				p.Phases["work"] = phase
				return nil
			}),
		).
		AddTransition(
			sdkstate.State("Middle"),
			sdkstate.State("End"),
			sdkstate.Event("Finish"),
			WithOnExit(func(_ *state.Project) error {
				actionsExecuted = append(actionsExecuted, "exit-Middle")
				return nil
			}),
			WithOnEntry(func(_ *state.Project) error {
				actionsExecuted = append(actionsExecuted, "enter-End")
				return nil
			}),
		).
		Build()

	proj := &state.Project{
		ProjectState: projectschema.ProjectState{
			Name:   "test-project",
			Type:   "with-actions",
			Branch: "test-branch",
			Phases: map[string]projectschema.PhaseState{
				"work": {
					Status:   "pending",
					Enabled:  true,
					Metadata: make(map[string]interface{}),
				},
			},
			Statechart: projectschema.StatechartState{
				Current_state: "Start",
				Updated_at:    time.Now(),
			},
		},
	}

	machine := config.BuildMachine(proj, sdkstate.State("Start"))

	// First transition: Start -> Middle
	err := machine.Fire(sdkstate.Event("Advance"))
	if err != nil {
		t.Fatalf("first transition failed: %v", err)
	}

	// Verify actions executed in correct order
	if len(actionsExecuted) != 2 {
		t.Errorf("expected 2 actions, got %d", len(actionsExecuted))
	}
	if len(actionsExecuted) >= 1 && actionsExecuted[0] != "exit-Start" {
		t.Errorf("first action should be exit-Start, got %s", actionsExecuted[0])
	}
	if len(actionsExecuted) >= 2 && actionsExecuted[1] != "enter-Middle" {
		t.Errorf("second action should be enter-Middle, got %s", actionsExecuted[1])
	}

	// Verify state mutations persisted
	phase := proj.Phases["work"]
	if phase.Metadata["left_start"] != true {
		t.Error("onExit action should have set left_start=true")
	}
	if phase.Metadata["entered_middle"] != true {
		t.Error("onEntry action should have set entered_middle=true")
	}

	// Verify machine transitioned
	if machine.State() != sdkstate.State("Middle") {
		t.Errorf("expected state=Middle, got %s", machine.State())
	}

	// Second transition: Middle -> End
	err = machine.Fire(sdkstate.Event("Finish"))
	if err != nil {
		t.Fatalf("second transition failed: %v", err)
	}

	// Verify all actions executed
	if len(actionsExecuted) != 4 {
		t.Errorf("expected 4 actions total, got %d", len(actionsExecuted))
	}

	// Verify final state
	if machine.State() != sdkstate.State("End") {
		t.Errorf("expected state=End, got %s", machine.State())
	}
}

// TestMultiplePhaseWorkflow verifies that a project type with multiple phases
// can be configured and transitioned through correctly.
func TestMultiplePhaseWorkflow(t *testing.T) {

	// Define project type with 3 phases
	config := NewProjectTypeConfigBuilder("multi-phase").
		WithPhase("planning",
			WithOutputs("task_list"),
		).
		WithPhase("implementation",
			WithInputs("task_list"),
			WithOutputs("code"),
			WithTasks(),
		).
		WithPhase("review",
			WithInputs("code"),
			WithOutputs("approval"),
		).
		SetInitialState(sdkstate.State("Planning")).
		AddTransition(
			sdkstate.State("Planning"),
			sdkstate.State("Implementation"),
			sdkstate.Event("StartWork"),
			WithGuard("task list approved", func(p *state.Project) bool {
				return p.PhaseOutputApproved("planning", "task_list")
			}),
		).
		AddTransition(
			sdkstate.State("Implementation"),
			sdkstate.State("Review"),
			sdkstate.Event("RequestReview"),
			WithGuard("all tasks complete", func(p *state.Project) bool {
				return p.AllTasksComplete()
			}),
		).
		AddTransition(
			sdkstate.State("Review"),
			sdkstate.State("Complete"),
			sdkstate.Event("Approve"),
			WithGuard("review approved", func(p *state.Project) bool {
				return p.PhaseOutputApproved("review", "approval")
			}),
		).
		Build()

	// Create project with all phases
	proj := &state.Project{
		ProjectState: projectschema.ProjectState{
			Name:   "test-project",
			Type:   "multi-phase",
			Branch: "test-branch",
			Phases: map[string]projectschema.PhaseState{
				"planning": {
					Status:  "completed",
					Enabled: true,
					Outputs: []projectschema.ArtifactState{
						{Type: "task_list", Path: "tasks.md", Approved: true, Created_at: time.Now()},
					},
				},
				"implementation": {
					Status:  "in_progress",
					Enabled: true,
					Inputs: []projectschema.ArtifactState{
						{Type: "task_list", Path: "tasks.md", Created_at: time.Now()},
					},
					Outputs: []projectschema.ArtifactState{
						{Type: "code", Path: "main.go", Approved: false, Created_at: time.Now()},
					},
					Tasks: []projectschema.TaskState{
						{
							Id:             "001",
							Status:         "completed",
							Name:           "Task 1",
							Phase:          "implementation",
							Created_at:     time.Now(),
							Updated_at:     time.Now(),
							Iteration:      1,
							Assigned_agent: "implementer",
						},
					},
				},
				"review": {
					Status:  "pending",
					Enabled: true,
				},
			},
			Statechart: projectschema.StatechartState{
				Current_state: "Planning",
				Updated_at:    time.Now(),
			},
		},
	}

	machine := config.BuildMachine(proj, sdkstate.State("Planning"))

	// Advance through states
	// Planning -> Implementation
	err := machine.Fire(sdkstate.Event("StartWork"))
	if err != nil {
		t.Fatalf("transition to Implementation failed: %v", err)
	}
	if machine.State() != sdkstate.State("Implementation") {
		t.Errorf("expected Implementation state, got %s", machine.State())
	}

	// Implementation -> Review (should work since all tasks complete)
	err = machine.Fire(sdkstate.Event("RequestReview"))
	if err != nil {
		t.Fatalf("transition to Review failed: %v", err)
	}
	if machine.State() != sdkstate.State("Review") {
		t.Errorf("expected Review state, got %s", machine.State())
	}

	// Review -> Complete (should fail without approval)
	can, err := machine.CanFire(sdkstate.Event("Approve"))
	if err != nil {
		t.Fatalf("CanFire(Approve) failed: %v", err)
	}
	if can {
		t.Error("transition to Complete should be blocked without approval")
	}

	// Add approval and try again
	phase := proj.Phases["review"]
	phase.Outputs = []projectschema.ArtifactState{
		{Type: "approval", Path: "approval.md", Approved: true, Created_at: time.Now()},
	}
	proj.Phases["review"] = phase

	err = machine.Fire(sdkstate.Event("Approve"))
	if err != nil {
		t.Fatalf("transition to Complete failed: %v", err)
	}
	if machine.State() != sdkstate.State("Complete") {
		t.Errorf("expected Complete state, got %s", machine.State())
	}
}

// TestGuardBlocksAndAllows verifies that guards correctly block and allow transitions.
func TestGuardBlocksAndAllows(t *testing.T) {
	t.Run("guard blocks transition", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("blocked").
			SetInitialState(sdkstate.State("Start")).
			AddTransition(
				sdkstate.State("Start"),
				sdkstate.State("End"),
				sdkstate.Event("Go"),
				WithGuard("always blocked", func(_ *state.Project) bool {
					return false // Always block
				}),
			).
			Build()

		proj := &state.Project{
			ProjectState: projectschema.ProjectState{
				Name:   "test-project",
				Type:   "blocked",
				Branch: "test-branch",
				Phases: make(map[string]projectschema.PhaseState),
				Statechart: projectschema.StatechartState{
					Current_state: "Start",
					Updated_at:    time.Now(),
				},
			},
		}

		machine := config.BuildMachine(proj, sdkstate.State("Start"))

		can, err := machine.CanFire(sdkstate.Event("Go"))
		if err != nil {
			t.Fatalf("CanFire failed: %v", err)
		}
		if can {
			t.Error("guard should block transition")
		}
	})

	t.Run("guard allows transition", func(t *testing.T) {
		config := NewProjectTypeConfigBuilder("allowed").
			SetInitialState(sdkstate.State("Start")).
			AddTransition(
				sdkstate.State("Start"),
				sdkstate.State("End"),
				sdkstate.Event("Go"),
				WithGuard("always allowed", func(_ *state.Project) bool {
					return true // Always allow
				}),
			).
			Build()

		proj := &state.Project{
			ProjectState: projectschema.ProjectState{
				Name:   "test-project",
				Type:   "allowed",
				Branch: "test-branch",
				Phases: make(map[string]projectschema.PhaseState),
				Statechart: projectschema.StatechartState{
					Current_state: "Start",
					Updated_at:    time.Now(),
				},
			},
		}

		machine := config.BuildMachine(proj, sdkstate.State("Start"))

		can, err := machine.CanFire(sdkstate.Event("Go"))
		if err != nil {
			t.Fatalf("CanFire failed: %v", err)
		}
		if !can {
			t.Error("guard should allow transition")
		}

		// Verify the transition actually works
		err = machine.Fire(sdkstate.Event("Go"))
		if err != nil {
			t.Fatalf("Fire failed: %v", err)
		}
		if machine.State() != sdkstate.State("End") {
			t.Errorf("expected state=End, got %s", machine.State())
		}
	})
}

// TestReviewBranchingWorkflow demonstrates complete branching workflow using AddBranch.
// This tests the auto-generation of transitions and event determiners for state-determined branching.
func TestReviewBranchingWorkflow(t *testing.T) {
	// Create complete project type with AddBranch
	config := NewProjectTypeConfigBuilder("review-test").
		WithPhase("review",
			WithOutputs("review"),
		).
		SetInitialState(sdkstate.State("ReviewActive")).
		AddBranch(
			sdkstate.State("ReviewActive"),
			BranchOn(func(p *state.Project) string {
				// Get review assessment from latest approved review artifact
				phase := p.Phases["review"]
				for i := len(phase.Outputs) - 1; i >= 0; i-- {
					artifact := phase.Outputs[i]
					if artifact.Type == "review" && artifact.Approved {
						if assessment, ok := artifact.Metadata["assessment"].(string); ok {
							return assessment
						}
					}
				}
				return ""
			}),
			When("pass",
				sdkstate.Event("ReviewPass"),
				sdkstate.State("FinalizeState"),
				WithDescription("Review approved - proceed to finalization"),
			),
			When("fail",
				sdkstate.Event("ReviewFail"),
				sdkstate.State("ReworkState"),
				WithDescription("Review failed - return to planning for rework"),
			),
		).
		Build()

	// Test 1: Review passes
	t.Run("review passes", func(t *testing.T) {
		// Create test project with pass assessment
		proj := &state.Project{
			ProjectState: projectschema.ProjectState{
				Name:   "review-test",
				Type:   "review-test",
				Branch: "test-branch",
				Phases: map[string]projectschema.PhaseState{
					"review": {
						Status:  "active",
						Enabled: true,
						Outputs: []projectschema.ArtifactState{
							{
								Type:     "review",
								Approved: true,
								Path:     "review.md",
								Metadata: map[string]interface{}{
									"assessment": "pass",
								},
								Created_at: time.Now(),
							},
						},
					},
				},
				Statechart: projectschema.StatechartState{
					Current_state: "ReviewActive",
					Updated_at:    time.Now(),
				},
			},
		}

		// Build machine starting in ReviewActive
		machine := config.BuildMachine(proj, sdkstate.State("ReviewActive"))

		// Determine event (should return ReviewPass)
		event, err := config.DetermineEvent(proj)
		if err != nil {
			t.Fatalf("DetermineEvent failed: %v", err)
		}
		if event != sdkstate.Event("ReviewPass") {
			t.Errorf("expected ReviewPass event, got %s", event)
		}

		// Fire event (should transition to FinalizeState)
		err = machine.Fire(event)
		if err != nil {
			t.Fatalf("Fire(ReviewPass) failed: %v", err)
		}
		if machine.State() != sdkstate.State("FinalizeState") {
			t.Errorf("expected FinalizeState, got %s", machine.State())
		}
	})

	// Test 2: Review fails
	t.Run("review fails", func(t *testing.T) {
		// Create test project with fail assessment
		proj := &state.Project{
			ProjectState: projectschema.ProjectState{
				Name:   "review-test",
				Type:   "review-test",
				Branch: "test-branch",
				Phases: map[string]projectschema.PhaseState{
					"review": {
						Status:  "active",
						Enabled: true,
						Outputs: []projectschema.ArtifactState{
							{
								Type:     "review",
								Approved: true,
								Path:     "review.md",
								Metadata: map[string]interface{}{
									"assessment": "fail",
								},
								Created_at: time.Now(),
							},
						},
					},
				},
				Statechart: projectschema.StatechartState{
					Current_state: "ReviewActive",
					Updated_at:    time.Now(),
				},
			},
		}

		// Build machine starting in ReviewActive
		machine := config.BuildMachine(proj, sdkstate.State("ReviewActive"))

		// Determine event (should return ReviewFail)
		event, err := config.DetermineEvent(proj)
		if err != nil {
			t.Fatalf("DetermineEvent failed: %v", err)
		}
		if event != sdkstate.Event("ReviewFail") {
			t.Errorf("expected ReviewFail event, got %s", event)
		}

		// Fire event (should transition to ReworkState)
		err = machine.Fire(event)
		if err != nil {
			t.Fatalf("Fire(ReviewFail) failed: %v", err)
		}
		if machine.State() != sdkstate.State("ReworkState") {
			t.Errorf("expected ReworkState, got %s", machine.State())
		}
	})

	// Test 3: Error on unmapped value
	t.Run("error on unmapped assessment", func(t *testing.T) {
		// Create test project with unknown assessment
		proj := &state.Project{
			ProjectState: projectschema.ProjectState{
				Name:   "review-test",
				Type:   "review-test",
				Branch: "test-branch",
				Phases: map[string]projectschema.PhaseState{
					"review": {
						Status:  "active",
						Enabled: true,
						Outputs: []projectschema.ArtifactState{
							{
								Type:     "review",
								Approved: true,
								Path:     "review.md",
								Metadata: map[string]interface{}{
									"assessment": "unknown",
								},
								Created_at: time.Now(),
							},
						},
					},
				},
				Statechart: projectschema.StatechartState{
					Current_state: "ReviewActive",
					Updated_at:    time.Now(),
				},
			},
		}

		// Determine event should return error
		_, err := config.DetermineEvent(proj)
		if err == nil {
			t.Error("expected error for unmapped assessment value")
		}
		// Verify error message is helpful
		expectedSubstrings := []string{
			"no branch defined",
			"\"unknown\"",
			"available:",
		}
		for _, substr := range expectedSubstrings {
			if !contains(err.Error(), substr) {
				t.Errorf("error message should contain %q, got: %s", substr, err.Error())
			}
		}
	})
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// createTestProject creates a minimal project state for testing.
func createTestProject(_ *testing.T, typeName string) *state.Project {
	return &state.Project{
		ProjectState: projectschema.ProjectState{
			Name:   "test-project",
			Type:   typeName,
			Branch: "test-branch",
			Phases: map[string]projectschema.PhaseState{
				"work": {
					Status:  "pending",
					Enabled: true,
					Outputs: []projectschema.ArtifactState{
						{Type: "result", Approved: false, Created_at: time.Now()},
					},
					Metadata: map[string]interface{}{
						"complexity": "medium",
					},
				},
			},
			Statechart: projectschema.StatechartState{
				Current_state: "Idle",
				Updated_at:    time.Now(),
			},
		},
	}
}
