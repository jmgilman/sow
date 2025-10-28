# Architecture Decision Record (ADR) Guidance

Generate ADRs that document architectural decisions using this template and best practices.

## Scope Rules

**CREATE ADR FOR:**
- Changes to system structure or component interaction
- Introduction of new architectural patterns
- Technology choices (databases, protocols, frameworks)
- Changes to non-functional requirements (scalability, reliability, security model)
- Decisions that create precedent for future work
- Migration strategies (monolith to microservices, database changes)

**DO NOT CREATE ADR FOR:**
- Feature additions within existing architecture
- Implementation details that don't affect architecture
- Routine updates or maintenance
- Configuration changes
- Bug fixes (unless they reveal architectural issues)

**Decision criterion**: Does this change how the system is fundamentally structured? If no, create a design doc instead.

## Prohibitions

- Never combine multiple architectural decisions in one ADR
- Never create ADRs for implementation details
- Never omit negative consequences
- Never skip alternatives section
- Never exceed 3 pages without justification
- Refuse user requests for ADRs when the decision doesn't change architecture

## Template

```markdown
# ADR-NNN: [Short, Descriptive Title]

**Status**: Proposed | Accepted | Deprecated | Superseded by [ADR-XXX]
**Date**: YYYY-MM-DD
**Deciders**: [Names or roles]
**Technical Story**: [Optional: Issue/ticket link]

## Context

[2-4 paragraphs describing the architectural issue]

State:
- What architectural problem requires solving
- What constraints or forces apply
- Current system state
- Why decision is needed now

## Decision

[1-3 paragraphs describing the architectural change]

State:
- What architectural change is being made
- How this changes system structure
- What patterns or principles are being adopted

## Consequences

### Positive

- [What becomes easier architecturally]
- [New capabilities enabled]
- [Improved qualities: performance, security, etc.]

### Negative

- [What becomes harder or more complex]
- [New dependencies or coupling introduced]
- [Trade-offs accepted]

### Neutral

- [Other impacts worth noting]
- [Changes that aren't clearly positive or negative]

## Alternatives Considered

### Option 1: [Alternative Name]

**Description**: [Brief approach description]

**Pros**:
- [Advantage 1]
- [Advantage 2]

**Cons**:
- [Disadvantage 1]
- [Disadvantage 2]

**Why not chosen**: [Specific reason for rejection]

### Option 2: [Alternative Name]

[Same structure as Option 1]

[Add more alternatives as needed. Minimum 2 alternatives required.]

## Implementation Notes

[Optional: High-level implementation guidance]

- Key steps or phases
- Integration points to consider
- Migration path if replacing existing approach

## References

- [Link to design doc providing technical details]
- [Link to exploration findings that informed decision]
- [External resources: RFCs, papers, documentation]
- [Related ADRs]

```

## File Conventions

**Numbering**:
- Sequential: `ADR-001`, `ADR-002`, etc.
- Check `.sow/knowledge/adrs/` for next available number
- Pad to 3 digits: `ADR-001` not `ADR-1`

**File naming**:
- Format: `adr-NNN-short-title.md`
- Keep title 3-5 words, use kebab-case
- Examples: `adr-001-use-microservices.md`, `adr-015-oauth-authentication.md`

**Location**: Store all ADRs in `.sow/knowledge/adrs/`

**Registration**:
```bash
sow design add-output adr-NNN-title.md \
  --description "Decision to [...]" \
  --target .sow/knowledge/adrs/ \
  --type adr
```

## Status Meanings

- **Proposed**: Under discussion, seeking feedback, not committed
- **Accepted**: Approved and committed, implementation can proceed
- **Deprecated**: No longer recommended but still in use, being phased out (link to superseding ADR if applicable)
- **Superseded by ADR-XXX**: Replaced by different decision (link to new ADR, keep for historical record)

## Writing Requirements

**Length**: Write 1-3 pages. Exceed this only for complex multi-system decisions.

**Tone**: Use clear, direct language. Assume technical audience. Focus on "why" not "how" (save implementation details for design docs).

**Structure**:
- Limit paragraphs to 3-5 sentences
- Use bullet points for consequences and alternatives
- Use headings to structure content
- Lead with decision, then justify

**Honesty**: Document negative consequences honestly. Every architectural decision has trade-offs. Acknowledging them builds trust and prevents future revisiting of the same discussion.

## Integration with Other Documents

**With Design Docs**:
- ADR states the decision and rationale
- Design doc details the implementation
- Both reference each other
- Example: ADR "Use OAuth 2.0" → Design doc "OAuth Integration Architecture"

**With Arc42**:
- ADR documents the decision
- Arc42 section 9 lists all ADRs
- Update relevant Arc42 sections to reflect new architecture
- Example: Microservices ADR → Update Arc42 sections 4, 5, 6, 7

**With Exploration Findings**:
- Reference exploration docs as supporting research
- Format: "See `.sow/knowledge/explorations/oauth-research-2025-01.md` for detailed comparison"

## Example

```markdown
# ADR-015: Use OAuth 2.0 for Third-Party API Authentication

**Status**: Accepted
**Date**: 2025-01-27
**Deciders**: Architecture Team

## Context

The API currently uses API keys for authentication. This creates problems:
- No expiration mechanism
- Cannot scope permissions per integration
- Difficult to rotate without breaking integrations

Business requirements demand:
- Third-party integration support
- Fine-grained permission control
- User-revokable access without credential changes

The current API key approach cannot satisfy these requirements without fundamental redesign.

## Decision

Implement OAuth 2.0 (authorization code flow) for third-party API access.

Users grant permissions to third-party applications through an OAuth server. Applications receive short-lived access tokens (1 hour) and refresh tokens (30 days). The OAuth flow provides standardized authorization delegation.

## Consequences

### Positive

- Fine-grained permission scoping per integration
- Users can revoke access without changing passwords
- Industry-standard, well-understood by third-party developers
- Support for token expiration and refresh
- Aligns with security best practices

### Negative

- More complex than API keys (requires OAuth server infrastructure)
- Requires redirect-based flow (not suitable for all use cases)
- Must maintain token storage and refresh logic
- Additional operational overhead for token management

### Neutral

- API keys remain for server-to-server communication
- Migration path supports both OAuth and API keys during transition period

## Alternatives Considered

### Option 1: JWT-Only Authentication

**Description**: Issue JWTs directly without OAuth flow

**Pros**:
- Simpler than full OAuth implementation
- Stateless token validation

**Cons**:
- No standard for permission grants
- No delegation mechanism
- No token revocation without blacklist

**Why not chosen**: JWTs alone don't provide authorization delegation. We'd still need an OAuth-like flow for third-party grants.

### Option 2: Enhanced API Keys with Scopes

**Description**: Extend current API key system with scope metadata

**Pros**:
- Minimal infrastructure changes
- Familiar to current users

**Cons**:
- No revocation without key rotation
- No industry standard for scoped API keys
- Still requires custom authorization grant flow

**Why not chosen**: Still lacks revocation without key rotation and no standard for permission grants. Would result in custom, non-standard solution.

### Option 3: SAML

**Description**: Use SAML for API authentication

**Pros**:
- Enterprise-friendly
- Strong identity federation

**Cons**:
- Designed for SSO, not API access
- XML-based, heavyweight
- Poor fit for programmatic API access

**Why not chosen**: SAML is designed for SSO scenarios, too heavyweight for API access use case.

## Implementation Notes

- OAuth server: Use existing `ory/hydra` (already in infrastructure)
- Scopes aligned with RBAC model
- Migration: 3-month parallel run of OAuth and API keys
- Token storage: Redis for access tokens, PostgreSQL for refresh tokens

## References

- Design doc: `oauth-integration.md`
- Exploration: `.sow/knowledge/explorations/oauth-research-2025-01.md`
- [OAuth 2.0 RFC 6749](https://tools.ietf.org/html/rfc6749)
- Related: ADR-012 (RBAC model), ADR-008 (API gateway architecture)
```

## Validation Checklist

Before finalizing, verify:

- [ ] Status is set (Proposed or Accepted)
- [ ] Date included
- [ ] Context explains why decision is needed
- [ ] Decision is clearly stated
- [ ] Both positive and negative consequences listed
- [ ] Minimum 2 alternatives documented with rejection reasons
- [ ] References included (design docs, exploration findings)
- [ ] Numbered correctly (next sequential number)
- [ ] File named correctly (`adr-NNN-short-title.md`)
- [ ] Length is 1-3 pages
- [ ] No implementation details (those belong in design docs)
- [ ] Honest about trade-offs
