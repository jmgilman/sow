# Schema Proposal - Revised Project State

**Status**: Draft Proposal (Step 3)
**Date**: 2025-10-14
**Purpose**: Define revised schemas for the new 5-phase, human-in-loop project model

---

## Overview

This document proposes revised CUE schemas to support the new sow project model:

- **Fixed 5 phases** (discovery, design, implementation, review, finalize)
- **Human-in-loop** for discovery and design phases
- **AI-autonomous** for implementation, review, finalize phases
- **Phase enablement flags** (discovery/design can be skipped)
- **Artifact tracking** with approval states
- **Simplified structure** (task presence = approval, no `pending_approvals`)

---

## Key Changes from Current Schema

### Project State Schema

**Removed Fields**:
- `complexity` - No longer using complexity assessment for progressive planning
- `active_phase` - Can be inferred from phase status (first phase with `in_progress`)

**Changed Fields**:
- `phases[]` - Now fixed structure with 5 specific phase types, each with different fields

**New Fields**:
- `phases.discovery.enabled` - Whether discovery phase is enabled
- `phases.design.enabled` - Whether design phase is enabled
- `phases.discovery.artifacts[]` - Discovery outputs with approval tracking
- `phases.design.artifacts[]` - Design outputs with approval tracking
- `phases.implementation.tasks[]` - Implementation task breakdown
- `phases.review.reports[]` - Review iteration reports
- `phases.review.iteration` - Current review iteration counter

### Task State Schema

**Minimal Changes**:
- Task state structure remains largely unchanged
- Still used for implementation phase tasks
- Same fields for iteration, feedback, references

---

## Proposed Schema: Project State

### Complete CUE Schema

```cue
// Project State Schema (Revised)
// Location: .sow/project/state.yaml
//
// Central state file for the new 5-phase project model.
// Phases are fixed: discovery, design, implementation, review, finalize.

package sow

import "time"

// ProjectState defines the complete project structure
#ProjectState: {
	project: #Project
	phases:  #Phases
}

// Project metadata
#Project: {
	// Project identifier (kebab-case recommended)
	name: string & =~"^[a-z0-9]+(-[a-z0-9]+)*$"

	// Git branch this project belongs to
	branch: string & =~"^[a-zA-Z0-9/_-]+$"

	// Timestamps (ISO 8601 format)
	created_at: string & time.Format(time.RFC3339)
	updated_at: string & time.Format(time.RFC3339)

	// Human-readable description from initial conversation
	description: string
}

// All five phases (fixed structure)
#Phases: {
	discovery:      #DiscoveryPhase
	design:         #DesignPhase
	implementation: #ImplementationPhase
	review:         #ReviewPhase
	finalize:       #FinalizePhase
}

// Discovery Phase (optional)
#DiscoveryPhase: {
	// Whether this phase is enabled for the project
	enabled: bool

	// Phase status
	// - "skipped" if enabled=false
	// - "pending" if enabled=true and not started
	// - "in_progress" if currently active
	// - "completed" if finished
	status: "skipped" | "pending" | "in_progress" | "completed"

	// Timestamps
	created_at:   string & time.Format(time.RFC3339)
	started_at:   null | (string & time.Format(time.RFC3339))
	completed_at: null | (string & time.Format(time.RFC3339))

	// Discovery type categorization
	// Set by /phase:discovery when it categorizes the work
	discovery_type?: "bug" | "feature" | "docs" | "refactor" | "general"

	// Artifacts produced during discovery
	// Each artifact requires human approval before phase can complete
	artifacts: [...#Artifact]
}

// Design Phase (optional)
#DesignPhase: {
	// Whether this phase is enabled for the project
	enabled: bool

	// Phase status
	status: "skipped" | "pending" | "in_progress" | "completed"

	// Timestamps
	created_at:   string & time.Format(time.RFC3339)
	started_at:   null | (string & time.Format(time.RFC3339))
	completed_at: null | (string & time.Format(time.RFC3339))

	// Whether architect agent was used (vs orchestrator doing design directly)
	architect_used?: bool

	// Artifacts produced during design
	// Each artifact requires human approval before phase can complete
	artifacts: [...#Artifact]
}

// Implementation Phase (required - always enabled)
#ImplementationPhase: {
	// Always true (implementation cannot be skipped)
	enabled: true

	// Phase status (cannot be "skipped")
	status: "pending" | "in_progress" | "completed"

	// Timestamps
	created_at:   string & time.Format(time.RFC3339)
	started_at:   null | (string & time.Format(time.RFC3339))
	completed_at: null | (string & time.Format(time.RFC3339))

	// Whether planner agent was used for task breakdown
	planner_used?: bool

	// Task breakdown
	// Presence in this list = human approval to execute
	tasks: [...#TaskSummary]

	// Task approval tracking
	// When orchestrator wants to add tasks mid-implementation, they're added here first
	// Human approval moves them to tasks[] array
	pending_task_additions?: [...#TaskSummary]
}

// Review Phase (required - always enabled)
#ReviewPhase: {
	// Always true (review cannot be skipped)
	enabled: true

	// Phase status (cannot be "skipped")
	status: "pending" | "in_progress" | "completed"

	// Timestamps
	created_at:   string & time.Format(time.RFC3339)
	started_at:   null | (string & time.Format(time.RFC3339))
	completed_at: null | (string & time.Format(time.RFC3339))

	// Review reports (numbered iterations: 001, 002, 003...)
	// Each iteration through review creates a new report
	reports: [...#ReviewReport]

	// Current review iteration (1-indexed, tracks which review we're on)
	iteration: int & >=1
}

// Finalize Phase (required - always enabled)
#FinalizePhase: {
	// Always true (finalize cannot be skipped)
	enabled: true

	// Phase status (cannot be "skipped")
	status: "pending" | "in_progress" | "completed"

	// Timestamps
	created_at:   string & time.Format(time.RFC3339)
	started_at:   null | (string & time.Format(time.RFC3339))
	completed_at: null | (string & time.Format(time.RFC3339))

	// Documentation updates made during finalize
	documentation_updates?: [...string]

	// Design artifacts moved to repository (e.g., ADRs)
	artifacts_moved?: [...{
		from: string
		to:   string
	}]

	// Whether project folder has been deleted (critical gate)
	project_deleted: bool

	// Pull request URL (if created via gh CLI)
	pr_url?: string
}

// Artifact (used in discovery and design phases)
#Artifact: {
	// Relative path from .sow/project/ root
	path: string

	// Human approval status
	// - false = not yet approved
	// - true = approved by human
	approved: bool

	// When artifact was created
	created_at: string & time.Format(time.RFC3339)
}

// Review Report (used in review phase)
#ReviewReport: {
	// Report number (e.g., "001", "002")
	id: string & =~"^[0-9]{3}$"

	// Relative path from .sow/project/phases/review/
	path: string

	// When report was created
	created_at: string & time.Format(time.RFC3339)

	// Overall assessment from report
	// - "pass" = no critical issues, ready to finalize
	// - "fail" = issues found, need to loop back to implementation
	assessment: "pass" | "fail"
}

// Task Summary (used in implementation phase)
#TaskSummary: {
	// Gap-numbered ID (e.g., "010", "020", "030")
	id: string & =~"^[0-9]{3}$"

	// Task name/description
	name: string

	// Task status
	status: "pending" | "in_progress" | "completed" | "abandoned"

	// Can run in parallel with other parallel tasks
	parallel: bool

	// Dependencies (task IDs that must complete before this one)
	dependencies?: [...string]
}
```

### Validation Rules

**Cross-Field Constraints**:

1. **Phase Status Logic**:
   ```cue
   // If enabled=false, status must be "skipped"
   // If enabled=true, status cannot be "skipped"

   // Discovery phase
   #DiscoveryPhase: {
   	if enabled == false {
   		status: "skipped"
   	}
   	if enabled == true {
   		status: "pending" | "in_progress" | "completed"
   	}
   }

   // Design phase (same logic)
   #DesignPhase: {
   	if enabled == false {
   		status: "skipped"
   	}
   	if enabled == true {
   		status: "pending" | "in_progress" | "completed"
   	}
   }
   ```

2. **Required Phases Always Enabled**:
   ```cue
   // Implementation, review, finalize must always have enabled=true
   #ImplementationPhase: {
   	enabled: true
   }

   #ReviewPhase: {
   	enabled: true
   }

   #FinalizePhase: {
   	enabled: true
   }
   ```

3. **Timestamp Ordering**:
   ```cue
   // For any phase:
   // - started_at must be after created_at (if not null)
   // - completed_at must be after started_at (if not null)
   ```

4. **Phase Ordering**:
   ```cue
   // A phase cannot be "in_progress" or "completed" if a previous enabled phase is "pending"
   // Order: discovery → design → implementation → review → finalize
   ```

5. **Artifact Approval**:
   ```cue
   // Discovery and Design phases cannot transition to "completed"
   // unless all artifacts have approved=true
   ```

6. **Project Deletion Gate**:
   ```cue
   // Finalize phase cannot have status="completed" unless project_deleted=true
   ```

---

## Example: Complete Project State

This example shows a project that went through all five phases, including a review loop-back where issues were found and additional implementation work was required.

```yaml
project:
  name: add-authentication
  branch: feat/add-auth
  description: Add JWT-based authentication system with user login and token management
  created_at: "2025-10-14T10:00:00Z"
  updated_at: "2025-10-14T17:00:00Z"

phases:
  discovery:
    enabled: true
    status: completed
    created_at: "2025-10-14T10:05:00Z"
    started_at: "2025-10-14T10:05:00Z"
    completed_at: "2025-10-14T11:20:00Z"
    discovery_type: feature
    artifacts:
      - path: phases/discovery/notes.md
        approved: true
        created_at: "2025-10-14T11:15:00Z"
      - path: phases/discovery/research/001-jwt-libraries.md
        approved: true
        created_at: "2025-10-14T10:30:00Z"

  design:
    enabled: true
    status: completed
    created_at: "2025-10-14T11:25:00Z"
    started_at: "2025-10-14T11:25:00Z"
    completed_at: "2025-10-14T13:00:00Z"
    architect_used: true
    artifacts:
      - path: phases/design/adrs/001-use-jwt-rs256.md
        approved: true
        created_at: "2025-10-14T12:30:00Z"
      - path: phases/design/design-docs/auth-system.md
        approved: true
        created_at: "2025-10-14T12:45:00Z"

  implementation:
    enabled: true
    status: completed
    created_at: "2025-10-14T13:05:00Z"
    started_at: "2025-10-14T13:10:00Z"
    completed_at: "2025-10-14T16:20:00Z"
    planner_used: false
    tasks:
      - id: "010"
        name: Create User model with validation
        status: completed
        parallel: false
      - id: "020"
        name: Implement JWT token generation service
        status: completed
        parallel: false
        dependencies: ["010"]
      - id: "030"
        name: Create login endpoint
        status: completed
        parallel: false
        dependencies: ["020"]
      - id: "040"
        name: Add authentication middleware
        status: completed
        parallel: false
        dependencies: ["020"]
      # Task added via fail-forward after first review found missing error handling
      - id: "050"
        name: Add comprehensive error handling to auth endpoints
        status: completed
        parallel: false
        dependencies: ["030"]

  review:
    enabled: true
    status: completed
    created_at: "2025-10-14T15:50:00Z"
    started_at: "2025-10-14T15:50:00Z"
    completed_at: "2025-10-14T16:30:00Z"
    iteration: 2  # Went through two review iterations
    reports:
      # First review found issues
      - id: "001"
        path: phases/review/reports/001-review.md
        created_at: "2025-10-14T15:55:00Z"
        assessment: fail
      # Second review after fix - passed
      - id: "002"
        path: phases/review/reports/002-review.md
        created_at: "2025-10-14T16:25:00Z"
        assessment: pass

  finalize:
    enabled: true
    status: completed
    created_at: "2025-10-14T16:35:00Z"
    started_at: "2025-10-14T16:35:00Z"
    completed_at: "2025-10-14T17:00:00Z"
    documentation_updates:
      - README.md
      - CHANGELOG.md
      - docs/api-reference.md
    artifacts_moved:
      - from: phases/design/adrs/001-use-jwt-rs256.md
        to: docs/adrs/003-use-jwt-rs256.md
    project_deleted: true
    pr_url: https://github.com/org/repo/pull/42
```

---

## Task State Schema (No Changes Required)

The task state schema **does not need changes** for the new model. It continues to work as-is:

**Reasoning**:
- Task states only exist in implementation phase
- Implementation phase still uses same task execution model
- Fields (iteration, feedback, references) remain relevant
- Worker agents read task state the same way

**Current Schema Remains Valid**:

```cue
// Task State Schema (No Changes)
// Location: .sow/project/phases/implementation/tasks/<id>/state.yaml

package sow

import "time"

#TaskState: {
	task: #Task
}

#Task: {
	id:          string & =~"^[0-9]{3}$"
	name:        string
	phase:       "implementation"  // Always implementation in new model
	status:      "pending" | "in_progress" | "completed" | "abandoned"
	created_at:  string & time.Format(time.RFC3339)
	started_at:  null | (string & time.Format(time.RFC3339))
	updated_at:  string & time.Format(time.RFC3339)
	completed_at: null | (string & time.Format(time.RFC3339))
	iteration:   int & >=1
	references: [...string]
	feedback:   [...#Feedback]
	files_modified: [...string]
}

#Feedback: {
	id:         string & =~"^[0-9]{3}$"
	created_at: string & time.Format(time.RFC3339)
	status:     "pending" | "addressed" | "superseded"
}
```

**Note**: The only change is semantic - `phase` is now always `"implementation"`. The `assigned_agent` field has been removed as it's redundant (all implementation tasks use the implementer agent).

---

## File Structure Implications

### Project Directory Structure

With the new schema, the complete project structure becomes:

```
.sow/project/
├── state.yaml                          # Revised project state (this proposal)
│
├── phases/
│   ├── discovery/                      # Only exists if discovery enabled
│   │   ├── log.md                     # Conversation log
│   │   ├── notes.md                   # Orchestrator's notes
│   │   ├── decisions.md               # Key decisions
│   │   └── research/                  # Researcher agent outputs
│   │       ├── 001-topic.md
│   │       └── 002-topic.md
│   │
│   ├── design/                         # Only exists if design enabled
│   │   ├── log.md                     # Conversation log
│   │   ├── notes.md                   # Design alignment notes
│   │   ├── requirements.md            # Formalized requirements (optional)
│   │   ├── adrs/                      # Architecture Decision Records
│   │   │   ├── 001-decision.md
│   │   │   └── 002-decision.md
│   │   ├── design-docs/               # Design documents
│   │   │   └── system-design.md
│   │   └── diagrams/                  # Diagrams (optional)
│   │       └── architecture.png
│   │
│   ├── implementation/                 # Always exists (required phase)
│   │   ├── log.md                     # Orchestrator coordination log
│   │   ├── implementation-plan.md     # Planner output (if used)
│   │   └── tasks/
│   │       ├── 010/
│   │       │   ├── state.yaml        # Task metadata
│   │       │   ├── description.md    # Task requirements
│   │       │   ├── log.md            # Implementer action log
│   │       │   └── feedback/         # Human corrections
│   │       │       └── 001.md
│   │       └── 020/
│   │           └── ...
│   │
│   ├── review/                         # Always exists (required phase)
│   │   ├── log.md                     # Review conversation log
│   │   └── reports/
│   │       ├── 001-review.md          # First review
│   │       └── 002-review.md          # Second review (if looped back)
│   │
│   └── finalize/                       # Always exists (required phase)
│       └── log.md                     # Finalize phase log
```

### Conditional Directory Creation

**Discovery and Design Directories**:
- Only created if `enabled: true` in state.yaml
- Skipped phases never create directories
- Keeps project structure minimal for simple projects

**Implementation, Review, Finalize Directories**:
- Always created (required phases)
- Even if work is trivial, these phases always execute

---

## Schema Finalized

All design decisions have been made:

1. ✅ **`pending_task_additions`** - Kept in project state under `implementation.pending_task_additions`
2. ✅ **Artifact approval** - Boolean only (`approved: bool`), no timestamp
3. ✅ **Review iteration counter** - Explicit field (`review.iteration`)

## Next Steps

**Proceed to Step 4** (Rewrite Orchestrator) to incorporate this finalized schema structure into the orchestrator agent behavior.

---

## References

- **PHASE_SPECIFICATIONS.md** - Complete phase behavior specifications
- **TRUTH_TABLE.md** - Project initialization decision flow
- **Current Schema**: `schemas/cue/project-state.cue` (v0.2.0)
- **Current Schema**: `schemas/cue/task-state.cue` (v0.2.0)
- **ADR 001**: Go CLI Architecture decisions
