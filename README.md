# sow - System of Work

> AI-powered framework for structured software development with Claude Code

**sow** is an opinionated framework that coordinates specialized AI agents through a 5-phase workflow, combining human-led planning with AI-autonomous execution for building software features.

## Features

- 🤖 **Multi-agent orchestration** - Specialized workers (researcher, architect, planner, implementer, reviewer) coordinated by an orchestrator agent
- 📋 **5-phase lifecycle** - Discovery → Design → Implementation → Review → Finalize with intelligent phase selection
- 🧠 **Human-AI collaboration** - Humans lead planning (discovery/design), AI executes autonomously (implementation/review/finalize)
- 💾 **Zero-context resumability** - Resume work anytime from disk state without conversation history
- 🔗 **External knowledge integration** - Reference style guides, conventions, and code examples during work
- 🌿 **One project per branch** - Clean git integration with project state committed to feature branches
- 🔀 **Multi-session concurrency** - Work on multiple branches simultaneously via git worktrees with isolated state

## Quick Start

```bash
# 1. Install the Claude Code plugin
# (In Claude Code)
claude plugin marketplace add https://github.com/jmgilman/sow
claude plugin install sow

# 2. Install the CLI
claude /sow:install

# 3. Initialize in your repository
cd your-project
sow init

# 4. Start Claude
sow start
```

## Installation

### Prerequisites

- [Claude Code](https://www.anthropic.com/claude/code) - AI-powered coding assistant
- Git repository

### Plugin Installation

The sow plugin provides Claude Code agents, slash commands, and hooks.

**Via Claude Code:**

```bash
claude plugin marketplace add https://github.com/jmgilman/sow
claude plugin install sow
```

This copies the plugin's execution layer (`.claude/`) to your local machine.

### CLI Installation

The sow CLI provides commands for logging, state management, and validation.

#### Option 1: Interactive Installation (Recommended)

After installing the plugin, use the built-in installer:

```bash
claude /sow:install
```

This will:
- Auto-detect your platform and architecture
- Prefer Homebrew if available, otherwise download binary
- Configure your PATH if needed
- Verify installation

#### Option 2: Homebrew (macOS/Linux)

```bash
brew install jmgilman/apps/sow
```

#### Option 3: Manual Download

1. Download the latest release for your platform from [GitHub Releases](https://github.com/jmgilman/sow/releases)
2. Extract the archive:
   ```bash
   tar -xzf sow_*_*.tar.gz  # macOS/Linux
   # or
   unzip sow_*_*.zip        # Windows
   ```
3. Move to your PATH:
   ```bash
   mkdir -p ~/.local/bin
   mv sow ~/.local/bin/
   chmod +x ~/.local/bin/sow
   ```
4. Ensure `~/.local/bin` is in your PATH

**Verify installation:**

```bash
sow version
```

## Basic Usage

### Initialize a Repository

```bash
cd your-project
sow init
```

This creates the `.sow/` directory structure for knowledge and project state.

### Start a New Project

On a feature branch, use Claude Code:

```
/project:new
```

The orchestrator will:
1. Ask questions to understand your work (bug fix, feature, etc.)
2. Recommend which phases to enable (discovery, design, implementation, review, finalize)
3. Create project structure and begin work

### Continue Existing Work

Resume work on a branch:

```
/project:continue
```

The orchestrator reads project state from disk and continues where you left off.

### Multi-Session Concurrency

Work on multiple branches simultaneously using git worktrees. When you start a project or exploration session, sow automatically creates an isolated worktree:

```bash
# Start work on multiple branches concurrently
sow project --branch feat/auth    # Creates worktree at .sow/worktrees/feat/auth
sow explore --branch explore/api  # Creates worktree at .sow/worktrees/explore/api
```

Each worktree has isolated session state (`.sow/project/`, `.sow/exploration/`, etc.) while sharing committed knowledge (`.sow/knowledge/`).

**Manage worktrees:**

```bash
# List active worktrees with session types
sow worktree list

# Remove a specific worktree
sow worktree remove .sow/worktrees/feat/auth

# Clean up orphaned worktree metadata
sow worktree prune
```

### Provide Feedback

During any phase, provide corrections or guidance:

```
"The authentication should use RS256, not HS256"
```

The orchestrator creates feedback, increments the task iteration, and respawns the worker with corrections.

### Complete and Clean Up

The orchestrator automatically:
- Runs review phase (mandatory quality check)
- Updates documentation (finalize phase)
- Runs final tests and linters
- Deletes `.sow/project/` folder
- Creates a pull request
- Hands off for merge

## Development

### Prerequisites

- [Go](https://go.dev/) 1.21+
- [just](https://github.com/casey/just) - Command runner
- [golangci-lint](https://golangci-lint.run/) - Linting
- [GoReleaser](https://goreleaser.com/) - Releases (optional)

### Setup

```bash
# Clone repository
git clone git@github.com:jmgilman/sow.git
cd sow

# Install dependencies
cd cli && go mod download

# Build the CLI
just build
```

### Development Commands

The project uses [`just`](https://github.com/casey/just) for common tasks. Run `just` to see all available commands:

```bash
just                  # List all commands
just test             # Run tests with coverage
just lint             # Run golangci-lint
just build            # Build sow binary
just ci               # Run all CI checks locally
just cue-validate     # Validate CUE schemas
just cue-generate     # Generate Go types from CUE
just fmt              # Format Go code
just tidy             # Tidy go.mod
just clean            # Clean build artifacts
just coverage         # Open coverage report in browser
```

### Testing

```bash
# Run all tests
just test

# Run all CI checks (same as GitHub Actions)
just ci

# View coverage report
just coverage
```

### Building

```bash
# Build for current platform
just build
# Binary: ./sow

# Test GoReleaser locally (creates dist/)
just release-local v1.0.0
```

### Project Structure

```
sow/
├── cli/                    # CLI implementation (Go)
│   ├── cmd/                # Commands (project, task, refs, etc.)
│   ├── internal/           # Internal packages
│   ├── schemas/            # CUE schemas (embedded)
│   └── main.go             # Entry point
├── plugin/                 # Claude Code plugin (distributed)
│   ├── agents/             # Agent definitions
│   ├── commands/           # Slash commands
│   └── hooks/              # Event hooks
├── docs/                   # Documentation
├── justfile                # Development commands
├── .goreleaser.yml         # Release configuration
└── README.md               # This file
```

### Release Process

Releases are automated via GitHub Actions when a tag is pushed:

1. **Create and push tag:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **GitHub Actions will:**
   - Run tests and linting
   - Build binaries for all platforms (Linux, macOS, Windows × amd64, arm64)
   - Create GitHub release with artifacts
   - Update Homebrew tap (`jmgilman/homebrew-apps`)

3. **Binary naming:**
   - Format: `sow_{VERSION}_{OS}_{ARCH}.tar.gz` (or `.zip` for Windows)
   - Example: `sow_v1.0.0_Darwin_arm64.tar.gz`

**Local snapshot build:**

```bash
just release-local v1.0.0
# Check dist/ directory
```

## Contributing

Contributions are welcome! This project is in active development.

**Areas of interest:**
- Plugin implementation and agent improvements
- CLI enhancements
- Documentation improvements
- Testing and validation
- Example projects and tutorials

Please open an issue to discuss major changes before submitting PRs.

## License

MIT - See [LICENSE](./LICENSE) for details

---

**Status:** Active development | **Latest release:** [GitHub Releases](https://github.com/jmgilman/sow/releases)
