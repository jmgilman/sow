# Research Methodology Guidance

You've invoked research guidance. Apply these best practices to your exploration work.

## Research Process

### 1. Start Broad, Then Focus

**Initial Phase** (Broad Survey):
- Identify the key technologies, approaches, or solutions in the space
- Get a high-level understanding of each option
- Note the major differences and use cases
- Document sources (URLs, docs, articles)

**Narrowing Phase** (Deep Dive):
- Pick the most promising 2-3 options based on requirements
- Research implementation details
- Look for real-world examples and case studies
- Identify potential issues or gotchas

### 2. Document as You Go

**Don't wait until the end to write**. Document continuously:

- **Research notes**: Raw findings, links, quotes
- **Summaries**: Synthesize what you've learned
- **Questions**: Track unknowns to investigate later
- **Insights**: Key realizations or eureka moments

### 3. Create Comparison Matrices

When evaluating multiple options, use structured comparisons:

```markdown
# OAuth vs JWT Comparison

| Aspect | OAuth 2.0 | JWT |
|--------|-----------|-----|
| Type | Authorization framework | Token format |
| Use Case | Third-party access | Stateless auth |
| Complexity | Higher (multiple flows) | Lower (just tokens) |
| Security | Delegated, fine-grained | Self-contained, expirable |
| Implementation | Requires auth server | Simpler, library-based |
```

Include:
- **Evaluation criteria** (performance, security, complexity, etc.)
- **Objective facts** (what each option actually does)
- **Trade-offs** (advantages vs disadvantages)
- **Context** (when to use each)

### 4. Trace Decisions

**Record your reasoning**:
- Why did you investigate option A vs option B?
- What criteria mattered most?
- What assumptions did you make?
- What constraints influenced the decision?

This helps others (and future you) understand the context.

## Documentation Patterns

### Research Notes

**Purpose**: Capture findings as you discover them

**Structure**:
```markdown
# [Technology/Approach] Research

## Overview
Brief description of what this is.

## Key Findings
- Finding 1: description + source
- Finding 2: description + source

## Examples
- Example 1: [description](link)
- Example 2: [description](link)

## Questions/Unknowns
- [ ] How does X handle Y scenario?
- [ ] What's the performance impact of Z?

## Sources
- [Source 1](url)
- [Source 2](url)
```

### Comparison Documents

**Purpose**: Side-by-side evaluation of options

**Structure**:
```markdown
# [Option A] vs [Option B] Comparison

## Quick Summary
One-paragraph overview of the comparison result.

## Detailed Comparison

### [Criterion 1]
- **Option A**: ...
- **Option B**: ...
- **Winner**: Option A (or depends on context)

### [Criterion 2]
...

## Recommendation
Based on [context/requirements], recommend [option] because [reasons].
```

### Findings Documents

**Purpose**: Synthesize research into actionable insights

**Structure**:
```markdown
# [Topic] Findings

## Context
What problem were we exploring?

## Key Insights
1. **Insight 1**: Description
   - Evidence: ...
   - Implications: ...

2. **Insight 2**: Description
   - Evidence: ...
   - Implications: ...

## Recommendations
- **Short-term**: What to do immediately
- **Long-term**: Future considerations
- **Risks**: What to watch out for
```

## File Naming Conventions

Use descriptive, hierarchical names:

✅ **Good**:
- `oauth-research.md` - Clear and specific
- `oauth-vs-jwt-comparison.md` - Shows it's a comparison
- `auth-findings-2025-01.md` - Includes date for temporal context
- `payment-gateway-options.md` - Describes content

❌ **Avoid**:
- `notes.md` - Too generic
- `temp.md` - Unclear purpose
- `research-1.md` - Numbered without context

## Tagging Strategy

Tags enable discoverability. Use multiple tags per file:

**Technology tags**: `oauth`, `jwt`, `postgresql`, `redis`
**Type tags**: `research`, `comparison`, `findings`, `spike`
**Domain tags**: `authentication`, `database`, `caching`, `payments`
**Status tags**: `in-progress`, `completed`, `needs-review`

Example:
```bash
sow exploration add-file oauth-research.md \
  --description "OAuth 2.0 flows and implementation" \
  --tags "oauth,authentication,research,in-progress"
```

## When to Create New Files vs Update Existing

### Create New File When:
- Researching a distinctly different topic
- Starting a formal comparison
- Synthesizing findings into a deliverable
- Creating a reference document

### Update Existing File When:
- Adding to ongoing research on same topic
- Refining a comparison with new information
- Correcting or clarifying existing content
- Adding examples to existing analysis

## Integration with Formal Artifacts

### Moving from Exploration to Formalization

When research is complete, formalize findings into permanent artifacts:

**Architecture Decision Records (ADRs)**:
- Document the decision made
- Explain context and alternatives considered
- Reference exploration files for detailed research

**Design Documents**:
- Describe the chosen approach in detail
- Reference exploration findings as supporting evidence
- Provide implementation guidance

**Example workflow**:
1. Research in exploration (`.sow/exploration/`)
2. Create ADR in team location (e.g., `docs/adrs/`)
3. ADR references exploration files for context
4. Create summary in `.sow/knowledge/explorations/`

## Keeping Context Small

**Problem**: Too many files blow up context window

**Solutions**:
1. **Use the index** - Agent reads index first, loads only relevant files
2. **One file per topic** - Don't create dozens of small files
3. **Consolidate when appropriate** - Merge related notes into comprehensive files
4. **Archive or delete** - Remove dead ends and outdated research

## Common Pitfalls

❌ **Not documenting sources**: Can't verify claims later
✅ **Include URLs and references**

❌ **Waiting to write**: Forget details, lose insights
✅ **Document continuously as you learn**

❌ **Generic descriptions**: "Notes about auth"
✅ **Specific descriptions**: "OAuth 2.0 authorization code flow implementation"

❌ **Missing tags**: Files become undiscoverable
✅ **Tag everything with multiple relevant keywords**

❌ **No structure**: Wall of text that's hard to navigate
✅ **Use headings, lists, tables for scannability**

---

**Now apply these practices to your current exploration work.**
