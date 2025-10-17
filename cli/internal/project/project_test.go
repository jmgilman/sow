package project

import (
	"strings"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProjectState(t *testing.T) {
	name := "my-feature"
	branch := "feat/my-feature"
	description := "Add new feature"

	state := NewProjectState(name, branch, description)

	// Test project metadata
	assert.Equal(t, name, state.Project.Name)
	assert.Equal(t, branch, state.Project.Branch)
	assert.Equal(t, description, state.Project.Description)
	assert.WithinDuration(t, time.Now(), state.Project.Created_at, time.Second)
	assert.WithinDuration(t, time.Now(), state.Project.Updated_at, time.Second)

	// Test discovery phase (disabled)
	assert.False(t, state.Phases.Discovery.Enabled)
	assert.Equal(t, "skipped", state.Phases.Discovery.Status)
	assert.Nil(t, state.Phases.Discovery.Discovery_type)
	assert.Empty(t, state.Phases.Discovery.Artifacts)

	// Test design phase (disabled)
	assert.False(t, state.Phases.Design.Enabled)
	assert.Equal(t, "skipped", state.Phases.Design.Status)
	assert.Nil(t, state.Phases.Design.Architect_used)
	assert.Empty(t, state.Phases.Design.Artifacts)

	// Test implementation phase (enabled)
	assert.True(t, state.Phases.Implementation.Enabled)
	assert.Equal(t, "pending", state.Phases.Implementation.Status)
	assert.Empty(t, state.Phases.Implementation.Tasks)
	assert.Nil(t, state.Phases.Implementation.Pending_task_additions)

	// Test review phase (enabled)
	assert.True(t, state.Phases.Review.Enabled)
	assert.Equal(t, "pending", state.Phases.Review.Status)
	assert.Equal(t, int64(1), state.Phases.Review.Iteration)
	assert.Empty(t, state.Phases.Review.Reports)

	// Test finalize phase (enabled)
	assert.True(t, state.Phases.Finalize.Enabled)
	assert.Equal(t, "pending", state.Phases.Finalize.Status)
	assert.False(t, state.Phases.Finalize.Project_deleted)
	assert.Nil(t, state.Phases.Finalize.Pr_url)
}

func TestValidateBranch(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		wantErr bool
	}{
		{
			name:    "valid feature branch",
			branch:  "feat/add-authentication",
			wantErr: false,
		},
		{
			name:    "valid fix branch",
			branch:  "fix/login-bug",
			wantErr: false,
		},
		{
			name:    "invalid main branch",
			branch:  "main",
			wantErr: true,
		},
		{
			name:    "invalid master branch",
			branch:  "master",
			wantErr: true,
		},
		{
			name:    "branch named mainline is ok",
			branch:  "mainline",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranch(tt.branch)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "protected branch")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	t.Run("new project with no tasks", func(t *testing.T) {
		state := NewProjectState("my-feature", "feat/my-feature", "Add new feature")

		output := FormatStatus(state)

		// Check key components are present
		assert.Contains(t, output, "Project: my-feature (on feat/my-feature)")
		assert.Contains(t, output, "Description: Add new feature")
		assert.Contains(t, output, "Phases:")
		assert.Contains(t, output, "Discovery")
		assert.Contains(t, output, "Design")
		assert.Contains(t, output, "Implementation")
		assert.Contains(t, output, "Review")
		assert.Contains(t, output, "Finalize")
		assert.Contains(t, output, "Tasks: none yet")

		// Discovery and Design should not be enabled (no ✓)
		lines := strings.Split(output, "\n")
		discoveryLine := ""
		designLine := ""
		for _, line := range lines {
			if strings.Contains(line, "Discovery") {
				discoveryLine = line
			}
			if strings.Contains(line, "Design") && !strings.Contains(line, "Design:")  {
				designLine = line
			}
		}
		assert.Contains(t, discoveryLine, "[ ]")
		assert.Contains(t, designLine, "[ ]")
	})

	t.Run("project with tasks", func(t *testing.T) {
		state := NewProjectState("my-feature", "feat/my-feature", "Add new feature")

		// Add some tasks
		state.Phases.Implementation.Tasks = []schemas.Task{
			{Id: "010", Name: "Task 1", Status: "completed", Parallel: false, Dependencies: nil},
			{Id: "020", Name: "Task 2", Status: "in_progress", Parallel: false, Dependencies: nil},
			{Id: "030", Name: "Task 3", Status: "pending", Parallel: false, Dependencies: nil},
		}

		output := FormatStatus(state)

		assert.Contains(t, output, "Tasks: 3 total")
		assert.Contains(t, output, "1 completed")
		assert.Contains(t, output, "1 in_progress")
		assert.Contains(t, output, "1 pending")
	})

	t.Run("project with enabled discovery and design", func(t *testing.T) {
		state := NewProjectState("my-feature", "feat/my-feature", "Add new feature")

		// Enable discovery and design
		state.Phases.Discovery.Enabled = true
		state.Phases.Discovery.Status = "completed"
		state.Phases.Design.Enabled = true
		state.Phases.Design.Status = "in_progress"

		output := FormatStatus(state)

		// Both should show as enabled (✓)
		lines := strings.Split(output, "\n")
		discoveryLine := ""
		designLine := ""
		for _, line := range lines {
			if strings.Contains(line, "Discovery") {
				discoveryLine = line
			}
			if strings.Contains(line, "Design") && !strings.Contains(line, "Design:") {
				designLine = line
			}
		}
		assert.Contains(t, discoveryLine, "[✓]")
		assert.Contains(t, designLine, "[✓]")
	})
}

func TestValidatePhase(t *testing.T) {
	tests := []struct {
		name    string
		phase   string
		wantErr bool
	}{
		{
			name:    "valid discovery phase",
			phase:   PhaseDiscovery,
			wantErr: false,
		},
		{
			name:    "valid design phase",
			phase:   PhaseDesign,
			wantErr: false,
		},
		{
			name:    "valid implementation phase",
			phase:   PhaseImplementation,
			wantErr: false,
		},
		{
			name:    "valid review phase",
			phase:   PhaseReview,
			wantErr: false,
		},
		{
			name:    "valid finalize phase",
			phase:   PhaseFinalize,
			wantErr: false,
		},
		{
			name:    "invalid phase name",
			phase:   "invalid-phase",
			wantErr: true,
		},
		{
			name:    "empty phase name",
			phase:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePhase(tt.phase)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid phase")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEnablePhase(t *testing.T) {
	t.Run("enable discovery with valid type", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		discoveryType := "feature"

		err := EnablePhase(state, PhaseDiscovery, &discoveryType)
		require.NoError(t, err)

		assert.True(t, state.Phases.Discovery.Enabled)
		assert.Equal(t, "pending", state.Phases.Discovery.Status)
		assert.NotNil(t, state.Phases.Discovery.Discovery_type)
		// Discovery_type is of type any, so compare as string
		if dt, ok := state.Phases.Discovery.Discovery_type.(*string); ok {
			assert.Equal(t, "feature", *dt)
		} else {
			t.Fatal("discovery_type is not a string pointer")
		}
	})

	t.Run("enable discovery without type fails", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")

		err := EnablePhase(state, PhaseDiscovery, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "discovery_type is required")
	})

	t.Run("enable discovery with empty type fails", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		emptyType := ""

		err := EnablePhase(state, PhaseDiscovery, &emptyType)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "discovery_type is required")
	})

	t.Run("enable discovery with invalid type fails", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		invalidType := "invalid-type"

		err := EnablePhase(state, PhaseDiscovery, &invalidType)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid discovery_type")
	})

	t.Run("enable design phase", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")

		err := EnablePhase(state, PhaseDesign, nil)
		require.NoError(t, err)

		assert.True(t, state.Phases.Design.Enabled)
		assert.Equal(t, "pending", state.Phases.Design.Status)
	})

	t.Run("enable already enabled discovery fails", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		discoveryType := "feature"

		// Enable first time
		err := EnablePhase(state, PhaseDiscovery, &discoveryType)
		require.NoError(t, err)

		// Try to enable again
		anotherType := "bug"
		err = EnablePhase(state, PhaseDiscovery, &anotherType)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already enabled")
	})

	t.Run("enable already enabled design fails", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")

		// Enable first time
		err := EnablePhase(state, PhaseDesign, nil)
		require.NoError(t, err)

		// Try to enable again
		err = EnablePhase(state, PhaseDesign, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already enabled")
	})

	t.Run("cannot enable implementation phase", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")

		err := EnablePhase(state, PhaseImplementation, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "always enabled")
	})

	t.Run("cannot enable review phase", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")

		err := EnablePhase(state, PhaseReview, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "always enabled")
	})

	t.Run("cannot enable finalize phase", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")

		err := EnablePhase(state, PhaseFinalize, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "always enabled")
	})

	t.Run("all valid discovery types", func(t *testing.T) {
		validTypes := []string{"bug", "feature", "docs", "refactor", "general"}

		for _, discoveryType := range validTypes {
			t.Run(discoveryType, func(t *testing.T) {
				state := NewProjectState("test", "feat/test", "test")
				dt := discoveryType

				err := EnablePhase(state, PhaseDiscovery, &dt)
				require.NoError(t, err)
				// Discovery_type is of type any, so compare as string
				if dtPtr, ok := state.Phases.Discovery.Discovery_type.(*string); ok {
					assert.Equal(t, discoveryType, *dtPtr)
				} else {
					t.Fatal("discovery_type is not a string pointer")
				}
			})
		}
	})
}

func TestValidatePhaseCompletion_Discovery(t *testing.T) {
	t.Run("not enabled", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")

		err := ValidatePhaseCompletion(state, PhaseDiscovery)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not enabled")
	})

	t.Run("already completed", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Discovery.Enabled = true
		state.Phases.Discovery.Status = "completed"

		err := ValidatePhaseCompletion(state, PhaseDiscovery)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already completed")
	})

	t.Run("artifact not approved", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Discovery.Enabled = true
		state.Phases.Discovery.Status = "in_progress"
		state.Phases.Discovery.Artifacts = []schemas.Artifact{
			{Path: "doc.md", Approved: false},
		}

		err := ValidatePhaseCompletion(state, PhaseDiscovery)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not approved")
	})

	t.Run("all artifacts approved", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Discovery.Enabled = true
		state.Phases.Discovery.Status = "in_progress"
		state.Phases.Discovery.Artifacts = []schemas.Artifact{
			{Path: "doc.md", Approved: true},
		}

		err := ValidatePhaseCompletion(state, PhaseDiscovery)
		require.NoError(t, err)
	})
}

func TestValidatePhaseCompletion_Design(t *testing.T) {
	t.Run("artifact not approved", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Design.Enabled = true
		state.Phases.Design.Status = "in_progress"
		state.Phases.Design.Artifacts = []schemas.Artifact{
			{Path: "design.md", Approved: false},
		}

		err := ValidatePhaseCompletion(state, PhaseDesign)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not approved")
	})
}

func TestValidatePhaseCompletion_Implementation(t *testing.T) {
	t.Run("task not completed", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Implementation.Status = "in_progress"
		state.Phases.Implementation.Tasks = []schemas.Task{
			{Id: "010", Name: "Task 1", Status: "completed", Parallel: false, Dependencies: nil},
			{Id: "020", Name: "Task 2", Status: "in_progress", Parallel: false, Dependencies: nil},
		}

		err := ValidatePhaseCompletion(state, PhaseImplementation)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not completed or abandoned")
	})

	t.Run("all tasks completed", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Implementation.Status = "in_progress"
		state.Phases.Implementation.Tasks = []schemas.Task{
			{Id: "010", Name: "Task 1", Status: "completed", Parallel: false, Dependencies: nil},
			{Id: "020", Name: "Task 2", Status: "completed", Parallel: false, Dependencies: nil},
		}

		err := ValidatePhaseCompletion(state, PhaseImplementation)
		require.NoError(t, err)
	})

	t.Run("tasks completed and abandoned", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Implementation.Status = "in_progress"
		state.Phases.Implementation.Tasks = []schemas.Task{
			{Id: "010", Name: "Task 1", Status: "completed", Parallel: false, Dependencies: nil},
			{Id: "020", Name: "Task 2", Status: "abandoned", Parallel: false, Dependencies: nil},
		}

		err := ValidatePhaseCompletion(state, PhaseImplementation)
		require.NoError(t, err)
	})
}

func TestValidatePhaseCompletion_Review(t *testing.T) {
	t.Run("no reports", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Review.Status = "in_progress"

		err := ValidatePhaseCompletion(state, PhaseReview)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no review reports exist")
	})

	t.Run("latest report not pass", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Review.Status = "in_progress"
		now := time.Now()
		state.Phases.Review.Reports = []schemas.ReviewReport{
			{
				Id:         "001",
				Path:       "phases/review/001.md",
				Created_at: now,
				Assessment: "fail",
			},
		}

		err := ValidatePhaseCompletion(state, PhaseReview)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be 'pass'")
	})

	t.Run("latest report pass", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Review.Status = "in_progress"
		now := time.Now()
		state.Phases.Review.Reports = []schemas.ReviewReport{
			{
				Id:         "001",
				Path:       "phases/review/001.md",
				Created_at: now,
				Assessment: "pass",
			},
		}

		err := ValidatePhaseCompletion(state, PhaseReview)
		require.NoError(t, err)
	})
}

func TestValidatePhaseCompletion_Finalize(t *testing.T) {
	t.Run("project not deleted", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Finalize.Status = "in_progress"
		state.Phases.Finalize.Project_deleted = false

		err := ValidatePhaseCompletion(state, PhaseFinalize)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "project must be deleted")
	})

	t.Run("project deleted", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Finalize.Status = "in_progress"
		state.Phases.Finalize.Project_deleted = true

		err := ValidatePhaseCompletion(state, PhaseFinalize)
		require.NoError(t, err)
	})
}

func TestValidatePhaseCompletion_Invalid(t *testing.T) {
	t.Run("invalid phase name", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")

		err := ValidatePhaseCompletion(state, "invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid phase")
	})
}

func TestCompletePhase(t *testing.T) {
	t.Run("complete discovery phase", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Discovery.Enabled = true
		state.Phases.Discovery.Status = "in_progress"

		err := CompletePhase(state, PhaseDiscovery)
		require.NoError(t, err)

		assert.Equal(t, "completed", state.Phases.Discovery.Status)
		assert.NotNil(t, state.Phases.Discovery.Completed_at)
		assert.NotNil(t, state.Phases.Discovery.Started_at)
		// Verify they are ISO 8601 strings
		assert.IsType(t, "", state.Phases.Discovery.Completed_at)
		assert.IsType(t, "", state.Phases.Discovery.Started_at)
	})

	t.Run("complete discovery without starting it", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Discovery.Enabled = true
		state.Phases.Discovery.Status = "pending"
		state.Phases.Discovery.Started_at = nil

		err := CompletePhase(state, PhaseDiscovery)
		require.NoError(t, err)

		assert.Equal(t, "completed", state.Phases.Discovery.Status)
		assert.NotNil(t, state.Phases.Discovery.Started_at)
		assert.NotNil(t, state.Phases.Discovery.Completed_at)
		// Verify they are ISO 8601 strings
		assert.IsType(t, "", state.Phases.Discovery.Started_at)
		assert.IsType(t, "", state.Phases.Discovery.Completed_at)
	})

	t.Run("complete design phase", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Design.Enabled = true
		state.Phases.Design.Status = "in_progress"

		err := CompletePhase(state, PhaseDesign)
		require.NoError(t, err)

		assert.Equal(t, "completed", state.Phases.Design.Status)
		assert.NotNil(t, state.Phases.Design.Completed_at)
		assert.IsType(t, "", state.Phases.Design.Completed_at)
	})

	t.Run("complete implementation phase", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Implementation.Status = "in_progress"

		err := CompletePhase(state, PhaseImplementation)
		require.NoError(t, err)

		assert.Equal(t, "completed", state.Phases.Implementation.Status)
		assert.NotNil(t, state.Phases.Implementation.Completed_at)
		assert.IsType(t, "", state.Phases.Implementation.Completed_at)
	})

	t.Run("complete review phase", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Review.Status = "in_progress"
		now := time.Now()
		state.Phases.Review.Reports = []schemas.ReviewReport{
			{
				Id:         "001",
				Path:       "phases/review/001.md",
				Created_at: now,
				Assessment: "pass",
			},
		}

		err := CompletePhase(state, PhaseReview)
		require.NoError(t, err)

		assert.Equal(t, "completed", state.Phases.Review.Status)
		assert.NotNil(t, state.Phases.Review.Completed_at)
		assert.IsType(t, "", state.Phases.Review.Completed_at)
	})

	t.Run("complete finalize phase", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Finalize.Status = "in_progress"
		state.Phases.Finalize.Project_deleted = true

		err := CompletePhase(state, PhaseFinalize)
		require.NoError(t, err)

		assert.Equal(t, "completed", state.Phases.Finalize.Status)
		assert.NotNil(t, state.Phases.Finalize.Completed_at)
		assert.IsType(t, "", state.Phases.Finalize.Completed_at)
	})

	t.Run("fail to complete if validation fails", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Discovery.Enabled = true
		state.Phases.Discovery.Status = "in_progress"
		state.Phases.Discovery.Artifacts = []schemas.Artifact{
			{Path: "doc.md", Approved: false},
		}

		err := CompletePhase(state, PhaseDiscovery)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not approved")

		// Status should not change
		assert.Equal(t, "in_progress", state.Phases.Discovery.Status)
	})

	t.Run("updates project timestamp", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		originalTime := state.Project.Updated_at
		time.Sleep(10 * time.Millisecond)

		state.Phases.Implementation.Status = "in_progress"
		err := CompletePhase(state, PhaseImplementation)
		require.NoError(t, err)

		assert.True(t, state.Project.Updated_at.After(originalTime))
	})
}

func TestFormatPhaseStatus(t *testing.T) {
	t.Run("new project with default phases", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")

		output := FormatPhaseStatus(state)

		assert.Contains(t, output, "Phases:")
		assert.Contains(t, output, "Discovery")
		assert.Contains(t, output, "Design")
		assert.Contains(t, output, "Implementation")
		assert.Contains(t, output, "Review")
		assert.Contains(t, output, "Finalize")

		// Check that discovery and design are not enabled
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Discovery") {
				assert.Contains(t, line, "[ ]")
			}
			if strings.Contains(line, "Design") && !strings.Contains(line, "Design:") {
				assert.Contains(t, line, "[ ]")
			}
			if strings.Contains(line, "Implementation") {
				assert.Contains(t, line, "[✓]")
			}
		}
	})

	t.Run("project with enabled discovery and design", func(t *testing.T) {
		state := NewProjectState("test", "feat/test", "test")
		state.Phases.Discovery.Enabled = true
		state.Phases.Discovery.Status = "completed"
		state.Phases.Design.Enabled = true
		state.Phases.Design.Status = "in_progress"

		output := FormatPhaseStatus(state)

		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Discovery") {
				assert.Contains(t, line, "[✓]")
				assert.Contains(t, line, "completed")
			}
			if strings.Contains(line, "Design") && !strings.Contains(line, "Design:") {
				assert.Contains(t, line, "[✓]")
				assert.Contains(t, line, "in_progress")
			}
		}
	})

	t.Run("does not include project metadata", func(t *testing.T) {
		state := NewProjectState("my-feature", "feat/my-feature", "test description")

		output := FormatPhaseStatus(state)

		// Should not contain project metadata
		assert.NotContains(t, output, "Project:")
		assert.NotContains(t, output, "Description:")
		assert.NotContains(t, output, "Tasks:")

		// Should only contain phase information
		assert.Contains(t, output, "Phases:")
	})
}
