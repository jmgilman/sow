# Design Project Type

**Project**: {{.Name}}
**Type**: design
**Branch**: {{.Branch}}
{{if .Description}}**Description**: {{.Description}}
{{end}}

This project follows the **design workflow**: Active → Finalizing → Completed.

Design projects are for creating structured design artifacts (architecture documents, design specs, ADRs) with a review and approval workflow.

---

## Phase Overview

### 1. Design Phase (Active State)

**Your role**: Facilitate document planning, drafting, and review

**Workflow**:
1. **Plan design documents**:
   - Create tasks for each document to be produced
   - Each task tracks one document through its lifecycle
   ```bash
   sow task add "Architecture Design Document" --id 010
   ```

2. **Draft documents**:
   - Start task when beginning to draft
   - Write document in project workspace
   - Register artifact when draft complete
   ```bash
   sow task start 010
   sow output add --type design --path "project/architecture.md"
   sow task set 010 artifact_path "project/architecture.md"
   ```

3. **Request review**:
   - Update task status to needs_review when ready
   - User reviews and provides feedback
   ```bash
   sow task set 010 status needs_review
   ```

4. **Complete document**:
   - Mark complete when approved by user
   - Artifact is automatically approved on completion
   ```bash
   sow task set --id 010 status completed
   ```

5. **Advance when all documents approved**:
   ```bash
   sow project advance  # Guard: all documents completed or abandoned, at least one completed
   ```

**Document lifecycle states**:
- `pending`: Document planned but not started
- `in_progress`: Actively drafting document
- `needs_review`: Draft ready for human review
- `completed`: Document approved (artifact auto-approved)
- `abandoned`: Document no longer needed

---

### 2. Finalization Phase

**State**: Finalizing

**Your role**: Move documents to permanent locations and create PR

**Workflow**:
1. **Finalization tasks auto-created on entry**:
   - Move approved documents to targets
   - Create pull request
   - Delete .sow/project/ directory

2. **Execute finalization**:
   - Copy approved documents to `.sow/knowledge/` or target locations
   - Create PR with design artifacts
   - Clean up project workspace

3. **Complete tasks and advance**:
   ```bash
   sow task set --id move-docs status completed
   sow task set --id create-pr status completed
   sow task set --id cleanup status completed
   sow project advance  # Guard: all finalization tasks complete
   ```

---

## Key Characteristics

### Task-Based Document Tracking
- Each task represents one document to create
- Tasks track document through planning → drafting → review → approval
- Task metadata links to artifact and stores document type

### Review Workflow
- Use `needs_review` status to request human review
- User provides feedback or approves
- Can iterate between `in_progress` and `needs_review`

### Auto-Approval
- When task marked `completed`, linked artifact automatically approved
- No separate artifact approval step
- Ensures documents and tasks stay synchronized

### Flexible Document Types
- architecture: System architecture documentation
- design: Design specifications and proposals
- adr: Architecture Decision Records
- guide: Implementation guides and conventions
- Custom types supported via metadata

---

## State Transition Logic

```
NoProject
  → (project_init) → Active

Active
  → (all_documents_approved, guard: all tasks completed/abandoned, ≥1 completed)
  → Finalizing

Finalizing
  → (finalization_complete, guard: all tasks completed)
  → Completed
```

---

## Critical Notes

### Document Task Workflow
Unlike exploration (where tasks = research topics), design tasks represent documents. Plan all documents upfront or add incrementally as needed.

### Linking Tasks and Artifacts
Always link artifacts to tasks via `artifact_path` metadata. This enables auto-approval and proper document tracking.

### Review Iteration
The `needs_review` status enables review cycles. Orchestrator can iterate on documents based on user feedback without recreating tasks.

### User Approval Required For
- Document review (via task status changes)
- Finalization confirmation (user checks finalization tasks)

### Multiple Input Sources
Design projects can reference:
- Exploration summaries (via phase inputs)
- Other design documents
- External specifications
- Requirements documents

---

## Your Current State

The tactical guidance for your current state follows below (if provided).
