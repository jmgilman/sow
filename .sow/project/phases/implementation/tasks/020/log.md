# Task Log

## 2025-11-05 - Implementation of Guard Functions and Helpers

### Context
Task 020: Implement guard functions and helper utilities for the design project type.

### Approach
Followed TDD (Test-Driven Development) methodology:
1. Read task requirements and reference implementations from exploration project type
2. Created comprehensive test suite first (guards_test.go)
3. Implemented guard and helper functions to pass the tests (guards.go)
4. Verified all tests pass

### Test Coverage Created (guards_test.go)

#### allDocumentsApproved tests (7 tests):
- Missing design phase
- No tasks exist
- Tasks with pending status
- Tasks with in_progress status
- All tasks abandoned (requires at least one completed)
- All tasks completed (success case)
- Mix of completed and abandoned tasks (success case)

#### allFinalizationTasksComplete tests (6 tests):
- Missing finalization phase
- No tasks exist
- Tasks with pending status
- Tasks with in_progress status
- Abandoned tasks not allowed in finalization
- All tasks completed (success case)

#### countUnresolvedTasks tests (3 tests):
- Missing design phase returns 0
- All resolved tasks returns 0
- Mixed statuses returns correct count

#### validateTaskForCompletion tests (7 tests):
- Missing design phase error
- Task not found error
- No metadata error
- Missing artifact_path error
- Invalid artifact_path type error
- Artifact not found at path error
- Successful validation

#### autoApproveArtifact tests (6 tests):
- Missing design phase error
- Task not found error
- Invalid artifact_path error
- Artifact not found error
- Successful approval
- Correctly updates project state

### Implementation (guards.go)

#### Guard Functions (Pure):
1. **allDocumentsApproved**: Guards Active → Finalizing transition
   - Checks all design tasks are completed or abandoned
   - Requires at least one completed task (prevents all abandoned)
   - Returns boolean

2. **allFinalizationTasksComplete**: Guards Finalizing → Completed transition
   - Checks all finalization tasks are completed
   - Does not allow abandoned status (stricter than design tasks)
   - Returns boolean

#### Helper Functions:
3. **countUnresolvedTasks**: Returns count of unresolved design tasks
   - Used for status messages and prompts
   - Returns 0 if phase doesn't exist

4. **validateTaskForCompletion**: Validates task can be marked completed
   - Checks task exists
   - Validates metadata structure
   - Validates artifact_path exists and is valid
   - Verifies artifact exists in phase outputs
   - Returns descriptive errors

5. **autoApproveArtifact**: Auto-approves artifact when task completes
   - Finds task by ID
   - Extracts artifact_path from metadata
   - Locates artifact in phase outputs
   - Sets Approved = true
   - Updates project state
   - Only non-pure function (mutates state)

### Key Design Decisions

1. **Guard Purity**: Guards are pure functions (no side effects) except autoApproveArtifact which explicitly mutates state as documented.

2. **Error Messages**: All error messages are clear and actionable, telling the user exactly what they need to do.

3. **Phase-Specific Logic**:
   - Design tasks can be abandoned (with at least one completed)
   - Finalization tasks must all be completed (no abandoning)

4. **Validation Philosophy**: Enforces "plan before draft before complete" workflow:
   - Must link artifact to task (artifact_path in metadata)
   - Artifact must exist before task completion
   - Completing task auto-approves artifact

### Test Results
All 32 tests passing:
- 4 existing tests (events, metadata, states)
- 28 new tests (guards and helpers)

### Files Created
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/design/guards_test.go` (31,488 bytes)
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/design/guards.go` (7,561 bytes)

### Status
Task complete. All acceptance criteria met:
- All guard functions implemented with correct edge case handling
- All helper functions implemented with comprehensive validation
- Error messages are clear and actionable
- Functions are pure (except autoApproveArtifact as documented)
- Comprehensive test coverage for all functions
- All tests passing

Ready for review.
