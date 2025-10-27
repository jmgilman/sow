# Worktree Integration Patterns

## Summary

Worktrees should be created after branch determination but before mode initialization. The natural injection point is between `ModeRunner.Run()` and mode state operations. Two-phase context approach: main repo context for git operations, worktree context for mode execution.

## Current Flow

### Command Execution (sow explore/project)

```
1. PersistentPreRunE (root.go:36-62)
   ├─ findRepoRoot(cwd) → main repo path
   ├─ NewContext(repoRoot) → main repo context
   └─ Store context in cobra.Command

2. runExplore/runProject
   ├─ Get context (main repo)
   └─ Create ModeRunner

3. ModeRunner.Run (runner.go:35-77)
   ├─ HandleBranchScenario
   │  ├─ git.Branches() - list branches
   │  ├─ git.CheckoutBranch(name) OR CreateBranch(name)
   │  └─ Determine shouldCreateNew
   │
   ├─ IF shouldCreateNew:
   │  └─ initFunc(ctx, topic, branch)
   │     └─ Creates .sow/exploration/index.yaml in main repo
   │
   └─ promptGenerator(ctx, topic, branch, prompt)
      └─ Reads .sow/exploration/ from main repo

4. launchClaudeCode (project.go:422-443)
   └─ exec.Command("claude", prompt)
      └─ Dir = ctx.RepoRoot() → main repo
```

### Key Observations

**Context creation timing:**
- Happens in `PersistentPreRunE` before command runs
- Too early to know which branch/worktree needed
- Commands receive pre-built context

**Branch operations:**
- Happen in `ModeRunner.Run` via `HandleBranchScenario`
- Git checkout/create in main repo
- After this, we know the target branch

**Mode state operations:**
- `initFunc` creates mode state (e.g., `.sow/exploration/`)
- `promptGenerator` reads mode state
- Both use context passed to `ModeRunner.Run`
- Currently operates on main repo

**Claude Code launch:**
- Working directory set to `ctx.RepoRoot()`
- Currently always main repo
- Claude Code inherits environment

## Worktree Integration Points

### Option A: Context Recreation (Recommended)

Create worktree after branch determination, recreate context for worktree, pass to mode operations.

**Modified flow:**

```
1. PersistentPreRunE
   ├─ findRepoRoot(cwd) → main repo
   ├─ NewContext(repoRoot) → MAIN REPO context
   └─ Store in cobra.Command

2. runExplore/runProject
   ├─ Get main repo context
   ├─ Create ModeRunner
   └─ Run mode

3. ModeRunner.Run
   ├─ HandleBranchScenario
   │  ├─ Uses main repo context
   │  ├─ Determines branch
   │  └─ Returns branch info
   │
   ├─ *** NEW: Create/switch worktree ***
   │  ├─ git.CreateWorktree(branch, ".sow/worktrees/<branch>")
   │  └─ Get worktree path
   │
   ├─ *** NEW: Create worktree context ***
   │  └─ worktreeCtx = NewContext(worktreePath)
   │
   ├─ IF shouldCreateNew:
   │  └─ initFunc(worktreeCtx, topic, branch)
   │     └─ Creates .sow/exploration/ IN WORKTREE
   │
   └─ promptGenerator(worktreeCtx, topic, branch)
      └─ Reads .sow/exploration/ FROM WORKTREE

4. launchClaudeCode
   └─ Dir = worktreePath (from worktree context)
```

**Changes needed:**
- `ModeRunner.Run` becomes worktree-aware
- Add worktree creation logic between branch determination and mode ops
- Create new context for worktree
- Pass worktree context to `initFunc` and `promptGenerator`
- Return worktree path in `RunResult`

**Pros:**
- Clean separation: main repo context for git, worktree context for modes
- Mode operations naturally work in worktree
- Minimal changes to existing code

**Cons:**
- Two contexts per command execution
- ModeRunner needs worktree logic

### Option B: Launch-time Context

Don't recreate context in CLI. Change Claude Code's working directory to worktree. Claude Code creates its own context.

**Modified flow:**

```
1. PersistentPreRunE
   └─ Create main repo context (unchanged)

2. runExplore/runProject
   └─ ModeRunner.Run (mostly unchanged)
      ├─ Branch determination
      ├─ Mode init (in main repo - PROBLEM)
      └─ Prompt generation (reads main repo - PROBLEM)

3. *** NEW: Create worktree ***
   ├─ After ModeRunner.Run returns
   └─ git.CreateWorktree(branch, path)

4. launchClaudeCode
   └─ Dir = worktreePath
      └─ Claude Code starts
         ├─ Its own PersistentPreRunE runs
         ├─ findRepoRoot() finds worktree .git file
         └─ Creates worktree context
```

**Problems:**
- Mode state created in main repo, but Claude Code runs in worktree
- State mismatch: main repo has `.sow/exploration/`, worktree doesn't
- Won't work without significant restructuring

**Conclusion:** Not viable.

### Option C: Worktree-First Approach

Create/switch to worktree BEFORE mode operations. ALL operations happen in worktree.

**Modified flow:**

```
1. PersistentPreRunE
   └─ Create main repo context

2. runExplore/runProject
   ├─ Determine branch name (from flag or current)
   │
   ├─ *** NEW: Create/switch worktree EARLY ***
   │  ├─ Check if branch exists
   │  ├─ Create worktree for branch
   │  └─ Get worktree path
   │
   ├─ *** NEW: Create worktree context ***
   │  └─ worktreeCtx = NewContext(worktreePath)
   │
   └─ ModeRunner.Run(worktreeCtx, ...)
      ├─ Branch already determined (no checkout needed)
      ├─ Mode ops happen in worktree
      └─ Return result

3. launchClaudeCode
   └─ Dir = worktreePath
```

**Changes needed:**
- Move branch determination BEFORE ModeRunner
- Create worktree immediately after knowing branch
- Pass worktree context to ModeRunner
- Simplify ModeRunner (no git checkout, just mode ops)

**Pros:**
- Single context throughout (worktree context)
- Cleaner: all operations in worktree
- ModeRunner doesn't need worktree awareness

**Cons:**
- Need branch name early (before ModeRunner)
- More changes to command structure
- Git operations might still need main repo context

## Recommended Approach

**Hybrid of A and C:**

1. **Early branch determination**: Extract branch name before ModeRunner
2. **Worktree creation**: Create/switch to worktree after branch determined
3. **Single worktree context**: Create worktree context, use throughout
4. **Simplified ModeRunner**: Receives worktree context, no git operations

### Detailed Design

**Command structure:**
```go
func runExplore(cmd *cobra.Command, branchName, initialPrompt string) error {
    mainCtx := cmdutil.GetContext(cmd.Context())  // Main repo context

    // 1. Determine target branch
    targetBranch, err := determineBranch(mainCtx, branchName)
    if err != nil {
        return err
    }

    // 2. Create/switch to worktree
    worktreePath, err := ensureWorktree(mainCtx, targetBranch)
    if err != nil {
        return err
    }

    // 3. Create worktree context
    worktreeCtx, err := sow.NewContext(worktreePath)
    if err != nil {
        return err
    }

    // 4. Run mode with worktree context
    result, err := runner.Run(worktreeCtx, targetBranch, initialPrompt)
    if err != nil {
        return err
    }

    // 5. Launch Claude Code from worktree
    return launchClaudeCode(cmd, worktreeCtx, result.Prompt, claudeFlags)
}
```

**New functions needed:**

```go
// determineBranch figures out which branch to use
// Handles: --branch flag, current branch, creation validation
func determineBranch(ctx *sow.Context, branchFlag string) (string, error)

// ensureWorktree creates worktree if needed, returns path
// Handles: worktree creation, branch checkout in worktree
func ensureWorktree(ctx *sow.Context, branch string) (string, error)
```

**ModeRunner changes:**
- Remove branch checkout/create logic (happens before ModeRunner)
- Receives worktree context
- Simplify to just: check exists, init if needed, generate prompt

## Worktree Manager Design

Encapsulate worktree operations in a dedicated type.

```go
// WorktreeManager handles git worktree operations
type WorktreeManager struct {
    git *sow.Git
    basePath string  // .sow/worktrees/
}

// Ensure creates worktree if needed, returns path
func (m *WorktreeManager) Ensure(branch string) (string, error) {
    worktreePath := m.worktreePath(branch)

    // Check if worktree exists
    if m.exists(worktreePath) {
        return worktreePath, nil
    }

    // Check if branch exists
    branches, err := m.git.Branches()
    if err != nil {
        return "", err
    }

    branchExists := containsBranch(branches, branch)

    if branchExists {
        // Worktree for existing branch
        return m.create(branch, worktreePath)
    }

    // Create branch and worktree
    return m.createWithBranch(branch, worktreePath)
}

// create adds worktree for existing branch
func (m *WorktreeManager) create(branch, path string) (string, error)

// createWithBranch creates new branch and worktree
func (m *WorktreeManager) createWithBranch(branch, path string) (string, error)

// List returns all active worktrees
func (m *WorktreeManager) List() ([]WorktreeInfo, error)

// Remove deletes worktree
func (m *WorktreeManager) Remove(branch string) error

// Prune cleans orphaned worktrees
func (m *WorktreeManager) Prune() error

// worktreePath converts branch name to worktree path
// Preserves forward slashes: feat/auth → .sow/worktrees/feat/auth
func (m *WorktreeManager) worktreePath(branch string) string
```

## Git Operations Layer

Add worktree support to `sow.Git`.

**New methods:**

```go
// CreateWorktree creates a new worktree for the given branch
func (g *Git) CreateWorktree(branch, path string) error

// CreateWorktreeWithNewBranch creates worktree and new branch
func (g *Git) CreateWorktreeWithNewBranch(branch, path string) error

// ListWorktrees returns all worktrees
func (g *Git) ListWorktrees() ([]WorktreeInfo, error)

// RemoveWorktree removes a worktree
func (g *Git) RemoveWorktree(path string) error

// PruneWorktrees removes orphaned worktrees
func (g *Git) PruneWorktrees() error
```

**Implementation notes:**
- Use `git worktree add <path> <branch>`
- Use `git worktree list --porcelain` for listing
- Use `git worktree remove <path>` for cleanup
- Use `git worktree prune` for orphan cleanup

## Opt-in vs Always-on

**Question:** Should worktrees be opt-in or automatic?

### Option 1: Always-on (Recommended for long-term)

All mode commands automatically use worktrees:
- `sow explore --branch X` → creates `.sow/worktrees/explore/X/`
- `sow project --branch Y` → creates `.sow/worktrees/feat/Y/`
- Main repo stays clean

**Pros:**
- Consistent experience
- True session isolation
- Enables concurrent work

**Cons:**
- Breaking change (migration needed)
- Disk usage (multiple working trees)
- Learning curve

### Option 2: Opt-in Flag

Add `--worktree` flag to mode commands:
- `sow explore --branch X --worktree`
- Without flag, current behavior (main repo)

**Pros:**
- Non-breaking
- Users choose when to use worktrees
- Gradual adoption

**Cons:**
- Two code paths to maintain
- Inconsistent behavior
- Less clear benefit

### Option 3: Configuration Setting

Global setting in `.sow/config.yaml`:
```yaml
features:
  worktrees:
    enabled: true
    path: .sow/worktrees
```

**Pros:**
- Per-repo control
- Single code path once enabled
- Non-breaking (default disabled)

**Cons:**
- Configuration complexity
- Migration still needed eventually

### Recommendation

**Phase 1 (Initial):** Opt-in flag
- Add `--worktree` flag to explore/project/design/breakdown
- Implement worktree path in parallel to current path
- Gather feedback

**Phase 2 (Stable):** Configuration setting
- Add config option (default: disabled)
- Deprecation notice for non-worktree mode

**Phase 3 (Future):** Always-on
- Remove configuration option
- Worktrees become default and only behavior
- Migration tool for existing projects

## Claude Code Launch Consideration

When Claude Code is launched from worktree:
- It runs its own `findRepoRoot()` → finds worktree `.git` file
- Stops at worktree boundary (correct behavior)
- Creates context for worktree
- All operations naturally scoped to worktree

**Key insight:** Once Claude Code is launched from worktree with worktree context, it Just Works™. No special handling needed in Claude Code itself.

## Open Questions

1. **Branch creation location**: Should branches be created in main repo or worktree?
   - Probably main repo (git operations there), then worktree references it

2. **Initial .sow/ in worktrees**: Should worktrees start with `.sow/` structure?
   - Yes, but only mode-specific parts (e.g., `.sow/exploration/`)
   - Shared parts (`.sow/knowledge/`, `.sow/sinks/`) stay in main repo

3. **Switching between worktrees**: How does user switch active session?
   - Multiple terminals (one per worktree)
   - Or `sow sessions switch <branch>` to cd and open new Claude Code

4. **Main repo role**: What can user do in main repo without worktree?
   - View knowledge, manage refs, list sessions
   - But no project/exploration work

5. **Cleanup timing**: When to auto-remove worktrees?
   - On exploration completion? (probably yes)
   - On project merge? (probably yes)
   - On branch deletion? (definitely yes)

## Next Steps

To move forward with implementation:
1. **Worktree lifecycle** - Detailed create/switch/cleanup flows
2. **Git operations** - Implement worktree methods in `sow.Git`
3. **Command restructuring** - Refactor explore/project to use worktrees
4. **Migration tooling** - Help users transition existing work
