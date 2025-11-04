# Custom Agents Exploration - Summary

**Date:** January 2025
**Branch:** explore/custom-agents
**Status:** Research complete, ready for implementation

## What We Explored

This exploration investigated how to extend sow beyond Claude Code to support multiple AI agent CLIs (Cursor, Windsurf, etc.) while maintaining consistent orchestration patterns and subscription-based economics.

**Three main research areas:**

1. **User-level configuration strategy** - How to introduce user preferences (agent selection) without breaking sow's repository-scoped design
2. **Agent abstraction layer** - How to decouple agent invocation from Claude Code specifics
3. **Cross-agent consistency mechanisms** - How to ensure reliable behavior across different models and CLIs

## Key Findings

### 1. Orchestrator is an Agent, Not Code

**Critical architectural insight**: The orchestrator is another AI agent (Claude Code, Cursor, etc.), not Go code. This fundamentally changes the consistency model.

**Implications:**
- Cannot programmatically enforce validation - orchestrator reads files and makes judgments
- Only hard enforcement point is `sow` CLI validating CUE schemas
- Everything else (validation, feedback, iteration) is guidance-based via prompt engineering
- Both orchestrator and workers need extremely explicit prompts

### 2. Worker Self-Initialization Pattern

Workers are not given complete instructions in spawn prompt. Instead:
- Spawn prompt provides minimal context (task location, iteration)
- Worker runs `sow prompt` commands to load detailed instructions
- Instructions stored in embedded prompt templates

**This simplifies abstraction** - orchestrator just needs to invoke CLI and provide task location. Worker handles the rest.

### 3. Session Resumption Enables Bidirectional Communication

Both Claude Code and Cursor support resumable sessions:
- `claude --session-id <uuid>` / `claude --resume <uuid>`
- `cursor-agent agent --chat-id <uuid>` / `cursor-agent agent --resume <uuid>`

**Session lifecycle:**
1. Generate session UUID before spawning
2. Persist to task `state.yaml` immediately (before subprocess invocation)
3. Spawn worker with session ID
4. Worker subprocess exits (paused/completed)
5. Orchestrator reads state to determine status
6. If paused, orchestrator resumes with `--resume <uuid>`

**Critical**: Session ID must be persisted before spawning. When subprocess exits, can't query for ID - must read from state.

### 4. State-Based Protocol

**Subprocess exit is the signal, state.yaml is the message**:
- Orchestrator waits for subprocess to exit (blocking)
- When process exits, orchestrator reads `state.yaml` for status
- Status indicates what happened: `needs_review`, `paused`, `failed`
- Works universally across all agent CLIs

### 5. Four-Layer Consistency Model

**Layer 1: Structural consistency (CLI-enforced)**
- CUE schema validation when agents use `sow` CLI commands
- Only programmatic enforcement point

**Layer 2: Behavioral guidance (prompt engineering)**
- Explicit, imperative instructions for both orchestrator and workers
- Example commands, validation checklists, common mistakes documented

**Layer 3: Validation (orchestrator guidance)**
- Orchestrator agent is guided to validate worker output
- Not code enforcement - another agent reading files and judging

**Layer 4: Iteration loop (self-correction)**
- Feedback mechanism when validation fails
- Iteration forces compliance through corrections

## User Configuration Format

### Location

`~/.config/sow/config.yaml` (cross-platform)

Following precedent from refs cache at `~/.cache/sow/refs/`.

### Complete Configuration Structure

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
      session:
        create_flag: "--session-id"
        resume_flag: "--resume"
      custom_args: []            # Additional CLI flags

    cursor:
      type: "cursor"
      settings:
        yolo_mode: true          # Maps to cursor's equivalent permission flag
      session:
        create_flag: "--chat-id"
        resume_flag: "--resume"
      custom_args: []

    windsurf:
      type: "windsurf"
      settings:
        yolo_mode: true
      session:
        create_flag: null        # Doesn't support sessions
        resume_flag: null
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

### Configuration Fields Explained

**Executor definition:**
- `type`: Executor implementation to use (`claude`, `cursor`, etc.)
- `settings`: Common settings abstracted across executors
  - `yolo_mode`: Skip permission prompts (mapped to executor-specific flags)
  - `model`: Which underlying AI model to use (if applicable)
- `session`: Session support configuration
  - `create_flag`: Flag to pass session ID when spawning
  - `resume_flag`: Flag to resume existing session
  - Set to `null` if executor doesn't support sessions
- `custom_args`: Executor-specific flags that don't fit common patterns

**Bindings:**
- Map agent role (implementer, architect, etc.) to executor name
- User preference - most modern agents can handle any role
- Allows mixing executors (Claude for orchestrator, Cursor for implementer)

### Configuration Priority

1. **Environment variable override**: `SOW_AGENTS_IMPLEMENTER=windsurf sow agent spawn implementer 010`
2. **User config**: `~/.config/sow/config.yaml`
3. **Defaults**: Claude Code for all agents

### No Repository-Level Agent Config

Agent preferences are user-specific (tied to subscriptions and preferences), not repository-specific. User config is never committed to git.

Repository config (`.sow/config.yaml`) remains for artifact paths and repository settings only.

## Agent Management Architecture

### Core Principles

**Constraints:**
- 12+ agents (lightweight, easily extensible)
- 2-3 executors (heavyweight, complex integrations)
- Agents = name + capabilities + prompt
- Executors = CLI invocation + session management

**Design decisions:**
- Agents are structs (data, not behavior)
- Executors are interface (varying behavior)
- Prompt templates embedded in binary
- No Session object - Executor methods block directly
- Simple file structure in `cli/internal/agents/`

### Agent Definition

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

**All agents defined in one file:**

```go
// cli/internal/agents/agents.go

package agents

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

    Planner = &Agent{
        Name:         "planner",
        Description:  "Task breakdown and planning",
        Capabilities: "Must be able to read/write files, search codebase",
        PromptPath:   "planner.md",
    }

    Researcher = &Agent{
        Name:         "researcher",
        Description:  "Research and information gathering",
        Capabilities: "Must be able to read files, search codebase, fetch web content",
        PromptPath:   "researcher.md",
    }

    // ... more agents (easily extensible - just add struct)
)

// StandardAgents returns all built-in agents
func StandardAgents() []*Agent {
    return []*Agent{
        Implementer,
        Architect,
        Reviewer,
        Planner,
        Researcher,
        // ...
    }
}
```

**Why struct, not interface:**
- Agents are lightweight configuration, not varying behavior
- Easy to add new agents - just define struct
- Can serialize/deserialize
- No method overhead

### Agent Registry

```go
// cli/internal/agents/registry.go

package agents

import "fmt"

type AgentRegistry struct {
    agents map[string]*Agent
}

func NewAgentRegistry() *AgentRegistry {
    r := &AgentRegistry{
        agents: make(map[string]*Agent),
    }

    // Register all standard agents
    for _, agent := range StandardAgents() {
        r.Register(agent)
    }

    return r
}

func (r *AgentRegistry) Register(agent *Agent) {
    r.agents[agent.Name] = agent
}

func (r *AgentRegistry) Get(name string) (*Agent, error) {
    agent, ok := r.agents[name]
    if !ok {
        return nil, fmt.Errorf("unknown agent: %s", name)
    }
    return agent, nil
}

func (r *AgentRegistry) List() []*Agent {
    agents := make([]*Agent, 0, len(r.agents))
    for _, agent := range r.agents {
        agents = append(agents, agent)
    }
    return agents
}
```

### Embedded Prompt Templates

```go
// cli/internal/agents/templates.go

package agents

import (
    "embed"
    "fmt"
    "io/fs"
)

//go:embed templates/*
var templatesFS embed.FS

// LoadPrompt loads an agent prompt template from embedded files
func LoadPrompt(promptPath string) (string, error) {
    data, err := fs.ReadFile(templatesFS, "templates/"+promptPath)
    if err != nil {
        return "", fmt.Errorf("failed to load prompt %s: %w", promptPath, err)
    }
    return string(data), nil
}
```

**Template example:**

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
    └── {id}.md
```

Start by reading state.yaml to get your task ID and iteration number.
```

### Executor Interface

```go
// cli/internal/agents/executor.go

package agents

import "context"

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

**Why interface:**
- Executors have varying behavior (different CLIs, session mechanisms)
- Easy to add new executors (implement interface)
- Type-safe invocation
- Complexity isolated per implementation

**Key design decision**: Methods block directly (no Session object returned). CLI is stateless - each invocation is independent.

### Executor Implementations

```go
// cli/internal/agents/executor.go (continued)

import (
    "fmt"
    "os"
    "os/exec"
    "strings"
)

// ClaudeExecutor invokes Claude Code
type ClaudeExecutor struct {
    yoloMode bool
    model    string
}

func NewClaudeExecutor(yoloMode bool, model string) *ClaudeExecutor {
    return &ClaudeExecutor{
        yoloMode: yoloMode,
        model:    model,
    }
}

func (e *ClaudeExecutor) Name() string {
    return "claude-code"
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

// CursorExecutor invokes Cursor CLI
type CursorExecutor struct {
    yoloMode bool
}

func NewCursorExecutor(yoloMode bool) *CursorExecutor {
    return &CursorExecutor{yoloMode: yoloMode}
}

func (e *CursorExecutor) Name() string {
    return "cursor"
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

### Executor Registry

```go
// cli/internal/agents/executor.go (continued)

// ExecutorRegistry manages available executors
type ExecutorRegistry struct {
    executors map[string]Executor
}

func NewExecutorRegistry() *ExecutorRegistry {
    return &ExecutorRegistry{
        executors: make(map[string]Executor),
    }
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

// LoadExecutors creates executors from user configuration
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

### File Structure

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

### CLI Command Example

```go
// cli/cmd/agent/spawn.go

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

### Task State Schema Update

```cue
// cli/schemas/task_state.cue

#TaskState: {
    task: {
        id: string
        name: string
        status: "pending" | "in_progress" | "needs_review" | "completed" | "abandoned"
        assigned_agent: string
        iteration: int

        // NEW: Session ID for resumable agent conversations
        // Set when worker spawned, cleared when task reaches terminal state
        session_id?: string @go(,optional=nillable)

        // ... other fields
    }
}
```

## Usage Flow

### 1. Orchestrator Spawns Worker

Orchestrator agent (Claude Code) is prompted to spawn implementer:

```bash
sow agent spawn implementer 010
```

### 2. CLI Execution

```go
// Load agent definition
agent := agentRegistry.Get("implementer")

// Get executor from user config
executor := executorRegistry.Get("cursor") // User configured Cursor for implementer

// Load/generate session ID
sessionID := getOrCreateSessionID(taskID)

// Build prompt
prompt := "Execute task 010. Location: .sow/project/phases/implementation/tasks/010/"

// Spawn (blocks)
executor.Spawn(ctx, agent, prompt, sessionID)
```

### 3. Worker Execution

Cursor subprocess:
1. Receives agent prompt template + task prompt
2. Runs `sow prompt guidance/implementer/base`
3. Reads `state.yaml`, `description.md`, `feedback/`
4. Performs work
5. Logs via `sow agent log`
6. Updates state via `sow agent task state set-status needs_review`
7. Exits

### 4. Orchestrator Validation

CLI returns (subprocess exited), orchestrator reads state:

```bash
cat .sow/project/phases/implementation/tasks/010/state.yaml
```

Status indicates outcome:
- `needs_review`: Worker completed successfully
- `paused`: Worker blocked, needs input
- `failed`: Worker encountered error

### 5. Resume if Paused

If worker paused with blocker:

```bash
sow agent resume 010 "Use RS256 algorithm for JWT signing"
```

CLI resumes with conversation history:

```go
executor.Resume(ctx, sessionID, "Use RS256...")
```

Worker continues from where it left off.

## Implementation Phases

### Phase 1: Foundation (MVP)

1. **User config loading**
   - Parse `~/.config/sow/config.yaml`
   - CUE schema for validation
   - Environment variable overrides

2. **Agent registry**
   - Define standard agents (implementer, architect, reviewer, etc.)
   - Simple registry with Get/List methods
   - Embed prompt templates

3. **Claude executor only**
   - Implement ClaudeExecutor
   - Session support (spawn + resume)
   - Test with existing sow workflows

4. **CLI commands**
   - `sow agent spawn <agent> <task-id>`
   - `sow agent resume <task-id> <prompt>`
   - Task state schema with session_id field

### Phase 2: Multi-Executor Support

1. **Cursor executor**
   - Implement CursorExecutor
   - Test session resumption
   - Validate cross-executor consistency

2. **Configuration flexibility**
   - Common settings abstraction (yolo_mode)
   - Custom args support
   - Per-executor configuration

3. **Testing**
   - Agent compatibility test suite
   - Benchmark tasks for quality comparison
   - Cross-executor validation

### Phase 3: Orchestrator Guidance

1. **Update orchestrator prompts**
   - Worker spawning instructions
   - Session management guidance
   - Validation checklists
   - Parallel execution patterns (bash `&`)

2. **Worker prompt templates**
   - Update to be executor-agnostic
   - Clear sow CLI command examples
   - Failure mode documentation

### Phase 4: Polish

1. **Error handling**
   - Better error messages
   - Executor-not-found guidance
   - Session resumption failures

2. **Documentation**
   - User guide for configuring executors
   - Agent compatibility matrix
   - Troubleshooting guide

3. **Optional enhancements**
   - Parallel spawn helper (if needed)
   - Progress monitoring (if needed)
   - Timeout support (if needed)

## Open Questions for Implementation

### 1. Should we validate executor availability at startup?

Check if `claude`, `cursor-agent`, etc. exist on PATH when loading config?

**Recommendation**: Yes, fail fast with clear error if configured executor not found.

### 2. How to handle executor-specific session flags?

Currently hardcoded (`--session-id` vs `--chat-id`). Should this come from config?

**Current approach**: Hardcode in executor implementations (simple, type-safe)
**Alternative**: Store in config (`session.create_flag`, `session.resume_flag`)

**Recommendation**: Start hardcoded, move to config if we need dynamic configuration.

### 3. Should agents be user-extensible?

Allow users to define custom agents in `~/.config/sow/agents/custom-agent.md`?

**Recommendation**: Defer to v2. Start with built-in agents only.

### 4. How to version agent prompts?

If we update implementer.md template, how do older projects handle it?

**Recommendation**: Embed templates in binary. Version via sow version. Projects on older sow versions use older templates. No migration needed.

## Success Criteria

**Must have:**
- ✅ User can configure which executor handles which agent role
- ✅ Works with Claude Code (maintain current functionality)
- ✅ Session resumption enables bidirectional communication
- ✅ CUE schema validation enforces structural consistency
- ✅ Simple to add new agents (just define struct + template)

**Nice to have:**
- ✅ Works with Cursor CLI
- ✅ Works with Windsurf (if they have CLI)
- ✅ Parallel task spawning (via bash or CLI helper)
- ✅ Live progress monitoring

**Non-goals:**
- Programmatic validation enforcement (impossible - orchestrator is agent)
- Identical output across all models (unrealistic - models vary)
- Agent composition/inheritance (defer to v2)

## References

Created during exploration:
- `user-config-strategy.md` - User-level configuration approach
- `agent-abstraction-layer.md` - Initial abstraction design
- `cross-agent-consistency.md` - Consistency mechanisms
- `agent-executor-architecture.md` - First architecture proposal
- `revised-architecture.md` - Simplified architecture (final)

## Participants

**Conducted:** January 2025
**Participants:** Josh Gilman, Claude (Sonnet 4.5)
