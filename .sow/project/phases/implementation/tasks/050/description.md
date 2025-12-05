# Task Schema Update

## Context

This task is part of building the Executor System for sow - a multi-agent orchestration framework. The task schema needs to be updated to support session management and paused workflows.

**Why this is needed:**
- Session IDs must be stored in task state to enable session resumption
- Workers can pause (block) waiting for orchestrator input, requiring a `paused` status
- Session ID must persist through review iterations until task reaches terminal state
- Schema updates enable the state-based subprocess protocol

**Design reference:** The multi-agent architecture design document specifies the `session_id` field and `paused` status for task state.

## Requirements

### 1. Add `session_id` Field to Task Schema

Update `cli/schemas/project/task.cue` to add an optional `session_id` field:

```cue
#TaskState: {
    // ... existing fields ...

    // session_id is the optional session identifier for resumable agent conversations.
    // Set when a worker is spawned, persists through review iterations.
    // Only cleared when task reaches a terminal state (completed, abandoned).
    // Used by executors that support session resumption (claude, cursor).
    session_id?: string
}
```

### 2. Add `paused` Status

Update the `status` enum in `cli/schemas/project/task.cue`:

```cue
#TaskState: {
    // status indicates the current state of the task.
    // Must be one of:
    // - "pending" (not started)
    // - "in_progress" (actively working)
    // - "paused" (blocked, waiting for orchestrator input)
    // - "needs_review" (worker finished, awaiting orchestrator review)
    // - "completed" (successfully finished)
    // - "abandoned" (cancelled/obsolete)
    status: "pending" | "in_progress" | "paused" | "needs_review" | "completed" | "abandoned"
}
```

### 3. Regenerate Go Types

After updating the CUE schema, regenerate the Go types:

```bash
go generate ./cli/schemas/...
```

This will update `cli/schemas/project/cue_types_gen.go` with:
- `Session_id` field on `TaskState` struct (optional string)

### 4. Verify Schema Compatibility

Ensure existing task state files remain valid:
- New fields are optional (have `?` modifier)
- Existing status values still work
- No breaking changes to required fields

## Acceptance Criteria

1. **CUE Schema Updated** (`cli/schemas/project/task.cue`):
   - `session_id?: string` field added with proper documentation
   - `status` enum includes `"paused"` option
   - All existing fields and validation unchanged

2. **Go Types Regenerated** (`cli/schemas/project/cue_types_gen.go`):
   - `TaskState` struct includes `Session_id string` field with `json:"session_id,omitempty"` tag
   - Status field can accept "paused" as valid value

3. **Schema Validation Works**:
   - Existing task state YAML files still validate
   - New task state with session_id validates
   - New task state with paused status validates

4. **Unit tests verify**:
   - Task state with session_id loads correctly
   - Task state without session_id loads correctly (backward compatible)
   - Task state with paused status loads correctly
   - All existing task status values still work

## Technical Details

### Schema File Location

- CUE Schema: `cli/schemas/project/task.cue`
- Generated Go: `cli/schemas/project/cue_types_gen.go`

### CUE Syntax Reference

Optional field with documentation:
```cue
// session_id is the optional session identifier for resumable agent conversations.
session_id?: string
```

Status enum with additional value:
```cue
status: "pending" | "in_progress" | "paused" | "needs_review" | "completed" | "abandoned"
```

### Go Generation Command

The project uses CUE's gengotypes for code generation. Run:
```bash
cd cli && go generate ./schemas/...
```

Or directly:
```bash
cue exp gengotypes ./cli/schemas/project/
```

### JSON Tags

The generated Go struct will have:
```go
type TaskState struct {
    // ... existing fields ...
    Session_id string `json:"session_id,omitempty"`
}
```

The `omitempty` tag ensures backward compatibility - tasks without session_id will not have the field in their YAML.

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/schemas/project/task.cue` - Current task schema to update
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/schemas/project/cue_types_gen.go` - Generated Go types (will be regenerated)
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/cli/schemas/project/schemas_test.go` - Existing schema tests
- `/Users/josh/code/sow/.sow/worktrees/feat/executor-system-99/.sow/knowledge/designs/multi-agent-architecture.md` - Design document specifying session_id field and paused status (lines 844-863)

## Examples

### Updated CUE Schema

```cue
package project

import "time"

// TaskState represents a discrete unit of work within a phase.
// ...existing doc...
#TaskState: {
    // id is the unique three-digit identifier for this task.
    id: string & =~"^[0-9]{3}$"

    // name is the human-readable name of the task.
    name: string & !=""

    // phase identifies which phase this task belongs to.
    phase: string & !=""

    // status indicates the current state of the task.
    status: "pending" | "in_progress" | "paused" | "needs_review" | "completed" | "abandoned"

    // created_at is the timestamp when this task was created.
    created_at: time.Time

    // started_at is the optional timestamp when work on this task began.
    started_at?: time.Time

    // updated_at is the timestamp of the last modification to this task.
    updated_at: time.Time

    // completed_at is the optional timestamp when this task finished.
    completed_at?: time.Time

    // iteration is the current iteration number for this task.
    iteration: int & >=1

    // assigned_agent identifies the type of agent responsible for this task.
    assigned_agent: string & !=""

    // session_id is the optional session identifier for resumable agent conversations.
    // Set when a worker is spawned, persists through review iterations.
    // Only cleared when task reaches a terminal state (completed, abandoned).
    session_id?: string

    // inputs is the list of artifacts that this task consumes.
    inputs: [...#ArtifactState]

    // outputs is the list of artifacts that this task produces.
    outputs: [...#ArtifactState]

    // metadata holds task-specific data.
    metadata?: {...}
}
```

### Valid Task State YAML (New)

```yaml
id: "010"
name: "Implement executor interface"
phase: "implementation"
status: "paused"
created_at: 2025-01-15T10:00:00Z
updated_at: 2025-01-15T11:30:00Z
iteration: 1
assigned_agent: "implementer"
session_id: "550e8400-e29b-41d4-a716-446655440000"
inputs: []
outputs: []
```

### Valid Task State YAML (Backward Compatible - No Session ID)

```yaml
id: "010"
name: "Implement executor interface"
phase: "implementation"
status: "in_progress"
created_at: 2025-01-15T10:00:00Z
updated_at: 2025-01-15T11:30:00Z
iteration: 1
assigned_agent: "implementer"
inputs: []
outputs: []
```

### Test Examples

```go
func TestTaskState_SessionID(t *testing.T) {
    tests := []struct {
        name      string
        yaml      string
        wantSID   string
        wantValid bool
    }{
        {
            name: "with session_id",
            yaml: `
id: "010"
name: "Test task"
phase: "implementation"
status: "in_progress"
created_at: 2025-01-15T10:00:00Z
updated_at: 2025-01-15T10:00:00Z
iteration: 1
assigned_agent: "implementer"
session_id: "abc-123"
inputs: []
outputs: []
`,
            wantSID:   "abc-123",
            wantValid: true,
        },
        {
            name: "without session_id (backward compatible)",
            yaml: `
id: "010"
name: "Test task"
phase: "implementation"
status: "in_progress"
created_at: 2025-01-15T10:00:00Z
updated_at: 2025-01-15T10:00:00Z
iteration: 1
assigned_agent: "implementer"
inputs: []
outputs: []
`,
            wantSID:   "",
            wantValid: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var task TaskState
            err := yaml.Unmarshal([]byte(tt.yaml), &task)
            if (err == nil) != tt.wantValid {
                t.Errorf("Unmarshal error = %v, wantValid %v", err, tt.wantValid)
            }
            if err == nil && task.Session_id != tt.wantSID {
                t.Errorf("Session_id = %q, want %q", task.Session_id, tt.wantSID)
            }
        })
    }
}

func TestTaskState_PausedStatus(t *testing.T) {
    yaml := `
id: "010"
name: "Test task"
phase: "implementation"
status: "paused"
created_at: 2025-01-15T10:00:00Z
updated_at: 2025-01-15T10:00:00Z
iteration: 1
assigned_agent: "implementer"
inputs: []
outputs: []
`
    var task TaskState
    err := yaml.Unmarshal([]byte(yaml), &task)
    if err != nil {
        t.Fatalf("failed to unmarshal paused task: %v", err)
    }
    if task.Status != "paused" {
        t.Errorf("Status = %q, want %q", task.Status, "paused")
    }
}
```

## Dependencies

- No dependencies on other tasks - schema is foundational
- However, executors (tasks 030, 040) will use the session_id field

## Constraints

- Must maintain backward compatibility with existing task state files
- New field MUST be optional (use `?` modifier in CUE)
- Do NOT modify field names - stick with snake_case for consistency
- Do NOT remove or rename existing status values
- Generated code file should NOT be manually edited
- Run `go generate` after CUE schema changes
