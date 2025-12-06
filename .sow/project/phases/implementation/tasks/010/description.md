# Agent Command Package and List Subcommand

## Context

This task is part of implementing CLI commands for the sow orchestrator to spawn and resume worker agents. The overall goal is to enable task delegation to specialized agents through a session management protocol.

The multi-agent architecture design specifies that:
- Agents are lightweight configuration (roles, prompts) stored in `cli/internal/agents/`
- Executors handle CLI invocation and session management
- The orchestrator (an AI agent) uses `sow agent` commands to spawn workers

This task creates the foundation: the parent `agent` command and the `list` subcommand. The spawn and resume commands will be added in subsequent tasks.

## Requirements

### Parent Command (`cli/cmd/agent/agent.go`)

Create a new command package following the established pattern in `cli/cmd/config/` and `cli/cmd/refs/`:

1. **Package declaration**: Package name must be `agent`
2. **NewAgentCmd function**: Export a function `NewAgentCmd() *cobra.Command` that creates the parent command
3. **Command configuration**:
   - Use: `"agent"`
   - Short: `"Manage AI agents"`
   - Long: Multi-line description explaining the command group (spawn, resume, list)
4. **Subcommand registration**: Add subcommands using `cmd.AddCommand()`

### List Subcommand (`cli/cmd/agent/list.go`)

Implement the `sow agent list` command that displays all available agents:

1. **Command structure**:
   - Use: `"list"`
   - Short: `"List available agents"`
   - Long: Explanation of what agents are and how they're used
   - No arguments required
   - RunE: `runList`

2. **Implementation**:
   - Create a new `AgentRegistry` using `agents.NewAgentRegistry()`
   - Get all agents using `registry.List()`
   - Sort agents alphabetically by name for consistent output
   - Print formatted output with name and description

3. **Output format**:
   ```
   Available agents:
     architect     System design and architecture decisions
     decomposer    Specialized for decomposing complex features into project-sized, implementable work units
     implementer   Code implementation using Test-Driven Development
     planner       Research codebase and create comprehensive implementation task breakdown
     researcher    Focused, impartial research with comprehensive source investigation and citation
     reviewer      Code review and quality assessment
   ```
   - Use consistent column widths for alignment
   - Sort alphabetically for predictable output

### Root Command Integration (`cli/cmd/root.go`)

1. Import the agent package: `"github.com/jmgilman/sow/cli/cmd/agent"`
2. Add command registration: `cmd.AddCommand(agent.NewAgentCmd())`
3. Place it alphabetically among existing commands (after `NewAdvanceCmd()`)

## Acceptance Criteria

### Functional Requirements
- `sow agent` shows help with available subcommands
- `sow agent list` displays all 6 standard agents with descriptions
- `sow agent list` output is sorted alphabetically by agent name
- `sow agent --help` shows the agent command documentation

### Test Requirements (TDD approach)

Write tests FIRST, then implementation. Create `cli/cmd/agent/list_test.go`:

1. **TestNewAgentCmd_Structure**: Verify parent command has correct Use, Short, Long
2. **TestNewAgentCmd_HasSubcommands**: Verify subcommands are registered
3. **TestNewListCmd_Structure**: Verify list command has correct Use, Short, Long
4. **TestRunList_OutputsAllAgents**: Verify all 6 standard agents are in output
5. **TestRunList_SortedAlphabetically**: Verify output is sorted by agent name
6. **TestRunList_IncludesDescriptions**: Verify agent descriptions are shown

Test pattern (following `cli/cmd/config/init_test.go`):
```go
func TestRunList_OutputsAllAgents(t *testing.T) {
    cmd := &cobra.Command{}
    var buf bytes.Buffer
    cmd.SetOut(&buf)

    err := runList(cmd, nil)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    output := buf.String()
    expectedAgents := []string{"architect", "decomposer", "implementer", "planner", "researcher", "reviewer"}
    for _, agent := range expectedAgents {
        if !strings.Contains(output, agent) {
            t.Errorf("expected output to contain agent %q", agent)
        }
    }
}
```

## Technical Details

### File Structure
```
cli/cmd/agent/
    agent.go       # NewAgentCmd() - parent command
    agent_test.go  # Tests for parent command structure
    list.go        # newListCmd(), runList() - list subcommand
    list_test.go   # Tests for list command
```

### Package Imports

For `agent.go`:
```go
import (
    "github.com/spf13/cobra"
)
```

For `list.go`:
```go
import (
    "fmt"
    "sort"

    "github.com/jmgilman/sow/cli/internal/agents"
    "github.com/spf13/cobra"
)
```

### Naming Conventions
- Exported function: `NewAgentCmd()` (follows `NewConfigCmd`, `NewRefsCmd` pattern)
- Unexported subcommand creators: `newListCmd()` (lowercase, package-internal)
- Unexported run functions: `runList()` (lowercase, package-internal)

## Relevant Inputs

- `cli/cmd/config/config.go` - Parent command pattern to follow
- `cli/cmd/config/init.go` - Subcommand implementation pattern
- `cli/cmd/config/init_test.go` - Test patterns for command testing
- `cli/cmd/refs/refs.go` - Alternative parent command pattern
- `cli/cmd/root.go` - Where to register the new command
- `cli/internal/agents/agents.go` - Agent definitions and StandardAgents()
- `cli/internal/agents/registry.go` - AgentRegistry and NewAgentRegistry()
- `.sow/project/context/issue-101.md` - Issue requirements

## Examples

### Parent command pattern (from config.go):
```go
func NewConfigCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "config",
        Short: "Manage user configuration",
        Long: `Manage sow user configuration...

Commands:
  init      Create configuration file with template
  ...`,
    }

    cmd.AddCommand(newInitCmd())
    cmd.AddCommand(newPathCmd())
    // ...

    return cmd
}
```

### Subcommand pattern (from init.go):
```go
func newInitCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "init",
        Short: "Create configuration file with template",
        Long:  `Create a configuration file...`,
        RunE:  runInit,
    }
}

func runInit(cmd *cobra.Command, _ []string) error {
    // implementation
}
```

### Root command registration (from root.go):
```go
cmd.AddCommand(NewAdvanceCmd())
cmd.AddCommand(agent.NewAgentCmd())  // Add this line
cmd.AddCommand(config.NewConfigCmd())
```

## Dependencies

- No task dependencies - this is the first task
- Requires existing `cli/internal/agents/` package (already implemented)

## Constraints

- Do NOT implement spawn or resume commands - those are separate tasks
- Follow existing command patterns exactly
- The agent package must be in `cli/cmd/agent/` directory
- Tests must be written first (TDD)
- Output must go through `cmd.Print*` methods, not `fmt.Print*` directly (enables testing)
- Sort agents alphabetically for consistent, predictable output
