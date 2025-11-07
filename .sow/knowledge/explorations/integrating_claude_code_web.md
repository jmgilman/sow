# Integrating sow with Claude Code Web: Exploration Summary

## Context

This exploration investigated how to make sow (AI-powered system of work) compatible with Claude Code's web environment, which provides browser-based coding in isolated, ephemeral VMs. The core challenge: sow currently assumes a persistent local development environment with installed tools (sow CLI, gh CLI) and relies on plugin marketplace installation for agents. Web VMs operate differently - they're temporary, sandboxed, and have limited pre-installed tools.

**Goals**:
1. Enable sow to work in Claude Code web VMs without manual setup
2. Preserve existing local CLI workflow (no breaking changes)
3. Maintain full functionality in both environments

## Executive Summary

**Sow can be fully integrated with Claude Code web** through four complementary solutions:

1. **CLI Installation**: Automated via SessionStart hooks that download and install sow binary on VM startup
2. **Agent Distribution**: Bundle agents with CLI and provide `sow claude init` command to create repository `.claude/` directories
3. **GitHub Integration**: Implement dual GitHub client (gh CLI + API) with automatic environment detection
4. **Zero-Setup Experience**: Combine all solutions into a SessionStart hook that makes web VMs "just work"

**Key Insight**: Web VMs require repository-committed configuration (`.claude/` directories, SessionStart hooks) rather than user-level installation, inverting sow's current distribution model while maintaining backward compatibility.

## Key Findings

### 1. Web VM Environment Characteristics

**What We Learned**:
- VMs are ephemeral, isolated Linux sandboxes that clone repositories fresh for each session
- Standard dev tools pre-installed (git, curl, Python, Node.js, Go) but not specialized CLIs (gh, sow)
- GitHub authentication via secure proxy (scoped credentials), not local gh auth
- Network access allowlisted (GitHub included, enabling binary downloads)
- SessionStart hooks execute during initialization, enabling automated setup

**Implication**: Tools must be installed on-demand, and configuration must live in the repository, not user home directories.

### 2. sow CLI Installation

**Problem**: sow CLI binary not present in web VMs.

**Solution**: SessionStart hooks + binary download
- Hook script detects web environment (`CLAUDE_CODE_REMOTE=true`)
- Downloads sow binary from GitHub releases (similar to existing install command pattern)
- Installs to `~/.local/bin` and adds to PATH via `CLAUDE_ENV_FILE`
- Idempotent (checks if already installed)

**Implementation Pattern**:
```bash
# scripts/install-sow.sh
if [ "$CLAUDE_CODE_REMOTE" = "true" ] && ! command -v sow &> /dev/null; then
  VERSION=$(curl -s https://api.github.com/repos/jmgilman/sow/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
  curl -L -o sow.tar.gz "https://github.com/jmgilman/sow/releases/download/${VERSION}/sow_${VERSION#v}_Linux_x86_64.tar.gz"
  tar -xzf sow.tar.gz
  mv sow ~/.local/bin/sow
  chmod +x ~/.local/bin/sow
  echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$CLAUDE_ENV_FILE"
fi
```

**Status**: Fully specified, ready to implement

**Reference**: Task 010 findings - `phases/exploration/tasks/010/findings.md`

### 3. Agent Distribution for Web VMs

**Problem**: sow's agents currently distributed via plugin marketplace (installed to `~/.claude/plugins/`), which doesn't work in ephemeral VMs.

**Discovery**: Claude Code supports **two mechanisms** for agents/commands:
1. **User-level plugins**: `~/.claude/plugins/` (marketplace installation) - local CLI only
2. **Repository-level**: `.claude/agents/` and `.claude/commands/` (cloned with repo) - works everywhere

**Solution**: Bundle agents with CLI, provide initialization command
- Embed `plugin/` directory in CLI binary using Go's `embed.FS` (same pattern as existing prompt templates)
- Add `sow claude init` command that writes embedded agents to `.claude/` directories
- SessionStart hook automatically runs `sow claude init` in web VMs
- Local users can choose plugin OR repository approach

**Benefits**:
- Web VMs: Zero installation, agents available immediately after clone
- Local CLI: Keep using plugin marketplace if preferred
- Repository `.claude/` takes precedence over user plugins (per Claude Code spec)
- ~5-6 files, <10KB - negligible binary size increase

**Implementation Checklist**:
1. Create `cli/internal/plugin/plugin.go` with `//go:embed ../../plugin`
2. Add `InitClaude()` function to copy embedded files to `.claude/`
3. Create `sow claude init` command
4. Update SessionStart hook to call `sow claude init` automatically

**Status**: Fully specified with code examples

**Reference**: Task 012 findings - `phases/exploration/tasks/012/findings.md`

### 4. GitHub Integration Challenge

**Problem**: sow uses `gh` CLI for all GitHub operations (issues, PRs, branch linking), which requires:
1. gh binary installed
2. gh authenticated (`gh auth login`)

Web VMs have GitHub authentication via proxy (for git operations), but this doesn't provide gh CLI access.

**Discovery**:
- GitHub proxy handles git-level operations (clone, push, PR creation) but not gh commands
- Installing gh is feasible but authenticating it is the blocker
- Claude Code web allows users to set environment variables via UI
- GitHub API provides equivalents for all gh operations sow uses

**Solution**: Dual implementation with auto-detection

**Architecture**:
```
GitHubClient (interface)
  ├── GitHubCLI (gh wrapper) - for local dev
  └── GitHubAPI (REST + GraphQL) - for web VMs

NewGitHubClient() factory:
  if GITHUB_TOKEN env var present → use API
  else → use gh CLI
```

**API Equivalents**:
- Issues/PRs: GitHub REST API (`go-github` library)
- Branch linking: GraphQL `createLinkedBranch` mutation (`githubv4` library)
- All operations have API equivalents (verified)

**User Experience**:
- **Web VM**: User sets `GITHUB_TOKEN` via Claude Code web UI → API mode
- **Local CLI**: User has gh authenticated → CLI mode (no change)
- **Explicit API**: Local user can set `GITHUB_TOKEN` to test API mode

**Migration Impact**:
- Extract `GitHubClient` interface from current `GitHub` struct
- Rename current implementation to `GitHubCLI`
- Add new `GitHubAPI` implementation
- Factory function chooses based on environment
- **Zero breaking changes** for consumers (interface methods identical)

**Dependencies**:
- `github.com/google/go-github/v66` - REST API
- `github.com/shurcooL/githubv4` - GraphQL
- `golang.org/x/oauth2` - Token auth
- ~3-4MB binary increase (acceptable)

**Status**: Complete design with code examples for interface, both implementations, and factory

**Reference**: Tasks 013 and 014 findings - `phases/exploration/tasks/013/findings.md`, `phases/exploration/tasks/014/findings.md`

### 5. Complete Integration Picture

**Combining all solutions** creates seamless web VM experience:

**Repository Setup** (one-time):
```
my-project/
├── .claude/
│   ├── settings.json         # SessionStart hook config
│   ├── agents/               # sow agents (implementer, planner, reviewer)
│   └── commands/             # sow slash commands
├── scripts/
│   └── install-sow.sh        # Automated installation script
└── .env.example              # Template showing GITHUB_TOKEN requirement
```

**SessionStart Hook** (`.claude/settings.json`):
```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          {
            "type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR\"/scripts/install-sow.sh"
          }
        ]
      }
    ]
  }
}
```

**Installation Script** (`scripts/install-sow.sh`):
```bash
#!/bin/bash
set -e

# Only run in web VMs
if [ "$CLAUDE_CODE_REMOTE" != "true" ]; then
  exit 0
fi

# Install sow CLI (task 010)
if ! command -v sow &> /dev/null; then
  # ... download and install sow binary ...
fi

# Initialize .claude directory (task 012)
if [ ! -d ".claude/agents" ]; then
  sow claude init
fi

# Done - sow is ready with agents available
```

**User Workflow**:
1. Clone repository in Claude Code web
2. SessionStart hook runs automatically (installs sow, creates `.claude/`)
3. User sets `GITHUB_TOKEN` environment variable via web UI
4. Start working: `sow project new`, agents available, GitHub operations work via API

**Zero manual steps** required beyond setting token.

## Research Topics Deep Dive

### Task 010: CLI Installation via Hooks

**Research Focus**: How to get sow binary onto fresh VMs

**Findings**:
- SessionStart hooks execute shell scripts during VM initialization
- Environment variables available: `CLAUDE_PROJECT_DIR`, `CLAUDE_CODE_REMOTE`, `CLAUDE_ENV_FILE`
- GitHub releases API provides latest version info
- Binary download/install follows same pattern as existing `/sow:install` command
- PATH persistence via `CLAUDE_ENV_FILE`

**Key Code Pattern**:
```bash
# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Fetch latest version
VERSION=$(curl -s https://api.github.com/repos/jmgilman/sow/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

# Download and install
curl -L -o sow.tar.gz "https://github.com/jmgilman/sow/releases/download/${VERSION}/sow_${VERSION#v}_${OS}_${ARCH}.tar.gz"
tar -xzf sow.tar.gz && mv sow ~/.local/bin/sow
```

**Validation**: Pattern verified against existing installation command, GoReleaser config, and web VM environment docs.

### Task 011: Plugin Distribution Mechanics

**Research Focus**: How agents get to web VMs (plugin marketplace vs repository)

**Critical Discovery**:
Claude Code has **two separate** agent/command discovery mechanisms:
1. **Marketplace plugins** → `~/.claude/plugins/` (user-level)
2. **Repository directories** → `.claude/agents/`, `.claude/commands/` (project-level)

**Why This Matters**:
- Web VMs are ephemeral - no persistent `~/.claude/` directory
- Marketplace installation requires user interaction (not available in VMs)
- Repository `.claude/` is **cloned automatically** and **works immediately**

**Implication**: For web compatibility, sow must shift from marketplace-only distribution to **also supporting repository-committed `.claude/` directories**.

**Solution Path**:
- Marketplace plugins remain for local CLI users (existing workflow)
- Add `sow claude init` to create repository `.claude/` from embedded templates
- Dual distribution: plugin marketplace (local) + repository (web)
- Project-level `.claude/` takes precedence over user-level plugins

**Validation**: Confirmed via Claude Code documentation on subagents and project-level commands.

### Task 012: Bundling Agents with CLI

**Research Focus**: How to embed plugin agents in CLI binary

**Finding**: sow **already uses this pattern** for templates and schemas

**Existing Embed Infrastructure**:
```go
// cli/internal/prompts/prompts.go
//go:embed templates
var FS embed.FS

// Usage:
templates.Render(prompts.FS, "templates/guidance/research.md", nil)
```

**Apply to Plugin Agents**:
```go
// cli/internal/plugin/plugin.go
//go:embed ../../plugin
var FS embed.FS

// Copy to repository:
func InitClaude(repoRoot string) error {
    // Create .claude/agents/ and .claude/commands/
    // Copy files from embedded FS
}
```

**Benefits**:
- Zero-installation: agents available immediately
- No marketplace dependency in web VMs
- Self-contained CLI binary
- Minimal size increase (~10KB for 5-6 agent files)

**Command Design**: `sow claude init`
- Separate command (not `sow init --claude`) for clarity
- Idempotent (checks if `.claude/` exists)
- Provides next steps (git add, commit, push)

**Validation**: Pattern matches existing template embedding; implementation straightforward.

### Task 013: GitHub CLI Dependency Analysis

**Research Focus**: Implications of gh CLI dependency for web VMs

**Current State**:
- sow uses gh CLI for **all** GitHub operations (ADR-010 decision)
- Rationale: leverage user's existing auth, simpler than API client
- Operations: list/create issues, link branches, create PRs

**Web VM Challenge**:
1. gh CLI likely **not pre-installed** (not in documented tool list)
2. Even if installed, **authentication is unclear** (proxy auth ≠ gh auth)
3. No interactive `gh auth login` available in VMs

**Discovery**:
- Web VM GitHub proxy handles **git operations** (clone, push, PR)
- This is **separate from gh CLI** - proxy doesn't provide gh access
- Claude Code web allows **environment variables via UI**
- Users can set `GITHUB_TOKEN` with Personal Access Token

**Key Distinction**:
- **git-level operations** (proxy): Clone, push, create PRs via git push
- **gh-specific operations** (CLI/API): List issues, create issues, link branches to issues

**Not all sow features need gh**:
- Core workflow (standard project): branch → code → push → PR ✅ (works via git proxy)
- Issue operations (breakdown, `--issue` flag): ❌ (requires gh or API)

**Solution Options Evaluated**:
- A: Make gh optional (core works, issue ops degrade)
- B: Install gh + user PAT (full functionality, setup burden)
- C: Dual implementation (gh CLI + API) ✅ **CHOSEN**
- D: Investigate proxy token (unknown viability)

**Recommendation**: Solution C with `GITHUB_TOKEN` environment variable.

### Task 014: GitHub Client Interface Design

**Research Focus**: How to implement dual GitHub client (gh CLI + API)

**Interface Design**:
```go
type GitHubClient interface {
    CheckAvailability() error
    ListIssues(label, state string) ([]Issue, error)
    GetIssue(number int) (*Issue, error)
    CreateIssue(title, body string, labels []string) (*Issue, error)
    GetLinkedBranches(number int) ([]LinkedBranch, error)
    CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error)
    CreatePullRequest(title, body string) (string, error)
}
```

**Two Implementations**:

1. **GitHubCLI**: Current code renamed
   - Wraps gh commands with exec
   - Uses `gh auth status` for authentication
   - Parses JSON output

2. **GitHubAPI**: New implementation
   - REST API for issues/PRs (`go-github` library)
   - GraphQL for branch linking (`githubv4` library)
   - Token-based authentication

**API Mapping Verified**:
- `gh issue list` → REST `/repos/{owner}/{repo}/issues` ✅
- `gh issue view` → REST `/repos/{owner}/{repo}/issues/{number}` ✅
- `gh issue create` → REST `POST /repos/{owner}/{repo}/issues` ✅
- `gh pr create` → REST `POST /repos/{owner}/{repo}/pulls` ✅
- `gh issue develop` → GraphQL `createLinkedBranch` mutation ✅

**Auto-Detection Factory**:
```go
func NewGitHubClient() (GitHubClient, error) {
    if os.Getenv("GITHUB_TOKEN") != "" {
        // API mode: Extract owner/repo from GITHUB_REPOSITORY or git remote
        return NewGitHubAPI(token, owner, repo), nil
    }
    // CLI mode: Use gh binary
    return NewGitHubCLI(exec.NewLocal("gh")), nil
}
```

**Migration Plan**:
1. Extract interface (refactor existing code)
2. Rename `GitHub` → `GitHubCLI`
3. Implement `GitHubAPI`
4. Add factory with detection
5. Update `Context` to use interface

**Impact**: Zero breaking changes, improved testability, ~10-15 hours implementation time.

## Recommendations

### Immediate Actions

1. **Implement CLI Installation Hook** (Task 010)
   - Create `scripts/install-sow.sh` with platform detection and binary download
   - Add `.claude/settings.json` with SessionStart hook configuration
   - Test in web VM environment

2. **Add `sow claude init` Command** (Task 012)
   - Embed `plugin/` directory in CLI using `//go:embed`
   - Create `InitClaude()` function to copy agents to `.claude/`
   - Add `claude init` subcommand to CLI
   - Update SessionStart hook to call it automatically

3. **Extract GitHub Client Interface** (Tasks 013, 014)
   - Create `GitHubClient` interface
   - Rename current implementation to `GitHubCLI`
   - Add compile-time interface compliance check
   - Verify all tests pass (no behavior change)

### Follow-Up Work

4. **Implement GitHub API Client** (Task 014)
   - Add dependencies: go-github, githubv4, oauth2
   - Create `GitHubAPI` with REST + GraphQL operations
   - Implement all interface methods
   - Add unit tests for API client

5. **Add Auto-Detection Factory** (Task 014)
   - Create `NewGitHubClient()` factory function
   - Implement environment detection (`GITHUB_TOKEN`)
   - Update `Context` creation to use factory
   - Add integration tests for both modes

6. **Create Documentation**
   - Web VM setup guide (setting `GITHUB_TOKEN`)
   - Example `.env` file template
   - SessionStart hook configuration guide
   - Migration guide for existing users

### Testing Strategy

**Unit Tests**:
- CLI installation script (mock curl/tar)
- `InitClaude()` function (embedded FS copy)
- GitHub API client (mock HTTP responses)
- Factory detection logic

**Integration Tests**:
- End-to-end in web VM (manual verification)
- Local with gh CLI (existing test suite)
- Local with API mode (set GITHUB_TOKEN)

**Validation**:
- Clone sow-enabled repo in Claude Code web
- Verify SessionStart hook installs CLI
- Verify `.claude/` agents available
- Verify GitHub operations work with token

### Rollout Plan

**Phase 1: Foundation** (Low Risk)
- Extract GitHub client interface
- Embed plugin agents in CLI
- Add `sow claude init` command
- **Deliverable**: New capabilities, zero breaking changes

**Phase 2: Web VM Support** (Medium Risk)
- Implement GitHub API client
- Add factory with auto-detection
- Create SessionStart hook
- **Deliverable**: sow works in web VMs

**Phase 3: Documentation** (Low Risk)
- Write setup guides
- Create example configurations
- Update README with web VM instructions
- **Deliverable**: Clear user onboarding

**Timeline**:
- Phase 1: ~15-20 hours development
- Phase 2: ~15-20 hours development + testing
- Phase 3: ~5-10 hours documentation
- **Total**: ~35-50 hours (~1-2 weeks)

### Success Criteria

Web VM support is successful when:
1. ✅ User clones sow-enabled repository in Claude Code web
2. ✅ SessionStart hook runs automatically (no manual intervention)
3. ✅ sow CLI installed and accessible
4. ✅ Agents available in `.claude/` directory
5. ✅ User sets `GITHUB_TOKEN` via web UI
6. ✅ `sow project new` works with full functionality
7. ✅ GitHub operations (issues, PRs) work via API
8. ✅ Local CLI users see no changes (backward compatible)

## Technical Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    Claude Code Web VM                    │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  1. SessionStart Hook                                    │
│     └─> install-sow.sh                                   │
│         ├─> Download sow binary from GitHub releases    │
│         ├─> Install to ~/.local/bin                      │
│         └─> Run: sow claude init                         │
│                                                          │
│  2. sow CLI (embedded agents)                            │
│     └─> InitClaude()                                     │
│         └─> Copy plugin/* to .claude/                    │
│                                                          │
│  3. .claude/ (repository-committed)                      │
│     ├─> agents/                                          │
│     │   ├─> implementer.md                               │
│     │   ├─> planner.md                                   │
│     │   └─> reviewer.md                                  │
│     └─> commands/                                        │
│                                                          │
│  4. GitHub Integration                                   │
│     └─> GitHubClient (interface)                         │
│         └─> GitHubAPI (GITHUB_TOKEN present)             │
│             ├─> REST API (issues, PRs)                   │
│             └─> GraphQL (branch linking)                 │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### Data Flow

**VM Initialization**:
1. User opens repository in Claude Code web
2. VM clones repository (includes `.claude/settings.json`, `scripts/install-sow.sh`)
3. SessionStart hook fires → runs `install-sow.sh`
4. Script installs sow CLI binary
5. Script runs `sow claude init` (creates `.claude/` if missing)
6. Agents available, sow ready

**Runtime Operation**:
1. User invokes sow command: `sow project new`
2. sow context initialization → `NewGitHubClient()` factory
3. Factory detects `GITHUB_TOKEN` → creates `GitHubAPI` instance
4. GitHub operations use REST/GraphQL instead of gh CLI
5. Everything works transparently

**Local CLI** (no changes):
1. User has gh CLI authenticated
2. `NewGitHubClient()` → no `GITHUB_TOKEN` → creates `GitHubCLI`
3. Existing behavior preserved

## References

### Detailed Findings

All research tasks contain comprehensive findings with code examples, API documentation, and implementation patterns:

- **[Task 010](phases/exploration/tasks/010/findings.md)**: CLI installation via SessionStart hooks
- **[Task 011](phases/exploration/tasks/011/findings.md)**: Plugin distribution mechanics for web VMs
- **[Task 012](phases/exploration/tasks/012/findings.md)**: Bundling agents with CLI and init command
- **[Task 013](phases/exploration/tasks/013/findings.md)**: gh CLI dependency analysis
- **[Task 014](phases/exploration/tasks/014/findings.md)**: GitHub client interface design

### External Documentation

**Claude Code**:
- [Claude Code on the Web](https://code.claude.com/docs/en/claude-code-on-the-web) - Web environment, SessionStart hooks, GitHub proxy
- [Plugin Marketplaces](https://code.claude.com/docs/en/plugin-marketplaces) - Plugin installation and repository `.claude/` directories
- [Subagents](https://code.claude.com/docs/en/sub-agents) - Agent discovery from `.claude/agents/`

**GitHub**:
- [GitHub CLI](https://cli.github.com/) - Installation and authentication
- [go-github](https://pkg.go.dev/github.com/google/go-github/v66/github) - REST API client
- [githubv4](https://pkg.go.dev/github.com/shurcooL/githubv4) - GraphQL client
- [GitHub GraphQL API](https://docs.github.com/en/graphql) - createLinkedBranch mutation

**Go**:
- [embed Package](https://pkg.go.dev/embed) - Embedding files in binaries
- [oauth2 Package](https://pkg.go.dev/golang.org/x/oauth2) - Token authentication

### Code Locations

**Relevant sow files**:
- `cli/internal/sow/github.go` - Current GitHub client (to be refactored)
- `cli/internal/prompts/prompts.go` - Embed pattern example
- `cli/cmd/init.go` - Init command pattern
- `plugin/agents/*.md` - Agents to embed

**New files to create**:
- `cli/internal/plugin/plugin.go` - Embedded plugin files
- `cli/internal/sow/github_client.go` - GitHubClient interface
- `cli/internal/sow/github_cli.go` - gh CLI implementation
- `cli/internal/sow/github_api.go` - API implementation
- `cli/internal/sow/github_factory.go` - Auto-detection factory
- `cli/cmd/claude.go` - `sow claude` commands
- `scripts/install-sow.sh` - Installation script
- `.claude/settings.json` - SessionStart hook config

## Conclusion

Integrating sow with Claude Code web is **fully feasible** with the four solutions identified. The architecture preserves backward compatibility for local CLI users while enabling a seamless web VM experience.

**Key Success Factors**:
1. **Leverage existing patterns**: SessionStart hooks, embed.FS, interface abstraction
2. **Incremental implementation**: Each component can be built and tested independently
3. **User experience first**: Zero manual setup for web VMs, no changes for local CLI
4. **Clear migration path**: Well-defined phases with measurable progress

The exploration revealed that web VMs require an **inverted distribution model** - instead of user-level installation (marketplace plugins, local tools), everything must be **repository-committed and auto-installed**. This insight drives all four solutions and creates a coherent integration strategy.

**Next Steps**: Move to implementation, starting with Phase 1 (foundation) to establish patterns and validate approach.
