---
description: Create an architecture design document
allowed-tools: Read, Write, Grep, Glob, Bash(git add:*), Bash(git commit:*)
argument-hint: "<feature-or-system-name>"
model: inherit
---

Create an architecture design document for: $ARGUMENTS

## Process

1. **Create design doc file**
   - Path: `.sow/knowledge/architecture/slugified-name.md`
   - Slugify: lowercase, hyphens for spaces, alphanumeric only
   - Example: `authentication-system.md`, `payment-processing.md`

2. **Fill template**
   - Use template below
   - Populate from conversation context and requirements
   - Keep focused on architectural decisions, not implementation details
   - Include minimal code examples (interfaces/contracts only)

## Design Document Template

```markdown
# Feature/System Name

**Status**: Draft | In Review | Approved | Implemented

**Date**: YYYY-MM-DD

**Author**: Architect Agent

---

## Overview

Brief 2-3 sentence description of what this feature/system does and why it exists.

## Goals and Non-Goals

### Goals
- Primary objective 1
- Primary objective 2
- Primary objective 3

### Non-Goals
- What this explicitly does NOT aim to solve
- Future considerations out of scope
- Alternatives we're not pursuing now

## Architecture

### High-Level Design

Describe the major components and how they interact. Use text or simple diagrams.

```
[Component A] ---> [Component B] ---> [Component C]
                         |
                         v
                   [Component D]
```

### Components

**Component Name**
- Responsibility: What it does
- Dependencies: What it needs
- Interface: How others interact with it

Repeat for each major component.

### Ports and Adapters

If using hexagonal architecture:

**Inbound Ports** (how external systems call us)
- Port name and purpose
- Example: `interface OrderService { ... }`

**Outbound Ports** (what we need from infrastructure)
- Port name and purpose
- Example: `interface OrderRepository { ... }`

**Adapters**
- Technology choices for each port
- Example: REST adapter for OrderService, PostgreSQL adapter for OrderRepository

## API Design

Define contracts between components. Use interface definitions, not implementations.

```typescript
// Example: minimal interface showing contract
interface UserService {
  createUser(email: string, name: string): Promise<UserId>
  getUser(id: UserId): Promise<User>
}
```

Keep examples to 10-20 lines. Focus on signatures, not implementations.

## Data Models

Core entities and their relationships. Show domain model, not database schema.

```
User
- id: UserId
- email: Email
- profile: UserProfile

Order
- id: OrderId
- userId: UserId
- items: OrderItem[]
- status: OrderStatus
```

## Integration Points

### External Systems
- What external services/APIs we depend on
- How we communicate with them (REST, gRPC, events)
- Error handling strategy

### Events
- What domain events this system publishes
- What events it consumes

## Security Considerations

- Authentication/authorization approach
- Data protection requirements
- Threat model highlights
- Security boundaries

## Performance Considerations

- Expected load and scale requirements
- Critical performance paths
- Caching strategy (if needed)
- Rate limiting (if applicable)

Avoid premature optimization. Document only known requirements.

## Testing Strategy

- How business logic will be tested (unit tests in isolation)
- Integration points that need testing
- Key test scenarios

## Open Questions

- Unresolved decisions
- Areas needing more research
- Dependencies on other teams

## References

- Related ADRs
- External documentation
- Proof of concepts
- Prior art
```

## Guidelines

**Keep It High-Level**
- Focus on components, boundaries, and contracts
- Avoid implementation details
- Show interfaces, not classes

**Emphasize Decisions**
- Why this design over alternatives
- What trade-offs we're making
- What principles guided choices (YAGNI, dependency inversion, etc.)

**Code Examples**
- Maximum 10-20 lines per example
- Show interfaces/contracts, not implementations
- Use pseudocode or simple syntax
- Annotate with comments explaining key points

**Goals vs Non-Goals**
- Be explicit about scope boundaries
- Call out what you're intentionally NOT solving
- Prevents scope creep and over-engineering

**Be Honest About Unknowns**
- Document open questions
- Note where you need more information
- Flag dependencies on other decisions

**Length**
- Aim for 400-800 words total
- Longer is acceptable if complexity demands it
- Shorter is better if problem is simple

## Example Topics

Good design docs:
- "Authentication and Authorization System"
- "Payment Processing Flow"
- "Event-Driven Order Management"
- "Multi-Tenant Data Isolation"

Avoid:
- "The Entire Application" (too broad)
- "How to Implement User Login" (too detailed)
- "Database Schema" (implementation, not design)