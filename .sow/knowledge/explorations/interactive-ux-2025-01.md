# Interactive UX for Launch Commands Exploration Summary

**Date:** January 2025
**Branch:** explore/interactive-ux
**Status:** Research complete

## Context

This exploration investigated improving the user experience for launching and continuing projects in sow. The current `sow project` command relies on flags with edge cases that can cause errors. The goal: design an interactive CLI flow that guides users through valid options while preserving power-user flag-based workflows.

## What We Researched

- Current repository state detection logic and edge cases
- Interactive CLI flow design and decision trees
- Terminal UI library options (huh, promptui, survey, bubbletea)
- State machine vs conditional logic for flow implementation
- Integration patterns with existing Cobra command structure

## Key Findings

### Current Implementation Has Five Edge Cases

From analyzing `cli/cmd/project.go`, the flag-based command has error-prone scenarios:

1. **Feature branch creation from wrong location**: Creating new branch via `--branch` requires being on protected branch first, otherwise error
2. **Issue with linked branch but no project**: Unexpected state causes error instead of recovery path
3. **Multiple linked branches**: No user control over which branch is selected
4. **Protected branch blocks creation**: No guidance on next steps when attempting project creation on main/master
5. **Uncommitted changes**: Not validated before branch switching, could lose work

### Repository State Space Defines Three Major Flows

Interactive flow must handle three primary repository states:

**1. On protected branch (main/master)**
- No project should exist (state corruption if it does)
- User needs to: create new project on new branch, switch to existing branch, or work on issue

**2. On feature branch with project**
- User needs to: continue existing project, delete and start new, or switch branches
- Display project context (name, phase, task progress) before decision

**3. On feature branch without project**
- User needs to: create project here, switch to different branch, or create on new branch

GitHub issue flow integrates as parallel subtree, handling linked branches and issue selection.

### Decision Trees Map All Valid Paths

Comprehensive decision trees documented covering:
- Entry point based on repository state detection
- Branch decision logic for protected vs feature branches
- Project continuation vs creation flows
- GitHub issue integration (listing, selection, branch linking)
- Validation points (working tree clean, branch exists, not on protected branch)

All paths preserve flag-based power-user workflows - flags skip corresponding prompts.

### Conditional Logic Better Than State Machine

Analysis of state machine vs conditional logic concluded:

**State machine characteristics that don't fit:**
- CLI flow is linear/wizard-style, not complex cyclic behavior
- No state persistence needed (single execution completes flow)
- Terminal UI libraries already handle flow control naturally
- State machine setup overhead outweighs benefits

**Conditional logic advantages:**
- Direct mapping from decision tree to code
- Natural integration with `huh` forms that return values
- Less boilerplate than configuring states/events/guards
- Easier to modify (add prompt = add conditional)
- Testable via well-structured functions with mocked inputs

**Recommendation**: Use conditional logic with structured helper functions. Optional hybrid approach: state machine for pure state detection (repo state, branch type, project existence), conditional logic for user interaction.

### charmbracelet/huh Provides Ideal Terminal UI Solution

Evaluated four Go terminal UI libraries:

**charmbracelet/huh** (recommended):
- ✅ External editor integration via Text field (critical requirement)
- ✅ Dynamic forms with OptionsFunc/TitleFunc for conditional flows
- ✅ Complete field types: Input, Text, Select, MultiSelect, Confirm, FilePicker, Note
- ✅ Built-in validation with inline error display
- ✅ Simple Cobra integration (no architectural changes)
- ✅ Testing support via Bubble Tea program options
- ✅ Accessibility mode for screen readers
- ✅ Active maintenance, part of Charmbracelet ecosystem

**manifoldco/promptui** (not suitable):
- ❌ No multi-line text input (blocker)
- ❌ No editor integration (blocker)
- ❌ No form abstraction (must chain prompts manually)
- ❌ 73 open issues suggesting maintenance burden

**AlecAivazis/survey** (not recommended):
- ⚠️ Archived April 2024 (read-only)
- ⚠️ Maintainer recommends Charmbracelet Bubbletea/huh as successor
- Had comprehensive features including editor support (historically)

**Bubbletea** (too heavy):
- Full MVU (Model-View-Update) framework for complex TUIs
- Overkill for wizard-style forms (requires significant boilerplate)
- Huh uses Bubbletea internally, providing simpler API for forms

### Implementation Pattern Preserves Existing Structure

New structure integrates cleanly:

```
cli/cmd/project.go
  └─ runProject() - Dispatch based on flags
       ├─ Flags provided → runFlagBasedProject() [existing logic]
       └─ No/incomplete flags → runInteractiveProject() [new]

cli/internal/modes/interactive_project.go [NEW]
  ├─ InteractiveProjectFlow struct
  ├─ Run() - Entry point, detect state, route to handler
  ├─ protectedBranchFlow() - Handle protected branch scenarios
  ├─ existingProjectFlow() - Continue/delete/switch options
  ├─ newProjectFlow() - Create project on current/new branch
  └─ issueFlow() - GitHub issue selection and linking
```

Power users keep full flag-based control:
```bash
sow project --type standard --branch feat/auth --prompt "Add JWT"
```

New users get guided interactive flow with state context and validation.

## Implementation Highlights

### Dynamic Form Example

Huh's OptionsFunc/TitleFunc enable conditional prompts:

```go
var projectType string

form := huh.NewForm(
    // First prompt: select type
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("What type of project?").
            Options(
                huh.NewOption("Standard", "standard"),
                huh.NewOption("Exploration", "exploration"),
                huh.NewOption("Design", "design"),
                huh.NewOption("Breakdown", "breakdown"),
            ).
            Value(&projectType),
    ),
    // Second prompt: branch name (title reflects selected type)
    huh.NewGroup(
        huh.NewInput().
            TitleFunc(func() string {
                return fmt.Sprintf("Branch name for %s project?", projectType)
            }, &projectType).
            Validate(validateBranchName).
            Value(&branchName),
    ),
)
```

### External Editor Integration

Text field spawns editor for multi-line input:

```go
huh.NewText().
    Title("Initial prompt (Ctrl+E to open editor)").
    Editor("vim"). // Or detect from $EDITOR
    CharLimit(2000).
    Value(&initialPrompt)
```

Solves multi-line initial prompt entry without custom implementation.

### Testing Strategy

Forms support mocked I/O for unit testing:

```go
input := bytes.NewBufferString("standard\nfeat/auth\nAdd JWT\n")
output := &bytes.Buffer{}

form.WithProgramOptions(
    tea.WithInput(input),
    tea.WithOutput(output),
).Run()

assert.Equal(t, "feat/auth", ctx.Git().CurrentBranch())
assert.True(t, projectExists(ctx))
```

## Design Decisions

**Decision 1: Use conditional logic instead of state machine**
- **Rationale**: Linear wizard flow doesn't need state machine complexity. Terminal UI forms already provide flow control. No state persistence required.
- **Alternative considered**: Hybrid approach with state machine for state detection only
- **Impact**: Simpler code, easier maintenance, natural huh integration

**Decision 2: Use charmbracelet/huh for terminal UI**
- **Rationale**: Only library with external editor support (critical requirement), dynamic forms, and clean Cobra integration. Active maintenance, accessibility support.
- **Alternatives considered**: promptui (missing features), survey (archived), bubbletea (overkill)
- **Impact**: Adds Bubbletea dependency, increases binary size slightly, but provides complete solution

**Decision 3: Preserve flag-based workflow**
- **Rationale**: Power users need non-interactive option. Flags bypass prompts for CI/automation.
- **Implementation**: Dispatch logic checks flags first, falls back to interactive if incomplete
- **Impact**: Maintains backward compatibility, no breaking changes

**Decision 4: Early project type selection**
- **Rationale**: Project type affects subsequent prompts (branch name suggestions, validation rules, initial prompt templates)
- **Placement**: After repository state detection, before branch/prompt questions
- **Impact**: Enables type-specific guidance throughout flow

## Architecture Benefits

1. **Eliminates edge cases**: Interactive flow provides explicit paths for all scenarios (protected branch, missing projects, multiple linked branches)

2. **Validates state transitions**: Check working tree clean, branch existence, permissions before any action

3. **Provides context**: Display current state (project info, branch, phase) before user decisions

4. **Guides users**: Clear prompts with filtered/searchable options remove guesswork

5. **Preserves flexibility**: Flags still work for scripting and power users

6. **Testable**: Mocked inputs enable unit testing of full flows

7. **Accessible**: Screen reader mode supports visually impaired users

## Open Questions

- [ ] Should editor choice respect project configuration (e.g., `.editorconfig`)?
- [ ] How to handle Ctrl+C during form flow (graceful exit vs error)?
- [ ] Should form theme be configurable via sow config?
- [ ] Need custom key bindings for common actions?
- [ ] When consolidating modes to projects, does each project type need custom prompts?

## Artifacts Created

1. **cli-flow-state-detection.md** - Repository state detection logic, comprehensive decision trees covering all scenarios (protected/feature branch, project exists/missing, GitHub issues), validation points, and flag preservation strategy

2. **cli-flow-state-machine-analysis.md** - Analysis of state machine vs conditional logic for CLI flow implementation, comparing complexity, testability, maintenance, with recommendation for conditional approach

3. **terminal-ui-library-evaluation.md** - Evaluation of four Go terminal UI libraries (huh, promptui, survey, bubbletea) with feature comparison matrix, pros/cons, implementation patterns, and testing examples

## Next Steps

Implementation phase would involve:

1. **Add huh dependency**: `go get github.com/charmbracelet/huh/v2`

2. **Create interactive flow package**: `cli/internal/modes/interactive_project.go` with flow struct and handler functions

3. **Update project command**: Add dispatch logic to `cli/cmd/project.go` routing to interactive flow when flags incomplete

4. **Implement validation helpers**: Branch name validation, working tree checks, GitHub CLI availability

5. **Add unit tests**: Table-driven tests for each handler function with mocked terminal I/O

6. **Update documentation**: Add interactive flow examples to README and user guide

## Recommendation

**Proceed with interactive CLI flow implementation using charmbracelet/huh and conditional logic.**

This approach delivers significant UX improvements (eliminates edge cases, provides guidance, validates state) while maintaining backward compatibility (flags still work) and keeping implementation complexity reasonable (no state machine overhead).

The research identified a clear technical path forward with proven libraries and testable patterns.

## Participants

**Conducted:** January 28, 2025
**Participants:** Josh Gilman, Claude
