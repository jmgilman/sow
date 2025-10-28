# ADR-002: Interactive Wizard for Project Initialization

**Status**: Proposed
**Date**: 2025-10-28
**Deciders**: Josh Gilman, Architecture Team
**Related Design**: [Interactive Project Launch Design](../knowledge/designs/interactive-project-launch-design.md)
**Related ADR**: [ADR-001: Consolidate Operating Modes into Project Types](001-consolidate-modes-to-projects.md)

## Context

The `sow project` command serves as the primary entry point for all project work in the consolidated project system (post-ADR-001). This command must handle multiple complex scenarios:

1. **Repository state detection** - Protected vs feature branch, existing project vs clean state
2. **Project type selection** - Standard, exploration, design, breakdown
3. **Branch management** - Create new, switch to existing, validate working tree
4. **Context collection** - GitHub issue linking, project naming, initial configuration
5. **Error recovery** - Handle edge cases gracefully without blocking users

**Current problems with flag-based approach:**

The existing implementation uses flags (`--branch`, `--type`, etc.) combined with subcommands. This creates five documented edge cases:

1. **Feature branch creation from wrong location** - Creating new branch via `--branch` requires being on protected branch first, otherwise cryptic error
2. **Issue with linked branch but no project** - Unexpected state causes error instead of offering recovery path
3. **Multiple linked branches** - No user control over which branch is selected
4. **Protected branch blocks creation** - No guidance on next steps when attempting project creation on main/master
5. **Uncommitted changes** - Not validated before branch switching, risks losing work

These edge cases share a root cause: **flags require users to understand repository state before choosing correct command structure**. Users must mentally validate preconditions that the CLI could validate programmatically.

**User experience impact:**

```bash
# User doesn't realize they're on protected branch
$ sow project --type exploration --branch explore/auth
Error: cannot create project on protected branch

# User doesn't know what to do next - must manually:
# 1. Figure out they need to create branch first
# 2. Run git checkout -b explore/auth
# 3. Re-run sow project
# 4. Hope no other edge cases occur
```

**Why this matters architecturally:**

The project initialization flow is the **first interaction** users have with sow's project system. Poor UX here creates negative first impressions and reduces adoption. More critically, flag-based interfaces that require state understanding **don't scale** as project types multiply (exploration, design, breakdown, future types each add complexity).

**Important context for automation concerns:**

`sow project` always culminates in **launching an interactive Claude Code session**. The wizard creates project state, then hands off to `claude` CLI with a state-specific prompt. This means the command's ultimate purpose is **human-in-the-loop AI collaboration**, not unattended automation. Concerns about "no scripting support" are largely theoretical - you wouldn't automate launching an interactive AI session. Any genuine automation needs (e.g., CI creating project structure) would likely stop before Claude launch, which could be addressed with targeted flags if demand emerges.

## Decision

Implement `sow project` as an **interactive wizard** with no flags, using context-aware prompts and validation at each decision point.

**Core principles:**

1. **Detect state before asking questions** - Inspect repository, determine context, route to appropriate flow
2. **Validate before transitioning** - Check preconditions (working tree clean, branch exists, etc.) before operations
3. **Provide recovery paths** - When edge cases detected, offer clear options instead of errors
4. **Guide, don't block** - Users should never be stuck with unclear error messages

**Implementation approach:**

- **No flags** for initial implementation (deferred, not rejected - see "Consequences")
- **Lightweight state machine** tracks wizard position for testing without `stateless` library overhead
- **`charmbracelet/huh`** for terminal UI with dynamic forms and validation
- **Exhaustive decision tree** covering all repository states (protected/feature, project/no-project)

**Example user flow:**

```bash
$ sow project

# Wizard detects: on protected branch (main)
┌─────────────────────────────────────────┐
│ You're on 'main' (protected branch).   │
│ Projects must be on feature branches.  │
└─────────────────────────────────────────┘

What would you like to do?
  1. Create new project on new branch
  2. Switch to existing feature branch
  3. Cancel

> 1

What type of project?
  1. Standard (feature work)
  2. Exploration (research)
  3. Design (architecture docs)
  4. Breakdown (decompose work)

> 2

Enter branch name:
Hint: Use prefix 'explore/' for exploration projects
> explore/auth-patterns

✓ Creating exploration project on explore/auth-patterns
✓ Launching orchestrator...
```

**Wizard validates automatically:**
- Working tree clean before branch creation
- Branch name valid (no spaces, not protected, doesn't exist)
- Project type matches branch prefix (warns if mismatch)
- All preconditions met before creating project

## Alternatives Considered

### Option 1: Enhanced Flags with Better Error Messages

**Description**: Keep flag-based interface, improve error messages to explain what went wrong and suggest corrections.

**Example:**
```bash
$ sow project --type exploration --branch explore/auth
Error: cannot create project on protected branch 'main'

Suggestion: Switch to a feature branch first:
  git checkout -b explore/auth
  sow project --type exploration

Or use the interactive wizard:
  sow project
```

**Pros:**
- Non-breaking change (preserves existing interface)
- Scriptable/automatable for CI
- Familiar to users expecting CLI tools

**Cons:**
- Doesn't solve root problem (user must understand state)
- Error messages become complex documentation
- Still susceptible to edge cases as new project types added
- Users must read errors, mentally model state, try again

**Why not chosen**: Addresses symptoms, not root cause. Better errors help but don't eliminate the need for users to understand complex state before choosing flags.

---

### Option 2: Subcommands for Each Action

**Description**: Split functionality into explicit subcommands: `sow project new`, `sow project resume`, `sow project switch`.

**Example:**
```bash
sow project new --type exploration --branch explore/auth
sow project resume
sow project switch --branch design/api-redesign
```

**Pros:**
- Clear intent (command name describes action)
- Scriptable/automatable
- Familiar Unix pattern

**Cons:**
- Users must still choose correct subcommand based on state
- Doesn't eliminate edge cases (e.g., `new` fails if already on feature branch with project)
- Multiplies commands (3+ subcommands × multiple flags each)
- No validation until after user commits to command

**Why not chosen**: Defers state understanding to user. Better command names don't eliminate the problem of users choosing wrong command for their current state.

---

### Option 3: Hybrid (Wizard + Flags)

**Description**: Provide both interactive wizard and flag-based interface, letting users choose their preference.

**Example:**
```bash
sow project                               # Interactive wizard
sow project --type exploration --non-interactive  # Flags for CI
```

**Pros:**
- Best of both worlds (UX for humans, automation for scripts)
- Gradual adoption (users can choose wizard)
- Future-proof for CI needs

**Cons:**
- Significant implementation complexity (two complete code paths)
- Must maintain both interfaces forever (breaking either affects users)
- Flag interface still has edge case problems
- Higher maintenance burden (changes must update both paths)
- Risk of divergence (one interface gains features the other lacks)

**Why not chosen**: Premature optimization. No evidence yet that CI/automation is needed for project initialization (versus project execution). Better to implement wizard well, add flags later if demand proven, rather than maintain two interfaces from day one.

---

## Consequences

### Positive

- **Edge cases eliminated** - State detection and validation prevent all five documented edge cases
- **Reduced cognitive load** - Users answer simple questions instead of remembering flag combinations
- **Clear recovery paths** - Wizard offers options when problems detected, no dead ends
- **Better discoverability** - New users can explore options via prompts instead of reading docs
- **Testable flows** - Lightweight state machine enables unit testing wizard logic without interactive forms
- **Consistent UX pattern** - Establishes precedent for other complex commands (exploration, design) to use wizards

### Negative

- **No scripting support initially** - CI/automation must wait for flag support or use workarounds (manual branch creation + state file manipulation). **However**: This concern is largely theoretical since `sow project` **always launches an interactive Claude Code session** upon completion. You wouldn't automate launching an interactive AI session anyway - the command's purpose is human-in-the-loop project work. Any genuine automation need would likely stop before the Claude launch (e.g., just create project structure), which could be addressed with targeted flags if demand emerges.
- **Longer interaction for expert users** - Users who know exactly what they want still step through prompts
- **Terminal UI dependency** - Requires `huh` library and functional terminal (not an issue for Claude Code's primary use case)
- **Testing complexity** - Interactive tests harder than flag-based command tests (mitigated by state machine design)

### Neutral

- **Flags deferred, not rejected** - Can add `--non-interactive` mode later if automation demand emerges
- **Pattern may extend** - Other commands may adopt wizard pattern, establishing consistency (or creating more terminal UI dependencies)

## Reversibility

This decision is **highly reversible**:

1. **Wizard can coexist with flags** - Implementation structure supports adding flag-based interface later without breaking wizard
2. **No schema changes** - Decision affects CLI interface only, not underlying project structure
3. **Non-breaking addition** - Adding flags later doesn't affect existing wizard users

**Adding flags later (if needed):**

```go
// Wizard flow (existing)
func runProject(cmd *cobra.Command, args []string) error {
    if !isInteractive() {  // Check for --non-interactive flag
        return runProjectNonInteractive(cmd, args)
    }

    wizard := newWizard(ctx)
    return wizard.run()
}

// Flag-based flow (future)
func runProjectNonInteractive(cmd *cobra.Command, args []string) error {
    // Validate all preconditions upfront
    // Create project directly
    // Launch orchestrator
}
```

**Criteria for adding flags:**
- Demonstrated CI/automation need (not theoretical)
- Multiple users request scriptable interface
- Clear use cases that wizard doesn't serve

Until then, **simpler implementation wins**.

## Implementation Notes

See [Interactive Project Launch Design](../knowledge/designs/interactive-project-launch-design.md) for complete implementation details including:
- Repository state detection logic
- Complete decision tree (all flows)
- Wizard state machine structure
- Testing approach with mocking
- UX principles and validation points

**Key implementation constraint**: Wizard must remain **single-purpose** - only handles project initialization. Resist temptation to add unrelated features ("while we're here, let's also..."). Keep wizard focused on getting user from repository state to initialized project.

## References

- **Interactive Project Launch Design** - Complete specification of wizard implementation
- **ADR-001** - Context for why project initialization is critical (mode consolidation)
- **Edge case analysis** - Documented in exploration research (January 2025)
- **charmbracelet/huh** - [GitHub](https://github.com/charmbracelet/huh) - Terminal UI library choice
