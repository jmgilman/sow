package state

import (
	"testing"

	"github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/stretchr/testify/assert"
)

func TestTasksComplete(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []phases.Task
		expected bool
	}{
		{
			name:     "empty task list returns false",
			tasks:    []phases.Task{},
			expected: false,
		},
		{
			name: "all completed returns true",
			tasks: []phases.Task{
				{Status: "completed"},
				{Status: "completed"},
			},
			expected: true,
		},
		{
			name: "all abandoned returns true",
			tasks: []phases.Task{
				{Status: "abandoned"},
				{Status: "abandoned"},
			},
			expected: true,
		},
		{
			name: "mix of completed and abandoned returns true",
			tasks: []phases.Task{
				{Status: "completed"},
				{Status: "abandoned"},
			},
			expected: true,
		},
		{
			name: "one pending returns false",
			tasks: []phases.Task{
				{Status: "completed"},
				{Status: "pending"},
			},
			expected: false,
		},
		{
			name: "one in_progress returns false",
			tasks: []phases.Task{
				{Status: "completed"},
				{Status: "in_progress"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TasksComplete(tt.tasks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestArtifactsApproved(t *testing.T) {
	approvedTrue := true
	approvedFalse := false

	tests := []struct {
		name      string
		artifacts []phases.Artifact
		expected  bool
	}{
		{
			name:      "empty artifact list returns false",
			artifacts: []phases.Artifact{},
			expected:  false,
		},
		{
			name: "all approved returns true",
			artifacts: []phases.Artifact{
				{Approved: &approvedTrue},
				{Approved: &approvedTrue},
			},
			expected: true,
		},
		{
			name: "one not approved returns false",
			artifacts: []phases.Artifact{
				{Approved: &approvedTrue},
				{Approved: &approvedFalse},
			},
			expected: false,
		},
		{
			name: "all not approved returns false",
			artifacts: []phases.Artifact{
				{Approved: &approvedFalse},
				{Approved: &approvedFalse},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArtifactsApproved(tt.artifacts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMinTaskCount(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []phases.Task
		min      int
		expected bool
	}{
		{
			name:     "empty list with min 0",
			tasks:    []phases.Task{},
			min:      0,
			expected: true,
		},
		{
			name:     "empty list with min 1",
			tasks:    []phases.Task{},
			min:      1,
			expected: false,
		},
		{
			name:     "1 task with min 1",
			tasks:    []phases.Task{{}},
			min:      1,
			expected: true,
		},
		{
			name:     "1 task with min 2",
			tasks:    []phases.Task{{}},
			min:      2,
			expected: false,
		},
		{
			name:     "3 tasks with min 2",
			tasks:    []phases.Task{{}, {}, {}},
			min:      2,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MinTaskCount(tt.tasks, tt.min)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasArtifactWithType(t *testing.T) {
	tests := []struct {
		name         string
		artifacts    []phases.Artifact
		artifactType string
		expected     bool
	}{
		{
			name:         "empty list returns false",
			artifacts:    []phases.Artifact{},
			artifactType: "task_list",
			expected:     false,
		},
		{
			name: "artifact with matching type returns true",
			artifacts: []phases.Artifact{
				{
					Metadata: map[string]interface{}{
						"type": "task_list",
					},
				},
			},
			artifactType: "task_list",
			expected:     true,
		},
		{
			name: "artifact with different type returns false",
			artifacts: []phases.Artifact{
				{
					Metadata: map[string]interface{}{
						"type": "review",
					},
				},
			},
			artifactType: "task_list",
			expected:     false,
		},
		{
			name: "artifact with no metadata returns false",
			artifacts: []phases.Artifact{
				{Metadata: nil},
			},
			artifactType: "task_list",
			expected:     false,
		},
		{
			name: "artifact with no type in metadata returns false",
			artifacts: []phases.Artifact{
				{
					Metadata: map[string]interface{}{
						"other": "value",
					},
				},
			},
			artifactType: "task_list",
			expected:     false,
		},
		{
			name: "multiple artifacts, one matches",
			artifacts: []phases.Artifact{
				{
					Metadata: map[string]interface{}{
						"type": "review",
					},
				},
				{
					Metadata: map[string]interface{}{
						"type": "task_list",
					},
				},
			},
			artifactType: "task_list",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasArtifactWithType(tt.artifacts, tt.artifactType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnyTaskInProgress(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []phases.Task
		expected bool
	}{
		{
			name:     "empty list returns false",
			tasks:    []phases.Task{},
			expected: false,
		},
		{
			name: "one in_progress returns true",
			tasks: []phases.Task{
				{Status: "in_progress"},
			},
			expected: true,
		},
		{
			name: "no in_progress returns false",
			tasks: []phases.Task{
				{Status: "completed"},
				{Status: "pending"},
			},
			expected: false,
		},
		{
			name: "multiple in_progress returns true",
			tasks: []phases.Task{
				{Status: "completed"},
				{Status: "in_progress"},
				{Status: "in_progress"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnyTaskInProgress(tt.tasks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGuard_ClosureCapture_SimpleBool verifies that guards capture simple boolean values.
func TestGuard_ClosureCapture_SimpleBool(t *testing.T) {
	stateA := State("Pending")
	stateB := State("Approved")
	eventApprove := Event("approve")

	// Variable to be captured
	approved := false

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventApprove, WithGuard(func() bool {
		return approved
	}))
	machine := builder.Build()

	// Guard should fail when approved=false
	can, err := machine.CanFire(eventApprove)
	assert.NoError(t, err)
	assert.False(t, can, "Guard should fail when approved=false")

	err = machine.Fire(eventApprove)
	assert.Error(t, err, "Fire should fail when guard returns false")
	assert.Equal(t, stateA, machine.State(), "State should not change when guard fails")

	// Change captured variable
	approved = true

	// Guard should now pass
	can, err = machine.CanFire(eventApprove)
	assert.NoError(t, err)
	assert.True(t, can, "Guard should pass when approved=true")

	err = machine.Fire(eventApprove)
	assert.NoError(t, err, "Fire should succeed when guard returns true")
	assert.Equal(t, stateB, machine.State(), "State should change when guard passes")
}

// TestGuard_ClosureCapture_Integer verifies that guards can capture integer values.
func TestGuard_ClosureCapture_Integer(t *testing.T) {
	stateA := State("InProgress")
	stateB := State("Complete")
	eventComplete := Event("complete")

	// Variable to be captured
	taskCount := 0
	requiredTasks := 5

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventComplete, WithGuard(func() bool {
		return taskCount >= requiredTasks
	}))
	machine := builder.Build()

	// Should fail with insufficient tasks
	can, err := machine.CanFire(eventComplete)
	assert.NoError(t, err)
	assert.False(t, can)

	// Update task count
	taskCount = 5

	// Should now pass
	can, err = machine.CanFire(eventComplete)
	assert.NoError(t, err)
	assert.True(t, can)
}

// TestGuard_ClosureCapture_Struct verifies that guards can capture complex structs.
func TestGuard_ClosureCapture_Struct(t *testing.T) {
	type ReviewStatus struct {
		Approved     bool
		ReviewerName string
		Score        int
	}

	stateA := State("UnderReview")
	stateB := State("Merged")
	eventMerge := Event("merge")

	// Complex struct to be captured
	review := ReviewStatus{
		Approved:     false,
		ReviewerName: "",
		Score:        0,
	}

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventMerge, WithGuard(func() bool {
		return review.Approved && review.Score >= 3
	}))
	machine := builder.Build()

	// Should fail initially
	can, err := machine.CanFire(eventMerge)
	assert.NoError(t, err)
	assert.False(t, can)

	// Approve but with low score
	review.Approved = true
	review.Score = 2

	can, err = machine.CanFire(eventMerge)
	assert.NoError(t, err)
	assert.False(t, can, "Should fail when score is too low")

	// Update score
	review.Score = 5

	can, err = machine.CanFire(eventMerge)
	assert.NoError(t, err)
	assert.True(t, can, "Should pass when approved and score is high")
}

// TestGuard_ClosureCapture_MultipleGuards verifies that multiple guards can capture
// the same variables independently.
func TestGuard_ClosureCapture_MultipleGuards(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	stateC := State("C")
	eventToB := Event("to_b")
	eventToC := Event("to_c")

	// Shared captured variable
	threshold := 5

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventToB, WithGuard(func() bool {
		return threshold > 3
	}))
	builder.AddTransition(stateA, stateC, eventToC, WithGuard(func() bool {
		return threshold <= 3
	}))
	machine := builder.Build()

	// threshold = 5, so to_b should be allowed, to_c should not
	can, err := machine.CanFire(eventToB)
	assert.NoError(t, err)
	assert.True(t, can)

	can, err = machine.CanFire(eventToC)
	assert.NoError(t, err)
	assert.False(t, can)

	// Update threshold
	threshold = 2

	// Now to_c should be allowed, to_b should not
	can, err = machine.CanFire(eventToB)
	assert.NoError(t, err)
	assert.False(t, can)

	can, err = machine.CanFire(eventToC)
	assert.NoError(t, err)
	assert.True(t, can)
}

// TestGuard_EvaluationTiming verifies that guards are evaluated when CanFire/Fire is called,
// not at registration time.
func TestGuard_EvaluationTiming(t *testing.T) {
	stateA := State("A")
	stateB := State("B")
	eventGo := Event("go")

	// Variable starts false
	condition := false

	// Register transition with guard
	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventGo, WithGuard(func() bool {
		return condition
	}))
	machine := builder.Build()

	// Guard should fail now
	can, err := machine.CanFire(eventGo)
	assert.NoError(t, err)
	assert.False(t, can)

	// Update condition AFTER machine is built
	condition = true

	// Guard should now pass because it's evaluated at call time, not registration time
	can, err = machine.CanFire(eventGo)
	assert.NoError(t, err)
	assert.True(t, can, "Guard should be evaluated at call time, not registration time")
}

// TestGuard_ClosureCapture_TaskList verifies guards work with domain objects like task lists.
func TestGuard_ClosureCapture_TaskList(t *testing.T) {
	stateA := State("Implementation")
	stateB := State("Review")
	eventComplete := Event("complete")

	// Captured task list
	tasks := []phases.Task{
		{Status: "pending"},
		{Status: "in_progress"},
	}

	builder := NewBuilder(stateA, nil)
	builder.AddTransition(stateA, stateB, eventComplete, WithGuard(func() bool {
		return TasksComplete(tasks)
	}))
	machine := builder.Build()

	// Should fail with incomplete tasks
	can, err := machine.CanFire(eventComplete)
	assert.NoError(t, err)
	assert.False(t, can)

	// Complete all tasks
	tasks[0].Status = "completed"
	tasks[1].Status = "completed"

	// Should now pass
	can, err = machine.CanFire(eventComplete)
	assert.NoError(t, err)
	assert.True(t, can)

	err = machine.Fire(eventComplete)
	assert.NoError(t, err)
	assert.Equal(t, stateB, machine.State())
}
