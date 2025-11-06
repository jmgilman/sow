---

## Guidance: Active Breakdown

You are in the **Active** state of a breakdown project. Your focus is identifying work units based on the discovery document, spawning decomposer agents to write specifications, and managing review cycles.

### Your Role as Orchestrator

In the Active state, you coordinate the breakdown process:

1. **Review discovery document** - Understand codebase context
2. **Identify work units** - Determine project-sized decomposition (2-3 days each)
3. **Create tasks** - One task per work unit
4. **Spawn decomposers** - One decomposer agent per task to write specs
5. **Manage reviews** - Coordinate review cycles and approval

### Identifying Work Units

Review the approved discovery document and identify work units. Each work unit should be:
- **Project-sized**: 2-3 days minimum of implementation work
- **Self-contained**: Can be implemented as a standalone project
- **Well-scoped**: Clear boundaries and deliverables

Create tasks for each work unit:

```bash
sow task add "Work unit name" --id <3-digit-id>
```

Examples (project-sized):
- "OAuth2 Authentication Flow Implementation"
- "API Rate Limiting and Throttling System"
- "User Profile Management with CRUD Operations"
- "Real-time Notification Delivery System"

**Important**: Each work unit becomes its own project with orchestrator + implementers. Size accordingly.

### Spawning Decomposer Agents

For each task, spawn a decomposer agent to write the specification. **IMPORTANT**: Decomposers start with zero context, so you must provide everything they need.

#### Step 1: Write Task Description File

Create a `description.md` file for the task that explains the work unit:

```bash
# Create task description file
# Location: .sow/project/phases/breakdown/tasks/<task-id>/description.md
```

The description file should include:
- High-level goal and user benefit
- Scope boundaries (what's in/out of scope)
- Key technical areas to address
- Any specific requirements or constraints

Example (`description.md`):
```markdown
# Work Unit 001: OAuth2 Authentication Flow

Implement OAuth2 authentication flow for user login.

## Scope
- Token exchange with OAuth2 provider
- Session management and storage
- Refresh token handling

## Integration Points
- Must integrate with existing UserService
- Must integrate with SessionManager

## Out of Scope
- Social login providers (separate work unit)
```

#### Step 2: Register Input Artifacts

Register all context the decomposer needs to read:

```bash
# Always register the discovery document
sow task input add --id <task-id> --type discovery --path project/discovery/analysis.md

# Register relevant ADRs
sow task input add --id <task-id> --type adr --path .sow/knowledge/adrs/005-oauth-provider.md

# Register relevant design docs
sow task input add --id <task-id> --type design --path .sow/knowledge/design/auth-system.md

# Register input artifacts from breakdown phase (if any)
sow task input add --id <task-id> --type design --path <path-from-phase-inputs>
```

#### Step 3: Spawn Decomposer

Now spawn the decomposer agent with the task context:

```bash
# Use the Task tool to spawn decomposer agent
# Pass the task ID so it knows what to work on
```

**Decomposer will:**
- Read task description and registered inputs
- Explore existing code
- Write comprehensive specification
- Register as work_unit_spec artifact
- Link to task via metadata.artifact_path
- Mark task as needs_review when complete

### Specification Structure

Each decomposer will create a specification with:

1. **Behavioral Goal** (User Story)
   - "As a [user], I need [capability] so that [benefit]"
   - Focus on intended behavior, not just technical details
   - Clear success criteria for reviewers

2. **Existing Code Context** (Dual Format)
   - **Explanatory**: "This work unit leverages X which handles Y. We'll extend Z..."
   - **Reference List**:
     ```
     Key Files:
     - path/to/file.go:45-120 (Service class)
     - path/to/other.go:78-95 (Interface)
     ```

3. **Existing Documentation Context**
   - Not just links: "ADR-005 chose Auth0 over custom due to compliance. This implements section 3..."
   - Explain what's relevant to this work unit

4. **Dependencies**
   - Which work units must complete first
   - Why the dependency exists

5. **Acceptance Criteria**
   - Objective, measurable completion criteria
   - What the reviewer will verify

### Declaring Dependencies

Decomposer agents declare dependencies in task metadata:

```bash
sow task set <id> metadata.dependencies "001,002,003"
```

**Critical rules**:
- Dependencies must form a directed acyclic graph (DAG) - no cycles
- All referenced task IDs must exist and be completed
- Dependencies are validated before advancing to Publishing
- Self-references are not allowed

Example dependency chain:
```
001 - Authentication System (no dependencies)
002 - User Profile API (depends on: 001)
003 - Admin Dashboard (depends on: 001, 002)
```

### Review Workflow

As orchestrator, you manage the review cycle for each work unit:

1. **Decomposer completes draft**:
   - Decomposer writes specification
   - Registers as work_unit_spec artifact
   - Links to task via metadata.artifact_path
   - Marks task as needs_review

2. **You (orchestrator) review specification**:
   - Read the specification document
   - Verify it meets quality standards
   - Check behavioral goal is clear
   - Ensure existing code/docs are referenced properly
   - Verify scope is project-sized (2-3 days minimum)

3. **Iterate if needed**:
   - If changes needed, spawn decomposer again with feedback:
     ```bash
     sow task set <id> status in_progress
     ```
   - Provide clear feedback on what needs revision
   - Decomposer revises and marks needs_review again

4. **Complete when approved**:
   ```bash
   sow task set --id <id> status completed
   ```

   **Note**: Completing the task automatically approves the linked artifact. No separate approval step needed.

### Abandoning Work Units

If a work unit is no longer needed:

```bash
sow task abandon <id> --reason "No longer required"
```

Abandoned work units don't count against advancement readiness.

### Working with Inputs

If your breakdown phase has input artifacts (e.g., design documents):
- Share them with decomposer agents as context
- Ensure decomposers reference them in specifications
- Build on requirements from previous design work
- Maintain traceability from inputs through discovery to work units

### Advancement Criteria

You can advance to Publishing when:
- All work unit tasks are resolved (completed or abandoned)
- At least one work unit is completed (not all abandoned)
- All dependencies form a valid DAG (no cycles or invalid references)

Ready to advance? Run:
```bash
sow project advance
```

The guard will check that:
1. All work units are completed or abandoned
2. At least one work unit is completed
3. All dependencies are valid and form a DAG

### Tips

- **Review discovery first**: Understand what already exists before identifying work units
- **Size appropriately**: Each work unit = 2-3 days minimum, becomes its own project
- **Start with core units**: Identify critical work units first, add supporting units as needed
- **Provide context to decomposers**: Share discovery doc, relevant ADRs, and clear scope
- **Review for quality**: Ensure specs have behavioral goals, code references, and doc context
- **Think dependencies**: Identify which work units must come first
- **Iterate freely**: The review workflow supports multiple revision cycles
- **Check cycles**: Dependency cycles will block advancement - review your dependency graph
- **Parallel decomposition**: Can spawn multiple decomposers simultaneously for different work units
