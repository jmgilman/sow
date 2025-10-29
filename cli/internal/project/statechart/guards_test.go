package statechart

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
				{Approved: true},
				{Approved: true},
			},
			expected: true,
		},
		{
			name: "one not approved returns false",
			artifacts: []phases.Artifact{
				{Approved: true},
				{Approved: false},
			},
			expected: false,
		},
		{
			name: "all not approved returns false",
			artifacts: []phases.Artifact{
				{Approved: false},
				{Approved: false},
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
