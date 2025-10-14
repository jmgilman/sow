# ADR 001: Go CLI Architecture

**Status**: Accepted

**Date**: 2025-10-13

**Context**: Milestone 1 (CLI Foundation & Schema System)

---

## Context

The sow system requires a standalone CLI binary that:
1. Embeds CUE schemas as the source of truth for all `.sow/` file formats
2. Provides fast operations for initialization, validation, and logging
3. Manages external knowledge (sinks) and linked repositories
4. Supports auto-detection of task vs project context
5. Ensures schema version aligns with CLI version
6. Works cross-platform (macOS, Linux, Windows)

The CLI must be performant (especially for logging, which agents use frequently) and maintain zero external dependencies for schema access.

## Decision

We will build a Go CLI with the following architecture:

### 1. Project Structure

```
sow/
├── cmd/
│   └── sow/
│       └── main.go              # Entry point, cobra setup
│
├── internal/
│   ├── commands/
│   │   ├── init.go              # sow init
│   │   ├── validate.go          # sow validate
│   │   ├── schema.go            # sow schema
│   │   ├── log.go               # sow log
│   │   ├── session.go           # sow session-info
│   │   ├── sinks.go             # sow sinks subcommands
│   │   └── repos.go             # sow repos subcommands
│   │
│   ├── schema/
│   │   ├── embed.go             # go:embed directives
│   │   ├── validate.go          # CUE validation logic
│   │   ├── materialize.go       # Default value materialization
│   │   └── version.go           # Schema version tracking
│   │
│   ├── context/
│   │   ├── detector.go          # Task vs project context detection
│   │   ├── filesystem.go        # .sow/ structure traversal
│   │   └── git.go               # Git branch detection
│   │
│   ├── logging/
│   │   ├── writer.go            # Log entry formatting and appending
│   │   ├── agent.go             # Agent ID construction
│   │   └── vocabulary.go        # Action vocabulary validation
│   │
│   ├── sinks/
│   │   ├── manager.go           # Sink installation and updates
│   │   ├── index.go             # Index.json management
│   │   └── interrogate.go       # LLM-based sink summarization
│   │
│   ├── repos/
│   │   ├── manager.go           # Repo cloning and syncing
│   │   ├── index.go             # Index.json management
│   │   └── symlink.go           # Symlink support
│   │
│   └── config/
│       └── version.go           # CLI version constant
│
├── pkg/
│   └── sowfs/
│       ├── paths.go             # Standard .sow/ path constants
│       ├── reader.go            # Safe YAML/JSON reading
│       └── writer.go            # Atomic file writing
│
└── schemas/
    └── cue/
        ├── project-state.cue    # Embedded schema
        ├── task-state.cue       # Embedded schema
        ├── sink-index.cue       # Embedded schema
        ├── repo-index.cue       # Embedded schema
        └── sow-version.cue      # Embedded schema
```

**Rationale**:
- `cmd/` contains only entry point (standard Go project layout)
- `internal/` prevents external imports (CLI is self-contained)
- `pkg/` for potentially reusable filesystem utilities
- `schemas/` colocated with code for easy embedding
- Clear separation of concerns by domain (commands, schema, context, logging, etc.)

### 2. Command Routing (Cobra)

We use [Cobra](https://github.com/spf13/cobra) for command-line parsing and routing.

**Command Structure**:
```
sow                           # Root command (shows help)
├── --version                 # Global flag
├── --help                    # Global flag
├── init                      # Initialize .sow/ structure
├── validate                  # Validate against schemas
├── schema
│   ├── show <type>          # Display specific schema
│   └── list                 # List all schemas
├── log                       # Create log entry
├── session-info              # Display session status
├── sinks
│   ├── install <source>     # Install sink
│   ├── update [name]        # Update sinks
│   ├── list                 # List sinks
│   ├── remove <name>        # Remove sink
│   └── reindex              # Rebuild index
├── repos
│   ├── add <source>         # Add repo
│   ├── sync [name]          # Sync repos
│   ├── list                 # List repos
│   ├── remove <name>        # Remove repo
│   └── reindex              # Rebuild index
└── sync                      # Sync both sinks and repos
```

**Implementation Pattern**:
```go
// cmd/sow/main.go
package main

import (
    "github.com/spf13/cobra"
    "sow/internal/commands"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "sow",
        Short: "AI-powered system of work",
        Long:  "sow - Structured software development with AI agents",
    }

    // Global flags
    rootCmd.PersistentFlags().Bool("quiet", false, "Suppress output")
    rootCmd.PersistentFlags().Bool("verbose", false, "Verbose output")
    rootCmd.PersistentFlags().Bool("no-color", false, "Disable colors")

    // Register commands
    rootCmd.AddCommand(commands.NewInitCmd())
    rootCmd.AddCommand(commands.NewValidateCmd())
    rootCmd.AddCommand(commands.NewSchemaCmd())
    rootCmd.AddCommand(commands.NewLogCmd())
    rootCmd.AddCommand(commands.NewSessionInfoCmd())
    rootCmd.AddCommand(commands.NewSinksCmd())
    rootCmd.AddCommand(commands.NewReposCmd())
    rootCmd.AddCommand(commands.NewSyncCmd())

    rootCmd.Execute()
}
```

**Rationale**:
- Cobra is mature, widely-used, and well-documented
- Supports subcommands naturally
- Built-in help generation
- Flag parsing with validation
- Minimal boilerplate

### 3. Schema Embedding Strategy

**go:embed Directive**:
```go
// internal/schema/embed.go
package schema

import _ "embed"

//go:embed ../../schemas/cue/project-state.cue
var projectStateCUE string

//go:embed ../../schemas/cue/task-state.cue
var taskStateCUE string

//go:embed ../../schemas/cue/sink-index.cue
var sinkIndexCUE string

//go:embed ../../schemas/cue/repo-index.cue
var repoIndexCUE string

//go:embed ../../schemas/cue/sow-version.cue
var sowVersionCUE string
```

**Schema Loading**:
```go
// internal/schema/validate.go
package schema

import (
    "cuelang.org/go/cue"
    "cuelang.org/go/cue/cuecontext"
)

var (
    ctx             *cue.Context
    projectSchema   cue.Value
    taskSchema      cue.Value
    sinkSchema      cue.Value
    repoSchema      cue.Value
    versionSchema   cue.Value
)

func init() {
    ctx = cuecontext.New()

    // Compile embedded CUE schemas
    projectSchema = ctx.CompileString(projectStateCUE)
    taskSchema = ctx.CompileString(taskStateCUE)
    sinkSchema = ctx.CompileString(sinkIndexCUE)
    repoSchema = ctx.CompileString(repoIndexCUE)
    versionSchema = ctx.CompileString(sowVersionCUE)
}

// ValidateProjectState validates a project state file
func ValidateProjectState(data []byte) error {
    value := ctx.CompileBytes(data)
    return projectSchema.Unify(value).Validate()
}

// ValidateTaskState validates a task state file
func ValidateTaskState(data []byte) error {
    value := ctx.CompileBytes(data)
    return taskSchema.Unify(value).Validate()
}

// ... similar functions for other schemas
```

**Schema Materialization** (for `sow init`):
```go
// internal/schema/materialize.go
package schema

// MaterializeProjectState creates a new project state with defaults
func MaterializeProjectState(projectName, branch string) ([]byte, error) {
    // Build CUE value with required fields
    value := ctx.CompileString(`
        project: {
            name: "` + projectName + `"
            branch: "` + branch + `"
            created_at: "` + time.Now().Format(time.RFC3339) + `"
            updated_at: "` + time.Now().Format(time.RFC3339) + `"
            description: ""
            complexity: {
                rating: 1
                metrics: {
                    estimated_files: 0
                    cross_cutting: false
                    new_dependencies: false
                }
            }
            active_phase: "implement"
        }
        phases: []
    `)

    // Unify with schema to apply defaults
    unified := projectSchema.Unify(value)

    // Export to YAML
    return yaml.Marshal(unified)
}
```

**Version Alignment**:
```go
// internal/config/version.go
package config

// Version is the CLI version (set at build time)
// This MUST match the embedded schema versions
const Version = "0.2.0"

// internal/schema/version.go
package schema

import "sow/internal/config"

// Version returns the schema version (same as CLI version)
func Version() string {
    return config.Version
}
```

**Rationale**:
- `go:embed` compiles schemas into binary (no external files needed)
- Schemas loaded once at init (fast subsequent access)
- CUE library provides validation, unification, and default materialization
- Version constant ensures CLI version = schema version
- Materialization enables `sow init` to create correct file structure

### 4. Context Detection

**Auto-Detection Logic**:
```go
// internal/context/detector.go
package context

import (
    "os"
    "path/filepath"
)

type Context struct {
    Type        string // "task" or "project"
    TaskDir     string // .sow/project/phases/{phase}/tasks/{id}
    ProjectDir  string // .sow/project
    RepoRoot    string // Repository root
}

// Detect determines context from current working directory
func Detect() (*Context, error) {
    cwd, err := os.Getwd()
    if err != nil {
        return nil, err
    }

    // Walk up directory tree looking for .sow/
    repoRoot := findRepoRoot(cwd)
    if repoRoot == "" {
        return nil, errors.New("not in a sow repository")
    }

    projectDir := filepath.Join(repoRoot, ".sow", "project")

    // Check if we're in a task directory
    if isTaskDir(cwd, projectDir) {
        return &Context{
            Type:       "task",
            TaskDir:    cwd,
            ProjectDir: projectDir,
            RepoRoot:   repoRoot,
        }, nil
    }

    return &Context{
        Type:       "project",
        ProjectDir: projectDir,
        RepoRoot:   repoRoot,
    }, nil
}

func findRepoRoot(dir string) string {
    for {
        sowDir := filepath.Join(dir, ".sow")
        if _, err := os.Stat(sowDir); err == nil {
            return dir
        }

        parent := filepath.Dir(dir)
        if parent == dir {
            return "" // Reached filesystem root
        }
        dir = parent
    }
}

func isTaskDir(cwd, projectDir string) bool {
    // Check if path matches: .sow/project/phases/{phase}/tasks/{id}
    rel, err := filepath.Rel(projectDir, cwd)
    if err != nil {
        return false
    }

    parts := strings.Split(rel, string(filepath.Separator))
    return len(parts) == 4 && parts[0] == "phases" && parts[2] == "tasks"
}
```

**Usage in Commands**:
```go
// internal/commands/log.go
func NewLogCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "log",
        Short: "Create log entry",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Detect context
            ctx, err := context.Detect()
            if err != nil {
                return err
            }

            // Determine log file location
            var logFile string
            if forceProject || ctx.Type == "project" {
                logFile = filepath.Join(ctx.ProjectDir, "log.md")
            } else {
                logFile = filepath.Join(ctx.TaskDir, "log.md")
            }

            // Write log entry
            return logging.WriteEntry(logFile, entry)
        },
    }

    cmd.Flags().Bool("project", false, "Force project log")
    return cmd
}
```

**Rationale**:
- Agents don't need to specify context explicitly
- Works from any directory in the repository
- Clear logic for determining task vs project
- Fallback to project log when ambiguous
- Override available via `--project` flag

### 5. Core Abstractions

**Validation Engine Interface**:
```go
// internal/schema/validate.go
package schema

type Validator interface {
    Validate(data []byte) error
    ValidateFile(path string) error
}

type SchemaType string

const (
    ProjectStateSchema SchemaType = "project"
    TaskStateSchema    SchemaType = "task"
    SinkIndexSchema    SchemaType = "sink-index"
    RepoIndexSchema    SchemaType = "repo-index"
    VersionSchema      SchemaType = "version"
)

func GetValidator(schemaType SchemaType) Validator {
    // Return appropriate validator implementation
}
```

**File System Operations**:
```go
// pkg/sowfs/writer.go
package sowfs

import "io/ioutil"

// AtomicWrite writes file atomically (write to temp, then rename)
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
    tmpPath := path + ".tmp"

    if err := ioutil.WriteFile(tmpPath, data, perm); err != nil {
        return err
    }

    return os.Rename(tmpPath, path)
}

// pkg/sowfs/paths.go
package sowfs

import "path/filepath"

type Paths struct {
    RepoRoot   string
    SowDir     string
    ProjectDir string
    Knowledge  string
    Sinks      string
    Repos      string
}

func NewPaths(repoRoot string) *Paths {
    sowDir := filepath.Join(repoRoot, ".sow")
    return &Paths{
        RepoRoot:   repoRoot,
        SowDir:     sowDir,
        ProjectDir: filepath.Join(sowDir, "project"),
        Knowledge:  filepath.Join(sowDir, "knowledge"),
        Sinks:      filepath.Join(sowDir, "sinks"),
        Repos:      filepath.Join(sowDir, "repos"),
    }
}
```

**Error Handling**:
```go
// internal/errors/errors.go
package errors

import "fmt"

type CLIError struct {
    Code    int    // Exit code
    Message string
    Hint    string // Helpful suggestion
}

func (e *CLIError) Error() string {
    if e.Hint != "" {
        return fmt.Sprintf("%s\n\n💡 %s", e.Message, e.Hint)
    }
    return e.Message
}

var (
    ErrNotInRepo = &CLIError{
        Code:    3,
        Message: "Not in a sow repository",
        Hint:    "Run 'sow init' to initialize .sow/ structure",
    }

    ErrValidationFailed = &CLIError{
        Code:    4,
        Message: "Validation failed",
        Hint:    "Run 'sow schema show <type>' to see expected format",
    }
)
```

**Rationale**:
- Clean interfaces enable testing and future extensions
- Atomic writes prevent corruption
- Centralized path management reduces bugs
- Rich error types provide helpful user guidance
- Exit codes enable scripting

### 6. Logging Implementation

**Fast Log Writing** (performance-critical):
```go
// internal/logging/writer.go
package logging

import (
    "fmt"
    "os"
    "time"
)

type Entry struct {
    Timestamp string
    AgentID   string
    Action    string
    Result    string
    Files     []string
    Notes     string
}

// WriteEntry appends log entry to file (fast, no file re-reading)
func WriteEntry(logPath string, entry Entry) error {
    // Format entry
    formatted := fmt.Sprintf(`
## %s - %s

**Action**: %s
**Result**: %s
**Files**: %s
**Notes**: %s

---
`, entry.Timestamp, entry.AgentID, entry.Action, entry.Result,
   formatFiles(entry.Files), entry.Notes)

    // Open file for append
    f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()

    // Write entry (single write, no seeking)
    _, err = f.WriteString(formatted)
    return err
}

// internal/logging/agent.go
package logging

import (
    "fmt"
    "gopkg.in/yaml.v3"
    "io/ioutil"
)

// ConstructAgentID reads task state and builds agent ID
func ConstructAgentID(taskDir string) (string, error) {
    statePath := filepath.Join(taskDir, "state.yaml")

    data, err := ioutil.ReadFile(statePath)
    if err != nil {
        return "", err
    }

    var state struct {
        Task struct {
            AssignedAgent string `yaml:"assigned_agent"`
            Iteration     int    `yaml:"iteration"`
        } `yaml:"task"`
    }

    if err := yaml.Unmarshal(data, &state); err != nil {
        return "", err
    }

    return fmt.Sprintf("%s-%d",
        state.Task.AssignedAgent,
        state.Task.Iteration), nil
}
```

**Rationale**:
- Append-only writes (no reading entire file)
- Single syscall (fast)
- Auto-constructs agent ID from state file
- Consistent formatting
- Minimal overhead (~10-50ms vs 30s for file editing)

### 7. Build and Distribution

**Build Configuration**:
```makefile
# Makefile
VERSION := 0.2.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

build:
	go build -ldflags "-X sow/internal/config.Version=$(VERSION)" \
	         -o bin/sow cmd/sow/main.go

build-all:
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X sow/internal/config.Version=$(VERSION)" \
	                                   -o bin/sow-macos cmd/sow/main.go

	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags "-X sow/internal/config.Version=$(VERSION)" \
	                                  -o bin/sow-linux cmd/sow/main.go

	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags "-X sow/internal/config.Version=$(VERSION)" \
	                                    -o bin/sow-windows.exe cmd/sow/main.go

test:
	go test ./...

install:
	go install cmd/sow/main.go
```

**Installation Script**:
```bash
#!/bin/bash
# install.sh
set -e

VERSION="0.2.0"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
INSTALL_DIR="${HOME}/.local/bin"

echo "Installing sow CLI v${VERSION}..."

# Download binary
if [ "$OS" = "darwin" ]; then
    BINARY="sow-macos"
elif [ "$OS" = "linux" ]; then
    BINARY="sow-linux"
else
    echo "Unsupported OS: $OS"
    exit 1
fi

curl -L "https://github.com/your-org/sow/releases/download/v${VERSION}/${BINARY}" \
     -o "${INSTALL_DIR}/sow"

chmod +x "${INSTALL_DIR}/sow"

echo "✓ Installed to ${INSTALL_DIR}/sow"
echo "✓ Run 'sow --version' to verify"
```

**Rationale**:
- Single binary distribution (no dependencies)
- Cross-platform support
- Version injected at build time
- Simple installation via curl
- Standard Unix conventions

## Consequences

### Positive

1. **Single Binary**: No external dependencies, easy distribution
2. **Schema Alignment**: Version consistency guaranteed by embedding
3. **Performance**: Fast operations, especially logging (<100ms)
4. **Cross-Platform**: Works on macOS, Linux, Windows
5. **Type Safety**: CUE provides compile-time validation
6. **Zero Config**: Schemas embedded, no setup needed
7. **Offline**: No network required for validation
8. **Clean Architecture**: Clear separation of concerns
9. **Testable**: Interfaces enable unit testing
10. **Maintainable**: Standard Go project layout

### Negative

1. **Binary Size**: Embedding schemas increases binary size (~5-10MB)
2. **CUE Dependency**: Requires CUE library (Go dependency)
3. **Migration Complexity**: Schema changes require migration logic
4. **Build Process**: Need to build for multiple platforms
5. **Learning Curve**: Team needs to understand Go, CUE, and Cobra

### Mitigations

1. **Binary Size**: 5-10MB is acceptable for modern systems
2. **CUE Dependency**: Well-maintained library, stable API
3. **Migration**: Document migration patterns, create migration templates
4. **Build**: Automate with Makefile and CI/CD
5. **Learning**: Comprehensive documentation, examples, and ADRs

## Alternatives Considered

### Python CLI with JSON Schema

**Rejected**:
- JSON Schema lacks CUE's constraint expressiveness
- Python distribution more complex (pip, virtualenv)
- Slower performance
- External schema files or bundling complexity

### Rust CLI

**Rejected**:
- Longer compile times during development
- Steeper learning curve
- Less mature schema validation ecosystem
- Overkill for this use case

### Bash Scripts

**Rejected**:
- No type safety
- Poor error handling
- Hard to maintain
- No cross-platform support
- No schema validation

### Node.js CLI

**Rejected**:
- Requires Node.js runtime
- Larger distribution size
- Slower startup time
- Limited schema validation libraries

## Implementation Notes

### Phase 1: Core CLI Structure
1. Set up Go module and project structure
2. Implement Cobra command routing
3. Create basic commands (--version, --help)
4. Set up build system

### Phase 2: Schema Embedding
1. Embed CUE schemas with go:embed
2. Implement CUE validation logic
3. Create schema materialization for init
4. Test validation with sample files

### Phase 3: Core Commands
1. Implement `sow init`
2. Implement `sow validate`
3. Implement `sow schema`
4. Implement `sow log` (performance-critical)

### Phase 4: Context Detection
1. Build context detector
2. Integrate with log command
3. Test auto-detection logic

### Phase 5: Sink and Repo Management
1. Implement `sow sinks` commands
2. Implement `sow repos` commands
3. Build index management
4. Test with real repositories

### Phase 6: Polish
1. Improve error messages
2. Add helpful hints
3. Cross-platform testing
4. Performance optimization
5. Documentation

## References

- [CLI_REFERENCE.md](/Users/josh/code/sow/docs/CLI_REFERENCE.md) - Complete CLI specifications
- [ARCHITECTURE.md](/Users/josh/code/sow/docs/ARCHITECTURE.md) - Schema embedding rationale
- [ROADMAP.md](/Users/josh/code/sow/ROADMAP.md) - Milestone 1 requirements
- [CUE Language Documentation](https://cuelang.org/docs/)
- [Cobra Documentation](https://github.com/spf13/cobra)
- [Go Embed Documentation](https://pkg.go.dev/embed)

---

**Decision Made By**: architect-1

**Review Status**: Awaiting implementation feedback
