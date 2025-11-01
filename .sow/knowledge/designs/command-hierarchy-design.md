# Unified Command Hierarchy Design

**Author**: Design Orchestrator
**Date**: 2025-11-01
**Status**: Draft
**Type**: Comprehensive Design Doc

## Executive Summary

Restructure the sow CLI around a unified container model where projects, phases, and tasks share identical patterns for data management. Replace specialized commands with three universal operations: `set` (modify fields), `input/output add/set/remove` (manage artifacts), and `advance` (progress state). This simplification reduces command surface area by ~60%, makes the CLI highly discoverable through consistent patterns, and enables future project types without CLI changes.

## Overview

Implement a unified command hierarchy based on containers with inputs/outputs. The current CLI has specialized commands per container type (projects have different commands than tasks), leading to inconsistent UX and high maintenance burden. This design introduces a universal container model where all entities (project, phase, task) share the same structure and operations.

Current state: The CLI has approximately 40 commands with inconsistent patterns (`sow project add-artifact`, `sow task add-reference`, `sow phase approve-tasks`). Each container type has unique commands, making the CLI difficult to learn and extend. This design consolidates to ~12 universal commands that work across all container types.

## Goals and Non-Goals

**Goals**:
- G1: Unified container model - all containers (project, phase, task) share identical structure
- G2: Universal operations - same commands work on any container type
- G3: Data flow clarity - inputs (consumed) and outputs (produced) explicit in structure
- G4: Minimal command surface - three core operations cover all use cases
- G5: Type extensibility - new project types add artifact types without CLI changes
- G6: Discoverable interface - consistent patterns make commands predictable

**Success Metrics**:
- Command count reduced from ~40 to ~12
- Zero CLI changes needed when adding new project types
- Learning curve reduced (measure via user testing)

**Non-Goals** (explicitly out of scope):
- State machine transition logic (covered in advance command design)
- Implementation plan/phasing (implementation detail)
- Backward compatibility (breaking change accepted)
- Command aliases or shortcuts (focus on clarity)

## Background

### Current State

The CLI organizes commands by container type with specialized operations:

**Project commands** (~15 commands):
- `sow project new <name>`
- `sow project continue`
- `sow project set <field> <value>`
- `sow project add-artifact <path>`
- `sow project approve-artifact <path>`
- `sow project create-pr`
- etc.

**Task commands** (~10 commands):
- `sow task add <name>`
- `sow task set-status <id> <status>`
- `sow task add-reference <id> <path>`
- `sow task add-feedback <id> <path>`
- `sow task list`
- etc.

**Phase commands** (~8 commands):
- `sow phase set <field> <value>`
- `sow phase approve-tasks`
- `sow agent complete`
- etc.

**Problems**:
1. **Inconsistent patterns**: `project add-artifact` vs `task add-reference` (same concept, different names)
2. **Type-specific commands**: Adding new project types requires new CLI commands
3. **High maintenance burden**: 40+ commands to maintain and test
4. **Difficult to learn**: No consistent patterns, must memorize each command
5. **Poor discoverability**: Can't infer command names from structure
6. **Scattered logic**: Similar operations implemented differently across types

### Requirements

**Functional Requirements**:
- FR1: Support all current functionality through universal commands
- FR2: Enable project types to define valid artifact types per container
- FR3: Maintain clear distinction between inputs (consumed) and outputs (produced)
- FR4: Support type-specific artifact metadata without schema changes
- FR5: Preserve index-based artifact references for simplicity

**Non-Functional Requirements**:
- NFR1: Command execution time unchanged (same file I/O operations)
- NFR2: State file format changes must be backward compatible with migration
- NFR3: Clear error messages when invalid operations attempted
- NFR4: Schema validation catches invalid artifact types per project type

## Design

### Conceptual Model

#### Three Container Types

All entities in sow are containers with identical structure:

1. **Project**: Top-level container with metadata (branch, description)
2. **Phase**: Workflow stage within project (planning, implementation, review, finalize, etc.)
3. **Task**: Unit of work within any phase (not limited to implementation)

#### Universal Container Structure

Every container has four components:

```
Container:
  - Singular fields (scalars)
  - Inputs (artifacts consumed)
  - Outputs (artifacts produced)
  - Sub-containers (nested containers)
```

**Example - Phase Container**:
- **Singular fields**: `status`, `enabled`, `created_at`, `started_at`, `completed_at`
- **Inputs**: Context documents, design docs, previous phase outputs
- **Outputs**: Task lists, review reports, other artifacts
- **Sub-containers**: Tasks (for implementation phase only)

**Example - Task Container**:
- **Singular fields**: `id`, `name`, `phase`, `status`, `iteration`, `assigned_agent`
- **Inputs**: References, feedback, context
- **Outputs**: Modified files, generated artifacts
- **Sub-containers**: None (tasks are leaf nodes)

#### Data Flow Model

**Inputs** represent artifacts consumed by the container:
- Planning phase inputs: Context from exploration
- Task inputs: Style guide references, design docs, human feedback

**Outputs** represent artifacts produced by the container:
- Planning phase outputs: Task list
- Task outputs: Modified source files
- Review phase outputs: Review report with assessment

This explicit directionality makes data flow self-documenting.

### Universal Operations

Three operations work on all containers:

#### 1. Set (Modify Scalar Fields)

**Pattern**: `sow <container> set <field> <value> [--selector]`

**Examples**:
```bash
# Project level
sow project set description "Add JWT authentication"

# Phase level
sow phase set tasks_approved true --phase implementation

# Task level
sow task set --id 010 status completed
```

**Behavior**: Updates singular fields, saves state, no events fired.

#### 2. Input/Output Management (Manage Artifacts)

**Pattern**:
```bash
sow <container> input add --type <type> [--selector] [fields...]
sow <container> input set --index <n> <field> <value> [--selector]
sow <container> input remove --index <n> [--selector]
sow <container> input list [--selector]

# Same for output
sow <container> output add --type <type> [--selector] [fields...]
```

**Examples**:
```bash
# Phase inputs/outputs
sow input add --type context --path "discovery/research.md" --phase planning
sow output add --type task_list --path "planning/tasks.md" --phase planning
sow output set --index 0 approved true --phase planning

# Task inputs/outputs
sow task input add --id 010 --type reference --path "sinks/style-guide.md"
sow task input add --id 010 --type feedback --path "feedback/001.md"
sow task output add --id 010 --type modified --path "src/auth/jwt.ts"
```

**Behavior**: Manages artifact lists, type-specific fields stored in metadata.

#### 3. Advance (State Progression)

**Pattern**: `sow advance`

**Example**:
```bash
sow advance  # Examines state, fires event, transitions
```

**Behavior**: Covered in advance command design doc. Single command for all state transitions.

### Artifact System

#### Universal Artifact Schema

All inputs and outputs use identical schema:

```yaml
artifact:
  type: string              # "context", "task_list", "review", "reference", etc.
  path: string              # Path relative to .sow/project/
  approved: boolean?        # Optional approval flag
  created_at: timestamp
  metadata: map?            # Extensible metadata for type-specific fields
```

**Key Principle**: Known fields (type, path, approved, created_at) are direct fields. Unknown fields automatically route to metadata.

#### Field Path Traversal

The `set` command accepts a field path as its first argument, supporting dot notation for nested access:

```bash
# Known field - sets approved directly
sow output set --index 0 approved true

# Metadata field - use dot notation for path traversal
sow output set --index 0 metadata.assessment pass

# Nested metadata field
sow task input set --id 010 --index 2 metadata.status addressed
```

**Pattern**: `<field>` or `<field>.<subfield>.<subfield>`

**Benefit**: Explicit path makes it clear where values are stored in the schema. Dot notation enables arbitrary nesting in metadata.

#### Index-Based References

Artifacts referenced by position in array:

```bash
sow output list
# [0] task_list: planning/tasks.md (not approved)
# [1] design_doc: planning/design.md (approved)

sow output set --index 0 approved true
```

**Rationale**:
- Simple and unambiguous
- Visible in state files (humans can verify index)
- No ID invention needed
- Natural for ordered lists

## Schema Definitions

### Updated Phase Schema

```yaml
phase:
  # Singular fields (scalars)
  name: string              # planning, implementation, review, finalize
  status: string            # pending, in_progress, completed
  enabled: boolean
  created_at: timestamp
  started_at: timestamp?
  completed_at: timestamp?

  # Metadata (extensible for phase-specific flags)
  # Example: tasks_approved, project_deleted
  metadata: map?

  # Input artifacts (consumed by this phase)
  inputs:
    - type: string          # Artifact type
      path: string
      approved: boolean?
      created_at: timestamp
      metadata: map?

  # Output artifacts (produced by this phase)
  outputs:
    - type: string
      path: string
      approved: boolean?
      created_at: timestamp
      metadata: map?        # assessment, etc. go here

  # Sub-containers (any phase can have tasks)
  # This is simplified task metadata - full state in separate files
  tasks:
    - id: string            # Gap-numbered: 010, 020, 030
      name: string
      status: string
      parallel: boolean
      dependencies: [string]
      metadata: map?
```

**Changes from current**:
- Added `inputs` and `outputs` arrays
- Moved `artifacts` into appropriate input/output category
- Added `metadata` map for extensibility
- Tasks remain as simple list (detailed state in task files)

### Updated Task State Schema

Full task state lives in `.sow/project/phases/<phase>/tasks/<id>/state.yaml`:

```yaml
task:
  # Identity and assignment
  id: string                # Gap-numbered: 010, 020, 030
  name: string
  phase: string             # Which phase this task belongs to (e.g., "implementation", "planning", "review")
  status: string            # pending, in_progress, completed, abandoned
  iteration: integer        # 1, 2, 3, ...
  assigned_agent: string    # implementer, reviewer, etc.

  # Timestamps
  created_at: timestamp
  started_at: timestamp?
  updated_at: timestamp
  completed_at: timestamp?

  # Input artifacts (consumed by this task)
  inputs:
    - type: string          # "reference", "feedback"
      path: string
      created_at: timestamp
      metadata: map?        # status: "pending" | "addressed" for feedback

  # Output artifacts (produced by this task)
  outputs:
    - type: string          # "modified"
      path: string
      created_at: timestamp
      metadata: map?
```

**Changes from current**:
- Replaced `references: [string]` with `inputs: [artifact]`
- Replaced `files_modified: [string]` with `outputs: [artifact]`
- Replaced `feedback: [Feedback]` with inputs of type "feedback"
- Uniform artifact structure throughout

### Artifact Schema

```yaml
artifact:
  type: string              # Required. Defined by project type.
  path: string              # Required. Relative to .sow/project/
  approved: boolean?        # Optional. For approval workflows.
  created_at: timestamp     # Required. Auto-set on creation.
  metadata: map?            # Optional. Type-specific fields.
    # Examples:
    # - assessment: "pass" | "fail"  (for review artifacts)
    # - status: "pending" | "addressed"  (for feedback artifacts)
    # - source: "human" | "generated"
```

## Command Structure

### Project Commands

```bash
sow project new --branch <branch> "<initial-prompt>"
sow project continue [--branch <branch>]
sow project set <field> <value>
sow project delete
```

**Note**:
- `--branch` identifies which project (maps to git branch)
- `continue` defaults to current git branch if `--branch` omitted
- Project-level inputs/outputs not used in standard project (may be used by future project types)

### Phase Commands

```bash
sow phase set <field-path> <value> [--phase <name>]
```

**Note**: `--phase` optional, defaults to currently active phase if omitted.

### Task Commands

```bash
sow task add <name> [--agent <agent>] [--description "..."]
sow task set --id <task_id> <field-path> <value>
sow task abandon --id <task_id>
sow task list
```

### Input Commands (Phase-Level)

```bash
sow input add --type <type> [--phase <name>] [--field value...]
sow input set --index <n> <field-path> <value> [--phase <name>]
sow input remove --index <n> [--phase <name>]
sow input list [--phase <name>]
```

**Notes**:
- `--phase` optional, defaults to currently active phase
- Inline field setting via flags (e.g., `--path "..."`, `--approved true`)
- Field paths support dot notation (e.g., `metadata.status`)

### Output Commands (Phase-Level)

```bash
sow output add --type <type> [--phase <name>] [--field value...]
sow output set --index <n> <field-path> <value> [--phase <name>]
sow output remove --index <n> [--phase <name>]
sow output list [--phase <name>]
```

**Notes**:
- `--phase` optional, defaults to currently active phase
- Inline field setting via flags
- Field paths support dot notation

### Task Input/Output Commands

```bash
sow task input add --id <task_id> --type <type> [--field value...]
sow task input set --id <task_id> --index <n> <field-path> <value>
sow task input remove --id <task_id> --index <n>
sow task input list --id <task_id>

sow task output add --id <task_id> --type <type> [--field value...]
sow task output set --id <task_id> --index <n> <field-path> <value>
sow task output remove --id <task_id> --index <n>
sow task output list --id <task_id>
```

**Notes**:
- Field paths support dot notation for nested metadata access

### Advance Command

```bash
sow advance
```

Single command for all state transitions. Examines state, determines event, fires it. See advance command design doc for details.

## Command Reference

Complete command listing with descriptions, arguments, and flags.

### Project Commands

#### `sow project new`

Create a new project on a specified branch.

**Syntax**: `sow project new --branch <branch-name> [--issue <issue_id>] "<initial-prompt>"`

**Arguments**:
- `<initial-prompt>`: Description of what the project will accomplish (optional)

**Flags**:
- `--branch <name>`: Git branch for the project (required)
- `--issue <issue_id>`: GitHub issue ID to initialize the project from

**Example**:
```bash
sow project new --branch feat/add-auth "Add JWT authentication to API endpoints"
```

#### `sow project continue`

Resume an existing project.

**Syntax**: `sow project continue --branch <branch-name>`

**Flags**:
- `--branch <name>`: Git branch for the project

**Example**:
```bash
sow project continue --branch feat/add-authh
```

#### `sow project set`

Modify project-level scalar fields.

**Syntax**: `sow project set <field-path> <value>`

**Arguments**:
- `<field-path>`: Field to modify (supports dot notation)
- `<value>`: New value for the field

**Example**:
```bash
sow project set description "Updated project description"
```

#### `sow project delete`

Delete the current project.

**Syntax**: `sow project delete`

**Example**:
```bash
sow project delete
```

### Phase Commands

#### `sow phase set`

Modify phase-level scalar fields or metadata.

**Syntax**: `sow phase set <field-path> <value> [--phase <name>]`

**Arguments**:
- `<field-path>`: Field to modify (supports dot notation for metadata)
- `<value>`: New value for the field

**Flags**:
- `--phase <name>`: Target phase (optional, defaults to currently active phase)

**Examples**:
```bash
sow phase set status in_progress
sow phase set metadata.tasks_approved true --phase implementation
sow phase set metadata.custom_flag value
```

### Task Commands

#### `sow task add`

Create a new task in the current phase.

**Syntax**: `sow task add <name> [--agent <agent-type>] [--description "<text>"]`

**Arguments**:
- `<name>`: Task name

**Flags**:
- `--agent <type>`: Agent type to assign (optional)
- `--description "<text>"`: Task description (optional)

**Example**:
```bash
sow task add "Implement JWT signing" --agent implementer --description "Create RS256 signing"
```

#### `sow task set`

Modify task-level scalar fields or metadata.

**Syntax**: `sow task set --id <task-id> <field-path> <value>`

**Arguments**:
- `<field-path>`: Field to modify (supports dot notation)
- `<value>`: New value for the field

**Flags**:
- `--id <id>`: Task ID (required)

**Examples**:
```bash
sow task set --id 010 status completed
sow task set --id 010 iteration 2
sow task set --id 010 metadata.custom_field value
```

#### `sow task abandon`

Mark a task as abandoned.

**Syntax**: `sow task abandon --id <task-id>`

**Flags**:
- `--id <id>`: Task ID (required)

**Example**:
```bash
sow task abandon --id 010
```

#### `sow task list`

List all tasks in the current phase.

**Syntax**: `sow task list`

**Example**:
```bash
sow task list
```

### Input Commands (Phase-Level)

#### `sow input add`

Add an input artifact to a phase.

**Syntax**: `sow input add --type <artifact-type> [--phase <name>] [--<field> <value>...]`

**Flags**:
- `--type <type>`: Artifact type (required)
- `--phase <name>`: Target phase (optional, defaults to active phase)
- `--<field> <value>`: Inline field setting (e.g., `--path`, `--approved`)

**Examples**:
```bash
sow input add --type context --path "discovery/research.md"
sow input add --type context --path "sinks/guide.md" --phase planning
sow input add --type context --path "doc.md" --approved true
```

#### `sow input set`

Modify a field on an existing input artifact.

**Syntax**: `sow input set --index <n> <field-path> <value> [--phase <name>]`

**Arguments**:
- `<field-path>`: Field to modify (supports dot notation)
- `<value>`: New value

**Flags**:
- `--index <n>`: Artifact index in inputs array (required)
- `--phase <name>`: Target phase (optional, defaults to active phase)

**Examples**:
```bash
sow input set --index 0 approved true
sow input set --index 1 metadata.source human --phase planning
```

#### `sow input remove`

Remove an input artifact from a phase.

**Syntax**: `sow input remove --index <n> [--phase <name>]`

**Flags**:
- `--index <n>`: Artifact index to remove (required)
- `--phase <name>`: Target phase (optional, defaults to active phase)

**Example**:
```bash
sow input remove --index 2
```

#### `sow input list`

List all input artifacts for a phase.

**Syntax**: `sow input list [--phase <name>]`

**Flags**:
- `--phase <name>`: Target phase (optional, defaults to active phase)

**Example**:
```bash
sow input list
sow input list --phase planning
```

### Output Commands (Phase-Level)

#### `sow output add`

Add an output artifact to a phase.

**Syntax**: `sow output add --type <artifact-type> [--phase <name>] [--<field> <value>...]`

**Flags**:
- `--type <type>`: Artifact type (required)
- `--phase <name>`: Target phase (optional, defaults to active phase)
- `--<field> <value>`: Inline field setting

**Examples**:
```bash
sow output add --type task_list --path "planning/tasks.md"
sow output add --type review --path "review/report.md" --metadata.assessment pass
sow output add --type task_list --path "tasks.md" --approved false --phase planning
```

#### `sow output set`

Modify a field on an existing output artifact.

**Syntax**: `sow output set --index <n> <field-path> <value> [--phase <name>]`

**Arguments**:
- `<field-path>`: Field to modify (supports dot notation)
- `<value>`: New value

**Flags**:
- `--index <n>`: Artifact index in outputs array (required)
- `--phase <name>`: Target phase (optional, defaults to active phase)

**Examples**:
```bash
sow output set --index 0 approved true
sow output set --index 0 metadata.assessment pass
sow output set --index 1 metadata.reviewed_by orchestrator
```

#### `sow output remove`

Remove an output artifact from a phase.

**Syntax**: `sow output remove --index <n> [--phase <name>]`

**Flags**:
- `--index <n>`: Artifact index to remove (required)
- `--phase <name>`: Target phase (optional, defaults to active phase)

**Example**:
```bash
sow output remove --index 1
```

#### `sow output list`

List all output artifacts for a phase.

**Syntax**: `sow output list [--phase <name>]`

**Flags**:
- `--phase <name>`: Target phase (optional, defaults to active phase)

**Example**:
```bash
sow output list
sow output list --phase review
```

### Task Input Commands

#### `sow task input add`

Add an input artifact to a task.

**Syntax**: `sow task input add --id <task-id> --type <artifact-type> [--<field> <value>...]`

**Flags**:
- `--id <id>`: Task ID (required)
- `--type <type>`: Artifact type (required)
- `--<field> <value>`: Inline field setting

**Examples**:
```bash
sow task input add --id 010 --type reference --path "sinks/style-guide.md"
sow task input add --id 010 --type feedback --path "feedback/001.md"
sow task input add --id 020 --type reference --path "docs/arch.md" --approved true
```

#### `sow task input set`

Modify a field on an existing task input artifact.

**Syntax**: `sow task input set --id <task-id> --index <n> <field-path> <value>`

**Arguments**:
- `<field-path>`: Field to modify (supports dot notation)
- `<value>`: New value

**Flags**:
- `--id <id>`: Task ID (required)
- `--index <n>`: Artifact index in task inputs array (required)

**Examples**:
```bash
sow task input set --id 010 --index 0 approved true
sow task input set --id 010 --index 2 metadata.status addressed
```

#### `sow task input remove`

Remove an input artifact from a task.

**Syntax**: `sow task input remove --id <task-id> --index <n>`

**Flags**:
- `--id <id>`: Task ID (required)
- `--index <n>`: Artifact index to remove (required)

**Example**:
```bash
sow task input remove --id 010 --index 1
```

#### `sow task input list`

List all input artifacts for a task.

**Syntax**: `sow task input list --id <task-id>`

**Flags**:
- `--id <id>`: Task ID (required)

**Example**:
```bash
sow task input list --id 010
```

### Task Output Commands

#### `sow task output add`

Add an output artifact to a task.

**Syntax**: `sow task output add --id <task-id> --type <artifact-type> [--<field> <value>...]`

**Flags**:
- `--id <id>`: Task ID (required)
- `--type <type>`: Artifact type (required)
- `--<field> <value>`: Inline field setting

**Examples**:
```bash
sow task output add --id 010 --type modified --path "src/auth/jwt.ts"
sow task output add --id 010 --type modified --path "tests/auth.test.ts"
```

#### `sow task output set`

Modify a field on an existing task output artifact.

**Syntax**: `sow task output set --id <task-id> --index <n> <field-path> <value>`

**Arguments**:
- `<field-path>`: Field to modify (supports dot notation)
- `<value>`: New value

**Flags**:
- `--id <id>`: Task ID (required)
- `--index <n>`: Artifact index in task outputs array (required)

**Example**:
```bash
sow task output set --id 010 --index 0 approved true
sow task output set --id 010 --index 0 metadata.reviewed true
```

#### `sow task output remove`

Remove an output artifact from a task.

**Syntax**: `sow task output remove --id <task-id> --index <n>`

**Flags**:
- `--id <id>`: Task ID (required)
- `--index <n>`: Artifact index to remove (required)

**Example**:
```bash
sow task output remove --id 010 --index 1
```

#### `sow task output list`

List all output artifacts for a task.

**Syntax**: `sow task output list --id <task-id>`

**Flags**:
- `--id <id>`: Task ID (required)

**Example**:
```bash
sow task output list --id 010
```

### State Progression Command

#### `sow advance`

Progress the project to the next state.

**Syntax**: `sow advance`

**Description**: Examines current state, determines appropriate event, fires state machine transition. This is the only command that triggers state transitions.

**Example**:
```bash
sow advance
```

See [State Progression via Advance Command Design](./advance-command-design.md) for complete details.

## Artifact Types for Standard Project

Project types define valid artifact types per container. For standard project:

### Phase Input Types

**Planning Phase Inputs**:
- `context`: Contextual documents (exploration summaries, design docs, requirements)

**Implementation Phase Inputs**:
- None (implementation phase doesn't consume phase-level inputs)

**Review Phase Inputs**:
- None (review examines implementation outputs)

**Finalize Phase Inputs**:
- None (finalize processes completed work)

### Phase Output Types

**Planning Phase Outputs**:
- `task_list`: Breakdown of implementation work (required, must be approved)

**Implementation Phase Outputs**:
- None (implementation outputs are at task level)

**Review Phase Outputs**:
- `review`: Review report with assessment
  - **Metadata fields**: `assessment` ("pass" | "fail")

**Finalize Phase Outputs**:
- None (finalize produces project completion, not artifacts)

### Task Input Types

All implementation tasks support:

- `reference`: Reference documents (style guides, conventions, code examples)
  - **Examples**: `.sow/sinks/style-guide.md`, `.sow/knowledge/architecture.md`

- `feedback`: Human feedback items
  - **Metadata fields**: `status` ("pending" | "addressed")
  - **Path pattern**: `phases/implementation/tasks/{id}/feedback/{n}.md`

### Task Output Types

All implementation tasks produce:

- `modified`: Files modified by the task
  - **Examples**: `src/auth/jwt.ts`, `tests/auth/jwt.test.ts`

**Note**: Task outputs are typically auto-tracked by worker agents logging their changes.

## Command Mapping Table

### Project Commands

| Current Command | New Command | Notes |
|----------------|-------------|-------|
| `sow project` | `sow project new --branch <branch> "<prompt>"` | Expanded with subcommands |
| N/A (new) | `sow project continue [--branch <branch>]` | New subcommand |
| N/A (new) | `sow project set <field-path> <value>` | New subcommand |
| `sow project add-artifact <path>` | `sow output add --type <type> --path <path>` | Must specify type |
| `sow project approve-artifact <path>` | `sow output set --index <n> approved true` | Use index instead of path |
| `sow project list-artifacts` | `sow output list` | Simpler name |
| `sow project delete` | `sow project delete` | Unchanged |

### Phase Commands

| Current Command | New Command | Notes |
|----------------|-------------|-------|
| `sow phase set <field> <value>` | `sow phase set <field-path> <value> [--phase <name>]` | Added optional --phase flag |
| `sow phase add-artifact <path>` | `sow output add --type <type> --path <path> --phase <name>` | Must specify type and phase |
| `sow phase approve-artifact <path>` | `sow output set --index <n> approved true --phase <name>` | Use index |
| `sow phase approve-tasks` | `sow phase set metadata.tasks_approved true` | Explicit field setting via metadata path |
| `sow agent complete` | `sow advance` | Unified state progression |

### Task Commands

| Current Command | New Command | Notes |
|----------------|-------------|-------|
| `sow task add <name>` | `sow task add <name>` | Unchanged |
| `sow task set-status <id> <status>` | `sow task set --id <id> status <status>` | Generic set with field path |
| `sow task set-iteration <id> <n>` | `sow task set --id <id> iteration <n>` | Generic set with field path |
| `sow task add-reference <id> <path>` | `sow task input add --id <id> --type reference --path <path>` | Explicit input/type |
| `sow task add-feedback <id> <path>` | `sow task input add --id <id> --type feedback --path <path>` | Explicit input/type |
| `sow task mark-feedback-addressed <id> <feedback_id>` | `sow task input set --id <id> --index <n> metadata.status addressed` | Use index + metadata path |
| `sow task list-references <id>` | `sow task input list --id <id>` | Generic input list |
| `sow task list` | `sow task list` | Unchanged |
| `sow task abandon <id>` | `sow task abandon <id>` | Unchanged |

### Removed Commands

| Current Command | Replacement | Reason |
|----------------|-------------|--------|
| `sow agent complete` | `sow advance` | Consolidated into advance |
| `sow project create-pr` | Manual workflow | Out of scope for simplified CLI |

## Data Flow Examples

### Example 1: Planning Phase

**Setup Phase**:
```bash
# Create project
sow project new --branch feat/add-auth "Add JWT authentication to API"

# State: PlanningActive
```

**Add Context Inputs**:
```bash
# Orchestrator adds exploration findings as inputs
# Note: --phase optional (defaults to active phase, which is planning)
sow input add --type context --path "discovery/jwt-research.md"
sow input add --type context --path "sinks/auth-patterns.md"
```

**State after inputs**:
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
    outputs: []
```

**Create Task List Output**:
```bash
# Orchestrator creates task list
sow output add --type task_list --path "planning/tasks.md"
```

**Approve and Advance**:
```bash
# List outputs to see index
sow output list
# [0] task_list: planning/tasks.md (not approved)

# Orchestrator asks human for approval, then runs:
sow output set --index 0 approved true

# Orchestrator advances to next phase
sow advance
# → Fires EventCompletePlanning
# → Transitions to ImplementationPlanning
```

### Example 2: Task Execution with Feedback

**Create Task**:
```bash
sow task add "Implement JWT signing" --agent implementer
# Creates task 010
```

**Add Task Inputs**:
```bash
sow task input add --id 010 --type reference --path "sinks/style-guide.md"
sow task input add --id 010 --type reference --path "knowledge/jwt-design.md"
```

**Worker Executes, Adds Outputs**:
```bash
# Worker creates implementation files
sow task output add --id 010 --type modified --path "src/auth/jwt.ts"
sow task output add --id 010 --type modified --path "src/auth/verify.ts"
```

**Orchestrator Provides Feedback on Implementation**:
```bash
# Orchestrator creates feedback file and adds as input
sow task input add --id 010 --type feedback --path "phases/implementation/tasks/010/feedback/001.md"
```

**Worker Addresses Feedback**:
```bash
# List inputs to find feedback index
sow task input list --id 010
# [0] reference: sinks/style-guide.md
# [1] reference: knowledge/jwt-design.md
# [2] feedback: phases/implementation/tasks/010/feedback/001.md

# Worker marks feedback as addressed using metadata path
sow task input set --id 010 --index 2 metadata.status addressed
```

**State after feedback addressed**:
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
```

### Example 3: Review Phase

**Add Review Report**:
```bash
# Reviewer creates review report with assessment metadata
sow output add --type review --path "review/report.md" --metadata.assessment pass
```

**Result in state file**:
```yaml
phases:
  review:
    outputs:
      - type: "review"
        path: "review/report.md"
        created_at: "2025-01-15T14:00:00Z"
        metadata:
          assessment: "pass"
```

**Approve Review**:
```bash
# Orchestrator asks human for approval, then runs:
sow output set --index 0 approved true
```

**Advance**:
```bash
sow advance
# → ReviewPhase.Advance() reads metadata.assessment
# → Fires EventReviewPass
# → Transitions to FinalizeDocumentation
```

## Error Handling

**Invalid Artifact Type**:
```bash
sow output add --type invalid_type --path "test.md" --phase planning
# Error: invalid artifact type 'invalid_type' for planning phase
# Valid types: context (input), task_list (output)
```

**Index Out of Range**:
```bash
sow output set --index 5 approved true
# Error: index 5 out of range (phase has 2 outputs)
```

**Missing Required Fields**:
```bash
sow output add --type task_list
# Error: --path is required for artifact creation
```

**Type-Specific Metadata Validation**:
```bash
sow output set --index 0 metadata.assessment invalid
# Error: metadata.assessment must be 'pass' or 'fail' for review artifacts
```

**Invalid Field Path**:
```bash
# Attempting to set non-existent nested field
sow output set --index 0 foo.bar.baz value
# Error: invalid field path 'foo.bar.baz' - 'foo' is not a valid artifact field
```

## Security/Privacy Considerations

**Artifact Paths**:
- All paths relative to `.sow/project/` (prevents path traversal)
- Validation rejects absolute paths or `..` segments

**Metadata Injection**:
- Metadata values sanitized before YAML serialization
- No code execution risk (YAML parsing safe)

**File Access**:
- CLI respects file permissions (no privilege escalation)
- References to external files (sinks, repos) validated before access

## Performance Considerations

**Command Execution**:
- Same number of file I/O operations as current system
- `set` operations: Single read + modify + write (identical to current)
- Input/output operations: Append to array + save (identical to current artifact operations)

**Index Lookups**:
- Array iteration for index access (O(n), acceptable for small arrays)
- Typical artifact counts: 1-10 per phase, 1-5 per task

**State File Size**:
- Inputs/outputs increase state file size ~10-20% (additional metadata)
- Still well within acceptable limits (<100KB typical project state)

## Testing Strategy

**Unit Tests**:

1. **Artifact Management**:
   - Add/remove artifacts from inputs/outputs
   - Field routing (known fields vs metadata)
   - Index-based updates
   - Approval workflows

2. **Command Parsing**:
   - Field specification in commands
   - Optional vs required parameters
   - Type validation

3. **Schema Validation**:
   - Invalid artifact types rejected
   - Type-specific metadata validation
   - Phase/task container validation

**Integration Tests**:

1. **Full Lifecycle**:
   - Planning with context inputs and task list output
   - Implementation with task inputs/outputs
   - Review with assessment
   - Feedback loop (add feedback, address, iterate)

2. **Command Equivalence**:
   - Old commands → new commands produce same state
   - Migration smoke tests

**Validation Tests**:

1. **Project Type Semantics**:
   - Standard project artifact type enforcement
   - Invalid type rejection per phase
   - Custom project type artifact types (future)

## Alternatives Considered

### Option 1: Keep Specialized Commands, Add Universal Alongside

**Description**: Maintain current specialized commands (`project add-artifact`), add universal commands as alternative interface.

**Pros**:
- Backward compatibility
- Gradual migration path
- Users can choose preferred style

**Cons**:
- Two ways to do everything (confusing)
- Double maintenance burden
- No reduction in CLI complexity
- Eventually requires deprecation cycle anyway

**Why not chosen**: Fails to achieve goal G4 (minimal command surface). Complexity doubles instead of reducing.

### Option 2: Container Hierarchy (Nested Containers)

**Description**: Fully hierarchical model where commands operate on paths like `project.planning.inputs[0]`.

**Example**:
```bash
sow set project.planning.inputs[0].approved true
```

**Pros**:
- Extremely generic (one command for everything)
- Mirrors state file structure exactly
- Ultimate flexibility

**Cons**:
- Verbose and difficult to type
- Poor error messages (path parsing errors)
- No type safety (everything is a path string)
- Loses semantic meaning (inputs/outputs not explicit)

**Why not chosen**: Trades simplicity for generality. Path-based access sacrifices usability for minimal gain.

### Option 3: Artifact IDs Instead of Indices

**Description**: Assign UUIDs or sequential IDs to artifacts, reference by ID instead of index.

**Example**:
```bash
sow output add --type task_list --path "tasks.md"
# Created artifact: art-abc123

sow output set --artifact art-abc123 approved true
```

**Pros**:
- Stable references across list modifications
- More "database-like" (familiar to developers)
- No index shifting issues

**Cons**:
- ID invention adds complexity
- Users must track IDs (harder to remember than indices)
- State files less readable (IDs instead of obvious ordering)
- Overkill for small lists (typical 1-5 artifacts)

**Why not chosen**: Adds complexity without sufficient benefit. Index-based references are simpler and sufficient for small artifact lists.

## Open Questions

- [ ] Should `sow input/output list` support filtering by type (`--type reference`)?
- [ ] Should metadata fields have schema validation per artifact type?
- [ ] Should we allow bulk operations (e.g., approve all outputs)?
- [ ] CLI flag naming: `--phase` vs `--phase-name` for clarity?
- [ ] Should project-level inputs/outputs be supported (reserved for future project types)?

## References

- [Exploration: Simplified Command Hierarchy](../../knowledge/explorations/simplified_command_hierarchy.md)
- [Design: State Progression via Advance Command](./advance-command-design.md)
- [Arc42 Section 5: Building Block View](../../docs/architecture/05-building-blocks.md)
- [Arc42 Section 8: Cross-cutting Concepts](../../docs/architecture/08-crosscutting-concepts.md)

## Future Considerations

- **Query language**: `sow query "outputs where type=review and approved=true"` for complex filtering
- **Bulk operations**: `sow output approve-all` for batch approvals
- **Artifact templates**: Predefined artifact bundles per project type
- **Validation rules**: Per-project-type validation of artifact metadata
- **Custom artifact types**: User-defined types with custom validation
- **Artifact relationships**: Dependencies between artifacts (task outputs → review inputs)
