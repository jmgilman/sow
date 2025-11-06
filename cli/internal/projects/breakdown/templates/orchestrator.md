# Breakdown Project Type

**Project**: {{.Name}}
**Type**: breakdown
**Branch**: {{.Branch}}
{{if .Description}}**Description**: {{.Description}}
{{end}}

This project follows the **breakdown workflow**: Active → Publishing → Completed.

Breakdown projects are for decomposing complex features or design documents into discrete, implementable work units that can be published as GitHub issues.

---

## Phase Overview

### 1. Breakdown Phase (Active State)

**Your role**: Decompose features into work units, specify requirements, manage review workflow

**Workflow**:
1. **Identify work units**:
   - Create tasks for each implementable unit of work
   - Each task represents one future GitHub issue
   ```bash
   sow task add "Work unit name" --id <3-digit-id>
   ```

2. **Specify work units**:
   - Write detailed specification for each unit
   - Include requirements, acceptance criteria, dependencies
   - Store in project workspace
   ```bash
   sow task start <id>
   # Write specification file
   sow output add --type work_unit_spec --path "project/work-units/<id>-name.md"
   sow task set <id> metadata.artifact_path "project/work-units/<id>-name.md"
   ```

3. **Declare dependencies**:
   - Set dependencies using task IDs
   - Dependencies must form a directed acyclic graph (no cycles)
   ```bash
   sow task set <id> metadata.dependencies "001,002"
   ```

4. **Request review**:
   - Update task status to needs_review when spec is ready
   - User reviews specification and provides feedback
   ```bash
   sow task set <id> status needs_review
   ```

5. **Complete work unit**:
   - Mark complete when approved by user
   - Artifact is automatically approved on completion
   ```bash
   sow task set --id <id> status completed
   ```

6. **Advance when all work units approved**:
   ```bash
   sow project advance  # Guard: all work units completed/abandoned, dependencies valid
   ```

**Work unit lifecycle states**:
- `pending`: Work unit identified but not specified
- `in_progress`: Actively writing specification
- `needs_review`: Specification ready for human review
- `completed`: Work unit approved (artifact auto-approved)
- `abandoned`: Work unit no longer needed

---

### 2. Publishing Phase

**State**: Publishing

**Your role**: Publish approved work units as GitHub issues in dependency order

**Workflow**:
1. **Determine publishing order**:
   - Read task dependencies to build dependency graph
   - Perform topological sort to determine safe publishing order
   - Publish parent work units before children

2. **Publish each work unit**:
   - For each work unit in dependency order:
     - Check if already published (metadata.published)
     - Create GitHub issue with specification as body
     - Update task metadata with issue number and URL
   ```bash
   gh issue create --title "<task-name>" --body "$(cat <spec-path>)" --label sow
   sow task set --id <id> metadata.published true
   sow task set --id <id> metadata.github_issue_number <number>
   sow task set --id <id> metadata.github_issue_url <url>
   ```

3. **Advance when all published**:
   ```bash
   sow project advance  # Guard: all completed work units have metadata.published == true
   ```

---

## Key Characteristics

### Task-Based Work Unit Tracking
- Each task represents one work unit to implement
- Tasks track work unit through planning → specification → review → approval
- Task metadata links to specification artifact and stores dependencies

### Review Workflow
- Use `needs_review` status to request human review
- User provides feedback or approves
- Can iterate between `in_progress` and `needs_review`

### Auto-Approval
- When task marked `completed`, linked artifact automatically approved
- No separate artifact approval step
- Ensures specifications and tasks stay synchronized

### Dependency Management
- Dependencies declared as list of task IDs in metadata
- Must form valid DAG (directed acyclic graph, no cycles)
- Publishing respects dependency order (topological sort)

### GitHub Integration
- Publishing phase creates one GitHub issue per completed work unit
- Issues created in dependency order to maintain coherence
- Resumability: already-published work units are skipped (check metadata.published)

---

## State Transition Logic

```
NoProject
  → (project_init) → Active

Active
  → (all_work_units_approved, guard: all tasks completed/abandoned, ≥1 completed, dependencies valid)
  → Publishing

Publishing
  → (all_work_units_published, guard: all completed tasks have metadata.published == true)
  → Completed
```

---

## Critical Notes

### Work Unit Task Workflow
Unlike design (documents) or exploration (research topics), breakdown tasks represent implementable work units. Plan work units as you discover them during decomposition.

### Linking Tasks and Artifacts
Always link artifacts to tasks via `artifact_path` metadata. This enables auto-approval and proper work unit tracking.

### Review Iteration
The `needs_review` status enables review cycles. Orchestrator can iterate on specifications based on user feedback without recreating tasks.

### Dependency Validation
Dependencies MUST form a valid DAG. The guard function validates:
- All dependency references point to valid completed tasks
- No cyclic dependencies (including self-references)
- Publishing order determined by topological sort

### Resumability
Publishing workflow supports interruption and resumption:
- Check `metadata.published` before creating each issue
- Skip already-published work units
- No duplicate issues created on retry

### User Approval Required For
- Work unit specification review (via task status changes)
- Dependency validation (automatic via guard)
- Publishing confirmation (user can review before advancing)

### Input Sources
Breakdown projects can reference:
- Design documents (via phase inputs)
- Feature specifications
- Architecture documents
- Requirements documents

---

## Your Current State

The tactical guidance for your current state follows below (if provided).
