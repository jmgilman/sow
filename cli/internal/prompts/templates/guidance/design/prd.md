# Product Requirements Document (PRD) Guidance

Generate Product Requirements Documents for new services. PRDs describe product requirements and user needs, not technical implementation.

## Scope Rules

**CREATE PRD FOR:**
- New services substantial enough for product-level documentation
- Services requiring Arc42 architecture documentation
- Services needing stakeholder alignment on goals and scope
- Services where problem definition precedes technical design

**DO NOT CREATE PRD FOR:**
- Simple internal tools or utilities
- Features added to existing services (use design docs)
- Technical changes without product functionality impact
- Work where team agrees scope is straightforward

**KEY DISTINCTION**: PRDs answer "what problem are we solving?" Technical implementation belongs in design docs and Arc42.

## Prohibitions

- Never include technical implementation details (database choices, frameworks, code)
- Never create PRD for features within existing services (use design docs)
- Never create PRD for simple internal tools
- Never omit success metrics from goals
- Never use vague requirements without acceptance criteria
- Never exceed 10 pages without splitting docs
- Refuse requests for PRD when design doc is appropriate

## Document Boundaries

| Document | Focus | Question Answered | Content Type |
|----------|-------|-------------------|--------------|
| **PRD** | Product requirements | What problem are we solving? | Product-centric |
| **Arc42** | Architecture | How is the system structured? | Technical structure |
| **Design Doc** | Technical implementation | How do we build this feature? | Implementation details |
| **ADR** | Decisions | What did we decide and why? | Architectural choices |

## PRD Structure

### Required Sections

1. **Executive Summary**: 2-3 sentences stating what the service is, what problem it solves, and expected impact

2. **Problem Statement**:
   - Current state (3-4 paragraphs describing current situation and problems)
   - Pain points (specific problems users face with business impact)
   - User stories (as [user type], I want [capability], so that [benefit])
   - Opportunity (why solve now, opportunity cost of not solving)

3. **Goals and Success Metrics**:
   - Business goals (measurable objectives with metrics and targets)
   - User goals (what users need to accomplish)
   - Success metrics table (metric, current baseline, target, timeline)

4. **Scope**:
   - In-scope: Phased (MVP/Phase 1, Future/Phase 2+)
   - Out-of-scope: Explicit list of what is NOT included with reasons

5. **User Personas** (2-3 personas):
   - Role (job title or user type)
   - Goals (what persona wants to accomplish)
   - Pain points (specific problems faced)
   - Technical proficiency (low/medium/high)

6. **Requirements**:
   - Functional requirements: MoSCoW prioritization (Must Have/Should Have/Could Have/Won't Have) with acceptance criteria for each
   - Non-functional requirements: Performance, reliability, security, scalability, usability (quantified)

7. **User Flows**:
   - Key flows with goal, steps, alternative paths
   - Describe user actions and system responses

8. **Design Considerations** (optional):
   - Key UX principles
   - Interface notes (high-level, not detailed designs)

9. **Constraints and Assumptions**:
   - Constraints: Technical, business, regulatory, budget
   - Assumptions: List assumptions with risk analysis if wrong

10. **Dependencies**:
    - Internal dependencies (teams, services)
    - External dependencies (third-party APIs, services)

11. **Risks and Mitigation**:
    - Risk table (risk, probability, impact, mitigation strategy)

12. **Release Plan**:
    - Phased approach with scope, success criteria, target dates per phase

13. **Open Questions**:
    - Checklist of unresolved issues requiring resolution

14. **Appendix**:
    - References (research, interviews, competitive analysis)
    - Change log (date, change, author)

## Integration with Other Documents

**PRD → Arc42**:
- PRD Problem Statement → Arc42 Section 1 (Introduction and Goals)
- PRD Goals → Arc42 Section 1 (Quality Goals)
- PRD Requirements → Arc42 Section 2 (Constraints)
- Arc42 references PRD for detailed requirements context

**PRD → Design Docs**:
- Design docs reference PRD for:
  - Why feature exists
  - What problem it solves
  - Success criteria
- Link PRD in design doc Background section

**Example Arc42 reference**:
```markdown
# Arc42 Section 1: Introduction and Goals

## Requirements Overview

As defined in the [Service Name PRD](../service-name-prd.md), this service
solves [problem] by providing [solution].

**Key requirements** (from PRD):
- Requirement 1
- Requirement 2
```

## Example: PRD Excerpt

```markdown
# Authentication Service Product Requirements

**Author**: Product Team
**Date**: 2025-01-27
**Status**: Approved
**Stakeholders**: Product, Engineering, Security, Customer Success

## Executive Summary

Build a centralized authentication service to replace fragmented per-application auth implementations. This will reduce onboarding time from 20 minutes to under 5 minutes and enable single sign-on across all company products, improving user experience and reducing support costs.

## Problem Statement

### Current State

Each product team has implemented separate authentication systems. Users maintain different credentials for each product. Onboarding requires multiple registrations. Password resets must be performed separately for each product. Support receives 200+ authentication-related tickets monthly.

**Pain Points**:
- Users forget which credentials go with which product (60% of support tickets)
- New users abandon onboarding due to complexity (25% drop-off rate)
- Engineering teams duplicate authentication work across products
- Security updates must be deployed to 8 separate codebases
- No single sign-on (SSO) capability for enterprise customers

### User Stories

**As an end user**, I want to authenticate once, so that I can access all products without re-entering credentials.

**As an enterprise admin**, I want to manage user access centrally, so that onboarding/offboarding is consistent across products.

**As a product engineer**, I want a standard authentication service, so that I don't build auth logic for each new product.

### Opportunity

Enterprise customers require SSO for contracts (losing $500k annual revenue). Current fragmentation costs 3 engineering months per new product. Centralizing auth enables faster product launches and improved security posture.

## Goals and Success Metrics

### Business Goals

1. **Reduce authentication-related support tickets**
   - Metric: Support tickets tagged "authentication"
   - Current: 200 tickets/month
   - Target: < 50 tickets/month
   - Timeline: 6 months post-launch

2. **Improve user onboarding conversion**
   - Metric: Percentage completing registration
   - Current: 75% (25% drop-off)
   - Target: 90% (10% drop-off)
   - Timeline: 3 months post-launch

3. **Enable enterprise SSO sales**
   - Metric: Enterprise contracts requiring SSO
   - Current: 0 (cannot support)
   - Target: 5+ contracts closed
   - Timeline: 12 months

### User Goals

1. Single set of credentials for all products
2. Fast, simple registration (< 2 minutes)
3. Reliable password reset (< 1 minute)

### Success Metrics

| Metric | Current | Target | Timeline |
|--------|---------|--------|----------|
| Avg onboarding time | 20 min | < 5 min | 3 months |
| Support tickets (auth) | 200/month | < 50/month | 6 months |
| User satisfaction (auth) | 6.5/10 | 8.5/10 | 6 months |
| API uptime | 99.5% | 99.9% | Ongoing |

## Scope

### In Scope

**Phase 1 (MVP)**:
- Email/password authentication
- OAuth 2.0 token issuance
- Password reset flow
- User profile management
- Integration with 3 existing products

**Phase 2 (Enhancements)**:
- Social login (Google, GitHub)
- Multi-factor authentication (TOTP)
- Integration with remaining 5 products

**Phase 3 (Enterprise)**:
- SAML SSO for enterprise customers
- Role-based access control (RBAC)
- Admin dashboard for user management

### Out of Scope

NOT INCLUDED:
- Biometric authentication (future consideration)
- Mobile SDK (web-only initially, API available for mobile teams)
- Custom branding per product (standard branding only)
- Legacy product migration (products migrate on own timeline)

## User Personas

### Persona 1: Sarah - End User

**Role**: Marketing manager at mid-size company

**Goals**:
- Access multiple company products quickly
- Avoid remembering multiple passwords
- Reset password easily when forgotten

**Pain Points**:
- Constantly forgets which password goes with which product
- Frustrated by multiple registration processes
- Abandons products requiring separate registration

**Technical Proficiency**: Medium (comfortable with web apps, not technical)

### Persona 2: Mike - Enterprise Admin

**Role**: IT administrator at enterprise customer

**Goals**:
- Centrally manage user access across all products
- Onboard/offboard employees efficiently
- Enforce security policies (password complexity, MFA)

**Pain Points**:
- No SSO support blocks product adoption
- Manual user management across products is time-consuming
- Cannot enforce company security policies

**Technical Proficiency**: High (manages enterprise identity systems)

## Requirements

### Functional Requirements

**Must Have (MVP)**:

- FR1: Users must authenticate with email/password
  - Acceptance Criteria:
    - Email validation (RFC 5322)
    - Password minimum 12 characters
    - Password strength indicator
    - Account creation < 2 seconds
    - Confirmation email sent

- FR2: System must issue OAuth 2.0 access tokens
  - Acceptance Criteria:
    - Access tokens expire after 1 hour
    - Refresh tokens expire after 30 days
    - JWT format with RS256 signing
    - Standard OAuth 2.0 flows supported

- FR3: Users must reset passwords via email
  - Acceptance Criteria:
    - Reset link expires after 1 hour
    - Reset completes in < 1 minute
    - Email delivered within 30 seconds

**Should Have (Phase 2)**:

- FR4: Users can authenticate via social providers (Google, GitHub)
- FR5: Users can enable multi-factor authentication (TOTP)

**Could Have (Phase 3)**:

- FR6: Enterprise admins can configure SAML SSO

### Non-Functional Requirements

**Performance**:
- API response time < 100ms (p95)
- Support 1000 concurrent authentication requests
- Handle 10,000 users online simultaneously

**Reliability**:
- 99.9% uptime SLA
- Zero downtime deployments
- Automated failover within 30 seconds

**Security**:
- Passwords hashed with bcrypt (cost factor 12)
- TLS 1.3 for all connections
- OWASP Top 10 compliance
- SOC 2 Type II compliant
- Rate limiting: 5 failed login attempts = 15-minute lockout

**Scalability**:
- Horizontal scaling to support user growth
- Database sharding strategy for 1M+ users

**Usability**:
- Registration completable in < 2 minutes
- Mobile-responsive design
- WCAG 2.1 AA accessibility compliance

## Constraints and Assumptions

### Constraints

- **Technical**: Must integrate with existing user database (PostgreSQL)
- **Business**: Launch date Q2 2025 for enterprise sales cycle
- **Regulatory**: GDPR compliance required (EU customers)
- **Budget**: $100k development budget, $2k/month operational budget

### Assumptions

- Assumption 1: Average user load is 100 authentication requests/minute
- Assumption 2: 80% of authentication happens during business hours (9am-5pm)
- Assumption 3: Products migrate to new auth service within 6 months

**Risks if assumptions wrong**:
- If authentication load is 10x higher, requires additional infrastructure ($5k/month)
- If product migrations delayed, ROI timeline extends by 6 months

## Risks and Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Product teams slow to integrate | High | Medium | Dedicated integration support, clear documentation, SDK provided |
| User data migration issues | Medium | High | Comprehensive testing, phased rollout, rollback plan |
| Performance bottlenecks at scale | Medium | High | Load testing before launch, horizontal scaling ready |
| Security vulnerability discovered | Low | Critical | Security audit before launch, bug bounty program |
| Regulatory compliance gaps | Low | High | Legal review, GDPR expert consultation |

## Release Plan

### Phase 1: MVP (Target: Q2 2025)

**Scope**:
- Email/password authentication
- OAuth 2.0 token issuance
- Password reset
- Integration with 3 products

**Success Criteria**:
- 99.9% uptime achieved
- < 100ms p95 response time
- 90% of beta users complete registration

**Beta Users**: Internal employees (500 users), 2 pilot customers (1000 users)

### Phase 2: Enhancements (Target: Q3 2025)

**Scope**:
- Social login (Google, GitHub)
- Multi-factor authentication
- Integration with remaining 5 products

### Phase 3: Enterprise (Target: Q4 2025)

**Scope**:
- SAML SSO
- RBAC
- Admin dashboard

## Open Questions

- [ ] Should password reset links expire after 1 hour or 24 hours? (Security vs UX trade-off)
- [ ] Do we support account recovery without email access? (If yes, requires support process)
- [ ] Should we support passwordless authentication (magic links)? (Adds scope, defers to Phase 2)
```

## Requirements

**Focus on why**: State the problem, who experiences it, why it matters, and what success looks like. Never propose solutions without explaining the problem first.

**Measurability**: Include baseline (current state), target (goal), metric (how to measure), and timeline (when) for every goal.

**Specificity**: Use specific, testable criteria. Replace "improve user experience" with "reduce onboarding time from 20 min to < 5 min". Replace "good performance" with "< 100ms p95 response time".

**Scope clarity**: Explicitly define what is in-scope (phased) and out-of-scope. Use MoSCoW prioritization (Must/Should/Could/Won't Have) for requirements.

**Acceptance criteria**: Every functional requirement must have testable acceptance criteria. State how to verify the requirement is met.

**Phased delivery**: Structure scope in phases: MVP (must-have), enhancements (should-have), future (could-have). Enable iterative delivery and early feedback.

**Product focus**: Describe what to build, not how to build it. Save technical implementation (databases, frameworks) for architecture and design docs.

## File Conventions

**Storage location**: `.sow/knowledge/`

**File naming**:
- Format: `service-name-prd.md`
- Use kebab-case (hyphens)
- Keep concise (2-4 words + "-prd")
- Examples: `authentication-service-prd.md`, `payment-processing-prd.md`, `notification-service-prd.md`

**Registration**:
```bash
sow design add-output service-name-prd.md \
  --description "Product requirements for [service]" \
  --target .sow/knowledge/ \
  --type prd
```

## Validation Checklist

- [ ] Problem statement clearly explains current pain with business impact
- [ ] Goals are specific and measurable (not vague)
- [ ] Success metrics include baseline, target, and timeline
- [ ] Scope explicitly defined (in-scope phased, out-of-scope listed)
- [ ] User personas included (2-3 personas with goals, pain points, proficiency)
- [ ] Functional requirements have acceptance criteria (testable)
- [ ] Non-functional requirements are quantified (not "fast" but "< 100ms p95")
- [ ] User flows describe key scenarios with steps
- [ ] Risks identified with mitigation strategies
- [ ] Release plan has phases with target dates
- [ ] Open questions documented as checklist
- [ ] Length appropriate (5-10 pages for most services, never exceed 10)
- [ ] No technical implementation details (no databases, frameworks, code)
- [ ] Focus on product (what/why), not technical (how)
