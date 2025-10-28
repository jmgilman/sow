# Interactive Project Launch Design

**Author**: Architecture Team
**Date**: 2025-10-28
**Status**: Proposed
**Related Exploration**: Findings incorporated from interactive UX exploration (not separately documented)

## Overview

This design implements an interactive wizard for `sow project` command to eliminate edge cases and guide users through project creation/resumption. The current flag-based approach has error-prone scenarios that confuse users. The wizard detects repository state, validates transitions, and provides context-aware prompts for all users.

The exploration research (January 2025) validated this approach through comprehensive analysis of repository states, terminal UI library evaluation, and decision tree design.

## Goals and Non-Goals

**Goals**:
- Eliminate 5 edge cases in current flag-based implementation
- Provide interactive wizard for project creation and resumption
- Validate repository state at each decision point
- Guide users with context-aware prompts
- Handle all repository states (protected branch, feature branch with/without project)
- Simplify UX by removing flag complexity

**Non-Goals** (explicitly out of scope):
- Changing underlying project structure or state management
- Modifying project type schemas or state machines
- Supporting flags (deferred - can reassess if truly needed later)
- Custom themes or extensive UI customization (use library defaults)

## Design

### Current Problems

The flag-based `sow project` command has five error-prone edge cases:

1. **Feature branch creation from wrong location**: Creating new branch via `--branch` requires being on protected branch first, otherwise error
2. **Issue with linked branch but no project**: Unexpected state causes error instead of recovery path
3. **Multiple linked branches**: No user control over which branch is selected
4. **Protected branch blocks creation**: No guidance on next steps when attempting project creation on main/master
5. **Uncommitted changes**: Not validated before branch switching, could lose work

These create frustrating user experiences where valid workflows are blocked with unclear error messages.

### Implementation Strategy

Replace flag-based approach with interactive wizard flow that:
1. Detects current repository state
2. Routes to appropriate flow based on context
3. Provides numbered prompts with clear options
4. Validates state at each step

Key architectural decisions:
- **No flags**: Wizard-only interface simplifies implementation and UX. Flags can be reconsidered later if CI/automation needs arise.
- **Lightweight state machine**: Uses simple state variable with conditional logic (not full `stateless` library). Provides structure for testing while remaining lightweight. See Implementation Notes for details.

### Complete Decision Tree

```
┌────────────────────────────────────────────────────────────────┐
│ User runs: sow project                                         │
└────────────────────────────────────────────────────────────────┘
                           │
                           ▼
                           ▼
                          ┌──────────────────────────────────────┐
                          │ STEP 1: Detect repository state      │
                          │                                      │
                          │ Check:                               │
                          │ - Current branch name                │
                          │ - Is protected? (main/master)        │
                          │ - .sow/project/ exists?              │
                          │ - Working tree clean?                │
                          │ - Linked GitHub issue?               │
                          └──────────────────────────────────────┘
                                         │
                      ┌──────────────────┼──────────────────┐
                      │                  │                  │
            [Protected branch]  [Feature + project]  [Feature + no project]
                      │                  │                  │
                      ▼                  ▼                  ▼
          ┌───────────────┐   ┌──────────────┐   ┌──────────────┐
          │ STATE 1       │   │ STATE 2      │   │ STATE 3      │
          │ Protected     │   │ Existing     │   │ No project   │
          │ branch        │   │ project      │   │ on branch    │
          └───────────────┘   └──────────────┘   └──────────────┘
                      │                  │                  │
                      └──────────────────┴──────────────────┘
                                         │
                                         ▼

════════════════════════════════════════════════════════════════
STATE 1: PROTECTED BRANCH FLOW
════════════════════════════════════════════════════════════════

┌────────────────────────────────────────────────────────────────┐
│ Display context:                                               │
│ "You're on '[branch]' (protected branch).                     │
│  Projects must be created on feature branches."               │
└────────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────┐
│ Prompt: "What would you like to do?"                          │
│   1. Create new project on new branch                         │
│   2. Switch to existing feature branch                        │
│   3. Cancel                                                    │
└────────────────────────────────────────────────────────────────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
       [Create new]  [Switch to]   [Cancel]
              │            │            │
              ▼            ▼            ▼
    ┌─────────────┐  ┌──────────┐   Exit
    │ Go to       │  │ List all │
    │ PROJECT     │  │ feature  │
    │ TYPE        │  │ branches │
    │ SELECTION   │  │          │
    └─────────────┘  │ Filter by│
                     │ prefix:  │
                     │ explore/ │
                     │ design/  │
                     │ breakdown│
                     │ feat/    │
                     │ fix/     │
                     │          │
                     │ Select   │
                     │ branch   │
                     │          │
                     │ Checkout │
                     │ branch   │
                     │          │
                     │ Check if │
                     │ project  │
                     │ exists   │
                     │          │
                     │ If yes:  │
                     │ Resume   │
                     │          │
                     │ If no:   │
                     │ Offer to │
                     │ create   │
                     │ project  │
                     └──────────┘
                           │
                           ▼
              ┌────────────────────────────┐
              │ PROJECT TYPE SELECTION     │
              └────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────┐
│ Prompt: "What type of project?"                               │
│   1. Standard (feature work, can link to GitHub issue)        │
│   2. Exploration (research and investigation)                 │
│   3. Design (architecture and design docs)                    │
│   4. Breakdown (decompose work into tasks)                    │
│   5. Cancel                                                    │
└────────────────────────────────────────────────────────────────┘
                           │
          ┌────────────────┼───────────────┬──────────┬────────┐
          │                │               │          │        │
     [Standard]      [Exploration]    [Design]  [Breakdown] [Cancel]
          │                │               │          │        │
          ▼                └───────────────┴──────────┘        ▼
┌──────────────────┐                     │                   Exit
│ STANDARD PROJECT │                     │
│ SOURCE SELECTION │                     │
└──────────────────┘                     │
          │                              │
          ▼                              │
┌────────────────────────────────────────────────────────────────┐
│ Prompt: "How do you want to create this standard project?"   │
│   1. From GitHub issue (link to existing issue)              │
│   2. From scratch (I'll provide details)                      │
│   3. Cancel                                                    │
└────────────────────────────────────────────────────────────────┘
          │                              │
    ┌─────┴─────┐                        │
    │           │                        │
[From issue] [From scratch]              │
    │           │                        │
    ▼           ▼                        │
┌─────────┐ ┌──────────┐                │
│ List    │ │ Go to    │                │
│ open    │ │ NEW      │                │
│ issues  │ │ BRANCH   │                │
│         │ │ FLOW     │                │
│ Select  │ │          │                │
│ issue   │ └──────────┘                │
│         │      │                      │
│ Detect  │      │                      │
│ or      │      │                      │
│ prompt  │      │                      │
│ branch  │      │                      │
│ name    │      │                      │
│         │      │                      │
│ Create/ │      │                      │
│ checkout│      │                      │
│ branch  │      │                      │
│         │      │                      │
│ Init    │      │                      │
│ standard│      │                      │
│ project │      │                      │
│ with    │      │                      │
│ issue   │      │                      │
│ context │      │                      │
└─────────┘      │                      │
                 │                      │
                 └──────────────────────┘
                           │
                           ▼
                  All types converge to
                  NEW BRANCH FLOW

════════════════════════════════════════════════════════════════
STATE 2: EXISTING PROJECT FLOW
════════════════════════════════════════════════════════════════

┌────────────────────────────────────────────────────────────────┐
│ Load and display project information:                         │
│                                                                 │
│ "Project: [name]                                              │
│  Type: [type]                                                 │
│  Branch: [branch]                                             │
│  Current state: [state]                                       │
│  Progress: [X/Y] tasks completed"                             │
└────────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────────┐
│ Prompt: "What would you like to do?"                          │
│   1. Resume this project                                      │
│   2. Delete project and create new on this branch             │
│   3. Switch to different branch                               │
│   4. Cancel                                                    │
└────────────────────────────────────────────────────────────────┘
                           │
          ┌────────────────┼─────────────────┬───────────┐
          │                │                 │           │
      [Resume]         [Delete]          [Switch]    [Cancel]
          │                │                 │           │
          ▼                ▼                 ▼           ▼
    ┌──────────┐    ┌──────────┐      ┌───────┐      Exit
    │ Launch   │    │ Confirm  │      │ List  │
    │ orches-  │    │ delete   │      │ all   │
    │ trator   │    │          │      │ feature│
    │          │    │ If yes:  │      │ branch│
    │ Load     │    │ Delete   │      │       │
    │ state-   │    │ .sow/    │      │ Filter│
    │ specific │    │ project/ │      │ by    │
    │ prompt   │    │          │      │ prefix│
    │          │    │ Go to    │      │       │
    │ Continue │    │ NEW      │      │ Select│
    │ from     │    │ PROJECT  │      │       │
    │ current  │    │ FLOW     │      │ Check-│
    │ state    │    │          │      │ out   │
    └──────────┘    └──────────┘      │       │
                                      │ Check │
                                      │ if    │
                                      │ proj  │
                                      │ exists│
                                      │       │
                                      │ Resume│
                                      │ or    │
                                      │ offer │
                                      │ create│
                                      └───────┘

════════════════════════════════════════════════════════════════
STATE 3: NEW PROJECT FLOW (No project on current branch)
════════════════════════════════════════════════════════════════

┌────────────────────────────────────────────────────────────────┐
│ Prompt: "Where should the new project be created?"            │
│   1. This branch ([current-branch])                           │
│   2. New branch (I'll provide name)                           │
│   3. Switch to existing feature branch                        │
│   4. Cancel                                                    │
└────────────────────────────────────────────────────────────────┘
                           │
          ┌────────────────┼────────────────┬─────────┐
          │                │                │         │
    [This branch]    [New branch]      [Switch]  [Cancel]
          │                │                │         │
          ▼                ▼                ▼         ▼
    ┌─────────┐      ┌─────────┐      ┌───────┐   Exit
    │ THIS    │      │ Go to   │      │ Same  │
    │ BRANCH  │      │ PROJECT │      │ as    │
    │ PATH    │      │ TYPE    │      │ State │
    │         │      │ SELECT  │      │ 2     │
    └─────────┘      └─────────┘      └───────┘
          │                │
          └────────────────┘
                  │
                  ▼
    ┌───────────────────────────┐
    │ PROJECT TYPE SELECTION    │
    │ (Same as State 1)         │
    └───────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────────────────────┐
│ Prompt: "What type of project?"                               │
│   1. Standard (feature work, can link to GitHub issue)        │
│   2. Exploration (research and investigation)                 │
│   3. Design (architecture and design docs)                    │
│   4. Breakdown (decompose work into tasks)                    │
│   5. Cancel                                                    │
└────────────────────────────────────────────────────────────────┘
                  │
     ┌────────────┼───────────────┬──────────┬────────┐
     │            │               │          │        │
[Standard]  [Exploration]    [Design]  [Breakdown] [Cancel]
     │            │               │          │        │
     ▼            └───────────────┴──────────┘        ▼
┌──────────┐                     │                  Exit
│ STANDARD │                     │
│ SOURCE   │                     │
└──────────┘                     │
     │                           │
     ▼                           │
┌────────────────────────────────────────────────────────────────┐
│ Prompt: "How do you want to create this standard project?"   │
│   1. From GitHub issue (link to existing issue)              │
│   2. From scratch (I'll provide details)                      │
│   3. Cancel                                                    │
└────────────────────────────────────────────────────────────────┘
     │                           │
┌────┴────┐                      │
│         │                      │
[Issue] [Scratch]                │
     │         │                 │
     ▼         ▼                 │
┌─────┐   Continue               │
│Issue│   to THIS                │
│flow │   BRANCH or              │
│same │   NEW BRANCH             │
│as   │   PATH based             │
│State│   on earlier             │
│1    │   choice                 │
└─────┘                          │
     │         │                 │
     └─────────┴─────────────────┘
                  │
                  ▼
        Both paths continue below

────────────────────────────────────────────────────────────────
THIS BRANCH PATH
────────────────────────────────────────────────────────────────

Check if on protected branch
         │
    ┌────┴────┐
    │         │
[Protected] [Feature]
    │         │
    ▼         ▼
┌───────┐  Detect project type from branch prefix
│ ERROR │    │
│ Cannot│    ▼
│ create│  ┌──────────────────────────────────┐
│ on    │  │ Branch prefix → Project type:    │
│ protec│  │ explore/  → exploration          │
│ ted   │  │ design/   → design               │
│ branch│  │ breakdown/→ breakdown            │
│       │  │ Other     → standard             │
│ Restart│ └──────────────────────────────────┘
└───────┘              │
                       ▼
              ┌────────────────────────────┐
              │ Confirm:                   │
              │ "Create [type] project    │
              │  on branch [branch]?"     │
              │                            │
              │ [Yes] [No] [Cancel]        │
              └────────────────────────────┘
                       │
          ┌────────────┼────────────┐
          │            │            │
       [Yes]         [No]       [Cancel]
          │            │            │
          ▼            ▼            ▼
    Initialize    ┌────────┐     Exit
    project       │ Conflict│
                  │ Branch  │
    - Load schema │ has type│
    - Create .sow/│ prefix  │
      project/    │ but user│
    - Init state  │ wants   │
    - Hand off to │ different│
      orchestrator│         │
                  │ Message:│
                  │ "Use    │
                  │ branch  │
                  │ with    │
                  │ correct │
                  │ prefix" │
                  │         │
                  │ Restart │
                  └────────┘

────────────────────────────────────────────────────────────────
NEW BRANCH FLOW
────────────────────────────────────────────────────────────────

Prompt for branch name
  │
  ▼
┌────────────────────────────────────┐
│ "Enter branch name:"               │
│                                    │
│ Hint: Use prefix to indicate type: │
│ - explore/  → exploration          │
│ - design/   → design               │
│ - breakdown/→ breakdown            │
│ - feat/     → standard             │
│                                    │
│ Input: ________________            │
└────────────────────────────────────┘
  │
  ▼
Validate branch name
  │
  ├─ Empty? → Error, retry
  ├─ Protected (main/master)? → Error, retry
  ├─ Contains spaces? → Error, retry
  ├─ Contains '..'? → Error, retry
  │
  ▼
Detect project type from name
  │
  ▼
┌────────────────────────────────────┐
│ "Create [type] project            │
│  on new branch [branch]?"         │
│                                    │
│ [Yes] [No] [Cancel]                │
└────────────────────────────────────┘
  │
  ├─ [No] → Back to branch name prompt
  ├─ [Cancel] → Exit
  │
  ▼
Check working tree status
  │
  ├─ Uncommitted changes?
  │  │
  │  ▼
  │ ┌────────────────────────────┐
  │ │ WARN:                      │
  │ │ "You have uncommitted      │
  │ │  changes. Switching        │
  │ │  branches may lose work."  │
  │ │                            │
  │ │ Proceed anyway?            │
  │ │ [Yes] [No]                 │
  │ └────────────────────────────┘
  │          │
  │          ├─ [No] → Exit
  │          │
  ▼          ▼
Create and checkout branch: git checkout -b [branch]
  │
  ▼
Check if project exists on new branch (EDGE CASE)
  │
  ├─ [Project exists]
  │  │
  │  ▼
  │ ┌────────────────────────────┐
  │ │ WARN:                      │
  │ │ "Project already exists    │
  │ │  on branch [branch]."      │
  │ │                            │
  │ │ Options:                   │
  │ │ 1. Resume existing project │
  │ │ 2. Choose different branch │
  │ │ 3. Cancel                  │
  │ └────────────────────────────┘
  │          │
  │          ├─ [Resume] → Load project, launch orchestrator
  │          ├─ [Different] → Back to branch name prompt
  │          └─ [Cancel] → Exit
  │
  └─ [No project]
     │
     ▼
Initialize project
  │
  ├─ Load project type schema
  ├─ Create .sow/project/ structure
  ├─ Initialize state.yaml
  ├─ Set initial state
  ├─ Create log entry
  │
  ▼
Hand off to orchestrator
  │
  ├─ Load state-specific prompt
  ├─ Launch orchestrator with context
  │
  ▼
Project workflow begins
```

### Repository State Detection

The wizard must detect three pieces of information to route correctly:

**State Properties**:
- **Current branch name**: Used for type detection and protected branch check
- **Is protected branch**: `main` or `master` (cannot create projects here)
- **Project exists**: `.sow/project/state.yaml` file exists
- **Working tree clean**: No uncommitted changes
- **Linked GitHub issue**: Branch linked to issue (optional, for standard project GitHub flow)

**Three Primary States**:
1. **Protected branch** (main/master) - Cannot create projects, must navigate to feature branch
2. **Feature branch with project** - Can resume, delete, or switch to different branch
3. **Feature branch without project** - Can create on this branch or create on new branch

### Project Type Selection

When creating a new project, user selects type first:

**Type selection applies to both State 1 (protected branch) and State 3 (feature branch without project)**:
1. **Standard** - Feature work, bug fixes. Can optionally link to GitHub issue.
2. **Exploration** - Research and investigation. No GitHub issue linking.
3. **Design** - Architecture and design docs. No GitHub issue linking.
4. **Breakdown** - Decompose work into tasks. No GitHub issue linking.

**GitHub issue integration**: Only available for standard projects. After selecting "Standard", wizard asks:
- "From GitHub issue" → Lists open issues, creates/switches to branch, initializes with issue context
- "From scratch" → Prompts for branch name, creates branch, initializes without issue context

**Non-standard types** skip GitHub issue question entirely and proceed directly to branch creation.

### Branch Discovery

**How "switch to existing branch" works:**

The wizard lists all available feature branches (excluding main/master) filtered by conventional prefixes:
- `explore/*`
- `design/*`
- `breakdown/*`
- `feat/*`
- `fix/*`
- Other non-protected branches

**Branch listing approach**:
1. Run `git branch` to get all local branches
2. Filter out protected branches (main, master)
3. Present filtered list to user for selection
4. After user selects branch, checkout and check if `.sow/project/` exists
5. If project exists: resume project
6. If no project: offer to create project on that branch

**Important**: The wizard does NOT detect which branches have existing projects upfront. It only lists branches by prefix. Project existence is checked after user selects a branch. This keeps implementation simple and avoids needing a structured project discovery mechanism.

**Future enhancement**: A more sophisticated project discovery system could scan branches for `.sow/project/state.yaml` and show project metadata (name, type, state) before selection. This is deferred as future work.

### Project Type Detection

Branch prefix automatically determines project type (convention-based):

| Branch Prefix | Project Type | Initial State |
|---------------|-------------|---------------|
| `explore/` | exploration | Gathering |
| `design/` | design | Drafting |
| `breakdown/` | breakdown | Decomposing |
| Other (feat/, fix/, etc.) | standard | Planning |

No manual type selection needed - convention drives behavior.

### Validation Points

The wizard validates state at each transition:

**Before creating branch**:
- Working tree must be clean (or user acknowledges risk)
- Branch name must be valid (no spaces, no `..`, not protected name)

**Before creating project**:
- Must not be on protected branch
- Project must not already exist on target branch
- Branch must exist (if switching to existing)

**Before resuming project**:
- Project state.yaml must be readable and valid
- Current branch must match project.branch

### Command Usage

Single command, no flags:

```bash
sow project
# Always enters interactive wizard
# Detects repository state and guides user through appropriate flow
```

**Rationale for no flags**: Parsing flags and managing all possible combinations adds significant complexity. The wizard provides better UX by detecting context and validating state. Flags can be reconsidered later if automation/CI needs arise, but are deferred for initial implementation.

### User Experience Principles

**Context before choices**: Always display current state before asking what to do. Show project info, branch name, progress.

**Numbered prompts**: All prompts provide numbered options (1, 2, 3...) for clarity.

**Cancel anywhere**: User can cancel at any point (Ctrl+C or explicit Cancel option).

**Actionable errors**: Error messages explain why operation blocked and what to do next.

**Validation before action**: Check preconditions before attempting any operation. Don't fail after partial work.

**Clear consequences**: Warn user about destructive actions (uncommitted changes, delete project) before executing.

## Implementation Notes

**Library choice**: Use `charmbracelet/huh` for terminal UI
- Only library with external editor support (needed for multi-line initial prompts)
- Dynamic forms with conditional fields
- Built-in validation with inline error display
- Simple integration with existing Cobra CLI structure
- Active maintenance, part of Charmbracelet ecosystem

**Wizard state machine**: Use lightweight custom state machine (not `stateless` library)
- Wizard uses simple state variable with conditional logic, not the project state machine
- State variable tracks current position in decision tree for testing/debugging
- Each state handled by dedicated function that shows forms and transitions
- More structured than pure conditionals, lighter than full `stateless` state machine

**Implementation structure**:
```go
// cli/cmd/project.go
type wizardState string

const (
    stateDetect        wizardState = "detect"
    stateProtected     wizardState = "protected_branch"
    stateExisting      wizardState = "existing_project"
    stateNewProject    wizardState = "new_project"
    stateTypeSelect    wizardState = "type_selection"
    stateBranchCreate  wizardState = "branch_create"
    stateBranchSelect  wizardState = "branch_select"
    stateComplete      wizardState = "complete"
    stateCancelled     wizardState = "cancelled"
)

type wizard struct {
    state   wizardState
    ctx     *sow.Context
    choices map[string]interface{} // Accumulated user choices
}

func (w *wizard) run() error {
    // Loop until complete or cancelled
    for w.state != stateComplete && w.state != stateCancelled {
        if err := w.handleState(); err != nil {
            return err
        }
    }

    if w.state == stateCancelled {
        return nil // User cancelled, no error
    }

    // Launch orchestrator with state-specific prompt
    return w.launchOrchestrator()
}

func (w *wizard) handleState() error {
    switch w.state {
    case stateDetect:
        w.state = w.detectRepositoryState()
    case stateProtected:
        return w.handleProtectedBranch()
    case stateExisting:
        return w.handleExistingProject()
    case stateNewProject:
        return w.handleNewProject()
    case stateTypeSelect:
        return w.handleTypeSelection()
    case stateBranchCreate:
        return w.handleBranchCreate()
    case stateBranchSelect:
        return w.handleBranchSelect()
    }
    return nil
}

func (w *wizard) handleProtectedBranch() error {
    // Show huh form for protected branch options
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("What would you like to do?").
                Options(
                    huh.NewOption("Create new project", "create"),
                    huh.NewOption("Switch to existing branch", "switch"),
                    huh.NewOption("Cancel", "cancel"),
                ).
                Value(&w.choices["action"]),
        ),
    )

    if err := form.Run(); err != nil {
        return err
    }

    // Transition based on choice
    switch w.choices["action"].(string) {
    case "create":
        w.state = stateTypeSelect
    case "switch":
        w.state = stateBranchSelect
    case "cancel":
        w.state = stateCancelled
    }

    return nil
}

// Additional handlers follow same pattern...
```

**Testing approach with state machine**:
```go
// Mock form interactions by directly setting wizard state and choices
func TestWizardProtectedBranchFlow(t *testing.T) {
    w := &wizard{
        state:   stateProtected,
        ctx:     testContext(),
        choices: make(map[string]interface{}),
    }

    // Simulate user choosing "create"
    w.choices["action"] = "create"

    // Execute state handler
    err := w.handleProtectedBranch()
    require.NoError(t, err)

    // Verify transition
    assert.Equal(t, stateTypeSelect, w.state)
}
```

**Key benefits**:
- State variable provides clear tracking of wizard position
- Each state handler testable independently without interactive forms
- Transition logic explicit and testable
- Natural fit with `huh` form sequences
- Simpler than full `stateless` state machine, more structured than pure conditionals

**Testing strategy**:
- Mock terminal I/O using Bubble Tea program options
- Table-driven tests for each state handler
- Test all edge cases (protected branch, existing project, uncommitted changes)
- Verify all paths through decision tree

## Testing Approach

**Unit tests**:
- Repository state detection (various branch configurations)
- Project type detection from branch prefixes
- Branch name validation (valid, invalid, protected)
- State routing logic

**Integration tests**:
- Full wizard flow from detection through project initialization
- Edge case handling (all scenarios in decision tree)
- Resume project workflow
- Working tree validation
- All three primary state flows

**Manual verification**:
- Test on actual repository with various states
- Verify terminal UI renders correctly
- Test keyboard navigation and cancellation
- Validate error messages are clear

## Alternatives Considered

### Option 1: Keep Flag-Only Interface

**Description**: Maintain current flag-based approach, improve error messages only.

**Pros**:
- No new dependencies
- Simpler implementation
- Familiar for power users

**Cons**:
- Doesn't solve edge case problems
- Error messages still require state understanding
- No validation before operations
- Steeper learning curve

**Why not chosen**: Doesn't address root cause (lack of state awareness). Better errors help but don't eliminate confusion.

### Option 2: State Machine-Based Wizard

**Description**: Implement wizard using formal state machine.

**Pros**:
- Formal flow modeling
- Clear state transitions
- Testable guards

**Cons**:
- Overkill for linear wizard
- Significant boilerplate
- Terminal UI library already provides flow control

**Why not chosen**: State machine complexity not justified. Conditional logic maps naturally to decision tree.

### Option 3: Full TUI with Bubble Tea

**Description**: Build full terminal UI using Bubble Tea MVU framework.

**Pros**:
- Maximum customization
- Rich interactions

**Cons**:
- Massive implementation overhead (MVU architecture)
- Overkill for wizard-style forms
- `huh` provides simpler API

**Why not chosen**: Complexity exceeds requirements. `huh` provides exactly what's needed without MVU boilerplate.

## References

- **Exploration findings**: [Interactive UX for Launch Commands](../knowledge/explorations/interactive-ux-2025-01.md) - Comprehensive research documenting edge cases, decision trees, library evaluation
- **charmbracelet/huh**: [GitHub Repository](https://github.com/charmbracelet/huh) - Terminal UI library
- **Related design**: `project-modes-design.md` - Defines project type schemas wizard creates

## Future Considerations

**Enhanced GitHub issue integration**: More sophisticated issue workflows (filter by label, assignee, milestone).

**Configuration-based customization**: Custom branch prefix mappings, editor choice, theme customization.

**Multi-repository support**: Detect linked repos, offer to create projects in linked repos, sync state across repos.

**Project archiving**: Archive completed projects for historical reference before deleting.
