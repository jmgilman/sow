# 2. Architecture Constraints

## Technical Constraints

| Constraint                       | Description                                                            | Rationale                                                                                      |
| -------------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| **Go 1.21+**                     | CLI implementation must use Go 1.21 or later                           | Leverage modern Go features (generics, improved type inference) and maintain toolchain support |
| **Claude Code Integration**      | Must integrate as Claude Code plugin with agent spawning via Task tool | Platform requirement - sow is built specifically for Claude Code ecosystem                     |
| **CUE Schema Language**          | State schemas defined in CUE, Go types generated                       | Type-safe validation with human-readable schema definitions                                    |
| **Git Repository Required**      | All operations require git repository context                          | Project state versioning and branch-based workflows depend on git                              |
| **GitHub CLI (gh)**              | GitHub operations via gh CLI, not API directly                         | Leverages existing authentication, simpler implementation, consistent with user's local setup  |
| **Filesystem-Based State**       | All context stored in filesystem (no external database)                | Enables zero-context resumability, git versioning, and transparency                            |
| **Markdown + YAML**              | State files must be human-readable markdown or YAML                    | Transparency requirement - humans must be able to read/debug state                             |
| **Single Project Per Branch**    | Only one project can exist per git branch                              | Simplifies orchestrator logic, natural git integration                                         |
| **Billy Filesystem Abstraction** | Use Billy for filesystem operations                                    | Cross-platform compatibility, testability with in-memory implementations                       |

## Organizational Constraints

| Constraint                           | Description                                                                       | Impact                                                                       |
| ------------------------------------ | --------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| **Two-Layer Architecture**           | Execution layer (`.claude/` plugin) separate from data layer (`.sow/` repository) | Plugin updates don't affect repository state; clear upgrade path             |
| **Open Source MIT License**          | Public repository with permissive licensing                                       | Community contributions encouraged, commercial use allowed                   |
| **Single Maintainer Initially**      | Project maintained by individual developer                                        | Release cadence may be slower, documentation critical for community adoption |
| **Claude Code Plugin Marketplace**   | Distribution via Claude Code plugin system                                        | Installation/upgrade managed by plugin infrastructure                        |
| **Homebrew Distribution (CLI)**      | CLI distributed via Homebrew tap for easy installation                            | Requires maintaining Homebrew formula, GoReleaser automation                 |
| **Backward Compatibility for State** | State structure changes require migration tooling                                 | Protects users from breaking changes, increases maintenance burden           |
| **GitHub-Centric Workflow**          | Issue and PR management assumes GitHub                                            | Limited support for GitLab, Bitbucket, etc. (gh CLI constraint)              |

## Conventions

### Code Conventions
- **Error Handling**: Sentinel errors in `errors.go` files, wrapped with context via `fmt.Errorf`
- **Package Structure**: Domain-driven design with `internal/` packages (sow, project, prompts, etc.)
- **Testing**: `_test.go` suffix, table-driven tests, testify assertions
- **Naming**: Packages use lowercase single-word names, interfaces use `I` prefix only when ambiguous
- **Logging**: CLI logs to stdout/stderr, structured logs via `internal/logging` to markdown files

### API Conventions
- **CLI Commands**: Verb-noun structure (`sow project continue`, `sow design add-input`)
- **State Files**: YAML with snake_case field names (generated from CUE schemas)
- **Paths**: Relative paths within `.sow/`, absolute paths at repository root
- **Timestamps**: RFC3339 format for all timestamp fields

### Documentation Conventions
- **Architecture Docs**: Arc42 structure in `.sow/knowledge/architecture/arc42/`
- **ADRs**: Numbered ADR-NNN-title.md in `.sow/knowledge/adrs/`
- **Log Format**: Structured markdown with YAML frontmatter (timestamp, agent, action, result)
- **Code Comments**: Package doc comments in `doc.go` or first file, exported functions always documented
