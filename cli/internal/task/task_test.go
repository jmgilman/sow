package task

import (
	"strings"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateNextTaskID(t *testing.T) {
	t.Run("first task", func(t *testing.T) {
		tasks := []schemas.Task{}
		id := GenerateNextTaskID(tasks)
		assert.Equal(t, "010", id)
	})

	t.Run("second task", func(t *testing.T) {
		tasks := []schemas.Task{
			{Id: "010", Name: "Task 1", Status: "pending", Parallel: false, Dependencies: nil},
		}
		id := GenerateNextTaskID(tasks)
		assert.Equal(t, "020", id)
	})

	t.Run("multiple tasks", func(t *testing.T) {
		tasks := []schemas.Task{
			{Id: "010", Name: "Task 1", Status: "pending", Parallel: false, Dependencies: nil},
			{Id: "020", Name: "Task 2", Status: "pending", Parallel: false, Dependencies: nil},
			{Id: "030", Name: "Task 3", Status: "pending", Parallel: false, Dependencies: nil},
		}
		id := GenerateNextTaskID(tasks)
		assert.Equal(t, "040", id)
	})

	t.Run("gaps in numbering", func(t *testing.T) {
		tasks := []schemas.Task{
			{Id: "010", Name: "Task 1", Status: "pending", Parallel: false, Dependencies: nil},
			{Id: "030", Name: "Task 3", Status: "pending", Parallel: false, Dependencies: nil},
		}
		id := GenerateNextTaskID(tasks)
		// Should generate 040, not fill the gap
		assert.Equal(t, "040", id)
	})

	t.Run("out of order tasks", func(t *testing.T) {
		tasks := []schemas.Task{
			{Id: "030", Name: "Task 3", Status: "pending", Parallel: false, Dependencies: nil},
			{Id: "010", Name: "Task 1", Status: "pending", Parallel: false, Dependencies: nil},
			{Id: "020", Name: "Task 2", Status: "pending", Parallel: false, Dependencies: nil},
		}
		id := GenerateNextTaskID(tasks)
		assert.Equal(t, "040", id)
	})
}

func TestValidateTaskID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid ID 010",
			id:      "010",
			wantErr: false,
		},
		{
			name:    "valid ID 020",
			id:      "020",
			wantErr: false,
		},
		{
			name:    "valid ID 015",
			id:      "015",
			wantErr: false,
		},
		{
			name:    "valid ID 990",
			id:      "990",
			wantErr: false,
		},
		{
			name:    "invalid ID too short",
			id:      "10",
			wantErr: true,
		},
		{
			name:    "invalid ID too long",
			id:      "0100",
			wantErr: true,
		},
		{
			name:    "invalid ID not numeric",
			id:      "abc",
			wantErr: true,
		},
		{
			name:    "invalid ID too small",
			id:      "000",
			wantErr: true,
		},
		{
			name:    "invalid ID too large",
			id:      "999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaskID(tt.id)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewTaskState(t *testing.T) {
	id := "010"
	name := "Add authentication"
	assignedAgent := "implementer"

	state := NewTaskState(id, name, assignedAgent)

	// Test task metadata
	assert.Equal(t, id, state.Task.Id)
	assert.Equal(t, name, state.Task.Name)
	assert.Equal(t, "implementation", state.Task.Phase)
	assert.Equal(t, StatusPending, state.Task.Status)
	assert.WithinDuration(t, time.Now(), state.Task.Created_at, time.Second)
	assert.Nil(t, state.Task.Started_at)
	assert.WithinDuration(t, time.Now(), state.Task.Updated_at, time.Second)
	assert.Nil(t, state.Task.Completed_at)
	assert.Equal(t, int64(1), state.Task.Iteration)
	assert.Equal(t, assignedAgent, state.Task.Assigned_agent)
	assert.Empty(t, state.Task.References)
	assert.Empty(t, state.Task.Feedback)
	assert.Empty(t, state.Task.Files_modified)
}

func TestAddTaskToProjectState(t *testing.T) {
	t.Run("add first task", func(t *testing.T) {
		projectState := createTestProjectState()

		err := AddTaskToProjectState(projectState, "010", "Task 1", false, nil)
		require.NoError(t, err)

		assert.Len(t, projectState.Phases.Implementation.Tasks, 1)
		assert.Equal(t, "010", projectState.Phases.Implementation.Tasks[0].Id)
		assert.Equal(t, "Task 1", projectState.Phases.Implementation.Tasks[0].Name)
		assert.Equal(t, StatusPending, projectState.Phases.Implementation.Tasks[0].Status)
		assert.False(t, projectState.Phases.Implementation.Tasks[0].Parallel)
		assert.Nil(t, projectState.Phases.Implementation.Tasks[0].Dependencies)
	})

	t.Run("add task with dependencies", func(t *testing.T) {
		projectState := createTestProjectState()
		_ = AddTaskToProjectState(projectState, "010", "Task 1", false, nil)

		err := AddTaskToProjectState(projectState, "020", "Task 2", false, []string{"010"})
		require.NoError(t, err)

		assert.Len(t, projectState.Phases.Implementation.Tasks, 2)
		task2 := projectState.Phases.Implementation.Tasks[1]
		assert.NotNil(t, task2.Dependencies)
	})

	t.Run("add parallel task", func(t *testing.T) {
		projectState := createTestProjectState()

		err := AddTaskToProjectState(projectState, "010", "Task 1", true, nil)
		require.NoError(t, err)

		assert.True(t, projectState.Phases.Implementation.Tasks[0].Parallel)
	})

	t.Run("duplicate ID fails", func(t *testing.T) {
		projectState := createTestProjectState()
		_ = AddTaskToProjectState(projectState, "010", "Task 1", false, nil)

		err := AddTaskToProjectState(projectState, "010", "Task 2", false, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("invalid ID format fails", func(t *testing.T) {
		projectState := createTestProjectState()

		err := AddTaskToProjectState(projectState, "abc", "Task 1", false, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task ID")
	})

	t.Run("missing dependency fails", func(t *testing.T) {
		projectState := createTestProjectState()

		err := AddTaskToProjectState(projectState, "010", "Task 1", false, []string{"020"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("updates project timestamp", func(t *testing.T) {
		projectState := createTestProjectState()
		originalTime := projectState.Project.Updated_at
		time.Sleep(10 * time.Millisecond)

		_ = AddTaskToProjectState(projectState, "010", "Task 1", false, nil)

		assert.True(t, projectState.Project.Updated_at.After(originalTime))
	})
}

func TestFindTaskByID(t *testing.T) {
	t.Run("find existing task", func(t *testing.T) {
		projectState := createTestProjectState()
		_ = AddTaskToProjectState(projectState, "010", "Task 1", false, nil)
		_ = AddTaskToProjectState(projectState, "020", "Task 2", false, nil)

		task := FindTaskByID(projectState, "020")
		require.NotNil(t, task)
		assert.Equal(t, "020", task.Id)
		assert.Equal(t, "Task 2", task.Name)
	})

	t.Run("task not found", func(t *testing.T) {
		projectState := createTestProjectState()
		_ = AddTaskToProjectState(projectState, "010", "Task 1", false, nil)

		task := FindTaskByID(projectState, "999")
		assert.Nil(t, task)
	})

	t.Run("empty task list", func(t *testing.T) {
		projectState := createTestProjectState()

		task := FindTaskByID(projectState, "010")
		assert.Nil(t, task)
	})
}

func TestUpdateTaskStatusInProject(t *testing.T) {
	t.Run("update to in_progress", func(t *testing.T) {
		projectState := createTestProjectState()
		_ = AddTaskToProjectState(projectState, "010", "Task 1", false, nil)

		err := UpdateTaskStatusInProject(projectState, "010", StatusInProgress)
		require.NoError(t, err)

		task := FindTaskByID(projectState, "010")
		assert.Equal(t, StatusInProgress, task.Status)
	})

	t.Run("update to completed", func(t *testing.T) {
		projectState := createTestProjectState()
		_ = AddTaskToProjectState(projectState, "010", "Task 1", false, nil)

		err := UpdateTaskStatusInProject(projectState, "010", StatusCompleted)
		require.NoError(t, err)

		task := FindTaskByID(projectState, "010")
		assert.Equal(t, StatusCompleted, task.Status)
	})

	t.Run("task not found", func(t *testing.T) {
		projectState := createTestProjectState()

		err := UpdateTaskStatusInProject(projectState, "999", StatusCompleted)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("invalid status", func(t *testing.T) {
		projectState := createTestProjectState()
		_ = AddTaskToProjectState(projectState, "010", "Task 1", false, nil)

		err := UpdateTaskStatusInProject(projectState, "010", "invalid-status")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})

	t.Run("updates project timestamp", func(t *testing.T) {
		projectState := createTestProjectState()
		_ = AddTaskToProjectState(projectState, "010", "Task 1", false, nil)
		originalTime := projectState.Project.Updated_at
		time.Sleep(10 * time.Millisecond)

		_ = UpdateTaskStatusInProject(projectState, "010", StatusInProgress)

		assert.True(t, projectState.Project.Updated_at.After(originalTime))
	})
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{name: "pending", status: StatusPending, wantErr: false},
		{name: "in_progress", status: StatusInProgress, wantErr: false},
		{name: "completed", status: StatusCompleted, wantErr: false},
		{name: "abandoned", status: StatusAbandoned, wantErr: false},
		{name: "invalid", status: "invalid-status", wantErr: true},
		{name: "empty", status: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatus(tt.status)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	t.Run("update to in_progress sets started_at", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		err := UpdateTaskStatus(taskState, StatusInProgress)
		require.NoError(t, err)

		assert.Equal(t, StatusInProgress, taskState.Task.Status)
		assert.NotNil(t, taskState.Task.Started_at)
		assert.IsType(t, "", taskState.Task.Started_at)
		assert.Nil(t, taskState.Task.Completed_at)
	})

	t.Run("update to completed sets completed_at and started_at", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		err := UpdateTaskStatus(taskState, StatusCompleted)
		require.NoError(t, err)

		assert.Equal(t, StatusCompleted, taskState.Task.Status)
		assert.NotNil(t, taskState.Task.Started_at)
		assert.NotNil(t, taskState.Task.Completed_at)
		assert.IsType(t, "", taskState.Task.Started_at)
		assert.IsType(t, "", taskState.Task.Completed_at)
	})

	t.Run("update to abandoned sets completed_at", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		err := UpdateTaskStatus(taskState, StatusAbandoned)
		require.NoError(t, err)

		assert.Equal(t, StatusAbandoned, taskState.Task.Status)
		assert.NotNil(t, taskState.Task.Completed_at)
		assert.IsType(t, "", taskState.Task.Completed_at)
	})

	t.Run("doesn't overwrite existing started_at", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		existingStarted := "2024-01-01T10:00:00Z"
		taskState.Task.Started_at = existingStarted

		err := UpdateTaskStatus(taskState, StatusCompleted)
		require.NoError(t, err)

		assert.Equal(t, existingStarted, taskState.Task.Started_at)
	})

	t.Run("updates updated_at timestamp", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		originalTime := taskState.Task.Updated_at
		time.Sleep(10 * time.Millisecond)

		_ = UpdateTaskStatus(taskState, StatusInProgress)

		assert.True(t, taskState.Task.Updated_at.After(originalTime))
	})

	t.Run("invalid status fails", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		err := UpdateTaskStatus(taskState, "invalid-status")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})
}

func TestRemoveTaskFromProject(t *testing.T) {
	t.Run("remove task successfully", func(t *testing.T) {
		projectState := createTestProjectState()
		_ = AddTaskToProjectState(projectState, "010", "Task 1", false, nil)
		_ = AddTaskToProjectState(projectState, "020", "Task 2", false, nil)

		err := RemoveTaskFromProject(projectState, "010")
		require.NoError(t, err)

		assert.Len(t, projectState.Phases.Implementation.Tasks, 1)
		assert.Equal(t, "020", projectState.Phases.Implementation.Tasks[0].Id)
	})

	t.Run("task not found", func(t *testing.T) {
		projectState := createTestProjectState()

		err := RemoveTaskFromProject(projectState, "999")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("cannot remove task with dependents", func(t *testing.T) {
		projectState := createTestProjectState()
		_ = AddTaskToProjectState(projectState, "010", "Task 1", false, nil)
		_ = AddTaskToProjectState(projectState, "020", "Task 2", false, []string{"010"})

		err := RemoveTaskFromProject(projectState, "010")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "depends on it")
	})

	t.Run("updates project timestamp", func(t *testing.T) {
		projectState := createTestProjectState()
		_ = AddTaskToProjectState(projectState, "010", "Task 1", false, nil)
		originalTime := projectState.Project.Updated_at
		time.Sleep(10 * time.Millisecond)

		_ = RemoveTaskFromProject(projectState, "010")

		assert.True(t, projectState.Project.Updated_at.After(originalTime))
	})
}

func TestFormatTaskList(t *testing.T) {
	t.Run("empty task list", func(t *testing.T) {
		tasks := []schemas.Task{}

		output := FormatTaskList(tasks)
		assert.Contains(t, output, "No tasks yet")
	})

	t.Run("single task", func(t *testing.T) {
		tasks := []schemas.Task{
			{Id: "010", Name: "Task 1", Status: StatusPending, Parallel: false, Dependencies: nil},
		}

		output := FormatTaskList(tasks)
		assert.Contains(t, output, "Tasks:")
		assert.Contains(t, output, "ID")
		assert.Contains(t, output, "Status")
		assert.Contains(t, output, "Name")
		assert.Contains(t, output, "010")
		assert.Contains(t, output, "pending")
		assert.Contains(t, output, "Task 1")
	})

	t.Run("multiple tasks", func(t *testing.T) {
		tasks := []schemas.Task{
			{Id: "010", Name: "Task 1", Status: StatusPending, Parallel: false, Dependencies: nil},
			{Id: "020", Name: "Task 2", Status: StatusInProgress, Parallel: false, Dependencies: nil},
			{Id: "030", Name: "Task 3", Status: StatusCompleted, Parallel: false, Dependencies: nil},
		}

		output := FormatTaskList(tasks)
		assert.Contains(t, output, "010")
		assert.Contains(t, output, "020")
		assert.Contains(t, output, "030")
		assert.Contains(t, output, "pending")
		assert.Contains(t, output, "in_progress")
		assert.Contains(t, output, "completed")
	})

	t.Run("sorts tasks by ID", func(t *testing.T) {
		tasks := []schemas.Task{
			{Id: "030", Name: "Task 3", Status: StatusCompleted, Parallel: false, Dependencies: nil},
			{Id: "010", Name: "Task 1", Status: StatusPending, Parallel: false, Dependencies: nil},
			{Id: "020", Name: "Task 2", Status: StatusInProgress, Parallel: false, Dependencies: nil},
		}

		output := FormatTaskList(tasks)
		lines := strings.Split(output, "\n")

		// Find the data lines (skip header)
		var dataLines []string
		for _, line := range lines {
			if strings.Contains(line, "010") || strings.Contains(line, "020") || strings.Contains(line, "030") {
				dataLines = append(dataLines, line)
			}
		}

		// Verify order
		assert.Contains(t, dataLines[0], "010")
		assert.Contains(t, dataLines[1], "020")
		assert.Contains(t, dataLines[2], "030")
	})
}

func TestFormatTaskStatus(t *testing.T) {
	t.Run("pending task", func(t *testing.T) {
		taskState := NewTaskState("010", "Add authentication", "implementer")

		output := FormatTaskStatus(taskState)
		assert.Contains(t, output, "Task: 010 - Add authentication")
		assert.Contains(t, output, "Status: pending")
		assert.Contains(t, output, "Phase: implementation")
		assert.Contains(t, output, "Iteration: 1")
		assert.Contains(t, output, "Assigned Agent: implementer")
		assert.Contains(t, output, "Started:   not started")
		assert.Contains(t, output, "Completed: not completed")
	})

	t.Run("in_progress task", func(t *testing.T) {
		taskState := NewTaskState("010", "Add authentication", "implementer")
		_ = UpdateTaskStatus(taskState, StatusInProgress)

		output := FormatTaskStatus(taskState)
		assert.Contains(t, output, "Status: in_progress")
		assert.NotContains(t, output, "Started:   not started")
	})

	t.Run("completed task", func(t *testing.T) {
		taskState := NewTaskState("010", "Add authentication", "implementer")
		_ = UpdateTaskStatus(taskState, StatusCompleted)

		output := FormatTaskStatus(taskState)
		assert.Contains(t, output, "Status: completed")
		assert.NotContains(t, output, "Completed: not completed")
	})

	t.Run("task with references", func(t *testing.T) {
		taskState := NewTaskState("010", "Add authentication", "implementer")
		taskState.Task.References = []string{"knowledge/adrs/001.md", "sinks/style-guide/auth.md"}

		output := FormatTaskStatus(taskState)
		assert.Contains(t, output, "References:")
		assert.Contains(t, output, "knowledge/adrs/001.md")
		assert.Contains(t, output, "sinks/style-guide/auth.md")
	})

	t.Run("task with feedback", func(t *testing.T) {
		taskState := NewTaskState("010", "Add authentication", "implementer")
		taskState.Task.Feedback = []schemas.Feedback{
			{Id: "001", Created_at: time.Now(), Status: "pending"},
			{Id: "002", Created_at: time.Now(), Status: "applied"},
		}

		output := FormatTaskStatus(taskState)
		assert.Contains(t, output, "Feedback: 2 items")
	})

	t.Run("task with files modified", func(t *testing.T) {
		taskState := NewTaskState("010", "Add authentication", "implementer")
		taskState.Task.Files_modified = []string{"src/auth.go", "src/auth_test.go"}

		output := FormatTaskStatus(taskState)
		assert.Contains(t, output, "Files Modified:")
		assert.Contains(t, output, "src/auth.go")
		assert.Contains(t, output, "src/auth_test.go")
	})
}

func TestIncrementTaskIteration(t *testing.T) {
	t.Run("increment from initial value", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		assert.Equal(t, int64(1), taskState.Task.Iteration)

		err := IncrementTaskIteration(taskState)
		require.NoError(t, err)

		assert.Equal(t, int64(2), taskState.Task.Iteration)
	})

	t.Run("increment multiple times", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		_ = IncrementTaskIteration(taskState)
		_ = IncrementTaskIteration(taskState)
		_ = IncrementTaskIteration(taskState)

		assert.Equal(t, int64(4), taskState.Task.Iteration)
	})

	t.Run("updates updated_at timestamp", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		originalTime := taskState.Task.Updated_at
		time.Sleep(10 * time.Millisecond)

		_ = IncrementTaskIteration(taskState)

		assert.True(t, taskState.Task.Updated_at.After(originalTime))
	})
}

func TestSetTaskAgent(t *testing.T) {
	t.Run("change agent successfully", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		assert.Equal(t, "implementer", taskState.Task.Assigned_agent)

		err := SetTaskAgent(taskState, "reviewer")
		require.NoError(t, err)

		assert.Equal(t, "reviewer", taskState.Task.Assigned_agent)
	})

	t.Run("empty agent name fails", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		err := SetTaskAgent(taskState, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("updates updated_at timestamp", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		originalTime := taskState.Task.Updated_at
		time.Sleep(10 * time.Millisecond)

		_ = SetTaskAgent(taskState, "architect")

		assert.True(t, taskState.Task.Updated_at.After(originalTime))
	})
}

func TestAddTaskReference(t *testing.T) {
	t.Run("add first reference", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		assert.Empty(t, taskState.Task.References)

		err := AddTaskReference(taskState, "refs/python-style/conventions.md")
		require.NoError(t, err)

		assert.Len(t, taskState.Task.References, 1)
		assert.Equal(t, "refs/python-style/conventions.md", taskState.Task.References[0])
	})

	t.Run("add multiple references", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		_ = AddTaskReference(taskState, "refs/python-style/conventions.md")
		_ = AddTaskReference(taskState, "knowledge/adrs/001.md")
		_ = AddTaskReference(taskState, "knowledge/architecture/system.md")

		assert.Len(t, taskState.Task.References, 3)
	})

	t.Run("duplicate reference is ignored", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		_ = AddTaskReference(taskState, "refs/python-style/conventions.md")
		err := AddTaskReference(taskState, "refs/python-style/conventions.md")
		require.NoError(t, err)

		assert.Len(t, taskState.Task.References, 1)
	})

	t.Run("empty path fails", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		err := AddTaskReference(taskState, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("updates updated_at timestamp", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		originalTime := taskState.Task.Updated_at
		time.Sleep(10 * time.Millisecond)

		_ = AddTaskReference(taskState, "refs/test.md")

		assert.True(t, taskState.Task.Updated_at.After(originalTime))
	})
}

func TestAddModifiedFile(t *testing.T) {
	t.Run("add first file", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		assert.Empty(t, taskState.Task.Files_modified)

		err := AddModifiedFile(taskState, "src/auth/jwt.py")
		require.NoError(t, err)

		assert.Len(t, taskState.Task.Files_modified, 1)
		assert.Equal(t, "src/auth/jwt.py", taskState.Task.Files_modified[0])
	})

	t.Run("add multiple files", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		_ = AddModifiedFile(taskState, "src/auth/jwt.py")
		_ = AddModifiedFile(taskState, "src/auth/jwt_test.py")
		_ = AddModifiedFile(taskState, "docs/auth.md")

		assert.Len(t, taskState.Task.Files_modified, 3)
	})

	t.Run("duplicate file is ignored", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		_ = AddModifiedFile(taskState, "src/auth/jwt.py")
		err := AddModifiedFile(taskState, "src/auth/jwt.py")
		require.NoError(t, err)

		assert.Len(t, taskState.Task.Files_modified, 1)
	})

	t.Run("empty path fails", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		err := AddModifiedFile(taskState, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("updates updated_at timestamp", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		originalTime := taskState.Task.Updated_at
		time.Sleep(10 * time.Millisecond)

		_ = AddModifiedFile(taskState, "src/test.go")

		assert.True(t, taskState.Task.Updated_at.After(originalTime))
	})
}

func TestGenerateNextFeedbackID(t *testing.T) {
	t.Run("first feedback", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		id := GenerateNextFeedbackID(taskState)
		assert.Equal(t, "001", id)
	})

	t.Run("second feedback", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		taskState.Task.Feedback = []schemas.Feedback{
			{Id: "001", Created_at: time.Now(), Status: "pending"},
		}

		id := GenerateNextFeedbackID(taskState)
		assert.Equal(t, "002", id)
	})

	t.Run("multiple feedback items", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		taskState.Task.Feedback = []schemas.Feedback{
			{Id: "001", Created_at: time.Now(), Status: "pending"},
			{Id: "002", Created_at: time.Now(), Status: "addressed"},
			{Id: "003", Created_at: time.Now(), Status: "pending"},
		}

		id := GenerateNextFeedbackID(taskState)
		assert.Equal(t, "004", id)
	})

	t.Run("gaps in numbering", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		taskState.Task.Feedback = []schemas.Feedback{
			{Id: "001", Created_at: time.Now(), Status: "pending"},
			{Id: "005", Created_at: time.Now(), Status: "pending"},
		}

		id := GenerateNextFeedbackID(taskState)
		// Should generate 006, not fill the gap
		assert.Equal(t, "006", id)
	})
}

func TestAddFeedback(t *testing.T) {
	t.Run("add first feedback", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		assert.Empty(t, taskState.Task.Feedback)

		err := AddFeedback(taskState, "001")
		require.NoError(t, err)

		assert.Len(t, taskState.Task.Feedback, 1)
		assert.Equal(t, "001", taskState.Task.Feedback[0].Id)
		assert.Equal(t, "pending", taskState.Task.Feedback[0].Status)
		assert.WithinDuration(t, time.Now(), taskState.Task.Feedback[0].Created_at, time.Second)
	})

	t.Run("add multiple feedback items", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		_ = AddFeedback(taskState, "001")
		_ = AddFeedback(taskState, "002")
		_ = AddFeedback(taskState, "003")

		assert.Len(t, taskState.Task.Feedback, 3)
	})

	t.Run("duplicate ID fails", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		_ = AddFeedback(taskState, "001")

		err := AddFeedback(taskState, "001")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("invalid ID length fails", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		err := AddFeedback(taskState, "01")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be 3 digits")
	})

	t.Run("non-numeric ID fails", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")

		err := AddFeedback(taskState, "abc")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be numeric")
	})

	t.Run("updates updated_at timestamp", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		originalTime := taskState.Task.Updated_at
		time.Sleep(10 * time.Millisecond)

		_ = AddFeedback(taskState, "001")

		assert.True(t, taskState.Task.Updated_at.After(originalTime))
	})
}

func TestMarkFeedbackAddressed(t *testing.T) {
	t.Run("mark feedback as addressed", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		_ = AddFeedback(taskState, "001")

		err := MarkFeedbackAddressed(taskState, "001")
		require.NoError(t, err)

		assert.Equal(t, "addressed", taskState.Task.Feedback[0].Status)
	})

	t.Run("mark specific feedback in multiple items", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		_ = AddFeedback(taskState, "001")
		_ = AddFeedback(taskState, "002")
		_ = AddFeedback(taskState, "003")

		err := MarkFeedbackAddressed(taskState, "002")
		require.NoError(t, err)

		assert.Equal(t, "pending", taskState.Task.Feedback[0].Status)
		assert.Equal(t, "addressed", taskState.Task.Feedback[1].Status)
		assert.Equal(t, "pending", taskState.Task.Feedback[2].Status)
	})

	t.Run("feedback not found fails", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		_ = AddFeedback(taskState, "001")

		err := MarkFeedbackAddressed(taskState, "999")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("updates updated_at timestamp", func(t *testing.T) {
		taskState := NewTaskState("010", "Task 1", "implementer")
		_ = AddFeedback(taskState, "001")
		originalTime := taskState.Task.Updated_at
		time.Sleep(10 * time.Millisecond)

		_ = MarkFeedbackAddressed(taskState, "001")

		assert.True(t, taskState.Task.Updated_at.After(originalTime))
	})
}

// Helper function to create a test project state.
func createTestProjectState() *schemas.ProjectState {
	now := time.Now()
	state := &schemas.ProjectState{}
	state.Project.Name = "test-project"
	state.Project.Branch = "feat/test"
	state.Project.Description = "Test project"
	state.Project.Created_at = now
	state.Project.Updated_at = now

	state.Phases.Implementation.Status = "pending"
	state.Phases.Implementation.Enabled = true
	state.Phases.Implementation.Tasks = []schemas.Task{}

	return state
}
