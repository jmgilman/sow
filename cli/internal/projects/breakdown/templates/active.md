---

## Guidance: Active Breakdown

You are in the **Active** state of a breakdown project. Your focus is decomposing features into work units, specifying requirements, and managing review cycles.

### Identifying Work Units

Create tasks for each implementable unit of work:

```bash
sow task add "Work unit name" --id <3-digit-id>
```

Examples:
- "JWT Token Generation Middleware"
- "User Authentication Endpoint"
- "Session Management Service"
- "Token Validation Helper"

**Important**: Create at least one work unit task before adding specifications. Each specification must link to a task.

### Specifying Work Units

For each work unit:

1. **Start drafting**:
   ```bash
   sow task start <id>
   ```

2. **Write the specification**:
   - Create specification in project workspace (e.g., `project/work-units/001-jwt.md`)
   - Use markdown format
   - Include clear sections:
     - **Overview**: Brief description
     - **Requirements**: What must be built
     - **Acceptance Criteria**: How to verify completion
     - **Dependencies**: Other work units this depends on (if any)
     - **Technical Notes**: Implementation guidance, constraints
   - Reference input sources when relevant

3. **Register artifact**:
   ```bash
   sow output add --type work_unit_spec --path "project/work-units/<id>-name.md"
   ```

4. **Link artifact to task**:
   ```bash
   sow task set <id> artifact_path "project/work-units/<id>-name.md"
   ```

### Declaring Dependencies

If a work unit depends on others:

```bash
sow task set <id> dependencies "001,002,003"
```

**Critical rules**:
- Dependencies must form a directed acyclic graph (DAG) - no cycles
- All referenced task IDs must exist and be completed
- Dependencies are validated before advancing to Publishing
- Self-references are not allowed

Example dependency chain:
```
001 - JWT Middleware (no dependencies)
002 - Auth Endpoint (depends on: 001)
003 - Protected Routes (depends on: 001, 002)
```

### Review Workflow

When specification draft is ready:

1. **Request review**:
   ```bash
   sow task set <id> status needs_review
   ```

2. **User reviews specification**:
   - User reads the specification document
   - User provides feedback if changes needed
   - User signals approval when satisfied

3. **Iterate if needed**:
   - If changes needed, return to `in_progress`:
     ```bash
     sow task set <id> status in_progress
     ```
   - Make revisions to specification
   - Request review again when ready

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
- Reference them in your specifications
- Build on requirements from previous design work
- Link to input sources for traceability

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

- **Start simple**: Begin with core work units, add more as you understand the feature
- **Be specific**: Write clear acceptance criteria that can be objectively verified
- **Review early**: Request review on partial specs to catch issues early
- **Link everything**: Always link artifacts to tasks for proper tracking
- **Think dependencies**: Identify which work units must come first
- **Iterate freely**: The review workflow supports multiple revision cycles
- **Reference inputs**: Build on existing design work or specifications
- **Check cycles**: Dependency cycles will block advancement - review your dependency graph
