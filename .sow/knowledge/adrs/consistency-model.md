# ADR: Four-Layer Consistency Model for Multi-Agent System

**Status:** Draft
**Date:** January 2025
**Deciders:** Josh Gilman, Claude (Sonnet 4.5)
**Context:** Multi-agent system architecture supporting multiple AI CLIs

## Context

Sow is extending its agent system to support multiple AI CLI tools (Claude Code, Cursor, Windsurf, etc.). A critical question arises: **How do we maintain consistency when different AI models execute the same agent roles?**

Different models have inherent variability:
- Different training data and capabilities
- Different reasoning patterns
- Different output quality and style
- Different instruction-following fidelity

**The central challenge:** Ensure that tasks completed by different executors (using different models) meet consistent quality standards and follow the same structural patterns, without requiring identical output.

## Critical Architectural Insight

**The orchestrator is another AI agent, not Go code.**

This fundamentally changes what "enforcement" means:

- The orchestrator (Claude Code, Cursor, etc.) is an AI agent reading files and making judgments
- It cannot programmatically enforce validation rules - it interprets guidance
- The only hard enforcement point is the `sow` CLI validating CUE schemas
- Everything else (validation, feedback, iteration) is **guidance-based**, not code-based

**Implication:** Consistency comes from prompt engineering, schema validation, and iteration loops - not from programmatic constraints that prevent invalid states.

## Decision

We adopt a **four-layer consistency model** that accepts AI agent autonomy while providing progressive constraints:

### Layer 1: Structural Consistency (CLI-Enforced)

**What:** CUE schema validation when agents use `sow` CLI commands

**Enforcement:** Programmatic - the only hard constraint layer

**Examples:**
```bash
sow task set --id 010 status invalid_status
# Error: invalid status value, must be one of: pending, in_progress, needs_review, completed, abandoned, paused
```

**What it validates:**
- Task state structure (`state.yaml` format)
- Status enum values
- Required vs optional fields
- Type constraints (string, int, bool)

**What it cannot validate:**
- Whether work was done correctly
- Whether requirements were met
- Code quality or completeness
- Adherence to conventions

**Rationale:** CUE schemas provide structural guarantees. When workers use `sow` commands, the CLI validates inputs before writing to disk. This prevents structurally invalid state.

### Layer 2: Behavioral Guidance (Prompt Engineering)

**What:** Explicit, imperative instructions in agent prompt templates

**Enforcement:** Guidance-based - agents interpret and follow

**Examples:**

```markdown
<!-- From implementer.md template -->

## Critical Rules

YOU MUST:
- Read task state.yaml to understand iteration and feedback
- Read description.md for requirements
- Log all actions to your task log.md
- Run tests before setting status to needs_review
- Update status via `sow task set --id <id> status <status>`

YOU MUST NOT:
- Modify project state.yaml (orchestrator only)
- Modify other tasks' files
- Create new tasks (orchestrator only)
- Skip writing tests

## Common Mistakes to Avoid

❌ Setting status to needs_review without running tests
✓ Run full test suite, verify all passing, then set needs_review

❌ Modifying state.yaml directly instead of using sow commands
✓ Always use `sow task set` commands for state updates
```

**Characteristics:**
- Explicit do/don't lists
- Common mistakes documented
- Example commands provided
- Validation checklists
- Failure mode descriptions

**Rationale:** Clear, imperative guidance shapes agent behavior. While not programmatically enforced, explicit instructions reduce variance across different models. Models generally follow clear directives.

### Layer 3: Validation (Orchestrator Guidance)

**What:** Orchestrator agent validates worker output via guidance

**Enforcement:** Guidance-based - orchestrator interprets validation instructions

**Orchestrator prompt includes:**

```markdown
## Worker Output Validation

After worker completes, you MUST validate:

1. **Structural validation:**
   - Task state.yaml updated correctly
   - Status set appropriately (needs_review, paused, failed)
   - Task log contains action summary

2. **Work quality validation:**
   - Read all modified code files
   - Run test suite: `npm test` or equivalent
   - Verify tests pass
   - Check code matches requirements in description.md

3. **Completeness validation:**
   - All acceptance criteria met
   - Edge cases handled
   - Error handling present
   - Documentation updated if needed

If validation fails:
- Use `sow agent resume <task-id> "<feedback>"` to provide corrections
- Be specific: cite line numbers, describe issues, suggest fixes
- Iterate until work meets standards

Only mark completed after validation passes.
```

**Characteristics:**
- Orchestrator reads and evaluates worker output
- Another AI agent making judgments, not code enforcement
- Validation criteria explicitly documented
- Feedback mechanism via session resumption

**Rationale:** The orchestrator acts as a reviewing agent, guided to check worker output. While judgment-based (not programmatic), explicit validation criteria increase consistency. Different orchestrator models may have different standards, but guidance aligns their evaluation.

### Layer 4: Iteration Loop (Self-Correction)

**What:** Feedback mechanism when validation fails

**Enforcement:** Process-based - review cycle with corrections

**Flow:**

```
Worker completes → needs_review
    ↓
Orchestrator validates
    ↓
Issues found? → Resume with feedback
    ↓
Worker corrects → needs_review
    ↓
Orchestrator validates again
    ↓
Approved? → completed
```

**Example iteration:**

```bash
# Iteration 1
Worker: Implements authentication → needs_review
Orchestrator: Finds security issue (plaintext password logging)

# Iteration 2
Orchestrator: Resumes with specific feedback
Worker: Fixes security issue → needs_review
Orchestrator: Finds missing test coverage

# Iteration 3
Orchestrator: Resumes with test requirements
Worker: Adds tests → needs_review
Orchestrator: All checks pass → completed
```

**Characteristics:**
- Session continuity across iterations
- Specific, actionable feedback
- Worker has full context from previous attempts
- Iteration counter tracks quality metrics

**Rationale:** Even with varying model capabilities, iteration drives convergence toward quality standards. Feedback loop allows correction without requiring perfect first-attempt execution. Over multiple iterations, output quality improves regardless of model.

## Four-Layer Model Summary

```
┌────────────────────────────────────────────────────────────┐
│ Layer 4: Iteration Loop                                    │
│ Feedback-driven convergence toward quality standards       │
│ Enforcement: Process-based                                 │
└────────────────────────────────────────────────────────────┘
                            ▲
                            │ If validation fails
                            │
┌────────────────────────────────────────────────────────────┐
│ Layer 3: Validation (Orchestrator Guidance)                │
│ Orchestrator evaluates worker output                       │
│ Enforcement: Guidance-based (AI judgment)                  │
└────────────────────────────────────────────────────────────┘
                            ▲
                            │ Validates output
                            │
┌────────────────────────────────────────────────────────────┐
│ Layer 2: Behavioral Guidance (Prompt Engineering)          │
│ Explicit instructions shape agent behavior                 │
│ Enforcement: Guidance-based (AI interpretation)            │
└────────────────────────────────────────────────────────────┘
                            ▲
                            │ Interprets guidance
                            │
┌────────────────────────────────────────────────────────────┐
│ Layer 1: Structural Consistency (CLI-Enforced)             │
│ CUE schema validation via sow CLI                          │
│ Enforcement: Programmatic (hard constraint)                │
└────────────────────────────────────────────────────────────┘
```

**Progressive enforcement:**
- **Layer 1** (strongest): Programmatic prevention of structural invalidity
- **Layer 2** (guidance): Explicit behavioral instructions
- **Layer 3** (validation): Guided quality evaluation
- **Layer 4** (process): Iterative refinement toward standards

**Each layer compensates for the one below:**
- Layer 1 prevents structural chaos
- Layer 2 reduces behavioral variance
- Layer 3 catches quality issues
- Layer 4 corrects what previous layers missed

## Consequences

### Positive

**1. Realistic about AI agent capabilities**
- Accepts that orchestrator is an agent, not enforcement code
- Doesn't attempt impossible programmatic control
- Works with AI strengths (following guidance, iteration) rather than fighting them

**2. Supports multiple executors with different models**
- No assumption of identical behavior
- Progressive constraints accommodate model variance
- Iteration loop compensates for capability differences

**3. Quality through process, not prevention**
- Focus on validation and correction rather than upfront prevention
- Iteration drives convergence toward standards
- Session resumption enables meaningful feedback

**4. Scales with more agents**
- Layer 1 (schemas) unchanged as agents added
- Layer 2 (prompts) per-agent customization
- Layer 3/4 (validation/iteration) universal pattern

**5. Measurable quality metrics**
- Iteration count per task (quality indicator)
- Validation pass rate by executor/model
- Common failure patterns identifiable

### Negative

**1. Cannot guarantee first-attempt correctness**
- Iteration may be required even with good prompts
- Different models may need different iteration counts
- Users must accept review cycles as normal

**2. Quality depends on orchestrator model**
- Better orchestrator models → better validation
- Weaker orchestrator models → may miss issues
- No programmatic quality floor beyond Layer 1

**3. Prompt maintenance burden**
- Layer 2 (behavioral guidance) requires ongoing refinement
- As models improve, prompts may need updates
- Common mistakes must be documented and added to prompts

**4. Difficult to debug consistency issues**
- When agents behave unexpectedly, root cause ambiguous:
  - Prompt unclear?
  - Model capability limit?
  - Validation criteria missing?
- Requires iterative prompt refinement

**5. Testing complexity**
- Integration tests must accept variance
- Cannot assert exact output, only structural conformance
- Quality benchmarks needed for cross-executor comparison

### Risks and Mitigations

**Risk 1: Orchestrator fails to validate properly**

*Mitigation:*
- Explicit validation checklists in orchestrator prompts
- Benchmark tasks with known-correct solutions
- User feedback when quality issues observed
- Iteration on orchestrator prompts based on failure patterns

**Risk 2: Workers ignore behavioral guidance**

*Mitigation:*
- Emphasize critical rules (YOU MUST / MUST NOT)
- Document common mistakes explicitly
- Iteration loop catches deviations
- Session resumption provides corrective feedback

**Risk 3: Quality variance too high across executors**

*Mitigation:*
- Establish quality benchmarks (test suite pass rate, iteration count)
- Disable executors that consistently underperform
- User choice allows switching to better-performing executors
- Document executor compatibility matrix

**Risk 4: Infinite iteration loops**

*Mitigation:*
- Iteration counter tracked in state.yaml
- Orchestrator guidance includes escalation rules:
  - After N iterations, flag for human review
  - After M iterations, mark task as failed
- Users can intervene and provide direction

## Alternatives Considered

### Alternative 1: Programmatic Enforcement of All Rules

**Approach:** Write Go code that validates worker output programmatically

**Example:**
```go
func ValidateImplementation(taskID string) error {
    // Run tests
    if err := runTests(); err != nil {
        return err
    }

    // Check code quality
    if !hasTests(taskID) {
        return errors.New("missing tests")
    }

    // Validate requirements
    if !meetsRequirements(taskID) {
        return errors.New("requirements not met")
    }

    return nil
}
```

**Rejected because:**
- Orchestrator is an AI agent, not Go code
- Validation runs in orchestrator's context (AI subprocess)
- Cannot inject Go validation logic into Claude/Cursor subprocess
- Would require building entire validation framework in Go
- Duplicates work - orchestrator can already read code and judge quality

**Fundamental mismatch:** Trying to programmatically enforce what an AI agent should judge.

### Alternative 2: Identical Output Requirement

**Approach:** Require all executors to produce identical output for same task

**Validation:** Diff output across Claude, Cursor, Windsurf - reject if different

**Rejected because:**
- Different models have inherent variability
- Multiple correct solutions exist for most tasks
- Style/approach differences don't indicate quality problems
- Overly restrictive - prevents using better models in better ways
- Unrealistic given current AI technology

**Better approach:** Validate correctness (tests pass, requirements met), not identity.

### Alternative 3: Single Executor Only

**Approach:** Only support Claude Code, don't attempt multi-executor

**Rejected because:**
- User preference matters (subscriptions, tool familiarity)
- Future models may be better for specific roles
- Vendor lock-in undesirable
- Cursor/Windsurf may excel at implementation while Claude excels at orchestration
- Limits value proposition of sow as universal orchestrator

**Trade-off:** Complexity of multi-executor support worth the flexibility and user choice.

### Alternative 4: Human Validation Required

**Approach:** Every task requires human approval before completion

**Rejected because:**
- Defeats automation purpose
- Doesn't scale to large projects
- Orchestrator validation (Layer 3) provides reasonable quality check
- Human can still intervene when needed (iteration count high, quality concerns)

**Middle ground:** Orchestrator validates, human intervenes on exceptions.

## Implementation Notes

### Layer 1 Implementation

**CUE schemas already exist:**
- `cli/schemas/task_state.cue` - Task state structure
- `cli/schemas/user_config.cue` - User configuration (new)

**Validation occurs in CLI commands:**
```go
func (c *TaskSetCommand) Run() error {
    // Validate against CUE schema
    if err := c.validateTaskState(state); err != nil {
        return err
    }

    // Write to disk only if valid
    return c.writeTaskState(state)
}
```

### Layer 2 Implementation

**Agent prompt templates:**
- Stored: `cli/internal/agents/templates/`
- Embedded: `//go:embed templates/*`
- Loaded: Via `LoadPrompt(agent.PromptPath)`

**Template structure:**
```markdown
# Agent Role Description

## Initialization
[sow prompt commands to load guidance]

## Critical Rules
[YOU MUST / MUST NOT lists]

## Common Mistakes
[Documented failure patterns]

## Validation Checklist
[What to verify before completion]
```

### Layer 3 Implementation

**Orchestrator prompt updates:**
- Location: TBD (orchestrator-specific guidance)
- Contains: Worker validation instructions
- References: Layer 2 rules (consistency)

**Validation pattern:**
```markdown
After worker subprocess exits:
1. Read state.yaml → Check status
2. Read modified files → Validate changes
3. Run test suite → Verify passing
4. Compare to requirements → Check completeness
5. If issues → Resume with feedback
6. If approved → Mark completed
```

### Layer 4 Implementation

**Session management:**
- Session ID persisted in `state.yaml`
- Resume command: `sow agent resume <task-id> "<feedback>"`
- Iteration counter: Incremented on each resume
- Feedback: Passed as prompt to `executor.Resume()`

**Orchestrator guidance includes:**
- When to resume vs when to approve
- How to write actionable feedback
- Escalation rules (iteration count thresholds)

## Quality Assurance Strategy

### Benchmark Tasks

Create standard benchmark tasks with known solutions:

```
benchmark/
├── simple-feature/          # Basic CRUD implementation
├── security-sensitive/      # Authentication, input validation
├── performance-critical/    # Optimization, profiling
├── complex-refactor/        # Multi-file architectural change
└── bug-fix/                 # Root cause analysis, targeted fix
```

**Metrics per executor:**
- First-attempt success rate (no iterations needed)
- Average iteration count to completion
- Test pass rate
- Time to completion
- Adherence to requirements (manual review score)

### Executor Compatibility Matrix

Document executor performance across agent roles:

```
| Executor      | Orchestrator | Implementer | Architect | Reviewer |
|---------------|--------------|-------------|-----------|----------|
| Claude Opus   | Excellent    | Excellent   | Excellent | Excellent|
| Claude Sonnet | Excellent    | Good        | Excellent | Good     |
| Cursor        | Good         | Excellent   | Good      | Good     |
| Windsurf      | Unknown      | Unknown     | Unknown   | Unknown  |
```

**Quality tiers:**
- **Excellent**: <1.5 avg iterations, >95% test pass rate
- **Good**: <2.5 avg iterations, >90% test pass rate
- **Acceptable**: <4 avg iterations, >80% test pass rate
- **Poor**: >4 avg iterations or <80% test pass rate

### Continuous Improvement

**Feedback loop for prompt refinement:**

1. Track common failure patterns
2. Document in "Common Mistakes" sections
3. Update behavioral guidance (Layer 2)
4. Re-run benchmarks to verify improvement
5. Iterate on validation criteria (Layer 3)

**Example:**

```
Observation: Workers frequently set needs_review without running tests
Action: Add to Layer 2 guidance:
  "❌ Setting needs_review without running tests
   ✓ Run test suite, verify all pass, then set needs_review"
Result: Test-skipping rate drops from 15% to 3%
```

## Measuring Success

### Quantitative Metrics

1. **Iteration distribution:**
   - Track: Histogram of iteration counts to completion
   - Target: >70% of tasks complete in ≤2 iterations

2. **Validation pass rate:**
   - Track: % of tasks passing orchestrator validation first try
   - Target: >60% first-try pass rate

3. **Cross-executor variance:**
   - Track: Iteration count difference between executors on same benchmark
   - Target: <1.5 iteration difference between executors

4. **Structural violations:**
   - Track: CUE schema validation failures (Layer 1)
   - Target: <1% of sow command invocations fail validation

### Qualitative Metrics

1. **User satisfaction:**
   - Survey: Confidence in multi-executor quality
   - Feedback: Complaints about inconsistent output

2. **Prompt maintenance burden:**
   - Track: Frequency of prompt updates needed
   - Goal: Stable prompts over 3+ month periods

3. **Failure mode diversity:**
   - Track: Unique failure patterns across executors
   - Goal: Converging set of failure modes (not expanding)

## Related Decisions

- **Multi-Agent System Architecture** - Overall system design
- **Session Management Protocol** - Enables Layer 4 iteration loop
- **User Configuration System** - Allows executor selection per role

## References

- Exploration summary: `.sow/knowledge/explorations/custom-agents.md`
- Architecture design: `.sow/project/multi-agent-architecture.md`
- Cross-agent consistency research: `cross-agent-consistency.md` (exploration artifact)

## Revision History

- **2025-01-22**: Initial draft - Four-layer consistency model
