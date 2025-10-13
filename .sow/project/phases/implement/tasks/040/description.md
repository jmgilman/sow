# Task: Implement utility commands

## Objective

Implement `sow log`, `sow session-info`, and `sow validate` commands for agent logging, context detection, and validation.

## Context

These utility commands support agent workflows and state validation.

## Requirements

### Command: `sow log`

1. **Functionality**:
   - Append structured log entry to appropriate log.md file
   - Auto-detect context (task vs project)
   - Support both orchestrator and worker logs
   - Fast performance (<1s)

2. **Usage**:
   - `sow log --action <action> --details "<details>"` (in project context)
   - `sow log --task-id 010 --action <action> --details "<details>"` (in task context)

3. **Features**:
   - ISO 8601 timestamp
   - Agent ID construction
   - Markdown formatting
   - Atomic file appends

### Command: `sow session-info`

1. **Functionality**:
   - Detect and report session context
   - Used by SessionStart hooks

2. **Output** (JSON):
   ```json
   {
     "context": "task|project|none",
     "task_id": "010",
     "phase": "design",
     "cli_version": "0.1.0"
   }
   ```

3. **Features**:
   - Fast detection (<100ms)
   - JSON output for parsing
   - Error handling for invalid states

### Command: `sow validate`

1. **Functionality**:
   - Validate file(s) against CUE schemas
   - Auto-detect file type or use --type flag
   - Support glob patterns

2. **Usage**:
   - `sow validate .sow/project/state.yaml`
   - `sow validate --type task-state .sow/project/phases/*/tasks/*/state.yaml`
   - `sow validate --all` (validate all known files)

3. **Output**:
   - Success/failure per file
   - Clear error messages
   - Exit code (0=valid, 1=invalid)

### All Commands

- Help text and examples
- Error handling
- Tests
- Performance requirements met

## References

- `docs/CLI_REFERENCE.md` - Command specifications

## Deliverables

- [ ] `sow log` command implemented
- [ ] `sow session-info` command implemented
- [ ] `sow validate` command implemented
- [ ] All performance requirements met
- [ ] Tests passing
