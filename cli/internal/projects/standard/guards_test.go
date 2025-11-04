package standard

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/schemas/project"
)

func TestPhaseOutputApproved(t *testing.T) {
	tests := getPhaseOutputApprovedTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := phaseOutputApproved(tt.project, tt.phaseName, tt.outputType); got != tt.want {
				t.Errorf("phaseOutputApproved() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getPhaseOutputApprovedTests() []struct {
	name       string
	project    *state.Project
	phaseName  string
	outputType string
	want       bool
} {
	return []struct {
		name       string
		project    *state.Project
		phaseName  string
		outputType string
		want       bool
	}{
		{
			name: "returns true when output exists and approved",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"planning": {
							Outputs: []project.ArtifactState{
								{Type: "task_list", Path: "tasks.md", Approved: true},
							},
						},
					},
				},
			},
			phaseName: "planning", outputType: "task_list", want: true,
		},
		{
			name: "returns false when output not approved",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"planning": {
							Outputs: []project.ArtifactState{
								{Type: "task_list", Path: "tasks.md", Approved: false},
							},
						},
					},
				},
			},
			phaseName: "planning", outputType: "task_list", want: false,
		},
		{
			name: "returns false when phase missing",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{},
				},
			},
			phaseName: "planning", outputType: "task_list", want: false,
		},
		{
			name: "returns false when output type not found",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"planning": {
							Outputs: []project.ArtifactState{
								{Type: "other", Path: "other.md", Approved: true},
							},
						},
					},
				},
			},
			phaseName: "planning", outputType: "task_list", want: false,
		},
		{
			name: "returns false when outputs array empty",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"planning": {Outputs: []project.ArtifactState{}},
					},
				},
			},
			phaseName: "planning", outputType: "task_list", want: false,
		},
		{
			name: "returns true when multiple outputs exist and correct one approved",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"planning": {
							Outputs: []project.ArtifactState{
								{Type: "design_doc", Path: "design.md", Approved: false},
								{Type: "task_list", Path: "tasks.md", Approved: true},
								{Type: "other", Path: "other.md", Approved: true},
							},
						},
					},
				},
			},
			phaseName: "planning", outputType: "task_list", want: true,
		},
	}
}

func TestPhaseMetadataBool(t *testing.T) {
	tests := getPhaseMetadataBoolTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := phaseMetadataBool(tt.project, tt.phaseName, tt.key); got != tt.want {
				t.Errorf("phaseMetadataBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:funlen // Test data generation function
func getPhaseMetadataBoolTests() []struct {
		name      string
		project   *state.Project
		phaseName string
		key       string
		want      bool
	} {
	return []struct {
		name      string
		project   *state.Project
		phaseName string
		key       string
		want      bool
	}{
		{
			name: "returns true when key exists and value is true",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"implementation": {
							Metadata: map[string]interface{}{
								"tasks_approved": true,
							},
						},
					},
				},
			},
			phaseName: "implementation",
			key:       "tasks_approved",
			want:      true,
		},
		{
			name: "returns false when key exists and value is false",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"implementation": {
							Metadata: map[string]interface{}{
								"tasks_approved": false,
							},
						},
					},
				},
			},
			phaseName: "implementation",
			key:       "tasks_approved",
			want:      false,
		},
		{
			name: "returns false when key missing",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"implementation": {
							Metadata: map[string]interface{}{
								"other_key": true,
							},
						},
					},
				},
			},
			phaseName: "implementation",
			key:       "tasks_approved",
			want:      false,
		},
		{
			name: "returns false when phase missing",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{},
				},
			},
			phaseName: "implementation",
			key:       "tasks_approved",
			want:      false,
		},
		{
			name: "returns false when value is wrong type (string)",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"implementation": {
							Metadata: map[string]interface{}{
								"tasks_approved": "true",
							},
						},
					},
				},
			},
			phaseName: "implementation",
			key:       "tasks_approved",
			want:      false,
		},
		{
			name: "returns false when value is wrong type (int)",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"implementation": {
							Metadata: map[string]interface{}{
								"tasks_approved": 1,
							},
						},
					},
				},
			},
			phaseName: "implementation",
			key:       "tasks_approved",
			want:      false,
		},
		{
			name: "returns false when metadata is nil",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"implementation": {
							Metadata: nil,
						},
					},
				},
			},
			phaseName: "implementation",
			key:       "tasks_approved",
			want:      false,
		},
	}
}

func TestAllTasksComplete(t *testing.T) {
	tests := getAllTasksCompleteTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := allTasksComplete(tt.project); got != tt.want {
				t.Errorf("allTasksComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:funlen // Test data generation function
func getAllTasksCompleteTests() []struct {
		name    string
		project *state.Project
		want    bool
	} {
	now := time.Now()
	return []struct {
		name    string
		project *state.Project
		want    bool
	}{
		{
			name: "returns true when all tasks completed",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"implementation": {
							Tasks: []project.TaskState{
								{
									Id:     "001",
									Status: "completed",
								},
								{
									Id:     "002",
									Status: "completed",
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "returns true when all tasks completed or abandoned",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"implementation": {
							Tasks: []project.TaskState{
								{
									Id:     "001",
									Status: "completed",
								},
								{
									Id:     "002",
									Status: "abandoned",
								},
								{
									Id:     "003",
									Status: "completed",
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "returns false when mix of completed and pending",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"implementation": {
							Tasks: []project.TaskState{
								{
									Id:     "001",
									Status: "completed",
								},
								{
									Id:     "002",
									Status: "pending",
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "returns false when implementation phase missing",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{},
				},
			},
			want: false,
		},
		{
			name: "returns false when no tasks exist",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"implementation": {
							Tasks: []project.TaskState{},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "returns false when task in_progress",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"implementation": {
							Tasks: []project.TaskState{
								{
									Id:         "001",
									Status:     "in_progress",
									Updated_at: now,
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "returns false when task has other status",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"implementation": {
							Tasks: []project.TaskState{
								{
									Id:     "001",
									Status: "completed",
								},
								{
									Id:     "002",
									Status: "pending",
								},
							},
						},
					},
				},
			},
			want: false,
		},
	}
}

func TestLatestReviewApproved(t *testing.T) {
	tests := getLatestReviewApprovedTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := latestReviewApproved(tt.project); got != tt.want {
				t.Errorf("latestReviewApproved() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:funlen // Test data generation function
func getLatestReviewApprovedTests() []struct {
		name    string
		project *state.Project
		want    bool
	} {
	now := time.Now()
	return []struct {
		name    string
		project *state.Project
		want    bool
	}{
		{
			name: "returns true when latest review approved",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"review": {
							Outputs: []project.ArtifactState{
								{
									Type:       "review",
									Path:       "review1.md",
									Approved:   true,
									Created_at: now,
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "returns false when latest review not approved",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"review": {
							Outputs: []project.ArtifactState{
								{
									Type:       "review",
									Path:       "review1.md",
									Approved:   false,
									Created_at: now,
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "returns true when multiple reviews and latest approved",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"review": {
							Outputs: []project.ArtifactState{
								{
									Type:       "review",
									Path:       "review1.md",
									Approved:   false,
									Created_at: now.Add(-2 * time.Hour),
								},
								{
									Type:       "review",
									Path:       "review2.md",
									Approved:   true,
									Created_at: now.Add(-1 * time.Hour),
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "returns false when multiple reviews and latest not approved",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"review": {
							Outputs: []project.ArtifactState{
								{
									Type:       "review",
									Path:       "review1.md",
									Approved:   true,
									Created_at: now.Add(-2 * time.Hour),
								},
								{
									Type:       "review",
									Path:       "review2.md",
									Approved:   false,
									Created_at: now.Add(-1 * time.Hour),
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "returns false when no reviews",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"review": {
							Outputs: []project.ArtifactState{},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "returns false when review phase missing",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{},
				},
			},
			want: false,
		},
		{
			name: "returns true when mix of artifact types and latest review approved",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"review": {
							Outputs: []project.ArtifactState{
								{
									Type:       "review",
									Path:       "review1.md",
									Approved:   false,
									Created_at: now.Add(-3 * time.Hour),
								},
								{
									Type:       "other",
									Path:       "other.md",
									Approved:   true,
									Created_at: now.Add(-2 * time.Hour),
								},
								{
									Type:       "review",
									Path:       "review2.md",
									Approved:   true,
									Created_at: now.Add(-1 * time.Hour),
								},
							},
						},
					},
				},
			},
			want: true,
		},
	}
}

func TestProjectDeleted(t *testing.T) {
	tests := []struct {
		name    string
		project *state.Project
		want    bool
	}{
		{
			name: "returns true when project_deleted is true",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"finalize": {
							Metadata: map[string]interface{}{
								"project_deleted": true,
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "returns false when project_deleted is false",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"finalize": {
							Metadata: map[string]interface{}{
								"project_deleted": false,
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "returns false when project_deleted key missing",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"finalize": {
							Metadata: map[string]interface{}{
								"other_key": true,
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "returns false when finalize phase missing",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{},
				},
			},
			want: false,
		},
		{
			name: "returns false when metadata is nil",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"finalize": {
							Metadata: nil,
						},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := projectDeleted(tt.project); got != tt.want {
				t.Errorf("projectDeleted() = %v, want %v", got, tt.want)
			}
		})
	}
}
