# Multi-Agent System Architecture

**Status:** Draft
**Date:** January 2025
**Context:** Extension of sow to support multiple AI agent CLIs beyond Claude Code

## Executive Summary

This document describes the architecture for extending sow's agent system to support multiple AI CLI tools (Claude Code, Cursor, Windsurf, etc.) while maintaining consistent orchestration patterns and subscription-based economics.

**Key architectural decisions:**

1. **Agent/Executor Split** - Separate lightweight agent definitions (roles, prompts) from heavyweight executor implementations (CLI invocation, session management)
2. **User-Level Configuration** - Agent preferences stored in `~/.config/sow/config.yaml`, never in repository
3. **Zero-Config Experience** - Works out of the box with defaults, configuration optional for customization
4. **State-Based Protocol** - Subprocess exit signals completion, `state.yaml` contains the outcome
5. **Session Resumption** - Bidirectional communication via resumable CLI sessions, supports review iteration
6. **Guidance-Based Consistency** - Four-layer consistency model (see separate ADR)

## Context and Problem Statement

### Current State

Sow currently hardcodes Claude Code as the agent CLI:
- Orchestrator spawns workers via direct `claude` command invocation
- Worker prompts embedded in `.claude/agents/` directory
- Session management non-existent (no resumption support)
- No abstraction layer for alternative CLIs

### Goals

1. **Support multiple agent CLIs** - Claude Code, Cursor, Windsurf, future tools
2. **User preference system** - Let users choose their preferred tools per agent role
3. **Maintain consistency** - Ensure reliable behavior across different models and CLIs
4. **Enable bidirectional communication** - Support paused workflows and iterative refinement
5. **Preserve simplicity** - Don't introduce unnecessary complexity

### Non-Goals

- Programmatic validation enforcement (impossible - orchestrator is an agent, not code)
- Identical output across all models (unrealistic - models have inherent variability)
- Repository-level agent configuration (conflicts with subscription-based model)

## Critical Architectural Insight

**The orchestrator is another AI agent, not Go code.**

This is the foundational insight that shapes the entire design:

- The orchestrator (Claude Code, Cursor, etc.) is an AI agent reading files and making judgments
- Cannot programmatically enforce validation rules - only guidance via prompts
- The only hard enforcement point is the `sow` CLI validating CUE schemas
- Everything else (validation, feedback, iteration) is guidance-based
- Both orchestrator and worker agents need extremely explicit prompts

**Implication:** Consistency comes from prompt engineering and iteration loops, not programmatic constraints.

## High-Level Architecture

### Agent/Executor Split

**Agents** (lightweight, data-focused):
- Represent roles in the sow system (implementer, architect, reviewer, etc.)
- Defined as simple structs: name, description, capabilities, prompt path
- 12+ agents expected (easily extensible)
- No behavioral logic - just configuration

**Executors** (heavyweight, behavior-focused):
- Handle CLI invocation and session management
- Implement interface: `Spawn()`, `Resume()`, `SupportsResumption()`
- 2-3 executors expected (Claude, Cursor, potentially Windsurf)
- Each executor encapsulates CLI-specific complexity

**Rationale:**
- Agents are configuration → struct (lightweight, serializable)
- Executors are varying behavior → interface (type-safe, extensible)
- Clear separation of concerns
- Easy to add new agents (define struct) or executors (implement interface)

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      Orchestrator Agent                      │
│                   (Claude Code, Cursor, etc.)                │
│                                                              │
│  Reads: state.yaml, task descriptions                       │
│  Writes: state.yaml, project log                            │
│  Spawns: Worker agents via `sow agent spawn`                │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ sow agent spawn implementer 010
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                         sow CLI                              │
│                                                              │
│  ┌──────────────────┐      ┌──────────────────────────┐    │
│  │  Agent Registry  │      │   Executor Registry      │    │
│  │                  │      │                          │    │
│  │  - Implementer   │      │  - ClaudeExecutor        │    │
│  │  - Architect     │      │  - CursorExecutor        │    │
│  │  - Reviewer      │      │  - WindsurfExecutor      │    │
│  │  - Planner       │      │                          │    │
│  │  - ...           │      │  (Interface-based)       │    │
│  └──────────────────┘      └──────────────────────────┘    │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │           User Configuration                          │  │
│  │  ~/.config/sow/config.yaml                           │  │
│  │                                                       │  │
│  │  agents:                                             │  │
│  │    bindings:                                         │  │
│  │      implementer: "cursor"                           │  │
│  │      orchestrator: "claude-code"                     │  │
│  └──────────────────────────────────────────────────────┘  │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ executor.Spawn(agent, prompt, sessionID)
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                       Worker Agent                           │
│                   (Cursor, Claude Code, etc.)                │
│                                                              │
│  Subprocess execution:                                       │
│  1. Receives agent prompt template + task context           │
│  2. Runs `sow prompt guidance/{agent}/base`                 │
│  3. Reads state.yaml, description.md, feedback/             │
│  4. Performs work, logs actions                             │
│  5. Updates state.yaml with status                          │
│  6. Exits (subprocess terminates)                           │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Agent Definition

Agents are simple structs representing roles:

```go
// Agent represents a role in the sow system
type Agent struct {
    // Name is the agent identifier (e.g., "implementer", "architect")
    Name string

    // Description explains what this agent does
    Description string

    // Capabilities describes what the agent must be able to do (prose)
    Capabilities string

    // PromptPath is the path to the embedded prompt template
    // Relative to cli/internal/agents/templates/
    PromptPath string
}
```

**All agents defined in one file** (`cli/internal/agents/agents.go`):

```go
var (
    Implementer = &Agent{
        Name:         "implementer",
        Description:  "Code implementation using Test-Driven Development",
        Capabilities: "Must be able to read/write files, execute shell commands, search codebase",
        PromptPath:   "implementer.md",
    }

    Architect = &Agent{
        Name:         "architect",
        Description:  "System design and architecture decisions",
        Capabilities: "Must be able to read/write files, search codebase",
        PromptPath:   "architect.md",
    }

    Reviewer = &Agent{
        Name:         "reviewer",
        Description:  "Code review and quality assessment",
        Capabilities: "Must be able to read files, search codebase, execute shell commands",
        PromptPath:   "reviewer.md",
    }

    // ... more agents (12+ expected)
)

func StandardAgents() []*Agent {
    return []*Agent{Implementer, Architect, Reviewer, /* ... */}
}
```

**Design rationale:**
- Struct (not interface) because agents are lightweight configuration
- All in one file for easy discovery and maintenance
- Easily extensible - just add new struct definition
- Can be serialized/deserialized for future use cases

### 2. Agent Registry

Simple registry for agent lookup:

```go
type AgentRegistry struct {
    agents map[string]*Agent
}

func NewAgentRegistry() *AgentRegistry {
    r := &AgentRegistry{agents: make(map[string]*Agent)}
    for _, agent := range StandardAgents() {
        r.Register(agent)
    }
    return r
}

func (r *AgentRegistry) Get(name string) (*Agent, error) {
    agent, ok := r.agents[name]
    if !ok {
        return nil, fmt.Errorf("unknown agent: %s", name)
    }
    return agent, nil
}
```

### 3. Embedded Prompt Templates

Agent prompts embedded in binary using `go:embed`:

```go
//go:embed templates/*
var templatesFS embed.FS

func LoadPrompt(promptPath string) (string, error) {
    data, err := fs.ReadFile(templatesFS, "templates/"+promptPath)
    if err != nil {
        return "", fmt.Errorf("failed to load prompt %s: %w", promptPath, err)
    }
    return string(data), nil
}
```

**Template structure:**

```
cli/internal/agents/templates/
├── implementer.md
├── architect.md
├── reviewer.md
├── planner.md
└── researcher.md
```

**Template content pattern** (worker self-initialization):

```markdown
<!-- cli/internal/agents/templates/implementer.md -->

You are a software implementer agent. Your instructions are provided
dynamically via the sow prompt system.

## Initialization

Run this command immediately to load your base instructions:

```bash
sow prompt guidance/implementer/base
```

The base prompt will guide you through:
1. Reading task context (state.yaml, description.md, feedback)
2. Loading mandatory TDD guidance
3. Inferring task scenario and loading appropriate workflow guidance
4. Executing the implementation

## Context Location

Your task context is located at:

```
.sow/project/phases/implementation/tasks/{task-id}/
├── state.yaml        # Task metadata, iteration, references
├── description.md    # Requirements and acceptance criteria
├── log.md            # Your action log (append here)
└── feedback/         # Corrections from previous iterations
```

Start by reading state.yaml to get your task ID and iteration number.
```

**Design rationale:**
- Workers self-initialize by running `sow prompt` commands
- Spawn prompt stays minimal (just task location, iteration)
- Detailed instructions stored in embedded prompt templates
- Simplifies executor abstraction - just invoke CLI with minimal context

### 4. Executor Interface

Executors handle CLI invocation and session management:

```go
// Executor invokes agent CLIs and manages sessions
type Executor interface {
    // Name returns the executor identifier (e.g., "claude-code", "cursor")
    Name() string

    // Spawn invokes an agent with the given prompt and session ID
    // Blocks until the subprocess exits
    Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error

    // Resume continues an existing session with additional prompt
    // Blocks until the subprocess exits
    Resume(ctx context.Context, sessionID string, prompt string) error

    // SupportsResumption indicates if this executor can resume sessions
    SupportsResumption() bool
}
```

**Key design decisions:**

1. **Methods block directly** - No Session object returned
   - CLI is stateless - each invocation is independent
   - Session state lives in task `state.yaml`, not Go memory
   - Simpler model: spawn → block → exit → read state

2. **Context-aware** - Supports cancellation and timeouts

3. **Minimal interface** - Only essential operations
   - Spawn: Create new session
   - Resume: Continue existing session
   - SupportsResumption: Capability check

### 5. Executor Implementations

Session management flags (`--session-id`, `--chat-id`, `--resume`) are hardcoded in each executor implementation. These flags are intrinsic to each CLI's design and never vary, so they're implementation details rather than user configuration.

**Claude Code Executor:**

```go
type ClaudeExecutor struct {
    yoloMode bool
    model    string
}

func (e *ClaudeExecutor) Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error {
    // Load agent prompt template
    agentPrompt, err := LoadPrompt(agent.PromptPath)
    if err != nil {
        return err
    }

    // Build full spawn prompt
    fullPrompt := fmt.Sprintf("%s\n\n%s", agentPrompt, prompt)

    // Build command with flags
    args := []string{}
    if e.yoloMode {
        args = append(args, "--dangerously-skip-permissions")
    }
    if e.model != "" {
        args = append(args, "--model", e.model)
    }
    if sessionID != "" {
        args = append(args, "--session-id", sessionID)
    }

    cmd := exec.CommandContext(ctx, "claude", args...)
    cmd.Stdin = strings.NewReader(fullPrompt)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // Run and block until exit
    return cmd.Run()
}

func (e *ClaudeExecutor) Resume(ctx context.Context, sessionID string, prompt string) error {
    cmd := exec.CommandContext(ctx, "claude", "--resume", sessionID)
    cmd.Stdin = strings.NewReader(prompt)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}

func (e *ClaudeExecutor) SupportsResumption() bool {
    return true
}
```

**Cursor Executor:**

```go
type CursorExecutor struct {
    yoloMode bool
}

func (e *CursorExecutor) Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error {
    agentPrompt, err := LoadPrompt(agent.PromptPath)
    if err != nil {
        return err
    }

    fullPrompt := fmt.Sprintf("%s\n\n%s", agentPrompt, prompt)

    args := []string{"agent"}
    if sessionID != "" {
        args = append(args, "--chat-id", sessionID)
    }

    cmd := exec.CommandContext(ctx, "cursor-agent", args...)
    cmd.Stdin = strings.NewReader(fullPrompt)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}

func (e *CursorExecutor) Resume(ctx context.Context, sessionID string, prompt string) error {
    cmd := exec.CommandContext(ctx, "cursor-agent", "agent", "--resume", sessionID)
    cmd.Stdin = strings.NewReader(prompt)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd.Run()
}

func (e *CursorExecutor) SupportsResumption() bool {
    return true
}
```

### 6. Executor Registry

Similar to agent registry, manages executor instances:

```go
type ExecutorRegistry struct {
    executors map[string]Executor
}

func (r *ExecutorRegistry) Register(name string, executor Executor) {
    r.executors[name] = executor
}

func (r *ExecutorRegistry) Get(name string) (Executor, error) {
    executor, ok := r.executors[name]
    if !ok {
        return nil, fmt.Errorf("unknown executor: %s", name)
    }
    return executor, nil
}
```

**Loading from user configuration:**

```go
func LoadExecutors(config *UserConfig) (*ExecutorRegistry, error) {
    registry := NewExecutorRegistry()

    for name, execConfig := range config.Agents.Executors {
        switch execConfig.Type {
        case "claude":
            executor := NewClaudeExecutor(
                execConfig.Settings.YoloMode,
                execConfig.Settings.Model,
            )
            registry.Register(name, executor)

        case "cursor":
            executor := NewCursorExecutor(
                execConfig.Settings.YoloMode,
            )
            registry.Register(name, executor)

        default:
            return nil, fmt.Errorf("unknown executor type: %s", execConfig.Type)
        }
    }

    return registry, nil
}
```

## User Configuration System

### Configuration Location

`~/.config/sow/config.yaml` (cross-platform standard)

**Rationale:**
- User preferences tied to subscriptions, not repositories
- Never committed to git
- Follows precedent from refs cache at `~/.cache/sow/refs/`

### Configuration Structure

```yaml
# ~/.config/sow/config.yaml

agents:
  # Executor definitions
  executors:
    claude-code:
      type: "claude"
      settings:
        yolo_mode: true          # Maps to --dangerously-skip-permissions
        model: "sonnet"          # Which Claude model to use
      custom_args: []            # Additional CLI flags

    cursor:
      type: "cursor"
      settings:
        yolo_mode: true          # Maps to cursor's equivalent permission flag
      custom_args: []

    windsurf:
      type: "windsurf"
      settings:
        yolo_mode: true
      custom_args: ["--cascade"]

  # Bindings: which executor handles which agent role
  bindings:
    orchestrator: "claude-code"
    implementer: "cursor"
    architect: "claude-code"
    reviewer: "claude-code"
    planner: "claude-code"
    researcher: "claude-code"
```

### Configuration Fields

**Executor definition:**
- `type`: Executor implementation to use (`claude`, `cursor`, etc.)
- `settings`: Common settings abstracted across executors
  - `yolo_mode`: Skip permission prompts (mapped to executor-specific flags)
  - `model`: Which underlying AI model to use (if applicable)
- `custom_args`: Executor-specific flags that don't fit common patterns

**Note on session flags:** Session management flags (`--session-id`, `--chat-id`, `--resume`) are hardcoded in executor implementations, not configured by users. These are intrinsic to each CLI's design and never change.

**Bindings:**
- Map agent role (implementer, architect, etc.) to executor name
- User preference - most modern agents can handle any role
- Allows mixing executors (Claude for orchestrator, Cursor for implementer)

### Configuration Priority

1. **Environment variable override**: `SOW_AGENTS_IMPLEMENTER=windsurf sow agent spawn implementer 010`
2. **User config**: `~/.config/sow/config.yaml`
3. **Defaults**: Claude Code for all agents

### No Repository-Level Agent Config

Agent preferences are user-specific (tied to subscriptions and preferences), not repository-specific.

Repository config (`.sow/config.yaml`) remains for artifact paths and repository settings only.

### Configuration Management

#### Missing Config File Behavior

**When `~/.config/sow/config.yaml` doesn't exist:**

- Accept defaults silently, no errors
- All agents use Claude Code executor
- Settings: `yolo_mode: false` (safe default)
- Zero-config experience for existing users

**Rationale:**
- Backward compatible - sow works out of the box
- Users only configure when they want to customize
- No filesystem clutter from auto-generated files

#### Configuration Initialization

**Command:**
```bash
sow config init
```

**Behavior:**
- Creates `~/.config/sow/config.yaml` with commented template
- Shows all available executors and agent roles
- Includes examples for common scenarios
- Prompts user if file already exists (don't overwrite without confirmation)

**Generated template:**
```yaml
# Sow Agent Configuration
# Location: ~/.config/sow/config.yaml
#
# This file configures which AI CLI tools handle which agent roles.
# If this file doesn't exist, all agents use Claude Code by default.
#
# Configuration priority:
#   1. Environment variables (SOW_AGENTS_IMPLEMENTER=cursor)
#   2. This config file
#   3. Built-in defaults (Claude Code)

agents:
  executors:
    # Claude Code executor
    # Uncomment and modify to customize Claude Code settings
    # claude-code:
    #   type: "claude"
    #   settings:
    #     yolo_mode: true    # Skip permission prompts
    #     model: "sonnet"    # or "opus", "haiku"
    #   custom_args: []

    # Cursor executor
    # Uncomment to enable Cursor
    # cursor:
    #   type: "cursor"
    #   settings:
    #     yolo_mode: true
    #   custom_args: []

    # Add more executors as needed (windsurf, etc.)

  # Bindings: which executor handles which agent role
  # Uncomment and modify to change from defaults
  bindings:
    orchestrator: "claude-code"
    implementer: "claude-code"    # Change to "cursor" to use Cursor for implementation
    architect: "claude-code"
    reviewer: "claude-code"
    planner: "claude-code"
    researcher: "claude-code"
```

#### Configuration Discovery

**Users can discover config location through multiple mechanisms:**

**1. Show config path:**
```bash
sow config path
# Output: /Users/josh/.config/sow/config.yaml

sow config path --exists
# Output: true (or false if doesn't exist)
```

**2. Help text references location:**
```bash
sow agent spawn --help
# Output includes:
#   Agent selection uses configuration from ~/.config/sow/config.yaml
#   Run 'sow config init' to create a configuration file.
#   Run 'sow config path' to see the full path.
```

**3. Error messages mention path:**
```
Error: unknown executor: copilot

Available executors:
  - claude-code
  - cursor

Check your configuration at ~/.config/sow/config.yaml
Run 'sow config path' to see the full path.
Run 'sow config init' to create a default configuration.
```

**4. Verbose output shows resolution:**
```bash
sow agent spawn implementer 010 --verbose
# Output:
#   Loading agent configuration...
#   Config file: /Users/josh/.config/sow/config.yaml (not found, using defaults)
#   Agent: implementer
#   Executor: claude-code (default)
#   Settings: yolo_mode=false (default)
```

#### Configuration Validation

**Command:**
```bash
sow config validate
```

**Checks:**
- ✓ Valid YAML syntax
- ✓ Valid CUE schema (executor types, settings structure)
- ✓ All referenced executors are registered types
- ✓ All agent bindings reference defined executors
- ⚠ Warning: executor binary 'cursor-agent' not found on PATH

**Output example:**
```
Validating configuration at /Users/josh/.config/sow/config.yaml...

✓ YAML syntax valid
✓ CUE schema valid
✓ Executor types valid (claude, cursor)
✓ All bindings reference defined executors
⚠ Warning: cursor-agent binary not found on PATH
  The 'cursor' executor requires cursor-agent CLI to be installed.
  Install: [link to installation instructions]

Configuration is valid with 1 warning.
```

#### Additional Config Commands

**Show effective configuration:**
```bash
sow config show
```

Displays current effective configuration (merged defaults + file + env vars):

```yaml
# Effective configuration (merged from defaults + file + environment)
# Config file: /Users/josh/.config/sow/config.yaml (exists)
# Environment overrides: SOW_AGENTS_IMPLEMENTER=cursor

agents:
  executors:
    claude-code:
      type: "claude"
      settings:
        yolo_mode: false    # (default)
        model: "sonnet"     # (from config file)
    cursor:
      type: "cursor"
      settings:
        yolo_mode: true     # (from config file)
  bindings:
    orchestrator: "claude-code"  # (from config file)
    implementer: "cursor"        # (from environment: SOW_AGENTS_IMPLEMENTER)
    architect: "claude-code"     # (from config file)
    # ...
```

**Edit configuration:**
```bash
sow config edit
```

- Opens config file in `$EDITOR`
- Creates file with template if doesn't exist
- Validates after edit (optional confirmation)

**Reset to defaults:**
```bash
sow config reset
```

- Removes user config file (with confirmation)
- Reverts to built-in defaults
- Preserves backup at `~/.config/sow/config.yaml.backup`

#### Cross-Platform Configuration Paths

**Path resolution uses Go's `os.UserConfigDir()`:**

- **Linux/Mac**: `~/.config/sow/config.yaml`
- **Windows**: `%APPDATA%\sow\config.yaml`
  - Typically: `C:\Users\<user>\AppData\Roaming\sow\config.yaml`

**Implementation:**
```go
import "os"

func GetUserConfigPath() (string, error) {
    configDir, err := os.UserConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(configDir, "sow", "config.yaml"), nil
}
```

**All error messages and help text show platform-appropriate paths.**

#### Configuration File Schema (CUE)

```cue
// cli/schemas/user_config.cue

#UserConfig: {
    agents: {
        executors: [string]: {
            type: "claude" | "cursor" | "windsurf"
            settings: {
                yolo_mode?: bool
                model?: string  // Only for claude type
            }
            custom_args?: [...string]
        }

        bindings: {
            orchestrator: string
            implementer: string
            architect: string
            reviewer: string
            planner: string
            researcher: string
            // Additional agents as they're added
        }
    }
}
```

**Validation ensures:**
- Executor types are recognized
- Bindings reference defined executors
- Settings match executor type (e.g., `model` only for Claude)
- Structure matches expected schema

## Session Management Protocol

### Session Lifecycle

**Session resumption enables bidirectional communication between orchestrator and workers.**

Both Claude Code and Cursor support resumable sessions:
- Claude: `--session-id <uuid>` / `--resume <uuid>`
- Cursor: `--chat-id <uuid>` / `--resume <uuid>`

**Lifecycle:**

1. **Orchestrator generates session UUID** before spawning
2. **Persist to task `state.yaml`** immediately (before subprocess invocation)
3. **Spawn worker** with session ID via executor
4. **Worker subprocess exits** (status: `paused`, `needs_review`, `failed`)
5. **Orchestrator reads state** to determine status
6. **If paused or needs corrections**:
   - Orchestrator resumes with `--resume <uuid>`
   - Worker continues with feedback/answer
   - Repeat from step 4
7. **If approved**:
   - Orchestrator sets status to `completed`
   - Session ID cleared (task reaches terminal state)

**Critical**: Session ID must be persisted before spawning. When subprocess exits, can't query for ID - must read from state. Session persists through review iterations until work is approved.

### Task State Schema Update

```cue
// cli/schemas/task_state.cue

#TaskState: {
    task: {
        id: string
        name: string
        status: "pending" | "in_progress" | "needs_review" | "completed" | "abandoned" | "paused"
        assigned_agent: string
        iteration: int

        // Session ID for resumable agent conversations
        // Set when worker spawned, persists through review iterations
        // Only cleared when task reaches approved terminal state (completed/abandoned)
        session_id?: string @go(,optional=nillable)

        // ... other fields
    }
}
```

### Paused Workflow Pattern

**Use case**: Worker needs input from orchestrator (e.g., architectural decision, clarification)

**Flow:**

1. Worker encounters blocker
2. Worker sets status to `paused` in `state.yaml`
3. Worker writes question/context to task log
4. Worker exits (subprocess terminates)
5. Orchestrator reads state, sees `paused` status
6. Orchestrator reads task log to understand blocker
7. Orchestrator provides answer/decision
8. Orchestrator resumes session with response:
   ```bash
   sow agent resume 010 "Use RS256 algorithm for JWT signing"
   ```
9. CLI calls `executor.Resume(sessionID, prompt)`
10. Worker continues from where it left off

**Benefits:**
- Enables complex workflows requiring back-and-forth
- Maintains conversation context across subprocess invocations
- Works universally across Claude Code and Cursor

### Review Iteration Pattern

**Use case**: Worker completes task, but orchestrator review identifies issues requiring corrections

**Flow:**

1. Worker completes work
2. Worker sets status to `needs_review` in `state.yaml`
3. Worker writes summary to task log
4. Worker exits (subprocess terminates)
5. Orchestrator reads state, sees `needs_review` status
6. Orchestrator validates work (reads code, runs tests, checks against requirements)
7. **If issues found**:
   - Orchestrator documents issues in task log or feedback file
   - Orchestrator resumes session with corrections:
     ```bash
     sow agent resume 010 "The authentication logic has a security issue: passwords are logged in plaintext on line 42. Remove the log statement and add a unit test to verify passwords are never logged."
     ```
8. CLI calls `executor.Resume(sessionID, prompt)`
9. Worker receives feedback, makes corrections
10. Worker sets status to `needs_review` again
11. **Repeat until approved**
12. When work is acceptable:
    - Orchestrator sets status to `completed`
    - Session ID cleared (task reaches terminal state)

**Critical design points:**

- **Session persists through review cycles** - Don't clear session_id when status is `needs_review`
- **Iteration counter tracks attempts** - Can be used for quality metrics or escalation
- **Feedback mechanism** - Orchestrator can provide detailed corrections via resume prompt
- **Conversation continuity** - Worker has full context from previous attempts

**Benefits:**
- Enables iterative refinement of work
- Maintains conversation context across correction cycles
- Worker learns from mistakes within same task
- Supports guidance-based consistency model (see Consistency ADR)

**Example iteration sequence:**

```
Iteration 1: Worker implements → needs_review → Issues found
Iteration 2: Worker corrects → needs_review → More issues found
Iteration 3: Worker corrects → needs_review → Approved → completed
```

## State-Based Subprocess Protocol

**Core principle**: Subprocess exit is the signal, `state.yaml` is the message.

### Protocol Flow

1. **Orchestrator spawns worker** via `sow agent spawn`
2. **CLI invokes executor** with agent, prompt, session ID
3. **Executor spawns subprocess** (blocking call)
4. **Worker executes**, updates `state.yaml` with status
5. **Worker subprocess exits** (process terminates)
6. **Executor.Spawn() returns** (unblocked)
7. **Orchestrator reads `state.yaml`** to understand outcome

### Status Codes

Worker sets status in `state.yaml` to communicate outcome:

- `needs_review`: Worker completed successfully, requesting review
- `paused`: Worker blocked, needs input from orchestrator
- `failed`: Worker encountered error, cannot proceed
- `completed`: Work approved and finalized (rare - usually orchestrator sets this)

### Why This Works

**Universal across all agent CLIs:**
- No special protocol or API required
- Works with any subprocess (Claude, Cursor, Windsurf, shell scripts)
- Simple file-based communication
- Zero-context resumability (all state on disk)

**Orchestrator reads, worker writes:**
- Clear ownership model
- No race conditions (subprocess exited before orchestrator reads)
- Validates against CUE schema when worker uses `sow task set status`

## File Structure

```
cli/internal/agents/
├── agents.go           # All agent definitions (Implementer, Architect, etc.)
├── registry.go         # AgentRegistry implementation
├── templates.go        # Embedded template loading (//go:embed)
├── executor.go         # Executor interface + implementations + ExecutorRegistry
└── templates/
    ├── implementer.md  # Agent prompt templates (embedded)
    ├── architect.md
    ├── reviewer.md
    ├── planner.md
    ├── researcher.md
    └── ... (12+ templates)
```

**Design rationale:**
- All agents in one file - easy discovery and maintenance
- Executors in one file - interface + implementations + registry together
- Templates embedded - versioned with sow binary, no external dependencies
- Simple structure - matches 12+ agents, 2-3 executors constraint

## CLI Commands

### Spawn Agent

```bash
sow agent spawn <agent-name> <task-id>
```

**Example:**
```bash
sow agent spawn implementer 010
```

**Implementation:**

```go
func runSpawn(cmd *cobra.Command, args []string) error {
    agentName := args[0]
    taskID := args[1]

    ctx := cmdutil.GetContext(cmd.Context())

    // Get agent definition
    agent, err := ctx.AgentRegistry.Get(agentName)
    if err != nil {
        return err
    }

    // Get executor from user config bindings
    executorName := ctx.UserConfig.Agents.Bindings[agentName]
    executor, err := ctx.ExecutorRegistry.Get(executorName)
    if err != nil {
        return err
    }

    // Load or create session ID
    taskState, err := loadTaskState(ctx, taskID)
    if err != nil {
        return err
    }

    if taskState.SessionID == "" {
        taskState.SessionID = uuid.New().String()
        if err := saveTaskState(ctx, taskState); err != nil {
            return err
        }
    }

    // Build task-specific prompt
    prompt := buildTaskPrompt(ctx, taskID)

    // Spawn agent (blocks until subprocess exits)
    if err := executor.Spawn(ctx, agent, prompt, taskState.SessionID); err != nil {
        return fmt.Errorf("agent execution failed: %w", err)
    }

    // Agent exited - orchestrator will check state.yaml for status
    return nil
}
```

### Resume Agent

```bash
sow agent resume <task-id> <prompt>
```

**Example:**
```bash
sow agent resume 010 "Use RS256 algorithm for JWT signing"
```

**Implementation:**

```go
func runResume(cmd *cobra.Command, args []string) error {
    taskID := args[0]
    prompt := args[1]

    ctx := cmdutil.GetContext(cmd.Context())

    // Load task state to get session ID and agent
    taskState, err := loadTaskState(ctx, taskID)
    if err != nil {
        return err
    }

    if taskState.SessionID == "" {
        return fmt.Errorf("no session found for task %s", taskID)
    }

    // Get executor from user config bindings
    executorName := ctx.UserConfig.Agents.Bindings[taskState.AssignedAgent]
    executor, err := ctx.ExecutorRegistry.Get(executorName)
    if err != nil {
        return err
    }

    if !executor.SupportsResumption() {
        return fmt.Errorf("executor %s does not support session resumption", executorName)
    }

    // Resume (blocks until subprocess exits)
    if err := executor.Resume(ctx, taskState.SessionID, prompt); err != nil {
        return fmt.Errorf("agent resumption failed: %w", err)
    }

    return nil
}
```

### Config Commands

Configuration management commands for user preferences.

#### Initialize Config

```bash
sow config init
```

Creates `~/.config/sow/config.yaml` with commented template.

#### Show Config Path

```bash
sow config path
sow config path --exists
```

Displays the configuration file location and optionally checks if it exists.

#### Validate Config

```bash
sow config validate
```

Validates configuration file:
- YAML syntax
- CUE schema conformance
- Executor type validity
- Binding references
- Binary availability (warnings)

#### Show Effective Config

```bash
sow config show
```

Displays merged configuration (defaults + file + env vars) with source annotations.

#### Edit Config

```bash
sow config edit
```

Opens configuration file in `$EDITOR`. Creates with template if doesn't exist.

#### Reset Config

```bash
sow config reset
```

Removes user configuration, reverts to defaults. Creates backup before removal.

## Complete Usage Flow

### Example: Spawn Implementer from Orchestrator

**Orchestrator (Claude Code) prompted to spawn implementer:**

```bash
sow agent spawn implementer 010
```

**CLI execution:**

1. Load agent definition: `agentRegistry.Get("implementer")`
2. Get executor from user config: `executorRegistry.Get("cursor")` (user configured Cursor for implementer)
3. Load/generate session ID from task state
4. Build task prompt: `"Execute task 010. Location: .sow/project/phases/implementation/tasks/010/"`
5. Invoke executor: `executor.Spawn(ctx, agent, prompt, sessionID)`

**Executor spawns subprocess:**

```bash
cursor-agent agent --chat-id <uuid>
```

**Worker (Cursor) subprocess:**

1. Receives agent prompt template + task prompt
2. Runs `sow prompt guidance/implementer/base`
3. Reads `state.yaml`, `description.md`, `feedback/`
4. Performs implementation work
5. Logs actions via `sow agent log "Implemented authentication"`
6. Updates state via `sow task set --id 010 status needs_review`
7. Exits

**CLI returns (subprocess exited):**

Orchestrator reads state:

```bash
cat .sow/project/phases/implementation/tasks/010/state.yaml
```

Status indicates outcome:
- `needs_review`: Worker completed successfully
- `paused`: Worker blocked, needs input
- `failed`: Worker encountered error

**If paused with blocker:**

```bash
sow agent resume 010 "Use RS256 algorithm for JWT signing"
```

CLI resumes conversation:

```go
executor.Resume(ctx, sessionID, "Use RS256...")
```

Worker continues from where it left off.

### Example: Review Iteration with Corrections

**Worker completes task:**

Status set to `needs_review`, subprocess exits.

**Orchestrator validates work:**

Orchestrator reads code, runs tests, finds security issue.

**Orchestrator resumes with feedback:**

```bash
sow agent resume 010 "The authentication logic has a security issue: passwords are logged in plaintext on line 42 of auth.go. Remove the log statement and add a unit test to verify passwords are never logged."
```

**CLI resumes session:**

```go
executor.Resume(ctx, sessionID, "The authentication logic...")
```

**Worker (Cursor) subprocess continues:**

1. Receives feedback in conversation context
2. Reads current code at `auth.go:42`
3. Removes plaintext password logging
4. Adds unit test for password logging
5. Runs tests to verify fix
6. Logs actions: `"Fixed security issue: removed password logging, added test"`
7. Updates state: `sow task set --id 010 status needs_review`
8. Exits

**Orchestrator validates again:**

Orchestrator reads updated code, runs tests, all checks pass.

**Orchestrator approves:**

```bash
sow task set --id 010 status completed
```

Session ID cleared, task reaches terminal state.

**Key points:**
- Session persisted through correction cycle
- Worker had full conversation context (original task + feedback)
- Iteration counter incremented (tracks quality metrics)
- Same pattern works for multiple correction rounds

## Integration Points

### With Existing Sow Components

**Project state management:**
- Task `state.yaml` updated to include `session_id` field
- CUE schema validation ensures structural consistency
- No changes to phase state or project state

**Prompt system:**
- Workers use `sow prompt` to load embedded guidance
- Agent templates reference prompt paths
- Orchestrator guidance updated to use `sow agent spawn/resume`

**Validation:**
- CUE schemas validate task state when workers use `sow task set`
- Orchestrator guided to validate worker output (guidance-based, not programmatic)

### With Claude Code

**Current orchestrator behavior:**
- Uses `.claude/agents/` directory for worker prompts
- Direct subprocess invocation
- No session management

**New orchestrator behavior:**
- Uses `sow agent spawn` command
- CLI handles executor selection and session management
- Session resumption enables bidirectional communication

**Migration:**
- Agent prompts migrated from `.claude/agents/` to `cli/internal/agents/templates/`
- Orchestrator prompt updated with new spawn/resume commands
- Worker prompts updated to be executor-agnostic

## Error Handling

### Executor Not Found

When user configures unknown executor:

```yaml
bindings:
  implementer: "copilot"  # Not registered
```

**Error message:**
```
Error: unknown executor: copilot

Available executors:
  - claude-code
  - cursor
  - windsurf

Check your configuration at ~/.config/sow/config.yaml
```

### Executor Binary Not Found

When configured executor binary doesn't exist on PATH:

```yaml
executors:
  cursor:
    type: "cursor"
```

But `cursor-agent` not installed.

**Error message:**
```
Error: executor binary not found: cursor-agent

Executor "cursor" requires the cursor-agent CLI to be installed and available on PATH.

Install cursor-agent: [link to installation instructions]
```

**Implementation**: Validate executor availability at CLI initialization (fail fast).

### Session Resumption Not Supported

When attempting to resume session with executor that doesn't support it:

```bash
sow agent resume 010 "Use RS256"
```

But task assigned to Windsurf (no session support).

**Error message:**
```
Error: executor "windsurf" does not support session resumption

Cannot resume task 010. Consider:
  - Spawning a new worker instance
  - Switching to an executor with session support (claude-code, cursor)
```

### Session ID Missing

When attempting to resume but no session exists:

**Error message:**
```
Error: no session found for task 010

Task has not been started or session has expired.
Use `sow agent spawn implementer 010` to start a new session.
```

## Testing Strategy

### Unit Tests

**Agent registry:**
- Test agent registration and lookup
- Test unknown agent error handling
- Test agent listing

**Executor implementations:**
- Mock subprocess execution
- Verify correct CLI flags passed
- Verify session ID handling
- Test error propagation

**User configuration:**
- Test config parsing
- Test priority order (env vars → user config → defaults)
- Test invalid config validation

### Integration Tests

**Spawn and resume flow:**
- Test spawning worker with session ID
- Verify session ID persisted to state
- Test resuming session with additional prompt
- Verify state-based protocol

**Cross-executor compatibility:**
- Same task executed by Claude vs Cursor
- Verify structural consistency (state.yaml format)
- Verify behavioral consistency (task completion)

**Configuration scenarios:**
- Default configuration (all Claude)
- Mixed configuration (Claude orchestrator, Cursor implementer)
- Environment variable overrides

### Acceptance Tests

**End-to-end workflows:**
- Feature implementation project using Cursor for implementer
- Design project using Claude for all agents
- Bug fix with paused workflow (resume session)

**Quality benchmarks:**
- Same task executed by multiple executors
- Compare output quality, completeness, correctness
- Establish baseline for cross-executor consistency

## Open Questions

### 1. Validate executor availability at startup?

**Question**: Should we check if `claude`, `cursor-agent`, etc. exist on PATH when loading config?

**Recommendation**: Yes, fail fast with clear error if configured executor not found.

**Implementation**: Add `ValidateAvailability()` to executor interface, call during registry initialization.

### 2. Executor-specific session flags

**Decision**: Session flags are hardcoded in executor implementations.

**Rationale**: Session flags (`--session-id`, `--chat-id`, `--resume`) are intrinsic to each CLI's design and never change. They're implementation details, not user preferences. Hardcoding them:
- Keeps user configuration simpler
- Prevents misconfiguration
- Makes executors easier to maintain
- Matches the actual usage pattern (flags don't vary per user)

### 3. User-extensible agents?

**Question**: Allow users to define custom agents in `~/.config/sow/agents/custom-agent.md`?

**Recommendation**: Defer to v2. Start with built-in agents only. Complexity not justified yet.

### 4. Agent prompt versioning?

**Question**: If we update `implementer.md` template, how do older projects handle it?

**Recommendation**:
- Embed templates in binary
- Version via sow version
- Projects on older sow versions use older templates
- No migration needed

## Success Criteria

**Must have:**
- ✅ User can configure which executor handles which agent role
- ✅ Works with Claude Code (maintain current functionality)
- ✅ Session resumption enables bidirectional communication
- ✅ CUE schema validation enforces structural consistency
- ✅ Simple to add new agents (define struct + template)
- ✅ Simple to add new executors (implement interface)

**Nice to have:**
- ✅ Works with Cursor CLI
- ✅ Works with Windsurf (if they have CLI)
- ⚠️ Parallel task spawning (via bash `&` or CLI helper) - deferred to implementation
- ⚠️ Live progress monitoring - deferred to implementation

**Non-goals:**
- ❌ Programmatic validation enforcement (impossible - orchestrator is agent)
- ❌ Identical output across all models (unrealistic - models vary)
- ❌ Agent composition/inheritance (defer to v2)

## References

**Exploration artifacts:**
- `.sow/knowledge/explorations/custom-agents.md` - Complete exploration summary
- `user-config-strategy.md` - User-level configuration approach
- `agent-abstraction-layer.md` - Initial abstraction design
- `cross-agent-consistency.md` - Consistency mechanisms
- `agent-executor-architecture.md` - First architecture proposal
- `revised-architecture.md` - Simplified architecture (final)

**Related ADRs:**
- Consistency Model ADR (in progress) - Four-layer consistency approach

## Appendix: Configuration Examples

### Minimal Configuration (Defaults)

```yaml
# Implicit defaults when ~/.config/sow/config.yaml doesn't exist
agents:
  executors:
    claude-code:
      type: "claude"
      settings:
        yolo_mode: false
  bindings:
    orchestrator: "claude-code"
    implementer: "claude-code"
    architect: "claude-code"
    reviewer: "claude-code"
    planner: "claude-code"
    researcher: "claude-code"
```

### Mixed Executor Configuration

```yaml
# Claude for orchestration, Cursor for implementation
agents:
  executors:
    claude-code:
      type: "claude"
      settings:
        yolo_mode: true
        model: "sonnet"

    cursor:
      type: "cursor"
      settings:
        yolo_mode: true

  bindings:
    orchestrator: "claude-code"
    implementer: "cursor"      # Use Cursor for implementation
    architect: "claude-code"
    reviewer: "claude-code"
    planner: "claude-code"
    researcher: "claude-code"
```

### Advanced Configuration

```yaml
# Multiple executors with custom settings
agents:
  executors:
    claude-opus:
      type: "claude"
      settings:
        yolo_mode: true
        model: "opus"

    claude-sonnet:
      type: "claude"
      settings:
        yolo_mode: true
        model: "sonnet"

    cursor:
      type: "cursor"
      settings:
        yolo_mode: true

  bindings:
    orchestrator: "claude-opus"    # Use Opus for orchestration
    implementer: "cursor"           # Use Cursor for implementation
    architect: "claude-opus"        # Use Opus for architecture
    reviewer: "claude-sonnet"       # Use Sonnet for reviews
    planner: "claude-sonnet"        # Use Sonnet for planning
    researcher: "claude-sonnet"     # Use Sonnet for research
```
