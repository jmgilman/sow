# Standard Project Advance: Simplified Command Hierarchy

**Summary of exploration:** A completely unified command structure based on containers with inputs/outputs and single-command state progression.

## Core Concepts

### Three Container Types

1. **Project** - Top-level container with metadata
2. **Phase** - Workflow stage with inputs/outputs
3. **Task** - Unit of work with inputs/outputs

### Container Structure

Every container has:
- **Singular fields** - scalar values (status, iteration, name, etc.)
- **Inputs** - artifacts consumed by the container
- **Outputs** - artifacts produced by the container
- **Sub-containers** - nested containers (phases in project, tasks in phase)

### Data Flow Model

- **Inputs** = artifacts consumed (references, feedback, context)
- **Outputs** = artifacts produced (modified files, task lists, reports)
- **Artifacts** = indexed list items with type and fields
- **Advance** = single command that progresses state

## Schemas

### Universal Artifact Schema

All inputs and outputs use the same artifact schema:

```yaml
artifact:
  type: string              # "context", "task_list", "review", "reference", "feedback", "modified"
  path: string              # Path relative to .sow/project/
  approved: boolean?        # Optional approval flag
  created_at: timestamp
  metadata: map?            # Extensible metadata for type-specific fields
    # Examples:
    # - assessment: "pass" | "fail"  (for review artifacts)
    # - status: "pending" | "addressed"  (for feedback artifacts)
    # - source: "human" | "generated"
    # ... any other fields
```

**Key principle:** Type-specific fields go in `metadata`, not as direct fields. This keeps the artifact schema universal and extensible.

### Phase Schema

```yaml
phase:
  # Singular fields
  name: string              # planning, implementation, review, finalize
  status: string            # pending, in_progress, completed
  enabled: boolean
  created_at: timestamp
  started_at: timestamp?
  completed_at: timestamp?

  # Metadata (extensible for phase-specific flags)
  metadata:
    tasks_approved: boolean?
    project_deleted: boolean?

  # Input artifacts (consumed by this phase)
  inputs:
    - type: string          # "context", "design", etc.
      path: string
      approved: boolean?
      created_at: timestamp
      metadata: map?

  # Output artifacts (produced by this phase)
  outputs:
    - type: string          # "task_list", "review", etc.
      path: string
      approved: boolean?
      created_at: timestamp
      metadata: map?        # assessment, etc. go here

  # Sub-containers (implementation phase only)
  tasks:
    - id: string            # Gap-numbered: 010, 020, 030
      name: string
      status: string        # pending, in_progress, completed, abandoned
      parallel: boolean
      dependencies: [string]
      metadata: map?
```

**Note:** The `tasks` field in phases is a simple list for task metadata. The full task state lives in separate task state files.

### Task State File Schema

Separate from the phase.tasks[] list, each task has a detailed state file at `.sow/project/phases/implementation/tasks/<id>/state.yaml`:

```yaml
task:
  # Identity and assignment
  id: string                # Gap-numbered: 010, 020, 030
  name: string
  phase: string             # Always "implementation" in standard project
  status: string            # pending, in_progress, completed, abandoned
  iteration: integer        # 1, 2, 3, ...
  assigned_agent: string    # implementer, reviewer, etc.

  # Timestamps
  created_at: timestamp
  started_at: timestamp?
  updated_at: timestamp
  completed_at: timestamp?

  # Context (simple string arrays in current design)
  references: [string]      # Paths to reference files
  feedback: [Feedback]      # Human feedback items
  files_modified: [string]  # Paths to modified files

# Feedback is currently a simple object in TaskState
feedback:
  id: string                # "001", "002", etc.
  created_at: timestamp
  status: string            # "pending", "addressed"
```

**Proposed change:** In the new design, `references`, `feedback`, and `files_modified` would become artifact lists using the universal Artifact schema:

```yaml
task:
  # ... (same fields as above)

  # Input artifacts (consumed by this task)
  inputs:
    - type: "reference"     # Reference document
      path: "sinks/style-guide.md"
      created_at: timestamp
    - type: "feedback"      # Human feedback
      path: "feedback/001.md"
      created_at: timestamp
      metadata:
        status: "pending"

  # Output artifacts (produced by this task)
  outputs:
    - type: "modified"      # Modified file
      path: "src/auth/jwt.ts"
      created_at: timestamp
```

## Command Structure

### Field and Metadata Handling

When setting fields on artifacts using `input set` or `output set`:
- **Known fields** (type, path, approved, created_at) are set directly on the artifact
- **Unknown fields** are automatically stored in the artifact's `metadata` map
- This allows type-specific fields (like `assessment` for reviews or `status` for feedback) to be set naturally without explicit metadata nesting

**Examples:**
```bash
# Sets approved field directly
sow output set --index 0 approved true

# Sets metadata.assessment automatically
sow output set --index 0 assessment pass

# Sets metadata.status automatically
sow task input set --id 010 --index 2 status addressed
```

This design keeps the artifact schema universal while allowing flexible, type-specific extensions through metadata.

### Project Commands

```bash
sow project new <name> --description "..."
sow project continue
sow project set <field> <value>
sow project delete
```

### Phase Commands

```bash
sow phase set <field> <value> [--phase <name>]
sow phase get <field> [--phase <name>]
```

### Task Commands

```bash
sow task add <name> [--agent <agent>] [--description "..."]
sow task set --id <task_id> <field> <value>
sow task get --id <task_id> <field>
sow task abandon --id <task_id>
sow task list
```

### Input Commands (Phase-Level)

```bash
sow input add --type <type> [--phase <name>] [fields...]
sow input set --index <n> <field> <value> [--phase <name>]
sow input remove --index <n> [--phase <name>]
sow input list [--phase <name>]
```

### Output Commands (Phase-Level)

```bash
sow output add --type <type> [--phase <name>] [fields...]
sow output set --index <n> <field> <value> [--phase <name>]
sow output remove --index <n> [--phase <name>]
sow output list [--phase <name>]
```

### Task Input Commands

```bash
sow task input add --id <task_id> --type <type> [fields...]
sow task input set --id <task_id> --index <n> <field> <value>
sow task input remove --id <task_id> --index <n>
sow task input list --id <task_id>
```

### Task Output Commands

```bash
sow task output add --id <task_id> --type <type> [fields...]
sow task output set --id <task_id> --index <n> <field> <value>
sow task output remove --id <task_id> --index <n>
sow task output list --id <task_id>
```

### State Progression

```bash
sow advance
```

Single command that:
- Examines current state
- Checks guards
- Fires appropriate event
- Transitions to next state

## Complete Lifecycle Example

### 1. Initialize Project

```bash
sow project new "add-auth" --description "Add JWT authentication to API"
# → Creates project, enters PlanningActive state
```

**State:**
```yaml
project:
  name: "add-auth"
  description: "Add JWT authentication to API"
  branch: "feat/add-auth"
  statechart:
    current_state: "PlanningActive"
```

---

### 2. Planning Phase

#### Add Context Inputs

```bash
sow input add --type context --path "discovery/jwt-research.md"
sow input add --type context --path "sinks/auth-patterns.md"
```

#### Create Task List Output

```bash
sow output add --type task_list --path "planning/tasks.md"
```

#### Review and Approve

```bash
sow output list
# [0] task_list: planning/tasks.md (not approved)

sow output set --index 0 approved true
```

**State:**
```yaml
phases:
  planning:
    status: "in_progress"
    inputs:
      - type: "context"
        path: "discovery/jwt-research.md"
        created_at: "2025-01-15T09:00:00Z"
      - type: "context"
        path: "sinks/auth-patterns.md"
        created_at: "2025-01-15T09:30:00Z"
    outputs:
      - type: "task_list"
        path: "planning/tasks.md"
        created_at: "2025-01-15T10:30:00Z"
        approved: true
```

#### Advance to Implementation

```bash
sow advance
# → PlanningPhase.Advance() returns EventCompletePlanning
# → State machine checks PlanningComplete guard (task list approved)
# → Transitions to ImplementationPlanning
```

---

### 3. Implementation Planning

#### Create Tasks

```bash
sow task add "Implement JWT signing" --agent implementer
sow task add "Implement JWT verification" --agent implementer
sow task add "Add auth middleware" --agent implementer
```

**State:**
```yaml
phases:
  implementation:
    status: "in_progress"
    tasks:
      - id: "010"
        name: "Implement JWT signing"
        status: "pending"
        assigned_agent: "implementer"
        iteration: 1
        inputs: []
        outputs: []
      - id: "020"
        name: "Implement JWT verification"
        status: "pending"
        assigned_agent: "implementer"
        iteration: 1
        inputs: []
        outputs: []
      - id: "030"
        name: "Add auth middleware"
        status: "pending"
        assigned_agent: "implementer"
        iteration: 1
        inputs: []
        outputs: []
```

#### Approve Tasks

```bash
sow phase set tasks_approved true
```

#### Advance to Execution

```bash
sow advance
# → ImplementationPhase.Advance() in ImplementationPlanning state
# → Returns EventTasksApproved
# → State machine checks TasksApproved guard
# → Transitions to ImplementationExecuting
```

---

### 4. Implementation Execution

#### Start Task

```bash
sow task set --id 010 status in_progress
```

#### Add Task Inputs

```bash
sow task input add --id 010 --type reference --path "sinks/style-guide.md"
sow task input add --id 010 --type reference --path "knowledge/jwt-design.md"
```

**State:**
```yaml
tasks:
  - id: "010"
    status: "in_progress"
    inputs:
      - type: "reference"
        path: "sinks/style-guide.md"
        created_at: "2025-01-15T11:00:00Z"
      - type: "reference"
        path: "knowledge/jwt-design.md"
        created_at: "2025-01-15T11:00:00Z"
```

#### Worker Executes Task

(Worker creates implementation, outputs are auto-tracked via logging)

**State after execution:**
```yaml
tasks:
  - id: "010"
    outputs:
      - type: "modified"
        path: "src/auth/jwt.ts"
        created_at: "2025-01-15T13:00:00Z"
      - type: "modified"
        path: "src/auth/verify.ts"
        created_at: "2025-01-15T13:30:00Z"
```

#### Human Provides Feedback

Human writes feedback to files, then adds them as artifacts:

```bash
# Human creates feedback files (e.g., via editor or command)
# .sow/project/phases/implementation/tasks/010/feedback/001.md
# .sow/project/phases/implementation/tasks/010/feedback/002.md

# Add feedback artifacts
sow task input add --id 010 --type feedback --path "phases/implementation/tasks/010/feedback/001.md"
sow task input add --id 010 --type feedback --path "phases/implementation/tasks/010/feedback/002.md"
```

#### Review Feedback (Worker or Human)

```bash
sow task input list --id 010
# Inputs:
# [0] reference: sinks/style-guide.md
# [1] reference: knowledge/jwt-design.md
# [2] feedback: phases/implementation/tasks/010/feedback/001.md (status: pending)
# [3] feedback: phases/implementation/tasks/010/feedback/002.md (status: pending)
```

#### Address Feedback

Worker marks feedback as addressed by setting status in metadata:

```bash
sow task input set --id 010 --index 2 status addressed
sow task input set --id 010 --index 3 status addressed
```

**Note:** Fields like `status` are automatically stored in the artifact's `metadata` field.

**State:**
```yaml
tasks:
  - id: "010"
    inputs:
      - type: "reference"
        path: "sinks/style-guide.md"
        created_at: "2025-01-15T11:00:00Z"
      - type: "reference"
        path: "knowledge/jwt-design.md"
        created_at: "2025-01-15T11:00:00Z"
      - type: "feedback"
        path: "phases/implementation/tasks/010/feedback/001.md"
        created_at: "2025-01-15T12:00:00Z"
        metadata:
          status: "addressed"
      - type: "feedback"
        path: "phases/implementation/tasks/010/feedback/002.md"
        created_at: "2025-01-15T12:30:00Z"
        metadata:
          status: "addressed"
```

#### Complete Task

```bash
sow task set --id 010 status completed
```

#### Complete Remaining Tasks

```bash
sow task set --id 020 status in_progress
# ... (similar workflow)
sow task set --id 020 status completed

sow task set --id 030 status in_progress
# ... (similar workflow)
sow task set --id 030 status completed
```

#### Advance to Review

```bash
sow advance
# → ImplementationPhase.Advance() in ImplementationExecuting state
# → Returns EventAllTasksComplete
# → State machine checks AllTasksComplete guard
# → Transitions to ReviewActive
```

---

### 5. Review Phase

#### Create Review Report

```bash
sow output add --type review --path "review/report.md" --assessment pass
```

**Note:** The `--assessment` flag automatically sets `metadata.assessment` in the artifact.

**State:**
```yaml
phases:
  review:
    status: "in_progress"
    outputs:
      - type: "review"
        path: "review/report.md"
        created_at: "2025-01-15T14:00:00Z"
        approved: false
        metadata:
          assessment: "pass"
```

#### Approve Review

```bash
sow output list
# [0] review: review/report.md (assessment: pass, not approved)

sow output set --index 0 approved true
```

#### Advance to Finalize

```bash
sow advance
# → ReviewPhase.Advance() in ReviewActive state
# → Checks LatestReviewApproved guard
# → Reads metadata.assessment field ("pass")
# → Returns EventReviewPass
# → Transitions to FinalizeDocumentation
```

If review failed:
```bash
# Would have set assessment to "fail"
sow output set --index 0 assessment fail
sow output set --index 0 approved true

sow advance
# → Would return EventReviewFail
# → Transitions back to ImplementationPlanning (loop for rework)
```

**Note:** Setting `assessment` via the command automatically stores it in `metadata.assessment`.

---

### 6. Finalize Phase

#### Documentation Substate

```bash
# (Orchestrator updates documentation)

sow advance
# → FinalizePhase.Advance() in FinalizeDocumentation state
# → Returns EventDocumentationDone
# → Transitions to FinalizeChecks
```

#### Checks Substate

```bash
# (Orchestrator runs tests, linters)

sow advance
# → FinalizePhase.Advance() in FinalizeChecks state
# → Returns EventChecksDone
# → Transitions to FinalizeDelete
```

#### Delete Substate

```bash
# (Orchestrator cleans up project folder)
sow phase set project_deleted true

sow advance
# → FinalizePhase.Advance() in FinalizeDelete state
# → Returns EventProjectDelete
# → State machine checks ProjectDeleted guard
# → Transitions to NoProject (project complete!)
```

---

## Key Design Principles

### 1. Single Command for State Progression

**Only `sow advance` fires state machine events.**

All other commands are pure data operations:
- `set` modifies fields
- `input/output add/set/remove` manages artifacts
- None trigger state transitions

**Benefit:** Predictable, explicit progression.

### 2. Inputs/Outputs Express Data Flow

**Semantic meaning:**
- Inputs = consumed by container
- Outputs = produced by container

**Examples:**
- Planning phase outputs task list
- Task inputs include references and feedback
- Task outputs include modified files
- Review phase outputs review report

**Benefit:** Clear directionality, self-documenting.

### 3. Index-Based Artifact References

Artifacts referenced by position in array:

```bash
sow output list
# [0] task_list: planning/tasks.md
# [1] design_doc: planning/design.md

sow output set --index 0 approved true
```

**Benefit:** Simple, unambiguous, visible in state files.

### 4. Set for All Scalar Modifications

All field updates use `set`:
```bash
sow task set --id 010 status completed
sow task set --id 010 iteration 2
sow phase set tasks_approved true
sow output set --index 0 approved true
```

No special commands for specific fields (no `approve`, `status`, etc.).

**Benefit:** Uniform, discoverable, consistent.

### 5. Project Types Define Semantics

The CLI provides universal commands. Project types define:
- Valid artifact types per container
- State machine transitions and guards
- Workflow prompts

**Standard project artifact types:**

**Phase inputs:**
- `context` - contextual documents

**Phase outputs:**
- `task_list` - planning output
- `review` - review phase output (metadata: `assessment` = "pass" | "fail")

**Task inputs:**
- `reference` - reference documents
- `feedback` - human feedback items (metadata: `status` = "pending" | "addressed")

**Task outputs:**
- `modified` - files changed by task

**Note:** All artifacts use the universal schema (type, path, approved, created_at, metadata). Type-specific fields are stored in metadata and can be set naturally via commands (e.g., `--assessment`, `--status`).

**Benefit:** Extensible to new project types without command changes.

## Benefits Summary

1. **Complete Uniformity** - Same patterns across all containers
2. **Predictable State Changes** - Only `advance` transitions state
3. **Semantic Clarity** - Inputs vs outputs express data flow
4. **Minimal Command Surface** - Three operations cover everything (set, add/remove, advance)
5. **Highly Discoverable** - Structure mirrors data model
6. **Extremely Extensible** - New artifact types work with existing commands
7. **Simple References** - Index-based, no ID invention needed
8. **Universal Artifact Schema** - Single schema for all artifacts with metadata extensibility for type-specific fields

## Command Summary Table

| Command | Purpose | Flags |
|---------|---------|-------|
| `sow project set <field> <value>` | Set project field | - |
| `sow phase set <field> <value>` | Set phase field | `--phase` (optional) |
| `sow task set --id <id> <field> <value>` | Set task field | `--id` (required) |
| `sow input add --type <type>` | Add phase input | `--phase` (optional), `--type` (required) |
| `sow output add --type <type>` | Add phase output | `--phase` (optional), `--type` (required) |
| `sow task input add --id <id> --type <type>` | Add task input | `--id` (required), `--type` (required) |
| `sow task output add --id <id> --type <type>` | Add task output | `--id` (required), `--type` (required) |
| `sow input set --index <n> <field> <value>` | Set phase input field | `--index` (required), `--phase` (optional) |
| `sow output set --index <n> <field> <value>` | Set phase output field | `--index` (required), `--phase` (optional) |
| `sow task input set --id <id> --index <n> <field> <value>` | Set task input field | `--id` (required), `--index` (required) |
| `sow task output set --id <id> --index <n> <field> <value>` | Set task output field | `--id` (required), `--index` (required) |
| `sow advance` | Progress state | - |

**Three operations. Infinite flexibility.**
