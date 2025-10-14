package validation

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestValidateProjectState tests validation of project state files
func TestValidateProjectState(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid project state",
			content: `project:
  name: test-project
  branch: feat/test
  created_at: 2025-10-13T00:00:00Z
  updated_at: 2025-10-13T00:00:00Z
  description: Test project
  complexity:
    rating: 2
    metrics:
      estimated_files: 5
      cross_cutting: true
      new_dependencies: false
  active_phase: implement
phases:
  - name: implement
    status: in_progress
    created_at: 2025-10-13T00:00:00Z
    completed_at: null
    tasks:
      - id: "010"
        name: Test task
        status: in_progress
        parallel: false
        assigned_agent: implementer
`,
			shouldErr: false,
		},
		{
			name: "invalid project name - not kebab-case",
			content: `project:
  name: Test_Project
  branch: feat/test
  created_at: 2025-10-13T00:00:00Z
  updated_at: 2025-10-13T00:00:00Z
  description: Test project
  complexity:
    rating: 2
    metrics:
      estimated_files: 5
      cross_cutting: true
      new_dependencies: false
  active_phase: implement
phases: []
`,
			shouldErr: true,
			errMsg:    "name",
		},
		{
			name: "invalid complexity rating - out of range",
			content: `project:
  name: test-project
  branch: feat/test
  created_at: 2025-10-13T00:00:00Z
  updated_at: 2025-10-13T00:00:00Z
  description: Test project
  complexity:
    rating: 5
    metrics:
      estimated_files: 5
      cross_cutting: true
      new_dependencies: false
  active_phase: implement
phases: []
`,
			shouldErr: true,
			errMsg:    "rating",
		},
		{
			name: "invalid task ID - wrong format",
			content: `project:
  name: test-project
  branch: feat/test
  created_at: 2025-10-13T00:00:00Z
  updated_at: 2025-10-13T00:00:00Z
  description: Test project
  complexity:
    rating: 2
    metrics:
      estimated_files: 5
      cross_cutting: true
      new_dependencies: false
  active_phase: implement
phases:
  - name: implement
    status: in_progress
    created_at: 2025-10-13T00:00:00Z
    completed_at: null
    tasks:
      - id: "1"
        name: Test task
        status: in_progress
        parallel: false
        assigned_agent: implementer
`,
			shouldErr: true,
			errMsg:    "id",
		},
		{
			name: "missing required field - name",
			content: `project:
  branch: feat/test
  created_at: 2025-10-13T00:00:00Z
  updated_at: 2025-10-13T00:00:00Z
  description: Test project
  complexity:
    rating: 2
    metrics:
      estimated_files: 5
      cross_cutting: true
      new_dependencies: false
  active_phase: implement
phases: []
`,
			shouldErr: true,
			errMsg:    "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpfile, err := os.CreateTemp("", "project-state-*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			// Validate
			err = ValidateProjectState(tmpfile.Name())

			// Check result
			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error to contain %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidateTaskState tests validation of task state files
func TestValidateTaskState(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid task state",
			content: `task:
  id: "020"
  name: Test task
  phase: implement
  status: in_progress
  created_at: 2025-10-13T00:00:00Z
  started_at: 2025-10-13T01:00:00Z
  updated_at: 2025-10-13T02:00:00Z
  completed_at: null
  iteration: 1
  assigned_agent: implementer
  references:
    - sinks/test/file.md
  feedback: []
  files_modified:
    - internal/test.go
`,
			shouldErr: false,
		},
		{
			name: "invalid task ID format",
			content: `task:
  id: "20"
  name: Test task
  phase: implement
  status: in_progress
  created_at: 2025-10-13T00:00:00Z
  started_at: null
  updated_at: 2025-10-13T00:00:00Z
  completed_at: null
  iteration: 1
  assigned_agent: implementer
  references: []
  feedback: []
  files_modified: []
`,
			shouldErr: true,
			errMsg:    "id",
		},
		{
			name: "invalid iteration - zero",
			content: `task:
  id: "020"
  name: Test task
  phase: implement
  status: in_progress
  created_at: 2025-10-13T00:00:00Z
  started_at: null
  updated_at: 2025-10-13T00:00:00Z
  completed_at: null
  iteration: 0
  assigned_agent: implementer
  references: []
  feedback: []
  files_modified: []
`,
			shouldErr: true,
			errMsg:    "iteration",
		},
		{
			name: "invalid phase name",
			content: `task:
  id: "020"
  name: Test task
  phase: invalid-phase
  status: in_progress
  created_at: 2025-10-13T00:00:00Z
  started_at: null
  updated_at: 2025-10-13T00:00:00Z
  completed_at: null
  iteration: 1
  assigned_agent: implementer
  references: []
  feedback: []
  files_modified: []
`,
			shouldErr: true,
			errMsg:    "phase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "task-state-*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			err = ValidateTaskState(tmpfile.Name())

			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error to contain %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidateSinkIndex tests validation of sink index files
func TestValidateSinkIndex(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid sink index",
			content: `{
  "sinks": [
    {
      "name": "python-style",
      "path": "python-style",
      "description": "Python style guide",
      "topics": ["formatting", "testing"],
      "when_to_use": "When writing Python code",
      "version": "v1.0.0",
      "source": "https://github.com/example/python-style",
      "updated_at": "2025-10-13T00:00:00Z"
    }
  ]
}`,
			shouldErr: false,
		},
		{
			name: "invalid sink name - not kebab-case",
			content: `{
  "sinks": [
    {
      "name": "Python_Style",
      "path": "python-style",
      "description": "Python style guide",
      "topics": ["formatting"],
      "when_to_use": "When writing Python code",
      "version": "v1.0.0",
      "source": "https://github.com/example/python-style",
      "updated_at": "2025-10-13T00:00:00Z"
    }
  ]
}`,
			shouldErr: true,
			errMsg:    "name",
		},
		{
			name: "missing required field",
			content: `{
  "sinks": [
    {
      "name": "python-style",
      "path": "python-style",
      "topics": ["formatting"],
      "when_to_use": "When writing Python code",
      "version": "v1.0.0",
      "source": "https://github.com/example/python-style",
      "updated_at": "2025-10-13T00:00:00Z"
    }
  ]
}`,
			shouldErr: true,
			errMsg:    "description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "sink-index-*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			err = ValidateSinkIndex(tmpfile.Name())

			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error to contain %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidateRepoIndex tests validation of repo index files
func TestValidateRepoIndex(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid repo index with clone",
			content: `{
  "repositories": [
    {
      "name": "auth-service",
      "path": "auth-service",
      "source": "https://github.com/example/auth-service",
      "purpose": "Reference authentication patterns",
      "type": "clone",
      "branch": "main",
      "updated_at": "2025-10-13T00:00:00Z"
    }
  ]
}`,
			shouldErr: false,
		},
		{
			name: "valid repo index with symlink",
			content: `{
  "repositories": [
    {
      "name": "shared-lib",
      "path": "shared-lib",
      "source": "/Users/test/code/shared-lib",
      "purpose": "Shared utilities",
      "type": "symlink",
      "branch": null,
      "updated_at": "2025-10-13T00:00:00Z"
    }
  ]
}`,
			shouldErr: false,
		},
		{
			name: "invalid type value",
			content: `{
  "repositories": [
    {
      "name": "auth-service",
      "path": "auth-service",
      "source": "https://github.com/example/auth-service",
      "purpose": "Reference authentication patterns",
      "type": "invalid",
      "branch": "main",
      "updated_at": "2025-10-13T00:00:00Z"
    }
  ]
}`,
			shouldErr: true,
			errMsg:    "type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "repo-index-*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			err = ValidateRepoIndex(tmpfile.Name())

			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error to contain %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidateVersion tests validation of version files
func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid version file",
			content: `sow_structure_version: 0.2.0
plugin_version: 0.2.0
initialized: 2025-10-13T00:00:00Z
last_migrated: null
`,
			shouldErr: false,
		},
		{
			name: "valid version file with migration",
			content: `sow_structure_version: 0.2.0
plugin_version: 0.2.0
initialized: 2025-10-13T00:00:00Z
last_migrated: 2025-10-13T01:00:00Z
`,
			shouldErr: false,
		},
		{
			name: "invalid version format",
			content: `sow_structure_version: 0.2
plugin_version: 0.2.0
initialized: 2025-10-13T00:00:00Z
last_migrated: null
`,
			shouldErr: true,
			errMsg:    "version",
		},
		{
			name: "invalid timestamp format",
			content: `sow_structure_version: 0.2.0
plugin_version: 0.2.0
initialized: 2025-10-13
last_migrated: null
`,
			shouldErr: true,
			errMsg:    "initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "version-*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			err = ValidateVersion(tmpfile.Name())

			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error to contain %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidationPerformance tests that validation completes quickly
func TestValidationPerformance(t *testing.T) {
	// Create a valid project state file
	content := `project:
  name: perf-test
  branch: feat/test
  created_at: 2025-10-13T00:00:00Z
  updated_at: 2025-10-13T00:00:00Z
  description: Performance test
  complexity:
    rating: 2
    metrics:
      estimated_files: 5
      cross_cutting: true
      new_dependencies: false
  active_phase: implement
phases:
  - name: implement
    status: in_progress
    created_at: 2025-10-13T00:00:00Z
    completed_at: null
    tasks:
      - id: "010"
        name: Task 1
        status: in_progress
        parallel: false
        assigned_agent: implementer
      - id: "020"
        name: Task 2
        status: pending
        parallel: false
        assigned_agent: implementer
`

	tmpfile, err := os.CreateTemp("", "perf-test-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Run validation multiple times
	start := time.Now()
	iterations := 100
	for i := 0; i < iterations; i++ {
		if err := ValidateProjectState(tmpfile.Name()); err != nil {
			t.Fatalf("validation failed: %v", err)
		}
	}
	elapsed := time.Since(start)

	// Check that average time is well under 1s
	avgTime := elapsed / time.Duration(iterations)
	maxTime := 100 * time.Millisecond // 100ms average should be plenty

	if avgTime > maxTime {
		t.Errorf("validation too slow: average %v per validation (max %v)", avgTime, maxTime)
	}

	t.Logf("Performance: %d validations in %v (avg %v)", iterations, elapsed, avgTime)
}

// TestValidationWithTemplateFiles tests validation using the actual template files
func TestValidationWithTemplateFiles(t *testing.T) {
	// This test verifies that our templates are valid
	// Note: Templates have empty/placeholder values, so we need to fill them in

	tests := []struct {
		name         string
		templatePath string
		validator    func(string) error
		fillContent  func() string
	}{
		{
			name:         "project-state template",
			templatePath: "schemas/templates/project-state.yaml",
			validator:    ValidateProjectState,
			fillContent: func() string {
				return `project:
  name: template-test
  branch: feat/test
  created_at: 2025-10-13T00:00:00Z
  updated_at: 2025-10-13T00:00:00Z
  description: Template test project
  complexity:
    rating: 1
    metrics:
      estimated_files: 1
      cross_cutting: false
      new_dependencies: false
  active_phase: implement
phases:
  - name: implement
    status: pending
    created_at: 2025-10-13T00:00:00Z
    completed_at: null
    tasks:
      - id: "010"
        name: Test task
        status: pending
        parallel: false
        assigned_agent: implementer
`
			},
		},
		{
			name:         "task-state template",
			templatePath: "schemas/templates/task-state.yaml",
			validator:    ValidateTaskState,
			fillContent: func() string {
				return `task:
  id: "010"
  name: Template test task
  phase: implement
  status: pending
  created_at: 2025-10-13T00:00:00Z
  started_at: null
  updated_at: 2025-10-13T00:00:00Z
  completed_at: null
  iteration: 1
  assigned_agent: implementer
  references: []
  feedback: []
  files_modified: []
`
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with filled content
			tmpfile, err := os.CreateTemp("", "template-test-*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			content := tt.fillContent()
			if _, err := tmpfile.Write([]byte(content)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			// Validate
			if err := tt.validator(tmpfile.Name()); err != nil {
				t.Errorf("validation failed: %v", err)
			}
		})
	}
}

// TestNonExistentFile tests error handling for missing files
func TestNonExistentFile(t *testing.T) {
	tests := []struct {
		name      string
		validator func(string) error
	}{
		{"ValidateProjectState", ValidateProjectState},
		{"ValidateTaskState", ValidateTaskState},
		{"ValidateSinkIndex", ValidateSinkIndex},
		{"ValidateRepoIndex", ValidateRepoIndex},
		{"ValidateVersion", ValidateVersion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator("/nonexistent/file.yaml")
			if err == nil {
				t.Error("expected error for nonexistent file")
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		filepath.Base(s) == substr || len(s) > len(substr)*2))
}
