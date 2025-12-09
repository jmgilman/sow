# libs/ Refactoring Summary

## Executive Summary

This document consolidates all exploration findings for refactoring `cli/internal/` packages into a new `libs/` directory structure. The refactoring achieves three primary goals:

1. **Decouple libraries from CLI constructs** - Library packages receive only specific dependencies they need (e.g., `core.FS`), not the full `Context`
2. **Unify SDK packages** - Combine `sdks/state` and `sdks/project` into a single `libs/project` package
3. **Enable reuse outside CLI** - Libraries become independent modules usable in other contexts

## Key Design Decisions

### 1. Context Stays in CLI

The `sow.Context` type remains in `cli/internal/sow/` as a CLI-specific convenience wrapper that bundles FS, Git, and GitHub for command convenience. Libraries never receive Context directly.

| Library Package | Receives | NOT |
|----------------|----------|-----|
| `libs/schemas` | (none - leaf) | - |
| `libs/exec` | (none - leaf) | - |
| `libs/config` | File path or bytes | Context |
| `libs/git` | Repo root path, `exec.Executor` | Context |
| `libs/project/state` | `core.FS` directly | Context |

### 2. Remove sow.FS Type Alias

The current `type FS = core.FS` alias in `sow` is eliminated. All code uses `github.com/jmgilman/go/fs/core.FS` directly.

### 3. Split Schemas Package

- **Move to libs/**: Project schemas (`ProjectState`, `PhaseState`, etc.) and config schemas (`Config`, `UserConfig`)
- **Keep in CLI**: Refs schemas (`RefsCommittedIndex`, `RefsLocalIndex`, `RefsCache`) since refs package stays in CLI

### 4. Unify SDK Packages

The `sdks/state` and `sdks/project` packages merge into `libs/project` since only `sdks/project` consumes `sdks/state`.

## Target Directory Structure

```
libs/
├── schemas/                    # Foundation - no dependencies
│   ├── config.cue              # Repo config schema
│   ├── user_config.cue         # User config schema
│   ├── cue_types_gen.go        # Generated: Config, UserConfig
│   ├── embed.go                # Embeds CUE files
│   ├── cue.mod/module.cue      # CUE module definition
│   └── project/
│       ├── project.cue         # ProjectState
│       ├── phase.cue           # PhaseState
│       ├── task.cue            # TaskState
│       ├── artifact.cue        # ArtifactState
│       └── cue_types_gen.go    # Generated Go types
│
├── exec/                       # No dependencies
│   ├── executor.go             # Executor interface + LocalExecutor
│   └── mock.go                 # MockExecutor for testing
│
├── config/                     # Depends on: libs/schemas
│   ├── repo.go                 # LoadRepoConfig(fs) or LoadRepoConfigFromBytes(data)
│   ├── user.go                 # LoadUserConfig(), GetUserConfigPath()
│   ├── defaults.go             # Default values
│   └── paths.go                # GetADRsPath(repoRoot, config)
│
├── git/                        # Depends on: libs/exec
│   ├── git.go                  # Git struct + operations
│   ├── github_client.go        # GitHubClient interface
│   ├── github_cli.go           # GitHubCLI implementation
│   ├── github_factory.go       # NewGitHubClient factory
│   ├── github_mock.go          # Mock for testing
│   ├── worktree.go             # EnsureWorktree(git, repoRoot, path, branch)
│   ├── types.go                # Issue, LinkedBranch
│   └── errors.go               # ErrGHNotInstalled, etc.
│
└── project/                    # Depends on: libs/schemas/project, core.FS
    ├── state.go                # State, Event types
    ├── machine.go              # Machine runtime wrapper
    ├── builder.go              # MachineBuilder (low-level)
    ├── config.go               # ProjectTypeConfig
    ├── project_builder.go      # ProjectTypeConfigBuilder (high-level)
    ├── options.go              # TransitionOption, PhaseOpt
    ├── branch.go               # AddBranch, BranchConfig
    ├── types.go                # GuardTemplate, Action, EventDeterminer
    ├── runtime.go              # BuildMachine, FireWithPhaseUpdates
    └── state/                  # Persistence layer
        ├── project.go          # Project wrapper (holds core.FS, not Context)
        ├── phase.go            # Phase helpers
        ├── task.go             # Task helpers
        ├── artifact.go         # Artifact type
        ├── collections.go      # PhaseCollection, TaskCollection
        ├── loader.go           # Load(fs core.FS), Save()
        ├── convert.go          # CUE type conversion
        ├── validate.go         # Validation helpers
        └── registry.go         # Project type registry
```

## CLI Structure After Refactoring

```
cli/internal/sow/
├── context.go      # Context bundles: core.FS, git.Git, git.GitHubCLI
├── fs.go           # NewFS() helper (no type alias)
├── sow.go          # Init(), DetectContext()
├── errors.go       # ErrNoProject, ErrNotInitialized
└── options.go      # PhaseOption, TaskOption patterns

cli/schemas/        # Only refs schemas remain
├── refs_cache.cue
├── refs_committed.cue
├── refs_local.cue
├── knowledge_index.cue
└── cue_types_gen.go    # Only Refs*, KnowledgeIndex types
```

## Dependency Graph

```
                    ┌──────────────────┐
                    │  libs/schemas    │  (foundation - no deps)
                    └────────┬─────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   │                   ▼
┌────────────────┐           │          ┌────────────────┐
│  libs/config   │           │          │  libs/project  │
└────────────────┘           │          └────────────────┘
                             │
                    ┌────────┴───────┐
                    │   libs/exec    │  (no deps)
                    └────────┬───────┘
                             │
                             ▼
                    ┌────────────────┐
                    │    libs/git    │
                    └────────────────┘

                 ┌─────────────────────────┐
                 │    cli/internal/sow     │
                 │  (imports all libs/*)   │
                 └─────────────────────────┘
```

No circular dependencies exist in this structure.

## Critical Interface Changes

### Project Persistence (Most Important)

```go
// BEFORE
func Load(ctx *sow.Context) (*Project, error)

type Project struct {
    project.ProjectState
    config  ProjectTypeConfig
    machine *sdkstate.Machine
    ctx     *sow.Context  // Full Context
}

// AFTER
func Load(fs core.FS) (*Project, error)

type Project struct {
    project.ProjectState
    config  ProjectTypeConfig
    machine *Machine
    fs      core.FS  // Only FS needed
}
```

### Config Loading

```go
// BEFORE
func LoadConfig(ctx *Context) (*schemas.Config, error)

// AFTER - Option A (accept FS)
func LoadRepoConfig(fs core.FS) (*schemas.Config, error)

// AFTER - Option B (accept bytes, more flexible)
func LoadRepoConfigFromBytes(data []byte) (*schemas.Config, error)
```

### Path Helpers

```go
// BEFORE
func GetADRsPath(ctx *Context, config *schemas.Config) string

// AFTER
func GetADRsPath(repoRoot string, config *schemas.Config) string
```

### Worktree Functions

```go
// BEFORE
func EnsureWorktree(ctx *Context, path, branch string) error

// AFTER
func EnsureWorktree(git *Git, repoRoot, path, branch string) error
```

## Migration Phases

### Phase 0: libs/schemas/ (Foundation)
**Files:** 10 | **Consumers:** 69

1. Create `libs/schemas/` and `libs/schemas/project/`
2. Copy schema files (*.cue)
3. Update CUE module path to `github.com/jmgilman/sow/libs/schemas`
4. Update embed.go for new paths
5. Regenerate Go types in both locations
6. Update 69 consumer import paths
7. Remove moved files from `cli/schemas/`

**Risk:** Medium - CUE module resolution must work correctly

### Phase 1: libs/exec/ (Leaf)
**Files:** 2 (243 lines) | **Consumers:** 8

1. Copy `executor.go` and `mock.go` to `libs/exec/`
2. Update 8 consumer import paths
3. Delete `internal/exec/`

**Risk:** None - pure file move

### Phase 2: libs/config/
**Files:** 2 (490 lines) | **Consumers:** 12

1. Move `user_config.go` (minimal changes)
2. Refactor `config.go`:
   - Change `LoadConfig(ctx)` to `LoadRepoConfig(fs)` or `LoadRepoConfigFromBytes(data)`
   - Change path helpers to accept `repoRoot string`
3. Update 12 consumer files

**Risk:** Low - config loading is well-isolated

### Phase 3: libs/git/
**Files:** 6 (~1100 lines) | **Consumers:** 20+

1. Move Git and GitHub files
2. Refactor worktree functions to accept `*Git` instead of `*Context`
3. Update `internal/sow/context.go` to import from `libs/git`
4. Update 20+ consumer files

**Risk:** Medium - worktree signature changes

### Phase 4: libs/project/ (Largest)
**Files:** 19 (~2730 lines) | **Consumers:** 82

1. Create unified structure from `sdks/state/` + `sdks/project/`
2. Critical: Change `Project.ctx` to `Project.fs`
3. Remove unused `Machine.fs` field
4. Update 82 consumer files
5. Delete `internal/sdks/`

**Risk:** High - many consumers, core data model change

### Phase 5: Cleanup
1. Remove `sow.FS` type alias
2. Update remaining uses to `core.FS`
3. Verify `internal/sow/` contains only CLI-specific code

## File Movement Summary

| Source | Destination | Lines |
|--------|-------------|-------|
| `cli/schemas/` (partial) | `libs/schemas/` | ~400 |
| `internal/exec/` | `libs/exec/` | 243 |
| `internal/sow/config.go`, `user_config.go` | `libs/config/` | 490 |
| `internal/sow/git*.go`, `worktree.go` | `libs/git/` | ~1100 |
| `internal/sdks/state/` | `libs/project/` | ~430 |
| `internal/sdks/project/` | `libs/project/` | ~1200 |
| `internal/sdks/project/state/` | `libs/project/state/` | ~1100 |
| **Total** | | **~5000 lines** |

## Consumer Update Counts

| Package | Files to Update |
|---------|-----------------|
| `libs/schemas` | 69 |
| `libs/exec` | 8 |
| `libs/config` | 12 |
| `libs/git` | 20+ |
| `libs/project` | 82 |

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| CUE schema migration breaks code generation | Test CUE generation in isolation before updating consumers |
| Large PRs difficult to review | Execute phases as separate PRs |
| Breaking changes in project type definitions | API stays same, only import paths change |
| Circular dependencies | Already verified - no cycles possible |
| Test coverage gaps | Move test files alongside source, verify coverage |

## Key Observations

1. **Project only uses ctx.FS()** - The entire Project SDK only needs filesystem access for persistence. This is the critical insight enabling the decoupling.

2. **Machine.fs is dead code** - The `Machine` struct has an `fs` field that is never read. Can be safely removed.

3. **User config already standalone** - The `LoadUserConfig()` function already uses `os.ReadFile` directly, not Context.

4. **Refs package stays in CLI** - The `internal/refs/` package is not moving to libs/, so its schemas stay in `cli/schemas/`.

## Source Documents

- Task 010: sow package refactoring map
- Task 020: exec package refactoring map
- Task 030: SDK unification feasibility
- Task 040: Complete architecture design (comprehensive)
- Task 050: Schemas package migration

All findings documents are located at:
`.sow/project/phases/exploration/tasks/{task-id}/findings.md`
