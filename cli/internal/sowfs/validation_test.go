package sowfs

import (
	"testing"

	"github.com/jmgilman/go/fs/billy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAll_EmptyStructure(t *testing.T) {
	// Create minimal .sow structure
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/knowledge", 0755)
	fs.MkdirAll(".sow/refs", 0755)

	// Create SowFS
	sowFS, err := NewSowFSWithFS(fs, "/test")
	require.NoError(t, err)
	defer sowFS.Close()

	// Validate all
	result := sowFS.ValidateAll()

	// Should have no errors (nothing to validate)
	assert.False(t, result.HasErrors())
	assert.Empty(t, result.Errors)
}

func TestValidateAll_ValidRefsIndex(t *testing.T) {
	// Create .sow structure with valid refs index
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/refs", 0755)

	validIndex := `{
		"version": "1.0.0",
		"refs": []
	}`
	fs.WriteFile(".sow/refs/index.json", []byte(validIndex), 0644)

	// Create SowFS
	sowFS, err := NewSowFSWithFS(fs, "/test")
	require.NoError(t, err)
	defer sowFS.Close()

	// Validate all
	result := sowFS.ValidateAll()

	// Should have no errors
	assert.False(t, result.HasErrors())
}

func TestValidateAll_InvalidRefsIndex(t *testing.T) {
	// Create .sow structure with invalid refs index
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/refs", 0755)

	invalidIndex := `{
		"version": "not-semver",
		"refs": []
	}`
	fs.WriteFile(".sow/refs/index.json", []byte(invalidIndex), 0644)

	// Create SowFS - should fail during construction due to validation
	sowFS, err := NewSowFSWithFS(fs, "/test")

	// Should get validation error during construction
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "refs/index.json")
	assert.Contains(t, err.Error(), "refs-committed")
	assert.Nil(t, sowFS)
}

func TestValidateAll_ValidLocalIndex(t *testing.T) {
	// Create .sow structure with valid local refs index
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/refs", 0755)

	validLocalIndex := `{
		"version": "1.0.0",
		"refs": []
	}`
	fs.WriteFile(".sow/refs/index.local.json", []byte(validLocalIndex), 0644)

	// Create SowFS
	sowFS, err := NewSowFSWithFS(fs, "/test")
	require.NoError(t, err)
	defer sowFS.Close()

	// Validate all
	result := sowFS.ValidateAll()

	// Should have no errors
	assert.False(t, result.HasErrors())
}

func TestValidateAll_InvalidLocalIndex(t *testing.T) {
	// Create .sow structure with invalid local refs index
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/refs", 0755)

	invalidLocalIndex := `{
		"version": "bad-version",
		"refs": []
	}`
	fs.WriteFile(".sow/refs/index.local.json", []byte(invalidLocalIndex), 0644)

	// Create SowFS - should fail during construction due to validation
	sowFS, err := NewSowFSWithFS(fs, "/test")

	// Should get validation error during construction
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "refs/index.local.json")
	assert.Contains(t, err.Error(), "refs-local")
	assert.Nil(t, sowFS)
}

func TestValidateAll_ValidProjectState(t *testing.T) {
	// Create .sow structure with valid project state
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/project", 0755)

	validProjectState := `
project:
  name: my-feature
  branch: feat/my-feature
  description: A test feature
  created_at: "2025-10-15T10:00:00Z"
  updated_at: "2025-10-15T10:00:00Z"
phases:
  discovery:
    enabled: false
    status: skipped
    created_at: "2025-10-15T10:00:00Z"
    artifacts: []
  design:
    enabled: false
    status: skipped
    created_at: "2025-10-15T10:00:00Z"
    artifacts: []
  implementation:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    tasks: []
  review:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    iteration: 1
    reports: []
  finalize:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    project_deleted: false
`
	fs.WriteFile(".sow/project/state.yaml", []byte(validProjectState), 0644)

	// Create SowFS
	sowFS, err := NewSowFSWithFS(fs, "/test")
	require.NoError(t, err)
	defer sowFS.Close()

	// Validate all
	result := sowFS.ValidateAll()

	// Should have no errors
	assert.False(t, result.HasErrors())
}

func TestValidateAll_InvalidProjectState(t *testing.T) {
	// Create .sow structure with invalid project state
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/project", 0755)

	// Invalid: name has uppercase and spaces
	invalidProjectState := `
project:
  name: "Invalid Name With Spaces"
  branch: ""
  description: "Test"
  created_at: "2025-10-15T10:00:00Z"
  updated_at: "2025-10-15T10:00:00Z"
`
	fs.WriteFile(".sow/project/state.yaml", []byte(invalidProjectState), 0644)

	// Create SowFS - should fail during construction due to validation
	sowFS, err := NewSowFSWithFS(fs, "/test")

	// Should get validation error during construction
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "project/state.yaml")
	assert.Contains(t, err.Error(), "project-state")
	assert.Nil(t, sowFS)
}

func TestValidateAll_ValidTaskState(t *testing.T) {
	// Create .sow structure with valid task state
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)

	validTaskState := `
task:
  id: "010"
  name: Test task
  phase: implementation
  status: pending
  created_at: "2025-10-15T10:00:00Z"
  updated_at: "2025-10-15T10:00:00Z"
  iteration: 1
  references: []
  feedback: []
  files_modified: []
`
	fs.WriteFile(".sow/project/phases/implementation/tasks/010/state.yaml", []byte(validTaskState), 0644)

	// Create SowFS
	sowFS, err := NewSowFSWithFS(fs, "/test")
	require.NoError(t, err)
	defer sowFS.Close()

	// Validate all
	result := sowFS.ValidateAll()

	// Should have no errors
	assert.False(t, result.HasErrors())
}

func TestValidateAll_InvalidTaskState(t *testing.T) {
	// Create .sow structure with valid project but invalid task state
	fs := billy.NewMemory()
	err := fs.MkdirAll(".sow/project", 0755)
	require.NoError(t, err)

	// Create a minimal valid project state so validateProject runs
	validProjectState := `project:
  name: test-project
  branch: feat/test
  description: Test
  created_at: "2025-10-15T10:00:00Z"
  updated_at: "2025-10-15T10:00:00Z"
phases:
  discovery:
    enabled: false
    status: skipped
    created_at: "2025-10-15T10:00:00Z"
    artifacts: []
  design:
    enabled: false
    status: skipped
    created_at: "2025-10-15T10:00:00Z"
    artifacts: []
  implementation:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    tasks: []
  review:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    iteration: 1
    reports: []
  finalize:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    project_deleted: false
`
	err = fs.WriteFile(".sow/project/state.yaml", []byte(validProjectState), 0644)
	require.NoError(t, err)

	// Create task directory with invalid task state
	err = fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)
	require.NoError(t, err)

	// Invalid: id doesn't match pattern (needs 3+ digits)
	invalidTaskState := `task:
  id: "10"
  name: Test task
  phase: implementation
  status: pending
  created_at: "2025-10-15T10:00:00Z"
  updated_at: "2025-10-15T10:00:00Z"
  iteration: 1
  references: []
  feedback: []
  files_modified: []
`
	err = fs.WriteFile(".sow/project/phases/implementation/tasks/010/state.yaml", []byte(invalidTaskState), 0644)
	require.NoError(t, err)

	// Create SowFS - should fail during construction due to validation
	sowFS, err := NewSowFSWithFS(fs, "/test")

	// Should get validation error during construction
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "tasks/010/state.yaml")
	assert.Contains(t, err.Error(), "task-state")
	assert.Nil(t, sowFS)
}

func TestValidateAll_TaskDirectoryWithoutStateFile(t *testing.T) {
	// Create .sow structure with task directory but no state.yaml
	fs := billy.NewMemory()

	// Create a minimal valid project state so validateProject runs
	validProjectState := `project:
  name: test-project
  branch: feat/test
  description: Test
  created_at: "2025-10-15T10:00:00Z"
  updated_at: "2025-10-15T10:00:00Z"
phases:
  discovery:
    enabled: false
    status: skipped
    created_at: "2025-10-15T10:00:00Z"
    artifacts: []
  design:
    enabled: false
    status: skipped
    created_at: "2025-10-15T10:00:00Z"
    artifacts: []
  implementation:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    tasks: []
  review:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    iteration: 1
    reports: []
  finalize:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    project_deleted: false
`
	fs.WriteFile(".sow/project/state.yaml", []byte(validProjectState), 0644)

	// Create task directory without state.yaml
	fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)
	// No state.yaml file created

	// Create SowFS - should fail during construction due to validation
	sowFS, err := NewSowFSWithFS(fs, "/test")

	// Should get validation error during construction
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "tasks/010/state.yaml")
	assert.Contains(t, err.Error(), "missing")
	assert.Nil(t, sowFS)
}

func TestValidateAll_MultipleTasks(t *testing.T) {
	// Create .sow structure with multiple tasks
	fs := billy.NewMemory()

	// Create a minimal valid project state so validateProject runs
	validProjectState := `project:
  name: test-project
  branch: feat/test
  description: Test
  created_at: "2025-10-15T10:00:00Z"
  updated_at: "2025-10-15T10:00:00Z"
phases:
  discovery:
    enabled: false
    status: skipped
    created_at: "2025-10-15T10:00:00Z"
    artifacts: []
  design:
    enabled: false
    status: skipped
    created_at: "2025-10-15T10:00:00Z"
    artifacts: []
  implementation:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    tasks: []
  review:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    iteration: 1
    reports: []
  finalize:
    enabled: true
    status: pending
    created_at: "2025-10-15T10:00:00Z"
    project_deleted: false
`
	fs.WriteFile(".sow/project/state.yaml", []byte(validProjectState), 0644)

	fs.MkdirAll(".sow/project/phases/implementation/tasks/010", 0755)
	fs.MkdirAll(".sow/project/phases/implementation/tasks/020", 0755)

	validTaskState := `task:
  id: "010"
  name: Task 1
  phase: implementation
  status: pending
  created_at: "2025-10-15T10:00:00Z"
  updated_at: "2025-10-15T10:00:00Z"
  iteration: 1
  references: []
  feedback: []
  files_modified: []
`
	fs.WriteFile(".sow/project/phases/implementation/tasks/010/state.yaml", []byte(validTaskState), 0644)

	invalidTaskState := `task:
  id: "20"
  name: Task 2
  phase: implementation
  status: pending
  created_at: "2025-10-15T10:00:00Z"
  updated_at: "2025-10-15T10:00:00Z"
  iteration: 1
  references: []
  feedback: []
  files_modified: []
`
	fs.WriteFile(".sow/project/phases/implementation/tasks/020/state.yaml", []byte(invalidTaskState), 0644)

	// Create SowFS - should fail during construction due to validation
	sowFS, err := NewSowFSWithFS(fs, "/test")

	// Should get validation error during construction
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "tasks/020/state.yaml")
	assert.Nil(t, sowFS)
}

func TestValidateAll_MultipleErrors(t *testing.T) {
	// Create .sow structure with multiple invalid files
	fs := billy.NewMemory()
	fs.MkdirAll(".sow/refs", 0755)
	fs.MkdirAll(".sow/project", 0755)

	// Invalid refs index
	invalidRefsIndex := `{"version": "bad", "refs": []}`
	fs.WriteFile(".sow/refs/index.json", []byte(invalidRefsIndex), 0644)

	// Invalid local refs index
	invalidLocalIndex := `{"version": "also-bad", "refs": []}`
	fs.WriteFile(".sow/refs/index.local.json", []byte(invalidLocalIndex), 0644)

	// Invalid project state
	invalidProjectState := `
project:
  name: "Invalid Name"
  branch: ""
  description: "Test"
  created_at: "2025-10-15T10:00:00Z"
  updated_at: "2025-10-15T10:00:00Z"
`
	fs.WriteFile(".sow/project/state.yaml", []byte(invalidProjectState), 0644)

	// Create SowFS - should fail during construction due to validation
	sowFS, err := NewSowFSWithFS(fs, "/test")

	// Should get validation error during construction with all 3 files
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	// All three invalid files should be mentioned
	assert.Contains(t, err.Error(), "refs/index.json")
	assert.Contains(t, err.Error(), "refs/index.local.json")
	assert.Contains(t, err.Error(), "project/state.yaml")
	assert.Nil(t, sowFS)
}

func TestValidationResult_Error(t *testing.T) {
	result := &ValidationResult{}

	// Empty result
	assert.False(t, result.HasErrors())
	assert.Equal(t, "", result.Error())

	// Add errors
	result.Add("file1.yaml", "project-state", assert.AnError)
	result.Add("file2.json", "refs-committed", assert.AnError)

	// Should format nicely
	assert.True(t, result.HasErrors())
	errStr := result.Error()
	assert.Contains(t, errStr, "validation failed with 2 error(s)")
	assert.Contains(t, errStr, "file1.yaml")
	assert.Contains(t, errStr, "file2.json")
}

func TestValidationResult_Merge(t *testing.T) {
	result1 := &ValidationResult{}
	result1.Add("file1.yaml", "project-state", assert.AnError)

	result2 := &ValidationResult{}
	result2.Add("file2.json", "refs-committed", assert.AnError)
	result2.Add("file3.json", "refs-local", assert.AnError)

	// Merge result2 into result1
	result1.Merge(result2)

	// Should have all 3 errors
	assert.True(t, result1.HasErrors())
	assert.Len(t, result1.Errors, 3)
}

func TestValidationResult_AddNilError(t *testing.T) {
	result := &ValidationResult{}

	// Adding nil error should not add to result
	result.Add("file.yaml", "project-state", nil)

	assert.False(t, result.HasErrors())
	assert.Empty(t, result.Errors)
}
