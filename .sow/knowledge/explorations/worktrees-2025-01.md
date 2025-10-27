# Git Worktrees Exploration

**Date:** January 2025
**Branch:** explore/worktrees
**Status:** Exploration complete

## Executive Summary

This exploration investigated how git worktrees could enable concurrent sow sessions (exploration, project, design, breakdown) through isolated working directories per branch. The research covered worktree isolation mechanics, state management patterns, integration points within sow's architecture, and the current state of go-git library support.

**Key Findings:**
- Worktrees provide clean isolation (separate working directories, shared git history)
- Integration is possible at the `sow.Context` boundary without downstream code changes
- go-git v5 and `github.com/jmgilman/go/git` wrapper lack native worktree support
- Git CLI provides full worktree functionality (add/list/remove)

## What Worktrees Are

Git worktrees allow multiple working directories from a single repository, each checked out to different branches. This is a git feature available via `git worktree add/list/remove` commands.

### Example Structure
```
.sow/worktrees/
├── explore/topic/          # Exploration session (branch: explore/topic)
│   ├── .git → main/.git/worktrees/explore/topic/
│   ├── .sow/exploration/   # Active exploration state
│   └── [working files]
│
└── feat/auth/              # Project session (branch: feat/auth)
    ├── .git → main/.git/worktrees/feat/auth/
    ├── .sow/project/       # Active project state
    └── [working files]
```

**Characteristics:**
- Each worktree is an independent working directory
- All worktrees share the same git object database (disk efficient)
- Git enforces one worktree per branch (prevents conflicts)
- Main repo and worktrees can coexist

## Research Findings

### Worktree Isolation Model

**Shared across all worktrees:**
- `.git/objects/` - Git history (disk efficient)
- `.git/refs/` - Branches, tags
- `.git/config` - Repository settings
- `.sow/knowledge/` - Architecture docs, ADRs (symlink or main repo)
- `.sow/sinks/` - External knowledge
- `.sow/repos/` - Linked repositories

**Isolated per worktree:**
- Working directory - Branch-specific code
- `.git` file (points to main `.git/worktrees/<name>/`)
- Git index (staging area)
- HEAD (current branch/commit)
- `.sow/project/` or `.sow/exploration/` - Session state

### Potential Integration Point

Research identified `sow.Context` creation as a potential integration point:

```
┌─────────────────────────────────────────────┐
│ PersistentPreRunE (cmd/root.go)             │
│ ├─ findRepoRoot(cwd) → main repo            │
│ ├─ NewContext(repoRoot) → main repo context │
│ └─ Pass to command                           │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│ runExplore/runProject                        │
│ ├─ Get main repo context                    │
│ ├─ Determine target branch                  │
│ ├─ Create/switch to worktree ← NEW          │
│ ├─ NewContext(worktreePath) ← NEW           │
│ └─ Pass worktree context to mode            │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│ ModeRunner.Run                               │
│ ├─ Uses worktree context (unchanged)        │
│ ├─ Mode operations in worktree (unchanged)  │
│ └─ Returns prompt                            │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│ launchClaudeCode                             │
│ └─ Dir = worktree path                       │
│    └─ Claude Code runs in worktree           │
└─────────────────────────────────────────────┘
```

**Finding:** The existing `findRepoRoot()` naturally stops at worktree boundaries (`.git` file vs directory), meaning worktree-based sessions would be automatically isolated without code changes to discovery logic.

### State Partitioning

**Per-session state** (lives in worktree):
```
.sow/worktrees/<branch>/.sow/
├── project/          # Project state, tasks, logs
├── exploration/      # Exploration files, index
├── design/           # Design session state
└── breakdown/        # Breakdown session state
```

**Shared knowledge** (stays in main repo or symlinked):
```
.sow/knowledge/
├── overview.md
├── architecture/
├── adrs/
└── explorations/     # Summaries of completed explorations
    └── worktrees-2025-01.md  # This document
```

**Observation:** If worktrees were used, completed exploration summaries could move from worktree-local `.sow/exploration/` to shared `.sow/knowledge/explorations/` for cross-session access.

## Analyzed Integration Patterns

### Hypothetical Command Flow

If worktrees were integrated, commands would need to determine target branch before worktree creation:

```go
func runExplore(cmd *cobra.Command, branchName, initialPrompt string) error {
    mainCtx := cmdutil.GetContext(cmd.Context())  // Main repo context

    // 1. Determine target branch
    targetBranch, err := determineBranch(mainCtx, branchName)
    if err != nil {
        return err
    }

    // 2. Ensure worktree exists (create if needed)
    worktreePath := filepath.Join(mainCtx.RepoRoot(), ".sow/worktrees", targetBranch)
    worktree, err := mainCtx.Git().Repository().CreateWorktree(
        worktreePath,
        git.WorktreeOptions{Branch: plumbing.NewBranchReferenceName(targetBranch)},
    )
    if err != nil {
        return err
    }

    // 3. Create worktree context
    worktreeCtx, err := sow.NewContext(worktree.Path())
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

**Observed patterns:**
- Branch determination currently happens within `ModeRunner.Run()`
- Worktree creation would need to happen after branch determination but before context is passed to mode operations
- This suggests potential restructuring of command flow

### Possible Rollout Approaches

Research identified several possible adoption strategies:

1. **Opt-in flag**: `--worktree` flag for testing without affecting existing workflows
2. **Configuration-based**: Enable per-repo via `.sow/config.yaml`
3. **Always-on**: Make worktrees the only supported mode

Each approach has different migration and compatibility implications.

## Library Support Analysis

### Go-git Worktree Support

**Finding:** go-git v5 does not support git worktree operations:
- No `git worktree add` equivalent
- No `git worktree list` functionality
- No `git worktree remove` capability

**Options identified:**

1. Shell out to git CLI for worktree operations
2. Use third-party fork (github.com/cooper/go-git) - unmaintained, quality unknown
3. Use go-git-cmd-wrapper package - thin wrapper around git CLI
4. Wait for native go-git support (Issue #396, marked "not planned")

**Hypothetical wrapper enhancement** (if worktree support were added):

```go
// In github.com/jmgilman/go/git
func (r *Repository) CreateWorktree(path string, opts WorktreeOptions) (*Worktree, error) {
    // Determine if branch exists
    // Shell out: git worktree add <path> <branch>
    // Return Worktree wrapper
}

func (r *Repository) ListWorktrees() ([]*Worktree, error) {
    // Shell out: git worktree list --porcelain
    // Parse and return Worktree objects
}

func (w *Worktree) Remove() error {
    // Shell out: git worktree remove <path>
}

func (r *Repository) PruneWorktrees() error {
    // Shell out: git worktree prune
}
```

**CLI approach characteristics:**
- Precedent: sow already shells out for `claude` and `gh`
- Git CLI universally available in git repositories
- `--porcelain` flag provides stable output format for parsing
- Forward-compatible with git updates

### Reading Existing Worktrees

While wrapper creates worktrees via CLI, go-git can read them:

```go
// Worktree created by CLI has .git file (not directory)
// .git content: gitdir: /path/to/main/.git/worktrees/<name>

// go-git opens it with:
repo, err := git.PlainOpenWithOptions(
    worktreePath,
    &git.PlainOpenOptions{EnableDotGitCommonDir: true},
)
```

All normal git operations work once worktree is opened.

## Lifecycle Patterns

### Hypothetical Creation Flow
If worktrees were used, creation might look like:
```bash
sow explore --branch explore/topic
# Would need to: create worktree, initialize .sow/ structure, launch Claude
```

### Cleanup Characteristics
- Git worktrees require manual removal: `git worktree remove <path>`
- Adds one step beyond branch deletion
- Git provides `git worktree prune` for orphaned metadata
- Consistent with git's manual lifecycle management

### Discovery Patterns
Worktrees could be discovered by:
- Walking `.sow/worktrees/` directory structure
- Reading git worktree metadata: `git worktree list --porcelain`
- Parsing mode-specific state files in each worktree

## Analyzed Tradeoffs

### Path Structure Options

**Option 1: Preserve forward slashes**
- `feat/auth` → `.sow/worktrees/feat/auth/`
- Natural directory grouping
- Mirrors git branch namespacing

**Option 2: Flatten with escaping**
- `feat/auth` → `.sow/worktrees/feat-auth/`
- Simpler filesystem structure
- Loses semantic grouping

### Main Repo Role Options

**Option 1: Main repo never runs sessions**
- Clean separation
- Forces worktree adoption
- Breaking change

**Option 2: Main repo and worktrees coexist**
- Gradual migration
- User choice
- Maintains compatibility

### Shared Knowledge Access

**Option 1: Symlinks from worktrees**
- Direct filesystem access
- Platform-specific (Windows compatibility)

**Option 2: Read from main repo path**
- Cross-platform
- Requires path resolution logic

**Option 3: Duplicate in each worktree**
- Simple, isolated
- Disk waste, sync issues

## Architectural Observations

### Context Handling
- Current: Single context per command
- With worktrees: Two contexts (main repo + worktree)
- Discovery logic unchanged (stops at `.git` file boundary)

### Command Flow Changes
- Branch determination must happen earlier
- Worktree creation inserted mid-flow
- Mode operations remain unchanged

### Compatibility
- Opt-in approach avoids breaking changes
- Configuration-based allows per-repo control
- Always-on requires migration strategy

## Files Reference

Detailed research documents in `.sow/exploration/`:

1. **worktree-isolation-model.md** - How worktrees share git data but isolate working state
2. **multi-session-state-management.md** - State partitioning, discovery, and session coordination
3. **worktree-integration-patterns.md** - Command flow, context creation, rollout strategy
4. **go-git-worktree-support.md** - Comprehensive analysis of go-git limitations and CLI approach

## Summary

This exploration investigated git worktrees as a mechanism for concurrent sow sessions. Key findings:

**Technical Feasibility:**
- Worktrees provide clean isolation at the filesystem level
- Integration point identified at `sow.Context` boundary
- Existing discovery logic compatible without modification

**Library Support:**
- go-git v5 lacks native worktree support (Issue #396, marked "not planned")
- `github.com/jmgilman/go/git` wrapper also lacks support
- Git CLI provides full functionality
- Multiple integration approaches possible

**Architectural Patterns:**
- Multiple viable approaches for state partitioning
- Tradeoffs identified between various implementation strategies
- No single "correct" design - depends on use case priorities

**Use Cases:**
- Concurrent exploration and project work
- Parallel development on multiple features
- Context isolation between session types
- Reduced branch switching overhead

This exploration provides foundational research for future design decisions regarding multi-session support in sow. The findings document what's possible, library constraints, and architectural considerations without prescribing a specific solution.

---

**Exploration conducted:** October 2024 - January 2025
**Related branches:** explore/worktrees
**Participants:** Josh Gilman (user), Claude (exploration assistant)
