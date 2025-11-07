# ADR: Explicit Event Selection in `sow advance` Command

**Status**: Proposed
**Date**: 2025-11-06
**Deciders**: Core Team
**Context**: Enhancing the `sow advance` command to support orchestrator-driven state transitions

---

## Context and Problem Statement

The `sow advance` command currently operates with auto-determination only: it calls `DetermineEvent()` to select which event to fire based on project state. This works well for linear state progressions and state-determined branching (where the decision can be derived from examining project state), but it fails to support **intent-based branching** where the orchestrator must make a decision that cannot be discovered from the project state alone.

**Example scenario**: In an exploration project, when research is in progress, the orchestrator needs to decide whether to:
- `finalize` - Research is complete, move to finalization
- `add_more_research` - More investigation needed, return to planning

This decision depends on the orchestrator's judgment of research completeness based on conversation context, user goals, and domain understanding—information not encoded in the project state.

**Current limitation**: There is no way for the orchestrator to explicitly specify which event to fire. The only option is `sow advance`, which requires a `DetermineEvent()` implementation that can examine state and choose automatically.

## Decision Drivers

1. **Orchestrator Autonomy**: AI orchestrators need to make intent-based decisions that cannot be encoded in state
2. **Flexibility**: Different project types have different branching patterns (some state-determined, some intent-based)
3. **Backward Compatibility**: Existing workflows using auto-determination must continue working unchanged
4. **Discoverability**: Orchestrators need to discover available options without trial-and-error
5. **Safety**: Invalid transitions should be caught before execution
6. **Clarity**: Logs and history should show explicit orchestrator decisions

## Considered Options

### Option 1: Explicit Event as Positional Argument

Add optional `[event]` positional argument to `sow advance`:

```bash
# Auto-determination (existing behavior)
sow advance

# Explicit event selection (new)
sow advance finalize
sow advance add_more_research
```

**Pros**:
- Simple, intuitive syntax
- Backward compatible (argument is optional)
- Clear intent in command history
- Natural fit for orchestrator scripting

**Cons**:
- Could be confused with other arguments if not documented clearly
- Event names might conflict with future flags (unlikely, but possible)

### Option 2: Explicit Event via Flag

Use `--event` flag for explicit selection:

```bash
# Auto-determination
sow advance

# Explicit event
sow advance --event finalize
```

**Pros**:
- Clear distinction between auto and explicit modes
- No risk of argument confusion
- Self-documenting in command syntax

**Cons**:
- More verbose
- Less natural for orchestrators (extra typing)
- Flag syntax doesn't convey that event selection is the primary purpose

### Option 3: Separate Command

Create new command `sow fire` for explicit events:

```bash
# Auto-determination
sow advance

# Explicit event
sow fire finalize
```

**Pros**:
- Clear semantic distinction
- No risk of confusion
- Could have different behavior/validation

**Cons**:
- Two commands doing similar things
- More cognitive load for users
- Orchestrators need to know when to use which command
- Code duplication

### Option 4: Interactive Selection

Prompt orchestrator to choose when multiple options available:

```bash
$ sow advance

Multiple transitions available:
1. finalize - Complete research and move to finalization
2. add_more_research - Return to planning for more topics

Choose transition [1-2]: _
```

**Pros**:
- No new syntax needed
- Discoverable (shows options automatically)
- Clear decision point

**Cons**:
- Doesn't work for non-interactive contexts (CI, scripts)
- Orchestrators are automation—they shouldn't require human interaction
- Breaks automated workflows
- Latency from interactive prompts

## Decision Outcome

**Chosen option: Option 1 (Explicit Event as Positional Argument)**

Reasoning:
1. **Simplest syntax** - `sow advance finalize` is clear and concise
2. **Backward compatible** - Optional argument preserves existing behavior
3. **Orchestrator-friendly** - Easy to script and automate
4. **Intent-clear** - Command history shows explicit decisions
5. **Natural extension** - Feels like a natural evolution of the existing command

## Implementation Details

### Command Signature

```bash
sow advance [event]
```

### Behavior

**No argument** (existing behavior preserved):
- Calls `DetermineEvent()` to auto-select event
- For linear states: fires the single configured event
- For state-determined branching: uses discriminator to choose
- For intent-based branching: shows available options and errors (prompts for explicit event)

**With event argument** (new behavior):
- Validates event is configured for current state
- Checks guard conditions
- Fires event if valid
- Errors with helpful message if invalid or blocked

### Discovery and Validation

To support orchestrators in discovering and validating explicit events:

**List mode**:
```bash
sow advance --list
```
Shows all available transitions with descriptions and guard requirements.

**Dry-run mode**:
```bash
sow advance --dry-run finalize
```
Validates transition without executing (pre-flight check).

### Error Handling

**Invalid event**:
```
Error: event finalize not configured from state Researching
```

**Guard failure**:
```
Error: transition blocked: all tasks complete

Use --list to see available transitions
```

**Auto-determination failure** (intent-based branching):
```
Cannot auto-determine transition

Available transitions (choose one):
  sow advance finalize
  sow advance add_more_research

Error: specify event explicitly
```

## Consequences

### Positive

- **Enables intent-based branching**: Orchestrators can make decisions based on conversation context
- **Maintains backward compatibility**: Existing projects and workflows unchanged
- **Improved discoverability**: `--list` shows all options with context
- **Better debugging**: `--dry-run` validates before execution
- **Clear audit trail**: Explicit events visible in logs and git history
- **Flexible project design**: Project types can choose state-determined, intent-based, or mixed patterns

### Negative

- **Slight increase in complexity**: Two modes (auto vs explicit) to understand
- **Potential for confusion**: Users might not know when to use explicit events
  - *Mitigation*: Clear documentation, helpful error messages, orchestrator prompts
- **Event namespace concerns**: Events must be unique per state (already a constraint)
  - *Mitigation*: This is enforced by the state machine library

### Neutral

- **No performance impact**: Event validation and execution same as before
- **Minimal code changes**: Adding positional argument is straightforward
- **Documentation updates needed**: CLI docs and orchestrator prompts need updates

## Alternatives Considered and Rejected

### Why not interactive selection (Option 4)?

Interactive prompts break automation. Orchestrators are automated agents, not humans. Requiring interaction defeats the purpose of automation and introduces latency.

### Why not separate command (Option 3)?

Two commands (`sow advance` and `sow fire`) create confusion about when to use which. The decision is about *how* to advance (auto vs explicit), not *what* to do (both advance the state machine). A single command with optional argument is clearer.

### Why not event flag (Option 2)?

The `--event` flag is more verbose and makes the event feel like a parameter rather than the primary intent. Compare:
- `sow advance finalize` (clear: "advance by firing finalize")
- `sow advance --event finalize` (wordy: "advance with event parameter finalize")

For a CLI used heavily by orchestrators, conciseness and clarity matter.

## Validation

### Success Criteria

- [ ] Orchestrators can explicitly fire events for intent-based branching
- [ ] Existing auto-determination continues working without changes
- [ ] `--list` shows all available transitions with descriptions
- [ ] `--dry-run` validates transitions without executing
- [ ] Error messages guide orchestrators to solutions
- [ ] Exploration project can use intent-based branching successfully

### Testing Plan

- Unit tests for argument parsing and routing
- Integration tests for auto vs explicit modes
- E2E tests with real orchestrator workflows
- Backward compatibility tests with existing projects

## Related Decisions

- **ADR: State Machine Branching Support** - Covers the `AddBranch()` API for state-determined branching
- **SDK Design: AddBranch API** - Technical design for declarative branching
- **CLI Design: Enhanced advance Command** - Detailed implementation specification

## References

- Exploration document: `.sow/knowledge/explorations/advance-branching/advance-cli.md`
- Stateless library documentation: https://github.com/qmuntal/stateless
- Project SDK architecture: `cli/internal/sdks/project/state/doc.go`

---

## Revision History

- **2025-11-06**: Initial proposal
