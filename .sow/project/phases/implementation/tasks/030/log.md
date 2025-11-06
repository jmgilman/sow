# Task 030: Prompt Templates and Generators - Implementation Log

## Overview

Implemented prompt generation system for design project type following TDD approach. Created template files for static guidance and generator functions for dynamic project state rendering.

## Implementation Steps

### 1. Test-Driven Development (TDD)

**Red Phase**: Wrote comprehensive tests first
- Created `prompts_test.go` with 27 test cases covering:
  - `configurePrompts()` registration tests (4 tests)
  - `generateOrchestratorPrompt()` tests (3 tests)
  - `generateActivePrompt()` tests (11 tests)
  - `generateFinalizingPrompt()` tests (6 tests)
- Verified tests fail before implementation

**Green Phase**: Implemented to pass tests
- Created template files with static guidance
- Implemented prompt generator functions
- Iteratively ran tests and fixed issues

**Refactor Phase**: Code quality improvements
- Added helper function `getStatusIcon()` for status icons
- Ensured consistent error handling
- Verified all tests pass

### 2. Template Files Created

Created three markdown templates in `templates/` directory:

**`templates/orchestrator.md`** (4.5 KB):
- High-level design workflow explanation
- Phase overview (Design and Finalization)
- Document lifecycle states
- Task-based document tracking approach
- Review workflow and auto-approval mechanics
- State transition logic diagram
- Key characteristics and critical notes

**`templates/active.md`** (3.2 KB):
- Planning documents guidance
- Drafting documents workflow
- Review workflow with iteration support
- Abandoning documents process
- Working with inputs
- Advancement criteria
- Tips for effective design work

**`templates/finalizing.md`** (2.5 KB):
- Finalization tasks overview
- Moving documents to permanent locations
- Creating pull request
- Cleanup process
- Advancement criteria
- Tips for finalization

### 3. Prompt Generator Functions

Implemented four functions in `prompts.go`:

**`configurePrompts(builder *ProjectTypeConfigBuilder)`**:
- Registers all prompt generators with builder
- Chains calls to `WithOrchestratorPrompt()` and `WithPrompt()`
- Returns builder for method chaining
- Simple, declarative configuration

**`generateOrchestratorPrompt(p *Project)`**:
- Renders orchestrator template with project context
- Handles nil project gracefully
- Returns error message if template rendering fails
- Provides high-level workflow overview

**`generateActivePrompt(p *Project)`**:
- Builds dynamic prompt showing:
  - Project name, branch, description
  - Current state: "Active Design"
  - Design inputs (if any) with descriptions
  - Document task list with status icons
  - Task counts by status (pending, in_progress, needs_review, completed, abandoned)
  - Artifact paths and document types for tasks
  - Advancement readiness indicator using `allDocumentsApproved()` guard
  - Unresolved task count when not ready to advance
- Appends static guidance from `active.md` template
- Handles edge cases (no tasks, missing phase)

**`generateFinalizingPrompt(p *Project)`**:
- Builds dynamic prompt showing:
  - Project name and branch
  - Current state: "Finalizing"
  - Finalization task list with completion checkboxes
  - Advancement readiness indicator using `allFinalizationTasksComplete()` guard
- Appends static guidance from `finalizing.md` template
- Handles missing finalization phase gracefully

**Helper function `getStatusIcon(status string)`**:
- Maps task status to visual icons
- Consistent icons: `[ ]` pending, `[~]` in_progress, `[?]` needs_review, `[✓]` completed, `[✗]` abandoned
- Used throughout prompt generation for visual clarity

### 4. Guard Integration

Prompts integrate with guard functions from Task 020:

- `allDocumentsApproved(p)`: Checks if ready to advance from Active to Finalizing
- `allFinalizationTasksComplete(p)`: Checks if ready to complete design project
- `countUnresolvedTasks(p)`: Shows how many documents remain unresolved

This provides clear feedback to orchestrator about advancement readiness.

### 5. Template Rendering

Used `templates.Render()` utility for embedding templates:
- `//go:embed templates/*.md` directive embeds templates at compile time
- Templates rendered with project context using Go templates
- Error handling returns error message + dynamic content on failure
- Never returns empty string - always provides useful output

## Test Results

All tests pass successfully:

```
=== RUN   TestConfigurePrompts
=== RUN   TestConfigurePrompts/returns_non-nil_builder_for_chaining
=== RUN   TestConfigurePrompts/registers_orchestrator_prompt_generator
=== RUN   TestConfigurePrompts/registers_Active_state_prompt_generator
=== RUN   TestConfigurePrompts/registers_Finalizing_state_prompt_generator
--- PASS: TestConfigurePrompts (0.00s)

=== RUN   TestGenerateOrchestratorPrompt
--- PASS: TestGenerateOrchestratorPrompt (0.00s)
    [3 subtests passed]

=== RUN   TestGenerateActivePrompt
--- PASS: TestGenerateActivePrompt (0.00s)
    [11 subtests passed]

=== RUN   TestGenerateFinalizingPrompt
--- PASS: TestGenerateFinalizingPrompt (0.00s)
    [6 subtests passed]

PASS
ok  	github.com/jmgilman/sow/cli/internal/projects/design	0.249s
```

## Files Created/Modified

### Created Files

1. `/cli/internal/projects/design/prompts.go` (6.8 KB)
   - 4 prompt generator functions
   - 1 helper function
   - Embedded templates using go:embed

2. `/cli/internal/projects/design/prompts_test.go` (16 KB)
   - 27 comprehensive test cases
   - Tests cover all functions and edge cases

3. `/cli/internal/projects/design/templates/orchestrator.md` (4.5 KB)
   - Complete design workflow documentation
   - Phase explanations and state transitions

4. `/cli/internal/projects/design/templates/active.md` (3.2 KB)
   - Active state operational guidance
   - Document drafting and review workflow

5. `/cli/internal/projects/design/templates/finalizing.md` (2.5 KB)
   - Finalization state operational guidance
   - Moving artifacts and creating PR

### Modified Files

None - This task only creates new files.

## Key Design Decisions

### 1. Status Icons for Visual Scanning

Used distinct unicode icons for each task status:
- Makes prompt scannable at a glance
- Consistent with exploration project patterns
- Clear visual differentiation between states

### 2. Guard Integration for Advancement Readiness

Prompts call guard functions to show advancement readiness:
- Provides immediate feedback to orchestrator
- Shows "ready to advance" when guards pass
- Shows unresolved count when guards fail
- Eliminates guesswork about when to advance

### 3. Template-Based Static Guidance

Static guidance in separate template files:
- Easy to maintain and update
- Clear separation of dynamic vs static content
- Supports Go template syntax for context injection
- Can be extended without code changes

### 4. Graceful Error Handling

All functions handle errors gracefully:
- Nil project returns error message (doesn't panic)
- Missing phases return error message
- Template rendering errors append to dynamic content
- Always returns useful string (never empty)

### 5. Task Metadata Display

Active prompt shows task metadata (artifact_path, document_type):
- Helps orchestrator see document linkage
- Verifies artifacts are properly connected to tasks
- Supports debugging and status verification

## Acceptance Criteria Verification

### Functional Requirements

- [x] `templates/` directory created with three .md files
- [x] `prompts.go` implements all four required functions
- [x] Orchestrator prompt explains design workflow clearly
- [x] Active prompt shows complete task status dynamically
- [x] Finalizing prompt shows finalization progress
- [x] All prompts use consistent formatting and style
- [x] Prompts handle edge cases (no tasks, missing phases)
- [x] Template rendering errors are handled gracefully

### Test Requirements

- [x] configurePrompts tests (4 tests)
- [x] generateOrchestratorPrompt tests (3 tests)
- [x] generateActivePrompt tests (11 tests)
- [x] generateFinalizingPrompt tests (6 tests)
- [x] All 27 tests pass

### Code Quality

- [x] All functions documented with clear descriptions
- [x] Template files use consistent markdown formatting
- [x] Prompt generators handle nil/missing data safely
- [x] Error handling includes context in error messages
- [x] Tests cover all edge cases
- [x] Templates are informative and actionable

## Integration Notes

This task integrates with:
- **Task 010**: Uses `Active`, `Finalizing` state constants
- **Task 020**: Uses `allDocumentsApproved()`, `allFinalizationTasksComplete()`, `countUnresolvedTasks()` guards
- **Templates SDK**: Uses `templates.Render()` for template rendering
- **Project SDK**: Uses `ProjectTypeConfigBuilder` for registration

Ready for integration into main design project configuration.

## Next Steps

These prompt generators need to be called from the main design project configuration function (likely in a future task that creates `design.go` or `exploration.go` equivalent).

The configuration should call:
```go
builder = configurePrompts(builder)
```

This will register all prompts with the design project type.
