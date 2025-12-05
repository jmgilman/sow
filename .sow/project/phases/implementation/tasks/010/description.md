# Task 010: CUE Schema and Go Type Generation

## Context

This task is part of work unit 004: User Configuration System for sow. The goal is to allow users to configure which AI CLI tools (Claude Code, Cursor, Windsurf) handle which agent roles via `~/.config/sow/config.yaml`.

The sow system uses CUE schemas to define data structures, which are then used for:
1. Runtime validation of YAML files
2. Go type generation via `cue exp gengotypes`

This task establishes the foundation by defining the CUE schema for user configuration.

## Requirements

### Create CUE Schema File

Create `cli/schemas/user_config.cue` that defines:

1. **UserConfig root structure** with optional `agents` field
2. **ExecutorConfig** for defining AI CLI tools:
   - `type`: Must be one of `"claude" | "cursor" | "windsurf"`
   - `settings`: Optional settings object with `yolo_mode` (bool) and `model` (string)
   - `custom_args`: Optional array of strings for additional CLI flags
3. **AgentsConfig** containing:
   - `executors`: Map of executor name to ExecutorConfig
   - `bindings`: Map of agent role to executor name

### Follow Existing Patterns

The schema should follow the patterns established in:
- `cli/schemas/config.cue` - Simple optional fields with `@go(,optional=nillable)`
- `cli/schemas/refs_cache.cue` - More complex nested structures

### Schema Definition

```cue
package schemas

// UserConfig defines the schema for the user configuration file at:
// ~/.config/sow/config.yaml
//
// This allows users to configure which AI CLI executors handle which agent roles.
#UserConfig: {
    // Agent configuration
    agents?: {
        // Executor definitions
        // Keys are executor names (e.g., "claude-code", "cursor")
        executors?: [string]: {
            // Type of executor
            type: "claude" | "cursor" | "windsurf"

            // Executor settings
            settings?: {
                // Skip permission prompts
                yolo_mode?: bool
                // AI model to use (only meaningful for claude type)
                model?: string
            } @go(,optional=nillable)

            // Additional CLI arguments
            custom_args?: [...string]
        }

        // Bindings from agent roles to executor names
        bindings?: {
            orchestrator?: string
            implementer?: string
            architect?: string
            reviewer?: string
            planner?: string
            researcher?: string
            decomposer?: string
        } @go(,optional=nillable)
    } @go(,optional=nillable)
}
```

### Regenerate Go Types

After creating the CUE schema:
1. Run `go generate ./cli/schemas` from the repository root
2. Verify that `cli/schemas/cue_types_gen.go` is updated with `UserConfig` type
3. Verify the generated types match the expected structure

### Expected Generated Types

The generated Go types should look similar to:

```go
type UserConfig struct {
    Agents *struct {
        Executors map[string]struct {
            Type     string `json:"type"`
            Settings *struct {
                Yolo_mode *bool   `json:"yolo_mode,omitempty"`
                Model     *string `json:"model,omitempty"`
            } `json:"settings,omitempty"`
            Custom_args []string `json:"custom_args,omitempty"`
        } `json:"executors,omitempty"`
        Bindings *struct {
            Orchestrator *string `json:"orchestrator,omitempty"`
            Implementer  *string `json:"implementer,omitempty"`
            Architect    *string `json:"architect,omitempty"`
            Reviewer     *string `json:"reviewer,omitempty"`
            Planner      *string `json:"planner,omitempty"`
            Researcher   *string `json:"researcher,omitempty"`
            Decomposer   *string `json:"decomposer,omitempty"`
        } `json:"bindings,omitempty"`
    } `json:"agents,omitempty"`
}
```

## Acceptance Criteria

- [ ] `cli/schemas/user_config.cue` created with schema definition
- [ ] Schema follows existing conventions (package name, comments, annotations)
- [ ] CUE schema validates correctly (`cue vet` passes)
- [ ] Go types generated successfully via `go generate ./cli/schemas`
- [ ] Generated types appear in `cli/schemas/cue_types_gen.go`
- [ ] Generated types support all required fields as optional
- [ ] Type constraints enforced: executor type must be "claude", "cursor", or "windsurf"

## Technical Details

### CUE Annotations

- Use `@go(,optional=nillable)` for optional struct fields that should be pointers in Go
- Use `?:` suffix for optional fields
- Use `[string]:` for map types
- Use `[...string]` for string arrays

### Code Generation

The generation is triggered by:
```bash
cd cli && go generate ./schemas
```

This runs the directive in `cli/schemas/doc.go`:
```go
//go:generate sh -c "cue exp gengotypes ./... && rm -f projects/cue_types_gen.go"
```

### Validation

Run CUE validation:
```bash
cd cli/schemas && cue vet ./...
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/cli/schemas/config.cue` - Existing simple CUE schema pattern
- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/cli/schemas/refs_cache.cue` - Complex nested schema pattern
- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/cli/schemas/cue_types_gen.go` - Generated Go types example
- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/cli/schemas/doc.go` - Code generation directive
- `/Users/josh/code/sow/.sow/worktrees/feat/user-configuration-system-98/.sow/project/context/issue-98.md` - Full requirements

## Examples

### Example YAML this schema should validate:

```yaml
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
    implementer: "cursor"
    architect: "claude-code"
    reviewer: "claude-code"
    planner: "claude-code"
    researcher: "claude-code"
    decomposer: "claude-code"
```

### Minimal valid config:

```yaml
# Empty is valid - all defaults apply
```

```yaml
agents:
  bindings:
    implementer: "claude-code"
```

## Dependencies

None - this is the foundation task.

## Constraints

- Do NOT modify other CUE schema files
- Do NOT modify the `go:generate` directive in doc.go
- The schema package name MUST be `schemas` to match existing files
- All fields should be optional to support zero-config experience
- Executor type enum MUST match exactly: "claude", "cursor", "windsurf"
