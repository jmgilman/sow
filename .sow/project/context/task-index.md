# Task Index

This index lists all tasks for the standard project type SDK migration in execution order.

## Task List

| ID  | Name | Description | Dependencies |
|-----|------|-------------|--------------|
| 010 | Create Package Structure | Set up directory structure and copy templates | None |
| 020 | Define States and Events | Create state/event constants using SDK types | 010 |
| 030 | Create Metadata Schemas | Define CUE schemas and Go embeddings | 010 |
| 040 | Create Prompt Functions | Extract prompt logic into SDK-compatible functions | 010, 020 |
| 050 | Implement Guard Functions (TDD) | Write tests first, then implement guards | 020 |
| 060 | Implement SDK Configuration (TDD) | Write integration tests first, then configure SDK | 020, 030, 040, 050 |
| 070 | Verify Old Implementation Untouched | Final verification of clean migration | 060 |

## Task Files

All task descriptions are located in `.sow/project/context/tasks/`:

- `010-create-package-structure.md`
- `020-define-states-and-events.md`
- `030-create-metadata-schemas.md`
- `040-create-prompt-functions.md`
- `050-implement-guard-functions-tdd.md`
- `060-implement-sdk-configuration-tdd.md`
- `070-verify-old-implementation-untouched.md`

## Execution Order

Tasks should be executed in numerical order (010 → 070). Dependencies are tracked in the table above.

**TDD Tasks**: Tasks 050 and 060 require Test-Driven Development:
1. Write tests FIRST (red phase)
2. Implement to make tests pass (green phase)
3. Refactor while keeping tests green

## Key Integration Points

- **Task 060** integrates all previous tasks (states, events, guards, prompts, schemas)
- **Task 070** verifies the entire migration left old code untouched

## Standards

All tasks must adhere to:
- TDD for tasks 050 and 060
- All tests pass
- Linters pass with no warnings
- Old `internal/project/standard/` completely untouched
- Both implementations coexist cleanly
