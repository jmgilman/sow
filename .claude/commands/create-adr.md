---
description: Create an Architecture Decision Record
allowed-tools: Read, Write, Grep, Glob, Bash(git add:*), Bash(git commit:*)
argument-hint: "<decision-title>"
model: inherit
---

Create an Architecture Decision Record (ADR) for: $ARGUMENTS

## Process

1. **Determine next ADR number**
   - Check `.sow/knowledge/adrs/` for existing ADRs
   - Use next sequential number (001, 002, 003, etc.)

2. **Create ADR file**
   - Path: `.sow/knowledge/adrs/NNN-slugified-title.md`
   - Slugify: lowercase, hyphens for spaces, alphanumeric only
   - Example: `003-use-postgresql-database.md`

3. **Fill template**
   - Use template below
   - Populate from conversation context
   - Status defaults to "Proposed"
   - Date uses ISO format (YYYY-MM-DD)

## ADR Template

```markdown
# NNN. Title in Title Case

**Status**: Proposed | Accepted | Rejected | Deprecated | Superseded

**Date**: YYYY-MM-DD

---

## Context

What is the issue we're addressing? What factors are relevant? What forces are at play?

Write 2-4 paragraphs describing:
- The problem or decision point
- Technical constraints
- Business requirements
- Alternative approaches considered

## Decision

What did we decide and why?

State the decision clearly in 1-2 sentences, then explain reasoning:
- Why this approach over alternatives
- What principles guided the choice
- What trade-offs we're accepting

## Consequences

### Positive

- Benefit 1
- Benefit 2
- Benefit 3

### Negative

- Cost 1
- Limitation 2
- Technical debt 3

### Neutral

- Changes that are neither good nor bad, but noteworthy

## References

- Link to related ADRs
- External documentation
- Code examples or prototypes
- Prior art or research
```

## Guidelines

**Context Section**
- Describe the situation, not the solution
- Include enough detail for future readers unfamiliar with the decision
- Cite technical constraints that influenced the decision

**Decision Section**
- Be explicit about what you're committing to
- Explain the "why" more than the "what"
- Reference principles from the architect role (YAGNI, dependency inversion, etc.)

**Consequences Section**
- Be honest about trade-offs
- Include both positive and negative impacts
- Consider operational, maintenance, and development costs

**Keep It Concise**
- ADRs are permanent records, not essays
- Aim for 200-400 words total
- Focus on decisions that are hard to reverse

## Example Titles

Good:
- "Use PostgreSQL for Primary Database"
- "Adopt Hexagonal Architecture for Core Domain"
- "Implement Event Sourcing for Audit Trail"

Avoid:
- "Database Stuff" (too vague)
- "Why We Chose PostgreSQL Over MySQL, MongoDB, and 5 Other Options" (too verbose)