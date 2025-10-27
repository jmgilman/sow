# Multi-Session State Management

## Summary

With worktrees in `.sow/worktrees/<branch>/ `(preserving `/` in branch names), each session gets isolated project state while sharing knowledge/sinks. Main repo stays clean. Session discovery via directory listing. No path resolution changes needed.

## Architecture

### Directory Structure

```
repo/
├── .git/
│   └── worktrees/                    # Git's worktree metadata
│       ├── explore/
│       │   └── topic/
│       └── feat/
│           └── auth/
│
├── .sow/
│   ├── knowledge/                    # SHARED across sessions
│   │   ├── architecture/
│   │   ├── adrs/
│   │   └── explorations/             # Summaries of past explorations (indexed)
│   ├── sinks/                        # SHARED across sessions
│   ├── repos/                        # SHARED across sessions
│   │
│   └── worktrees/                    # GIT-IGNORED
│       ├── explore/                  # Forward slash preserved
│       │   └── topic/                # Session 1: exploration
│       │       ├── .git              # File pointing to main .git/worktrees/
│       │       ├── .sow/
│       │       │   ├── project/      # Session state
│       │       │   └── exploration/  # Active exploration files
│       │       └── [working files]
│       │
│       └── feat/                     # Forward slash preserved
│           └── auth/                 # Session 2: feature project
│               ├── .git              # File pointing to main .git/worktrees/
│               ├── .sow/
│               │   └── project/      # Session state
│               └── [working files]
│
└── [main repo working files - kept clean]
```

### State Partitioning

**Per-session (isolated):**
- `.sow/project/` - Project state, tasks, logs (one per branch, per design)
- `.sow/exploration/` - Active exploration files (one per branch, per design)
- Working directory - Branch-specific code
- Git index - Staging area

**Shared (all sessions):**
- `.sow/knowledge/` - Architecture docs, ADRs
- `.sow/knowledge/explorations/` - Summaries of completed explorations (indexed)
- `.sow/sinks/` - External knowledge
- `.sow/repos/` - Linked repositories
- Git object database - Commits, history

### Exploration Mode

Explorations are **per-session** (one per branch, by design):
```
.sow/worktrees/explore/topic/
└── .sow/
    └── exploration/          # Active exploration files for this branch
```

**Completed exploration summaries** live in shared knowledge:
```
.sow/knowledge/explorations/
├── index.json                # Index for quick search
├── worktrees-2025-01.md     # Summary of worktrees exploration
└── auth-patterns-2024-12.md # Summary of auth exploration
```

This enables:
- Clean per-branch isolation during active exploration
- Shared access to past findings via indexed summaries
- Context gathering for future work

## Session Discovery

### Listing Active Sessions

With forward slashes preserved, discovery is trivial:
```bash
# Simple: list .sow/worktrees/ subdirectories
$ ls -R .sow/worktrees/
explore/topic/
feat/auth/

# Sow provides structured view:
$ sow sessions list
Active sessions:
  • explore/topic  (exploration mode)
  • feat/auth      (orchestrator, implementation phase)
```

Implementation: Walk `.sow/worktrees/`, preserving directory structure. Exploration sessions naturally grouped under `explore/`.

### Finding Current Session Context

**No changes needed.** Existing discovery works unchanged because:

1. Each worktree has `.git` file (not directory) pointing to main repo
2. `findRepoRoot()` walks up looking for `.git` - stops at worktree boundary
3. `NewContext(repoRoot)` looks for `.sow/` within that root
4. Agents execute from within worktree → discovery finds worktree's `.sow/`

See `cli/cmd/root.go:95-132` for current implementation.

### Switching Sessions

User workflow:
```bash
# Terminal 1: Start exploration session
$ sow start --branch explore/topic
# → Creates/switches to .sow/worktrees/explore/topic/
# → Opens Claude Code rooted at that directory

# Terminal 2: Start project session (concurrent)
$ sow start --branch feat/auth
# → Creates/switches to .sow/worktrees/feat/auth/
# → Opens Claude Code rooted at that directory
```

Key: Claude Code is rooted in the worktree, NOT main repo. This enables concurrent sessions without interference.

## Gitignore Requirements

Add to `.gitignore`:
```
.sow/worktrees/
```

This prevents:
- Committing session working directories
- Worktree state leaking between branches
- Conflicts when switching branches in main repo

## Path Resolution

**No changes required.** The existing mechanism works unchanged:

1. `findRepoRoot()` walks up looking for `.git`
2. Worktrees have `.git` file → stops at worktree root
3. `NewContext()` creates FS chrooted to `<repoRoot>/.sow/`
4. Each worktree has its own `.sow/` → agents use worktree-local state

All agents execute from within the worktree, so discovery naturally finds the correct `.sow/` directory.

## Edge Cases

### Branch Name Handling
- Git enforces unique worktree branches (prevents conflicts)
- Directory structure: **Preserve forward slashes** (`feat/auth` → `.sow/worktrees/feat/auth/`)
- Provides natural grouping (all explorations under `explore/`, features under `feat/`, etc.)

### Orphaned Worktrees
- Worktree deleted manually but git metadata remains
- Existing cleanup logic needs modification to also clean worktrees
- Commands: `git worktree prune` + `sow sessions clean`
- Consider automatic cleanup on exploration/project completion

### Concurrent Git Operations
- Git handles locking (refs, index)
- Each worktree has isolated index → no conflicts
- Shared object DB is thread-safe

### Main Repo Role
- No `.sow/project/` in main repo
- User must create session to work
- Main repo for: reviewing knowledge, managing sessions, non-project tasks

## Implementation Implications

### Commands That Need Updates

**`sow start`**
- `sow start` itself doesn't worry about branches (that's for modes/projects)
- When creating session for mode/project:
  - Create worktree if needed: `git worktree add .sow/worktrees/<branch> <branch>`
  - Launch Claude Code rooted in worktree directory
  - Forward slashes preserved in path

**`sow project`** / **`sow explore`** / **`sow design`** / **`sow breakdown`**
- No changes needed
- Discovery mechanism works identically from within worktree
- Each mode reads its own `.sow/project/` state

**New: `sow sessions`**
- `sow sessions list` - Show active sessions
- `sow sessions switch <branch>` - Switch to existing session
- `sow sessions clean` - Remove orphaned worktrees

### No Changes Needed

- Path resolution - `findRepoRoot()` naturally stops at worktree `.git` file
- Worker agents - Already read `.sow/project/` via discovery
- Knowledge management - Shared across sessions via main repo `.sow/knowledge/`
- Sinks/repos - Shared across sessions via main repo `.sow/{sinks,repos}/`
- All mode commands - Discovery works identically

## Open Questions

1. **Branch auto-creation:** Should `sow start --branch new-branch` auto-create branch if it doesn't exist?
   - Probably yes, with confirmation

2. **Session cleanup automation:** When should worktrees be auto-cleaned?
   - On exploration completion (`sow exploration set-status completed`)?
   - On project merge/completion?
   - Manual only for safety?
   - Likely: Modify existing cleanup logic to handle worktrees

3. **Exploration summary workflow:** How/when to create summaries in `.sow/knowledge/explorations/`?
   - Prompt user on exploration completion?
   - Manual via dedicated command?
   - Agent creates automatically?

4. **Sessions command scope:** Should `sow sessions` be a new top-level command or subcommand?
   - `sow sessions list|switch|clean` vs `sow session-list` etc.

## Next Steps

To validate integration patterns:
- **Sow integration patterns** - Actual command UX and workflows
- **Lifecycle management** - Detailed create/switch/cleanup flows
