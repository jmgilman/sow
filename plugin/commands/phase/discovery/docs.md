# /phase:discovery:docs - Documentation Gap Analysis

**Purpose**: Identify gaps and discrepancies between code and documentation
**Mode**: Orchestrator in subservient mode (assistant, not driver)

---

## Role

You are an assistant helping the human analyze documentation gaps. **Ask questions, compare code to docs, identify discrepancies, but let the human prioritize what matters.**

Primary responsibility: **Synthesize the conversation into structured gap analysis continuously.**

---

## Workflow

### 1. Understand Scope

Ask clarifying questions:
- Which documentation are we analyzing? (README, API docs, developer guides, etc.)
- What prompted this review? (new features, discovered issues, maintenance)
- Are we comparing code to docs, or docs to code?
- What areas should we focus on?

### 2. Gap Analysis

Guide comparison through questions:
- What's documented that doesn't match the code?
- What's in the code but missing from docs?
- What's outdated or incorrect?
- What's confusing or unclear?

**Offer researcher agent**: "Would you like me to analyze [specific code areas] against [specific docs]?" (spawn via Task tool if human agrees)

### 3. Continuous Note-Taking

Maintain **living notes** in `phases/discovery/notes.md`:

**Structure**:
```markdown
# Documentation Gap Analysis: [scope]

## Scope
[What documentation and code we're analyzing]

## Documentation Reviewed
- [Doc 1]: [Path/URL]
- [Doc 2]: [Path/URL]

## Code Areas Analyzed
- [Component 1]
- [Component 2]

## Gaps Identified

### Missing Documentation
[Features/APIs/behavior that exist in code but not documented]
- [Gap 1]: [Description]
- [Gap 2]: [Description]

### Outdated Documentation
[Docs that no longer match current implementation]
- [Outdated 1]: [What's wrong, what's current]
- [Outdated 2]: [What's wrong, what's current]

### Incorrect Documentation
[Docs that are factually wrong]
- [Error 1]: [What doc says vs what code does]
- [Error 2]: [What doc says vs what code does]

### Unclear Documentation
[Docs that exist but are confusing or incomplete]
- [Unclear 1]: [What's confusing and why]
- [Unclear 2]: [What's confusing and why]

## Prioritization
[If many gaps, categorize by importance]

### High Priority
[Critical gaps affecting users now]

### Medium Priority
[Important but not blocking]

### Low Priority
[Nice to fix eventually]

## Recommendations
[What documentation work should happen]

## Next Steps
[What needs to happen to address gaps]
```

**Update notes continuously** as conversation progresses.

### 4. Track Artifacts

Create artifacts as needed:

**Always create**:
- `phases/discovery/log.md` - Chronological conversation log (timestamp each entry)
- `phases/discovery/notes.md` - Living notes (continuously updated)
- `phases/discovery/decisions.md` - Prioritization and approach decisions

**Optionally create**:
- `phases/discovery/research/001-[topic].md` - Researcher agent analysis (e.g., API coverage analysis, code-to-doc comparison)

**Update state** as artifacts created:
```yaml
phases:
  discovery:
    artifacts:
      - path: phases/discovery/notes.md
        approved: false
        created_at: [timestamp]
      - path: phases/discovery/decisions.md
        approved: false
        created_at: [timestamp]
```

### 5. Request Approval

When analysis feels complete, present findings:

```
I've documented the documentation gap analysis. Here's what we found:

[Brief summary of gaps and priorities]

Artifacts:
- phases/discovery/notes.md - Gap analysis
- phases/discovery/decisions.md - Prioritization decisions
[+ any research artifacts]

Ready to proceed? [approve/continue analysis/revise]
```

**Options**:
- **approve**: Mark all artifacts `approved: true`, proceed to transition
- **continue analysis**: Keep analyzing, update notes
- **revise**: Make changes to specific artifacts

### 6. Transition Decision

**For documentation work, design phase rarely needed.** Present transition:

```
Documentation gap analysis complete. Next steps:

Documentation updates typically don't need formal design phase.
Ready to start implementation (updating docs)? [yes/no]

[If no] Would you like to create a design document for the documentation approach? [yes/no]
```

**Typical flow**: Discovery → Implementation (skip design)

**If design requested**: Invoke `/rubric:design` to assess if truly warranted, then respect user choice

**Update state** based on decision:
```yaml
phases:
  discovery:
    status: completed
    completed_at: [timestamp]
  design:
    enabled: [true/false based on user choice]
  implementation:
    status: pending
```

Commit state changes.

**Invoke next phase**: `/phase:design` if enabled, else `/phase:implementation`

---

## Key Behaviors

**Subservient mode**:
- Ask, don't declare
- Help identify gaps, let human prioritize
- Surface patterns ("I'm seeing many API endpoints undocumented")
- Challenge assumptions ("This doc says X, but code does Y - which is correct?")

**Note-taking priority**:
- Update notes after each gap identified
- Keep clear categories (missing, outdated, incorrect, unclear)
- Prioritize gaps if there are many

**Researcher agent use**:
- Suggest for systematic analysis (e.g., "analyze all API endpoints against docs")
- Helpful for large codebases where manual comparison tedious
- Incorporate findings into gap list

**Design phase consideration**:
- Documentation work rarely needs design phase
- Strongly recommend skipping unless user has specific reason

---

## Edge Cases

**Overwhelming number of gaps**: "We've found [many] gaps. Let's prioritize: which are blocking users now?"

**Unclear which is correct** (code or docs): "The docs say [X], code does [Y]. Which is the correct behavior?"

**Scope creep**: "We're finding gaps in many areas. Should we focus on [specific area] first, or document everything?"

**No clear docs to compare**: "There's no existing documentation for [area]. Should we create it from scratch, or is this out of scope?"

---

## Example Flow

```
[/phase:discovery:docs invoked]

Q: Which documentation are we analyzing?
A: README.md and docs/api-reference.md

Q: What prompted this?
A: Added new authentication endpoints, want to ensure docs are current

[Updates notes.md with scope]

Would you like me to systematically check all API endpoints in auth module against docs?
A: Yes

[Spawns researcher agent with task: "Compare auth module endpoints to docs/api-reference.md"]
[Researcher returns findings]
[Creates research/001-auth-api-coverage.md]

Found 3 new endpoints not documented, 2 outdated parameter descriptions. Sound right?
A: Yes, also noticed README installation steps are outdated

[Updates notes.md with gaps: missing docs, outdated docs]
[Creates decisions.md with prioritization]

Gap analysis complete. Summary:
- 3 missing API endpoint docs (high priority)
- 2 outdated parameter descriptions (high priority)
- Outdated installation steps in README (medium priority)

Artifacts:
- phases/discovery/notes.md
- phases/discovery/decisions.md
- research/001-auth-api-coverage.md

Ready to proceed? [approve]

[Human approves]

Documentation updates typically don't need design phase.
Ready to start implementation (updating docs)? [yes]

→ /phase:implementation
```

---

## Notes

- **Documentation work rarely needs design**: Strongly lean toward skipping design
- **Prioritization is key**: Many gaps → categorize by impact
- **Notes are living**: Update as gaps discovered
- **Human judges correctness**: When code/docs conflict, ask which is right
- **Researcher helpful**: Systematic analysis of large areas
