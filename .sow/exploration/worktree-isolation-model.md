# Git Worktree Isolation Model

## Summary

Git worktrees enable multiple working directories from a single repository, each with an isolated working tree and index but sharing the object database and refs. This maps well to sow's multi-session needs: each session gets independent file state while sharing git history.

## What's Shared

- **Object database** (`.git/objects/`) - All commits, trees, blobs
- **Refs** (`.git/refs/`) - Branches, tags, remotes
- **Configuration** (`.git/config`) - Repository settings
- **Hooks** (`.git/hooks/`) - Git hooks apply to all worktrees

## What's Isolated

- **Working directory** - Completely independent file trees
- **Index** (staging area) - Each worktree has its own `.git/worktrees/<name>/index`
- **HEAD** - Each worktree can be on different branch or commit
- **Checkout state** - Branch protection: can't checkout same branch in multiple worktrees

## Filesystem Layout

### Main repository:
```
/path/to/repo/
├── .git/
│   ├── objects/          # Shared
│   ├── refs/             # Shared
│   ├── config            # Shared
│   └── worktrees/        # Worktree metadata
│       └── session-1/
│           ├── HEAD
│           ├── index
│           └── gitdir
└── [working files]
```

### Additional worktree:
```
/path/to/worktree-dir/
├── .git                  # File pointing to main .git/worktrees/session-1/
└── [working files]
```

## Branch Protection Mechanism

Git prevents checking out the same branch in multiple worktrees by maintaining lock files. Attempting to checkout an already-checked-out branch fails with:

```
fatal: 'branch-name' is already checked out at '/path/to/other/worktree'
```

This is **good for sow**: Enforces one-session-per-branch, preventing state conflicts.

## Implications for Sow Multi-Session

### ✅ Enables
- Multiple concurrent sessions (different branches)
- Isolated `.sow/project/` directories (each worktree has independent files)
- Shared git history (performance, disk efficiency)
- Natural branch isolation (git enforces it)

### ⚠️ Constraints
- Can't run multiple sessions on same branch (by design)
- Worktrees must be on different branches or detached HEAD
- Path-based: each session needs a distinct directory

### 🤔 Open Questions
- Where should worktree directories live? (e.g., `/tmp/sow-sessions/`, `.sow/worktrees/`, user-specified?)
- How does `sow` discover/list active sessions across worktrees?
- Should main repo remain a "coordination" space, or also run sessions?

## Relevance to Use Case

**Your goal**: Run exploration mode on one branch, orchestrator project on another, simultaneously.

**Worktree model**: Each session gets:
- Independent working directory → separate `.sow/project/` states ✅
- Different branch → exploration vs project branches ✅
- Shared git data → can reference commits, switch contexts ✅
- Isolated staging → no conflict when committing ✅

**Feasibility**: High. Worktree isolation matches sow's needs.

## Next Steps

To validate integration:
1. **Multi-session state management** - How sow discovers and manages sessions
2. **Integration patterns** - Path resolution, session creation UX
3. **Lifecycle management** - Creating, switching, cleaning up worktrees
