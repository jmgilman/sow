# Interactive Wizard UX Flow

**Document Type**: UX Design Specification
**Task**: 010 (Part 1 of 2)
**Author**: Architecture Team
**Date**: 2025-01-06
**Status**: Draft
**Related**: [Technical Implementation](./wizard-technical-implementation.md), VISION.md, DIFF.md

## Overview

This document specifies the complete user experience flow for the interactive project wizard in `sow`. The wizard replaces the current flag-based interface with a guided, interactive experience that leverages the worktree architecture to simplify project creation and continuation.

**Companion document**: [Technical Implementation](./wizard-technical-implementation.md) covers implementation details, code patterns, and the huh library integration.

### Goals

1. **Simplify project management**: Single command (`sow project`) for all project operations
2. **Eliminate confusion**: No ambiguity about branches, states, or operations
3. **Provide clear guidance**: Interactive prompts with validation at each step
4. **Support both creation modes**: GitHub issues and manual project creation
5. **Enable continuation**: Easy discovery and resumption of existing projects

### Non-Goals

- Backward compatibility with flags (replaced entirely)
- Advanced project discovery (remote projects, cross-repository)
- Custom branch naming patterns (convention-based only)
- Theme customization (use library defaults)

### Key Architectural Insight

**Current branch is irrelevant.** The wizard operates entirely on target branches (where worktrees are created). Users can run `sow project` from any branch (main, feature branches, anywhere) and the experience is identical. This dramatically simplifies the UX and eliminates entire categories of edge cases.

---

## Complete Wizard Flow

### Entry Point

**Command**: `sow project`

**Single entry screen** - no state detection complexity:

```
╔══════════════════════════════════════════════════════════╗
║                     Sow Project Manager                  ║
╚══════════════════════════════════════════════════════════╝

What would you like to do?

  ○ Create new project
  ○ Continue existing project
  ○ Cancel

[Use arrow keys to navigate, Enter to select]
```

---

## Path 1: Create New Project

### Screen 1.1: Project Source Selection

After user selects "Create new project":

```
╔══════════════════════════════════════════════════════════╗
║                   Create New Project                     ║
╚══════════════════════════════════════════════════════════╝

How do you want to create this project?

  ○ From GitHub issue
  ○ From branch name
  ○ Cancel

[Use arrow keys to navigate, Enter to select]
```

---

### Path 1A: From GitHub Issue

#### Screen 1A.1: Issue Selection

```
╔══════════════════════════════════════════════════════════╗
║                   Select GitHub Issue                    ║
╚══════════════════════════════════════════════════════════╝

Select an issue (filtered by 'sow' label):

  ○ #123: Add JWT authentication
  ○ #124: Refactor database schema
  ○ #125: Implement rate limiting
  ○ #126: Add logging middleware
  ○ Cancel

[Use arrow keys to navigate, Enter to select]
[Issues fetched via: gh issue list --label sow]
```

**On selection**:
- Check if issue has linked branch
- If linked branch exists: Show error (see Error Messages section)
- If no linked branch: Proceed to type selection

#### Screen 1A.2: Project Type Selection

```
╔══════════════════════════════════════════════════════════╗
║                   Select Project Type                    ║
╚══════════════════════════════════════════════════════════╝

Issue: #123 - Add JWT authentication

What type of project?

  ○ Standard - Feature work and bug fixes
  ○ Exploration - Research and investigation
  ○ Design - Architecture and design documents
  ○ Breakdown - Decompose work into tasks
  ○ Cancel

[Use arrow keys to navigate, Enter to select]
```

**Type Configuration**:
| Type | Prefix | Description |
|------|--------|-------------|
| Standard | `feat/` | Feature work and bug fixes |
| Exploration | `explore/` | Research and investigation |
| Design | `design/` | Architecture and design documents |
| Breakdown | `breakdown/` | Decompose work into tasks |

**On selection**:
- Branch name created via `gh issue develop <issue> --name <prefix><issue-slug>-<number>`
- Example: Issue "Add JWT authentication" + type Standard → `feat/add-jwt-authentication-123`
- Proceed to prompt entry

#### Screen 1A.3: Initial Prompt Entry

```
╔══════════════════════════════════════════════════════════╗
║                    Initial Prompt                        ║
╚══════════════════════════════════════════════════════════╝

Issue: #123 - Add JWT authentication
Branch: feat/add-jwt-authentication-123

Enter your task or question for Claude (optional):
Press Ctrl+E to open $EDITOR for multi-line input

┌────────────────────────────────────────────────────────┐
│                                                        │
│                                                        │
│                                                        │
│                                                        │
│                                                        │
└────────────────────────────────────────────────────────┘

[Enter to submit, Ctrl+E for editor, Esc to skip]
```

**Notes**:
- Large text area for multi-line input
- Support $EDITOR invocation via Ctrl+E
- Optional - user can skip
- Example: "Focus on middleware implementation and integration tests"

**On submission**: Proceed to finalization

---

### Path 1B: From Branch Name

#### Screen 1B.1: Project Type Selection

```
╔══════════════════════════════════════════════════════════╗
║                   Select Project Type                    ║
╚══════════════════════════════════════════════════════════╝

What type of project?

  ○ Standard - Feature work and bug fixes
  ○ Exploration - Research and investigation
  ○ Design - Architecture and design documents
  ○ Breakdown - Decompose work into tasks
  ○ Cancel

[Use arrow keys to navigate, Enter to select]
```

#### Screen 1B.2: Project Name Entry

```
╔══════════════════════════════════════════════════════════╗
║                     Project Name                         ║
╚══════════════════════════════════════════════════════════╝

Type: Exploration

Enter project name:

┌────────────────────────────────────────────────────────┐
│ Web Based Agents                                       │
└────────────────────────────────────────────────────────┘

Preview: explore/web-based-agents

[Enter to continue, Esc to go back]
```

**Name Normalization**:
1. Convert to lowercase: "Web Based Agents" → "web based agents"
2. Replace spaces with hyphens: "web based agents" → "web-based-agents"
3. Remove invalid git characters: Only allow `a-z`, `0-9`, `-`, `_`
4. Remove leading/trailing hyphens
5. Collapse multiple consecutive hyphens

**Full Branch Name**: `<prefix>/<normalized-name>`
- Example: `explore/web-based-agents`

**Preview**: Show computed branch name in real-time as user types

**Validation** (on submit):
- Not empty or whitespace
- Not a protected branch name (main, master)
- Valid git branch name after normalization

**Validation Errors** (inline):
```
┌────────────────────────────────────────────────────────┐
│ main                                                   │
└────────────────────────────────────────────────────────┘

Preview: feat/main
⚠ Cannot use protected branch name
```

#### Screen 1B.3: Branch State Check

After user submits name, validate branch state:
- Does branch exist?
- Does worktree exist for branch?
- Does worktree have project?

**If branch already has project**: Show error (see Error Messages section)
**If branch available**: Proceed to prompt entry

#### Screen 1B.4: Initial Prompt Entry

```
╔══════════════════════════════════════════════════════════╗
║                    Initial Prompt                        ║
╚══════════════════════════════════════════════════════════╝

Type: Exploration
Branch: explore/web-based-agents

Enter your task or question for Claude (optional):
Press Ctrl+E to open $EDITOR for multi-line input

┌────────────────────────────────────────────────────────┐
│ Research the landscape of web-based agent frameworks   │
│ and compare their architectures. Focus on:             │
│ - Multi-agent coordination                             │
│ - Browser automation capabilities                      │
│ - Integration patterns                                 │
└────────────────────────────────────────────────────────┘

[Enter to submit, Ctrl+E for editor, Esc to skip]
```

**On submission**: Proceed to finalization

---

### Finalization Summary

Both creation paths complete with these steps:

1. **Conditional uncommitted changes check** (if current branch == target branch)
2. **Create/ensure branch and worktree exist**
3. **Initialize project** with type and optional issue context
4. **Generate 3-layer prompt** (orchestrator + project type + initial state + user prompt)
5. **Launch Claude** in worktree directory

**Success message**:
```
✓ Initialized project '<name>' on branch <branch>
✓ Launching Claude in worktree...
```

---

## Path 2: Continue Existing Project

### Screen 2.1: Project Selection

```
╔══════════════════════════════════════════════════════════╗
║                  Continue Existing Project               ║
╚══════════════════════════════════════════════════════════╝

Select a project to continue:

  ○ feat/auth - Add JWT authentication
    [Standard: implementation, 3/5 tasks completed]

  ○ design/cli-ux - CLI UX improvements
    [Design: active, 1/3 tasks completed]

  ○ explore/web-agents - Web Based Agents
    [Exploration: gathering, 4/7 tasks completed]

  ○ Cancel

[Use arrow keys to navigate, Enter to select]
```

**Project Discovery**:
- List directories in `.sow/worktrees/`
- Read `.sow/project/state.yaml` for each
- Display: `<branch> - <project-name>` with progress

**Progress Info Format**:
- Format: `<Type>: <phase>[, x/y tasks completed]`
- Tasks portion only shown if phase has tasks

**On selection**:
- Validate project still exists
- If missing: Show error (see Error Messages section)
- If valid: Proceed to prompt entry

### Screen 2.2: Continuation Prompt Entry

```
╔══════════════════════════════════════════════════════════╗
║                   Continue Project                       ║
╚══════════════════════════════════════════════════════════╝

Project: Add JWT authentication
Branch: feat/auth
State: Standard: implementation, 3/5 tasks completed

What would you like to work on? (optional):
Press Ctrl+E to open $EDITOR for multi-line input

┌────────────────────────────────────────────────────────┐
│ Let's focus on the token refresh logic and write      │
│ integration tests                                      │
└────────────────────────────────────────────────────────┘

[Enter to continue, Ctrl+E for editor, Esc to skip]
```

**On submission**: Proceed to finalization

### Finalization Summary

Continue path completes with:

1. **No uncommitted changes check needed** (worktree already exists)
2. **Ensure worktree exists** (idempotent)
3. **Load project state** from worktree
4. **Generate continue prompt** (3-layer + user prompt if provided)
5. **Launch Claude** in worktree directory

**Success message**:
```
✓ Continuing project '<name>' on branch <branch>
✓ Launching Claude in worktree...
```

---

## Validation Rules

### Branch Name Validation

**Rules**:
1. Not empty or whitespace-only
2. Not a protected branch (main, master)
3. Valid git ref name:
   - No spaces
   - No `..` sequences
   - No consecutive slashes `//`
   - No leading/trailing slashes
   - Only valid characters: `a-z`, `A-Z`, `0-9`, `-`, `_`, `/`
   - No special sequences: `~`, `^`, `:`, `?`, `*`, `[`

### Uncommitted Changes Check

**When to check**:
- Only when creating a new project
- Only when current branch == target branch
- Never when continuing (worktree already exists)

**Why conditional?**
- Git worktrees can't have same branch checked out twice
- If current == target, `EnsureWorktree()` must switch main repo to master/main
- If current != target, no switching needed, uncommitted changes irrelevant

**What to check**:
- Modified files (M)
- Deleted files (D)
- Staged changes
- **Ignore untracked files** (?) - they don't affect worktree creation

### GitHub Issue Validation

**Requirements**:
- Issue must have `sow` label
- Issue must not have linked branch (for creation)
- `gh` CLI must be available

---

## Error Messages

All error messages follow this pattern:
1. **What went wrong**: Brief description
2. **How to fix**: Specific commands or actions
3. **Next steps**: What user should do now

### Protected Branch Error

```
╔══════════════════════════════════════════════════════════╗
║                         Error                            ║
╚══════════════════════════════════════════════════════════╝

Cannot create project on protected branch 'main'

Projects must be created on feature branches.

Action: Choose a different project name

[Press Enter to retry]
```

### Issue Already Linked

```
╔══════════════════════════════════════════════════════════╗
║                         Error                            ║
╚══════════════════════════════════════════════════════════╝

Issue #123 already has a linked branch: feat/add-jwt-auth

To continue working on this issue:
  Select "Continue existing project" from the main menu

[Press Enter to return to issue list]
```

### Branch Already Has Project

```
╔══════════════════════════════════════════════════════════╗
║                         Error                            ║
╚══════════════════════════════════════════════════════════╝

Branch 'explore/web-agents' already has a project

To continue this project:
  Select "Continue existing project" from the main menu

To create a different project:
  Choose a different project name (currently: "web agents")

[Press Enter to retry name entry]
```

### Uncommitted Changes

```
╔══════════════════════════════════════════════════════════╗
║                         Error                            ║
╚══════════════════════════════════════════════════════════╝

Repository has uncommitted changes

You are currently on branch 'feat/add-jwt-auth-123'.
Creating a worktree requires switching to a different branch first.

To fix:
  Commit: git add . && git commit -m "message"
  Or stash: git stash

[Press Enter to exit wizard]
```

### Inconsistent State

```
╔══════════════════════════════════════════════════════════╗
║                         Error                            ║
╚══════════════════════════════════════════════════════════╝

Worktree exists but project missing

Branch 'feat/xyz' has a worktree at .sow/worktrees/feat/xyz
but no .sow/project/ directory.

To fix:
  1. Remove worktree: git worktree remove feat/xyz
  2. Delete directory: rm -rf .sow/worktrees/feat/xyz
  3. Try creating project again

[Press Enter to return to project list]
```

### GitHub CLI Missing

```
╔══════════════════════════════════════════════════════════╗
║                         Error                            ║
╚══════════════════════════════════════════════════════════╝

GitHub CLI not found

The 'gh' command is required for GitHub issue integration.

To install:
  macOS: brew install gh
  Linux: See https://cli.github.com/

Or select "From branch name" instead.

[Press Enter to return to source selection]
```

---

## Example User Journeys

### Journey 1: Create from GitHub Issue

```
$ sow project

[Entry screen]
→ User selects "Create new project"

[Source selection]
→ User selects "From GitHub issue"

[Issue list]
→ User selects "#124: Refactor database schema"
→ Wizard validates: no linked branch ✅

[Type selection]
→ User selects "Standard"

[Initial prompt]
→ User enters: "Focus on normalizing the users table and adding proper indexes. Start with a migration plan."
→ User presses Enter

[Finalization]
✓ Initialized project 'Refactor database schema' on branch feat/refactor-database-schema-124
✓ Launching Claude in worktree...

[Claude launches in .sow/worktrees/feat/refactor-database-schema-124/]
```

### Journey 2: Create Exploration Project

```
$ sow project

[Entry screen]
→ User selects "Create new project"

[Source selection]
→ User selects "From branch name"

[Type selection]
→ User selects "Exploration"

[Name entry]
→ User types: "Web Based Agents"
→ Preview shows: explore/web-based-agents
→ Wizard validates: branch available ✅

[Initial prompt]
→ User presses Ctrl+E to open editor
→ User writes multi-line prompt in $EDITOR, saves and exits

[Finalization]
✓ Initialized project 'Web Based Agents' on branch explore/web-based-agents
✓ Launching Claude in worktree...
```

### Journey 3: Continue Existing Project

```
$ sow project

[Entry screen]
→ User selects "Continue existing project"

[Project list]
→ List shows:
  - feat/auth - Add JWT authentication [Standard: implementation, 3/5 tasks completed]
  - design/cli-ux - CLI UX improvements [Design: active, 1/3 tasks completed]
  - explore/web-agents - Web Based Agents [Exploration: gathering, 4/7 tasks completed]

→ User selects "design/cli-ux"
→ Wizard validates: project exists ✅

[Continuation prompt]
→ User enters: "Let's work on the terminal UI technology selection ADR document next"
→ User presses Enter

[Finalization]
✓ Continuing project 'CLI UX improvements' on branch design/cli-ux
✓ Launching Claude in worktree...
```

### Journey 4: Error - Issue Already Linked

```
$ sow project

[Entry screen]
→ User selects "Create new project"

[Source selection]
→ User selects "From GitHub issue"

[Issue list]
→ User selects "#123: Add JWT authentication"
→ Wizard checks: issue has linked branch 'feat/auth' ❌

[Error screen]
Error: Issue #123 already has branch 'feat/auth'. Use continue instead.

→ User presses Enter
→ Returns to issue list

→ User presses Esc to cancel
→ Returns to main menu

→ User selects "Continue existing project"
→ Successfully continues the project
```

### Journey 5: Error - Branch Has Project

```
$ sow project

[Entry screen]
→ User selects "Create new project"

[Source selection]
→ User selects "From branch name"

[Type selection]
→ User selects "Exploration"

[Name entry]
→ User types: "web agents"
→ Preview shows: explore/web-agents
→ Wizard checks: branch has existing project ❌

[Error screen]
Error: Branch 'explore/web-agents' already has a project.

→ User presses Enter
→ Returns to name entry with field cleared

→ User types: "web agents v2"
→ Preview shows: explore/web-agents-v2
→ Wizard validates: branch available ✅
→ Continues successfully
```

---

## Design Rationale

### Why Single Entry Point?

- **Discoverability**: User doesn't need to remember separate commands
- **Context switching**: One mental model for all project operations
- **Guidance**: Wizard can route user to correct flow based on intent

### Why Explicit Type Selection?

- **Clarity**: User knows exactly what type of project they're creating
- **Education**: Descriptions explain what each type is for
- **No guessing**: Eliminates confusion from branch prefix inference

### Why Branch Name Preview?

- **Transparency**: User sees exactly what branch will be created
- **Correction**: User can adjust name if preview isn't what they want
- **Convention reinforcement**: Preview shows the prefix convention in action

### Why Optional Prompts?

- **Flexibility**: Quick start for users who want to get right in
- **Context**: Power users can provide detailed initial guidance
- **Continuation**: Makes sense to allow "just continue where I left off"

### Why Conditional Uncommitted Changes Check?

- **Least restrictive**: Only validate when technically necessary
- **Better UX**: Don't block users unnecessarily
- **Clear rationale**: When error does appear, explain why it matters

---

## References

- **[Technical Implementation](./wizard-technical-implementation.md)**: Implementation details and code patterns
- **[huh Library Verification](../../context/huh-library-verification.md)**: Library capability verification
- **VISION.md**: High-level UX vision and design decisions
- **DIFF.md**: Architecture changes from original design
- **worktree.go**: Worktree creation implementation
- **new.go**: Current project creation logic
- **continue.go**: Current project continuation logic
