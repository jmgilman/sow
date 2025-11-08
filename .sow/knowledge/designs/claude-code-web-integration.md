# Architecture Design: sow Integration with Claude Code Web

## Context

sow is an AI-powered framework for structured software development that coordinates work through phases using state machines and specialized agents. Currently, sow assumes a persistent local development environment with installed tools (sow CLI, gh CLI) and relies on Claude Code's plugin marketplace for agent distribution.

Claude Code web provides browser-based coding in isolated, ephemeral Linux VMs. These environments differ fundamentally from local development:

- **Ephemeral**: Fresh VM for each session, no persistent user directories
- **Sandboxed**: Limited pre-installed tools (git, curl, Python, Node.js, Go)
- **Repository-cloned**: Each session starts with a clean repository clone
- **GitHub authenticated**: Secure proxy for git operations, not general CLI access

This architecture describes how to make sow fully functional in Claude Code web environments while maintaining backward compatibility with local CLI workflows.

## Goals

1. **Zero-setup experience**: Users clone repository and start working, no manual installation
2. **Full functionality**: All sow features work in web VMs (issues, PRs, branch linking)
3. **Backward compatibility**: No breaking changes for local CLI users
4. **Maintainability**: Single codebase supports both environments with minimal complexity

## System Architecture

### High-Level Overview

The integration consists of five complementary solutions:

```
┌─────────────────────────────────────────────────────────────┐
│                   Claude Code Web VM                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │ 1. Automated Installation (SessionStart Hook)      │    │
│  │    - Detects web environment                       │    │
│  │    - Downloads sow binary from GitHub releases     │    │
│  │    - Installs to ~/.local/bin                      │    │
│  │    - Persists PATH via CLAUDE_ENV_FILE             │    │
│  └────────────────────────────────────────────────────┘    │
│                            ↓                                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │ 2. Agent Distribution (Embedded + Init)            │    │
│  │    - Agents embedded in CLI binary (embed.FS)      │    │
│  │    - sow claude init creates .claude/ directories  │    │
│  │    - Repository-committed agents                   │    │
│  └────────────────────────────────────────────────────┘    │
│                            ↓                                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │ 3. Claude Code Integration                         │    │
│  │    - .claude/agents/ (implementer, planner)        │    │
│  │    - .claude/commands/ (slash commands)            │    │
│  │    - Available immediately after clone             │    │
│  └────────────────────────────────────────────────────┘    │
│                            ↓                                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │ 4. GitHub Integration (Dual Client)                │    │
│  │    - GitHubClient interface                        │    │
│  │    - GitHubAPI (REST + GraphQL) for web VMs        │    │
│  │    - GitHubCLI (gh wrapper) for local dev          │    │
│  │    - Auto-detection via GITHUB_TOKEN               │    │
│  └────────────────────────────────────────────────────┘    │
│                            ↓                                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │ 5. Project Initialization (Dual Paths)             │    │
│  │    - sow project init (in-place, no worktree)      │    │
│  │    - General mode prompt (operator ↔ orchestrator) │    │
│  │    - Smart SessionStart prompt selection           │    │
│  │    - Mode switching based on .sow/project/         │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Component Architecture

#### 1. Automated Installation

**Purpose**: Get sow CLI binary onto fresh VMs automatically

**Components** (created by `sow claude init`):
- `.sow/scripts/install-sow.sh`: Shell script for platform detection and binary download (embedded in CLI)
- `.claude/settings.json`: SessionStart hook configuration (embedded in CLI)
- Environment detection via `CLAUDE_CODE_REMOTE=true`

**Flow**:
```
SessionStart hook fires
    ↓
Check CLAUDE_CODE_REMOTE
    ↓
Detect platform (uname)
    ↓
Fetch latest release from GitHub API
    ↓
Download platform-specific tarball
    ↓
Extract and install to ~/.local/bin
    ↓
Persist PATH via CLAUDE_ENV_FILE
    ↓
Run sow claude init
```

**Key Design Decisions**:
- Embedded in CLI binary (no external dependencies for users)
- Use existing GitHub releases infrastructure (no new distribution)
- Idempotent script (safe to run multiple times)
- Platform-aware (Linux x86_64 for web VMs, others for local testing)
- Minimal dependencies (curl, tar, standard POSIX tools)
- Created by `sow claude init` (zero manual configuration)

#### 2. Agent Distribution

**Purpose**: Make sow agents available without plugin marketplace

**Components**:
- `cli/internal/claude/claude.go`: Embedded agent files using `//go:embed`
- `sow claude init` command: Copies embedded agents to `.claude/`
- Repository `.claude/` directories: Committed agent definitions

**Embedded Content**:
```
cli/internal/claude/
├── claude.go              # Implementation of sow claude init
├── agents/                # Embedded agent files
│   ├── implementer.md
│   ├── planner.md
│   └── reviewer.md
├── commands/              # Embedded command files
│   ├── implement-feature.md
│   ├── fix-bug.md
│   └── create-adr.md
├── prompts/               # Embedded prompt templates
│   ├── general.md         # General mode operator prompt
│   ├── standard.md        # Standard project orchestrator prompt
│   ├── design.md          # Design project orchestrator prompt
│   └── exploration.md     # Exploration project orchestrator prompt
├── scripts/               # Embedded scripts
│   ├── install-sow.sh           # CLI installation
│   └── session-start-prompt.sh  # Smart prompt selector
└── config/                # Embedded hook configuration
    └── settings.json      # SessionStart hook template
```

**Embedding Pattern** (existing pattern from templates):
```go
// cli/internal/claude/claude.go
//go:embed agents commands prompts scripts config
var FS embed.FS

func Init(repoRoot string) error {
    // Create .claude/agents/ and .claude/commands/
    // Create .sow/scripts/install-sow.sh
    // Create .sow/scripts/session-start-prompt.sh
    // Create or merge .claude/settings.json with SessionStart hook
    // Copy all files from embedded FS to repository
    // Make idempotent (skip if exists, merge settings)
}
```

**Key Design Decisions**:
- All files embedded in CLI binary (agents, commands, scripts, hooks)
- Single command setup: `sow claude init` creates everything
- Dual distribution: Marketplace (local) + repository (web)
- Repository `.claude/` takes precedence (per Claude Code spec)
- Minimal size impact (~15KB total for all embedded files)
- Self-contained CLI binary (zero external dependencies)

#### 3. Claude Code Integration

**Purpose**: Provide sow agents and commands in Claude Code environment

**Repository Structure** (created by `sow claude init`):
```
repository/
├── .claude/
│   ├── settings.json          # SessionStart hook config (embedded)
│   ├── agents/                # sow agents (embedded)
│   │   ├── implementer.md
│   │   ├── planner.md
│   │   └── reviewer.md
│   └── commands/              # sow slash commands (embedded)
│       ├── implement-feature.md
│       ├── fix-bug.md
│       └── create-adr.md
└── .sow/
    └── scripts/
        └── install-sow.sh     # Installation automation (embedded)
```

**Discovery Mechanism**:
- Claude Code scans `.claude/agents/` for agent definitions
- Repository agents available immediately (no installation)
- Precedence: Repository > User plugins > Built-in

**Key Design Decisions**:
- Repository-committed configuration (survives VM recreation)
- Zero installation for agents (just clone)
- Consistent with Claude Code's repository-level customization model

#### 4. GitHub Integration (Dual Client)

**Purpose**: Enable GitHub operations in both web VMs and local CLI

**Interface Design**:
```go
type GitHubClient interface {
    CheckAvailability() error
    ListIssues(label, state string) ([]Issue, error)
    GetIssue(number int) (*Issue, error)
    CreateIssue(title, body string, labels []string) (*Issue, error)
    GetLinkedBranches(number int) ([]LinkedBranch, error)
    CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error)
    CreatePullRequest(title, body string, draft bool) (number int, url string, error)
    UpdatePullRequest(number int, title, body string) error
    MarkPullRequestReady(number int) error
}
```

**Two Implementations**:

1. **GitHubCLI** (local development):
   - Wraps `gh` commands via exec
   - Uses `gh auth status` for authentication check
   - Parses JSON output from gh commands
   - Current implementation (renamed)

2. **GitHubAPI** (web VMs):
   - REST API for issues/PRs (go-github library)
   - GraphQL for branch linking (githubv4 library)
   - Token-based authentication via `GITHUB_TOKEN`
   - New implementation

**Auto-Detection Factory**:
```go
func NewGitHubClient() (GitHubClient, error) {
    if token := os.Getenv("GITHUB_TOKEN"); token != "" {
        // API mode: Extract owner/repo from git remote
        owner, repo := getRepoInfo()
        return NewGitHubAPI(token, owner, repo), nil
    }
    // CLI mode: Use gh binary
    return NewGitHubCLI(exec.NewLocal("gh")), nil
}
```

**API Mapping**:
| Operation | gh CLI | GitHub API |
|-----------|--------|------------|
| List issues | `gh issue list` | REST `/repos/{owner}/{repo}/issues` |
| Get issue | `gh issue view` | REST `/repos/{owner}/{repo}/issues/{number}` |
| Create issue | `gh issue create` | REST `POST /repos/{owner}/{repo}/issues` |
| Create PR | `gh pr create` | REST `POST /repos/{owner}/{repo}/pulls` |
| Create draft PR | `gh pr create --draft` | REST `POST /repos/{owner}/{repo}/pulls` with `draft: true` |
| Update PR | `gh pr edit N` | REST `PATCH /repos/{owner}/{repo}/pulls/{number}` |
| Mark PR ready | `gh pr ready N` | REST `PATCH /repos/{owner}/{repo}/pulls/{number}` with `draft: false` |
| Link branch | `gh issue develop` | GraphQL `createLinkedBranch` mutation |

**Key Design Decisions**:
- Interface abstraction (enables dual implementation)
- Environment-based auto-detection (no configuration needed)
- Zero breaking changes (interface preserves existing API)
- Improved testability (mock interface, not exec calls)

### 5. Project Initialization for Web VMs

**Purpose**: Enable project creation in web VMs where worktrees and Claude Code launching don't apply

**The Challenge**:

Web VMs fundamentally differ from local development in how projects start:

| Aspect | Local CLI | Web VM |
|--------|-----------|--------|
| **Starting point** | User runs `sow project new` | User enters prompt in web UI |
| **Branch creation** | CLI creates branch + worktree | VM clones existing branch |
| **Claude launch** | CLI launches Claude Code with prompt | Claude Code already running |
| **Worktrees** | Yes - `.sow/worktrees/<branch>` | No - works on cloned branch directly |

**Current State**:

Currently, `sow project new` assumes local workflow:
- Creates git worktree
- Initializes `.sow/project/` in the worktree
- Launches Claude Code with project-specific prompt

This doesn't work in web VMs where:
- Claude Code is already running (can't launch it)
- No worktrees (repository cloned directly on branch)
- User's initial prompt comes from web UI, not CLI

**Proposed Solution: Dual Initialization Paths**

**Path 1: Local Development** (existing + proposed enhancement)

*Current:* `sow project new` with command-line flags
```bash
sow project new \
  --type standard \
  --name "my-project" \
  --desc "Description" \
  --issue 123
```

*Proposed:* Interactive wizard (see `.sow/knowledge/designs/interactive-project-launch/`)
```bash
sow project
# Interactive prompts guide user through:
# - Project type selection
# - Name/description input
# - Issue linking
# → Creates worktree
# → Launches Claude Code with project prompt
```

**Path 2: Web VM** (new requirement)

New command: `sow project init` for in-place initialization
```bash
sow project init \
  --type standard \
  --name "implement-jwt-auth" \
  --desc "Add JWT authentication" \
  [--issue N]
```

**Characteristics**:
- **Always** works on current branch (no worktree creation)
- **Never** launches Claude Code (already running)
- **Programmatic** (all options via flags, no interactivity)
- Creates `.sow/project/` in current repository location

**Usage in Web VMs**:

1. User opens repository in Claude Code web with prompt: "Implement JWT authentication"
2. SessionStart hook runs, installs sow, provides "general mode" prompt
3. Claude analyzes user's prompt and suggests: "This looks like a feature. Want to create a structured project?"
4. User confirms
5. Claude runs: `sow project init --type standard --name "jwt-auth" --desc "Implement JWT authentication"`
6. `.sow/project/` created on current branch
7. Mode automatically switches from operator to orchestrator

**Smart SessionStart Prompt**:

The SessionStart hook provides a context-aware prompt:

```bash
# Embedded in cli/internal/claude/scripts/session-start-prompt.sh
if [ -d ".sow/project" ]; then
    # Project exists - load project-specific orchestrator prompt
    TYPE=$(grep "^type:" .sow/project/state.yaml | cut -d: -f2 | tr -d ' ')
    sow prompt project/$TYPE
else
    # No project - load general mode operator prompt
    sow prompt general
fi
```

**General Mode Prompt** (embedded at `cli/internal/claude/prompts/general.md`):

```markdown
You are the sow operator. The user's prompt: "{USER_PROMPT}"

Analyze the prompt and determine if this would benefit from structured orchestration:
- Feature implementation → standard project
- Design/architecture work → design project
- Research/investigation → exploration project
- Complex feature decomposition → breakdown project

If structured orchestration would help, ask the user if they want to create a project.
If yes, run:

sow project init --type TYPE --name NAME --desc "DESCRIPTION"

This will create .sow/project/ and you'll switch to orchestrator mode.

If the user prefers to work without orchestration, proceed as a helpful coding assistant.
```

**Implementation Requirements**:

1. **New command**: `sow project init` with in-place semantics
2. **Conditional worktree logic**: Detect whether to use worktrees
3. **General mode prompt**: Embedded template for operator mode
4. **Smart prompt selector**: SessionStart script that chooses appropriate prompt
5. **Project prompts**: Embedded templates for each project type (already proposed)

**Key Design Decisions**:
- Clear separation: `new` = interactive + worktree, `init` = programmatic + in-place
- Mode switching automatic (based on `.sow/project/` existence)
- General mode enables both operator work and project initialization
- Single codebase supports both flows with minimal branching

## Integration Patterns

### Pattern 1: Web VM Initialization

```
User opens repository in Claude Code web
    ↓
VM clones repository (includes .claude/, scripts/)
    ↓
SessionStart hook fires → runs install-sow.sh
    ↓
Script detects CLAUDE_CODE_REMOTE=true
    ↓
Downloads and installs sow binary
    ↓
Runs sow claude init (creates .claude/ if missing)
    ↓
Agents available, sow ready
```

### Pattern 2: GitHub Operation (Web VM)

```
User runs: sow project new --issue 123
    ↓
sow initializes context
    ↓
NewGitHubClient() factory called
    ↓
Detects GITHUB_TOKEN environment variable
    ↓
Creates GitHubAPI instance
    ↓
Calls GetIssue(123) → REST API request
    ↓
Returns issue data transparently
```

### Pattern 3: GitHub Operation (Local CLI)

```
User runs: sow project new --issue 123
    ↓
sow initializes context
    ↓
NewGitHubClient() factory called
    ↓
No GITHUB_TOKEN, checks for gh CLI
    ↓
Creates GitHubCLI instance
    ↓
Calls GetIssue(123) → gh issue view 123
    ↓
Returns issue data (existing behavior)
```

### Pattern 4: Agent Invocation

```
Orchestrator spawns implementer agent
    ↓
Claude Code searches for agent:
  1. Check .claude/agents/implementer.md (✓ found)
  2. Check ~/.claude/plugins/sow/agents/implementer.md (web: N/A)
  3. Built-in agents
    ↓
Loads .claude/agents/implementer.md
    ↓
Agent executes with sow-specific prompts
```

## Data Flows

### Web VM Environment Flow

```
┌─────────────┐
│   User      │
└──────┬──────┘
       │ Opens repository in Claude Code web
       ↓
┌─────────────────────────────────────────┐
│  Ephemeral Linux VM                     │
├─────────────────────────────────────────┤
│  1. Clone repository                    │
│     - .claude/settings.json             │
│     - .claude/agents/                   │
│     - .sow/scripts/install-sow.sh       │
│  ┌────────────────────────────────────┐ │
│  │ 2. SessionStart Hook               │ │
│  │    - install-sow.sh executes       │ │
│  │    - Downloads sow binary          │ │
│  │    - Installs to ~/.local/bin      │ │
│  └────────────────────────────────────┘ │
│  ┌────────────────────────────────────┐ │
│  │ 3. sow Ready                       │ │
│  │    - CLI in PATH                   │ │
│  │    - Agents already available      │ │
│  │    - GitHub via API (GITHUB_TOKEN) │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
       │
       ↓
┌─────────────────────────────────────────┐
│  sow Workflow                           │
│  - sow project new                      │
│  - Orchestrator spawns agents           │
│  - GitHub operations via REST/GraphQL   │
│  - Agents read/write code               │
└─────────────────────────────────────────┘
```

### Local CLI Environment Flow

```
┌─────────────┐
│   User      │
└──────┬──────┘
       │ Has sow installed locally
       ↓
┌─────────────────────────────────────────┐
│  Local Development Environment          │
├─────────────────────────────────────────┤
│  - sow CLI installed (homebrew/binary)  │
│  - gh CLI authenticated                 │
│  - Agents via marketplace plugins       │
│  - (Optional) .claude/ in repository    │
└─────────────────────────────────────────┘
       │
       ↓
┌─────────────────────────────────────────┐
│  sow Workflow                           │
│  - sow project new                      │
│  - Orchestrator spawns agents           │
│  - GitHub operations via gh CLI         │
│  - Agents read/write code               │
└─────────────────────────────────────────┘

Note: Identical workflow, different implementation
```

## Deployment Model

### Repository Setup (One-Time)

**For sow-enabled repositories**:

1. **Initialize Claude Code integration**:
   ```bash
   sow claude init
   ```

   This single command creates everything needed:
   - `.claude/agents/` - sow agents (implementer, planner, reviewer)
   - `.claude/commands/` - sow slash commands
   - `.claude/settings.json` - SessionStart hook configuration
   - `.sow/scripts/install-sow.sh` - Automated installation script

   **What the installation script does**:
   ```bash
   # Embedded at cli/internal/claude/scripts/install-sow.sh
   #!/bin/bash
   set -e

   # Only run in web VMs
   if [ "$CLAUDE_CODE_REMOTE" != "true" ]; then
     exit 0
   fi

   # Install sow CLI if not present
   if ! command -v sow &> /dev/null; then
     # Download and install sow binary from GitHub releases
     # Add to PATH via CLAUDE_ENV_FILE
   fi
   ```

   **What the SessionStart hook does**:
   ```json
   // Embedded at cli/internal/claude/config/settings.json
   {
     "hooks": {
       "SessionStart": [
         {
           "matcher": "startup",
           "hooks": [
             {
               "type": "command",
               "command": "\"$CLAUDE_PROJECT_DIR\"/.sow/scripts/install-sow.sh"
             }
           ]
         }
       ]
     }
   }
   ```

2. **Commit and push**:
   ```bash
   git add .claude/ .sow/
   git commit -m "Add Claude Code web support for sow"
   git push
   ```

**That's it!** The repository is now ready for Claude Code web.

### User Workflow

**Web VM**:
1. Open repository in Claude Code web
2. Set `GITHUB_TOKEN` environment variable (via web UI)
3. Start working - everything auto-configured

**Local CLI** (unchanged):
1. Install sow (existing methods)
2. Authenticate gh CLI (existing flow)
3. Start working - existing workflow

### Environments Supported

| Environment | sow CLI | Agents | GitHub | Setup |
|-------------|---------|--------|--------|-------|
| **Web VM** | Auto-install | Repository | API | Set GITHUB_TOKEN |
| **Local (marketplace)** | Manual install | Plugins | gh CLI | None (existing) |
| **Local (repository)** | Manual install | Repository | gh CLI | sow claude init |

## Success Criteria

Integration is successful when:

1. ✅ **Zero-setup web VMs**: User clones repository, SessionStart hook completes, sow ready
2. ✅ **Full functionality**: All sow commands work (project new, issue linking, PR creation)
3. ✅ **Transparent operation**: User doesn't notice difference between web/local
4. ✅ **Backward compatible**: Local CLI users see no changes
5. ✅ **Single codebase**: No web-specific forks or branches
6. ✅ **Maintainable**: Clear separation of concerns, testable components

### Performance Targets

- SessionStart hook completion: <30 seconds (binary download + init)
- GitHub API operations: <2 seconds (vs gh CLI: <1 second)
- Binary size increase: <5MB (embed.FS + go-github)
- Zero runtime overhead for local CLI

## Risk Mitigation

### Risk: SessionStart Hook Failures

**Mitigation**:
- Idempotent script (safe retries)
- Early exit if not web VM
- Clear error messages
- Fallback: Manual `sow claude init`

### Risk: GitHub API Rate Limits

**Mitigation**:
- User-provided token (higher rate limits)
- Caching where appropriate
- Rate limit error handling
- Fallback guidance in errors

### Risk: Binary Size Growth

**Mitigation**:
- Minimal dependencies (go-github, githubv4)
- No unnecessary includes
- Monitor binary size in CI
- Target: <5MB increase

### Risk: Maintenance Burden

**Mitigation**:
- Interface abstraction (swap implementations)
- Comprehensive test coverage
- Clear documentation
- Single code path for business logic

## Implementation Phases

### Phase 1: Foundation (Low Risk)

**Goals**: Establish patterns, zero breaking changes

**Tasks**:
1. Extract GitHubClient interface from existing code
2. Rename current implementation to GitHubCLI
3. Add compile-time interface compliance checks
4. Embed all files in CLI (`//go:embed agents commands prompts scripts config`)
5. Create `sow claude init` command (copies embedded files to repository)
6. Create installation script template (`cli/internal/claude/scripts/install-sow.sh`)
7. Create session start prompt selector (`cli/internal/claude/scripts/session-start-prompt.sh`)
8. Create SessionStart hook template (`cli/internal/claude/config/settings.json`)
9. Create general mode prompt template (`cli/internal/claude/prompts/general.md`)
10. Create project-specific prompt templates (`cli/internal/claude/prompts/*.md`)
11. Create `sow project init` command (in-place initialization, no worktree)
12. Verify existing tests pass

**Deliverable**: New capabilities, existing workflow unchanged

### Phase 2: Web VM Support (Medium Risk)

**Goals**: Enable full functionality in web VMs

**Tasks**:
1. Implement GitHubAPI (REST + GraphQL)
2. Add factory with auto-detection
3. Test `sow claude init` creates all necessary files
4. Test `sow project init` creates project in-place
5. Test general mode → orchestrator mode transition
6. Integration testing in web VM
7. Verify SessionStart hook installs sow correctly
8. Verify smart prompt selector works

**Deliverable**: sow works in Claude Code web

### Phase 3: Documentation (Low Risk)

**Goals**: Clear user onboarding and developer guidance

**Tasks**:
1. Write web VM setup guide
2. Create example configurations
3. Update README with web support
4. Document GitHub API implementation
5. Migration guide for repository owners

**Deliverable**: Complete documentation

## Testing Strategy

### Unit Tests

- Installation script (mock curl/tar)
- `claude.Init()` function (embedded FS copy)
- GitHub API client (mock HTTP responses)
- Factory detection logic
- Interface compliance

### Integration Tests

- End-to-end in web VM (manual verification)
- Local with gh CLI (existing test suite)
- Local with API mode (GITHUB_TOKEN set)
- SessionStart hook execution

### Validation Checklist

- [ ] Clone sow-enabled repo in Claude Code web
- [ ] Verify SessionStart hook installs CLI
- [ ] Verify `.claude/` agents available
- [ ] Verify general mode prompt loads when no project exists
- [ ] Verify `sow project init` creates project in-place
- [ ] Verify mode switches to orchestrator after init
- [ ] Verify issue linking works via API
- [ ] Verify PR creation works (draft, update, mark ready)
- [ ] Verify local CLI unchanged (interactive wizard still works)

## Open Questions

### Resolved

- ✅ Can we download binaries in web VMs? **Yes** (allowlisted)
- ✅ Can we persist PATH? **Yes** (CLAUDE_ENV_FILE)
- ✅ Do repository `.claude/` directories work? **Yes** (documented)
- ✅ Do all GitHub operations have API equivalents? **Yes** (verified)

### Pending

- ⏳ Should we cache GitHub API responses to reduce rate limit impact?
- ⏳ Should we support multiple token sources (env, file, keychain)?
- ⏳ Should `sow claude init` be idempotent or error if `.claude/` exists?

## References

### Related Documents

- Exploration Summary: `.sow/knowledge/explorations/integrating_claude_code_web.md`
- Task 010 Findings: CLI installation patterns
- Task 012 Findings: Agent embedding and init command
- Task 014 Findings: GitHub client interface design

### External Documentation

- [Claude Code on the Web](https://code.claude.com/docs/en/claude-code-on-the-web)
- [Plugin Marketplaces](https://code.claude.com/docs/en/plugin-marketplaces)
- [go-github](https://pkg.go.dev/github.com/google/go-github/v66/github)
- [githubv4](https://pkg.go.dev/github.com/shurcooL/githubv4)
- [GitHub GraphQL API](https://docs.github.com/en/graphql)

### Code Locations

**Existing**:
- `cli/internal/sow/github.go` - Current GitHub client
- `cli/internal/prompts/prompts.go` - Embed pattern example
- `plugin/agents/*.md` - Marketplace agents (separate from embedded)
- `cli/cmd/project.go` - Current project commands

**New**:
- `cli/internal/claude/claude.go` - Init command implementation
- `cli/internal/claude/agents/*.md` - Agent files to embed
- `cli/internal/claude/commands/*.md` - Command files to embed
- `cli/internal/claude/prompts/general.md` - General mode operator prompt
- `cli/internal/claude/prompts/*.md` - Project-specific orchestrator prompts
- `cli/internal/claude/scripts/install-sow.sh` - Installation script to embed
- `cli/internal/claude/scripts/session-start-prompt.sh` - Smart prompt selector
- `cli/internal/claude/config/settings.json` - Hook configuration to embed
- `cli/internal/sow/github_client.go` - Interface
- `cli/internal/sow/github_cli.go` - CLI implementation
- `cli/internal/sow/github_api.go` - API implementation
- `cli/internal/sow/github_factory.go` - Auto-detection
- `cli/cmd/claude.go` - sow claude commands
- `cli/cmd/project_init.go` - sow project init (in-place initialization)

## Conclusion

This architecture enables sow to work seamlessly in both Claude Code web VMs and local CLI environments through five complementary solutions: automated installation, embedded agent distribution, Claude Code integration, dual GitHub client implementation, and dual project initialization paths.

**Key Innovations**:

1. **Everything embedded**: All files needed for web VM support (agents, commands, prompts, scripts, hooks) are embedded in the CLI binary
2. **Single command setup**: `sow claude init` creates complete repository configuration
3. **Dual initialization paths**:
   - Local: `sow project new` (interactive wizard, worktree, launches Claude)
   - Web VM: `sow project init` (programmatic, in-place, no launch)
4. **Smart mode switching**: General mode prompt enables seamless operator ↔ orchestrator transitions based on `.sow/project/` existence

The design prioritizes:
- **User experience**: Single command setup, zero manual configuration, context-aware prompts
- **Compatibility**: No breaking changes for local CLI, interactive wizard proposal preserved
- **Maintainability**: Clear abstractions, single codebase, all files embedded
- **Flexibility**: Supports both operator (ad-hoc) and orchestrator (structured) workflows in web VMs
- **Incremental delivery**: Independent phases with measurable progress

By leveraging existing patterns (SessionStart hooks, embed.FS, interface abstraction) and respecting environment boundaries, we achieve full functionality in ephemeral web VMs while preserving the robust local development experience.
