# Project Initialization Truth Table

**Status**: Complete - Ready for Implementation
**Date**: 2025-10-14
**Last Updated**: 2025-10-14

---

## Overview

This document defines the decision tree that guides the orchestrator through project initialization. The orchestrator asks questions, interprets answers, and recommends a phase plan for human approval.

**Core Principle**: The orchestrator gracefully leads the human through questions to infer the right phase plan, then ASKS for confirmation - never decides unilaterally.

---

## Decision Flow Structure

### Entry Point: User Wants to Start Work

```
Orchestrator: "What would you like to work on today?
              1. A new project
              2. An existing project
              3. A one-off task"
```

**Branches**:
- **Option 2 (Existing project)** → `/continue` flow (out of scope for this document)
- **Option 3 (One-off task)** → Orchestrator direct mode (no project structure)
- **Option 1 (New project)** → Continue to project initialization flow below

---

## Project Initialization Flow

### Question 1: What Are You Trying to Accomplish?

```
Orchestrator: "Great! What would you like to accomplish in this project?"
```

**Purpose**: Understand the nature of the work to infer likely phase needs.

**Common Responses** (examples):
- "Fix a bug where users can't login"
- "Add user authentication with JWT"
- "Refactor the authentication module - it's gotten messy"
- "Create a CLI tool to help with X"
- "Implement the design described in docs/auth-design.md"

**Orchestrator's Job**:
- Listen to the description
- Infer work type (bug, feature, refactor, etc.)
- Use this to guide next questions

---

### Question 2: Existing Context Assessment

```
Orchestrator: "Do you have any existing context, documents, or notes about this?"
```

**Purpose**: Determine if discovery/design phases are needed or can be skipped.

**Response Categories**:

#### A. "No" / "Not really" / "Just the idea"
- **Implication**: Likely needs discovery and/or design
- **Next**: Ask more questions to determine phase needs

#### B. "Yes, I have [design doc/detailed notes/requirements]"
- **Implication**: May skip discovery, may skip design
- **Next**: Ask specific follow-up about what they have

#### C. "There's existing documentation in the repo at X"
- **Implication**: Similar to B, likely can skip phases
- **Next**: Confirm they want to implement what's documented

---

### Question 3: Discovery Phase Decision

**Asked when**: Response to Q2 indicates limited context (Category A)

```
Orchestrator: "It sounds like we could benefit from some discovery work to
              better understand [the problem/requirements/current state].

              Would you like to start with a discovery phase? This could include:
              - Investigating the current codebase
              - Researching existing solutions
              - Understanding the problem space better

              [yes/no/tell me more]"
```

**Branches**:
- **Yes** → Discovery enabled, continue to design question after discovery complete
- **No** → Discovery disabled, continue to design question now
- **Tell me more** → Orchestrator explains discovery phase benefits, then re-asks

---

### Question 4: Design Phase Decision

**Asked when**:
- After discovery phase completes (if enabled)
- After Q2 if user has some context but no formal design
- Skipped if user references existing design docs

```
Orchestrator: "Based on [discovery findings/your description], do you want to
              create formal design documents before implementation?

              This could include:
              - Architecture Decision Records (ADRs)
              - Design documents
              - API specifications

              [yes/no/I'm not sure]"
```

**Branches**:
- **Yes** → Design enabled
- **No** → Design disabled
- **I'm not sure** → Orchestrator provides guidance based on project complexity/scope

**Orchestrator Guidance for "I'm not sure"**:

The orchestrator should think critically about whether design is warranted:

**Recommend Design When**:
- Large scope (10+ implementation tasks expected)
- New system component being added
- Architectural changes involved
- Cross-cutting concerns identified
- User explicitly uncertain about approach

**Recommend Skipping Design When**:
- Small scope (1-5 implementation tasks expected)
- Bug fix (even complex ones)
- Small feature addition
- Straightforward refactor
- Discovery notes provide sufficient detail

---

## Truth Table Matrix

| Q1: Work Type | Q2: Existing Context | Q3: Discovery? | Q4: Design? | Result: Phases Enabled |
|---------------|---------------------|----------------|-------------|------------------------|
| Bug fix | None | Yes | No | Discovery: true, Design: false |
| Bug fix | None | No | No | Discovery: false, Design: false |
| Bug fix | Investigation notes | N/A (skip Q3) | No | Discovery: false, Design: false |
| Small feature | None | Yes | No | Discovery: true, Design: false |
| Small feature | None | No | No | Discovery: false, Design: false |
| Small feature | Detailed notes | N/A (skip Q3) | No | Discovery: false, Design: false |
| Large feature | None | Yes | Yes | Discovery: true, Design: true |
| Large feature | None | Yes | No | Discovery: true, Design: false |
| Large feature | Design doc exists | N/A (skip Q3) | N/A (skip Q4) | Discovery: false, Design: false |
| Refactor | None | Yes | Maybe | Discovery: true, Design: user decides |
| Refactor | Analysis notes | N/A (skip Q3) | Maybe | Discovery: false, Design: user decides |
| New component | None | Yes | Yes | Discovery: true, Design: true |
| New component | Requirements doc | Maybe | Yes | Discovery: user decides, Design: true |
| Implement existing design | Design doc provided | N/A (skip Q3) | N/A (skip Q4) | Discovery: false, Design: false |

**Note**: Implementation, Review, and Finalize are always enabled (not shown in table).

---

## Example Walkthroughs

### Example 1: Bug Fix (No Context)

```
Orchestrator: What would you like to work on today?
Human: A new project

Orchestrator: Great! What would you like to accomplish in this project?
Human: Fix a bug where users can't login after password reset

Orchestrator: Do you have any existing context, documents, or notes about this?
Human: No, I just noticed the issue

Orchestrator: It sounds like we should investigate this bug to understand the
              root cause. Would you like to start with a discovery phase to
              dig into what's happening?
Human: Yes, that makes sense

Orchestrator: Perfect. After we complete discovery, we'll likely go straight
              to implementation since bug fixes typically don't need formal
              design documents. Does that sound right?
Human: Agreed

Orchestrator: Great! Let me create the project structure with:
              - Discovery: enabled
              - Design: disabled
              - Implementation: enabled (required)
              - Review: enabled (required)
              - Finalize: enabled (required)

              Ready to proceed? [yes/modify]
Human: Yes

[Orchestrator creates project, enters discovery phase]
```

**Result**: Discovery: true, Design: false

---

### Example 2: Large Feature (No Context)

```
Orchestrator: What would you like to work on today?
Human: A new project

Orchestrator: Great! What would you like to accomplish in this project?
Human: Add a comprehensive notification system with email, SMS, and push notifications

Orchestrator: Do you have any existing context, documents, or notes about this?
Human: No, just the high-level idea

Orchestrator: This sounds like a substantial feature. I think we should start
              with discovery to research notification approaches and understand
              the requirements better. Sound good?
Human: Yes

[After discovery phase completes...]

Orchestrator: I've completed the discovery research. Based on what we learned,
              this is a complex system with multiple components. I recommend
              creating formal design documents before implementation to define
              the architecture, API contracts, and integration points.

              Would you like to proceed with a design phase?
Human: Yes, let's create the design docs

[Orchestrator enters design phase]
```

**Result**: Discovery: true, Design: true

---

### Example 3: Implement Existing Design

```
Orchestrator: What would you like to work on today?
Human: A new project

Orchestrator: Great! What would you like to accomplish in this project?
Human: Implement the authentication system described in docs/auth-design.md

Orchestrator: Do you have any existing context, documents, or notes about this?
Human: Yes, the design doc at docs/auth-design.md has all the details

Orchestrator: Perfect! Let me review that document...
              [reads docs/auth-design.md]

              This design doc looks comprehensive. Since you already have the
              design, we can skip discovery and design phases and go straight
              to implementation. Does that work?
Human: Yes

Orchestrator: Great! Let me create the project structure with:
              - Discovery: disabled (design exists)
              - Design: disabled (design exists)
              - Implementation: enabled (required)
              - Review: enabled (required)
              - Finalize: enabled (required)

              Ready to proceed? [yes/modify]
Human: Yes

[Orchestrator creates project, enters implementation phase]
```

**Result**: Discovery: false, Design: false

---

### Example 4: Refactor (Uncertain About Design Need)

```
Orchestrator: What would you like to work on today?
Human: A new project

Orchestrator: Great! What would you like to accomplish in this project?
Human: Refactor the authentication module - it's become a 2000-line mess

Orchestrator: Do you have any existing context, documents, or notes about this?
Human: Not really, I just know it needs work

Orchestrator: Let's start with a discovery phase to analyze the current code
              and understand what needs to change. This will help us identify
              the refactoring strategy.
Human: Sounds good

[After discovery phase completes...]

Orchestrator: I've completed the analysis. The authentication module has several
              issues: tight coupling, mixed responsibilities, and duplicated logic.

              I recommend a refactoring approach that extracts 3 new modules:
              AuthValidator, TokenManager, and SessionHandler.

              This is a fairly complex refactor. Would you like to create a design
              document to formalize the new structure before implementing?
Human: No, the discovery notes are detailed enough. Let's just implement it.

Orchestrator: Understood. Proceeding to implementation phase.

[Orchestrator enters implementation phase]
```

**Result**: Discovery: true, Design: false

---

### Example 5: Small Feature (Skip Everything)

```
Orchestrator: What would you like to work on today?
Human: A new project

Orchestrator: Great! What would you like to accomplish in this project?
Human: Add a --version flag to the CLI that prints the version number

Orchestrator: Do you have any existing context, documents, or notes about this?
Human: No, it's a straightforward addition

Orchestrator: This is a simple feature. Do you want to do any discovery or just
              go straight to implementation?
Human: Let's just implement it

Orchestrator: Sounds good. No need for formal design docs either, right?
Human: Correct

Orchestrator: Great! Let me create the project structure with:
              - Discovery: disabled
              - Design: disabled
              - Implementation: enabled (required)
              - Review: enabled (required)
              - Finalize: enabled (required)

              This should be a quick implementation. Ready to proceed? [yes/modify]
Human: Yes

[Orchestrator creates project, enters implementation phase]
```

**Result**: Discovery: false, Design: false

---

## Orchestrator Decision-Making Guidelines

### When to Suggest Discovery Phase

**Suggest Discovery When**:
- User provides minimal context ("I want to add an API")
- Bug needs investigation (root cause unknown)
- User says "I need to understand X better"
- Feature requires domain knowledge user may not have
- Technical decision needs research (library comparison, approach evaluation)
- User is uncertain about feasibility
- Current state of codebase is unclear

**Don't Suggest Discovery When**:
- User has detailed notes/requirements already
- User references existing design documents
- Work is trivial/straightforward (simple additions)
- User explicitly wants to skip ("Let's just implement it")
- Implementation path is completely clear

### When to Suggest Design Phase

**Suggest Design When**:
- Large scope (10+ tasks expected in implementation)
- New system component being introduced
- Architectural changes involved
- Multiple integration points
- Cross-cutting concerns (affects many parts of system)
- User is uncertain about approach
- Complex API contracts or data models
- Discovery revealed significant complexity

**Don't Suggest Design When**:
- Bug fixes (just implement the fix)
- Small features (1-5 tasks)
- Minor refactors
- Discovery notes provide sufficient implementation detail
- User references existing design ("implement what's in docs/X")
- Work is straightforward and well-understood

### Confidence Levels in Recommendations

**High Confidence (Strong Recommendation)**:
```
Orchestrator: "I recommend [doing/skipping] [phase] because [clear reason]."
```

**Medium Confidence (Suggestion)**:
```
Orchestrator: "I think we should [do/skip] [phase], but I'm open to your thoughts."
```

**Low Confidence (Ask)**:
```
Orchestrator: "I'm not sure if we need [phase]. What do you think?"
```

**Guiding Principle**: When uncertain, ask the human rather than making assumptions.

---

## Branch Management Integration

After phase plan is confirmed, orchestrator checks branch status:

```
Orchestrator: "Great! Before we create the project, let me check your git branch..."

[Checks current branch]

Case 1: On main/master
  "You're currently on the main branch. I'll create a new feature branch for
   this project. Suggested name: feat/[inferred-from-description]

   Is this branch name okay? [yes/suggest different name]"

Case 2: On existing feature branch with no project
  "You're on branch '[branch-name]'. I'll create the project on this branch.
   Does that work? [yes/create new branch instead]"

Case 3: On existing feature branch WITH existing project
  "You're on branch '[branch-name]' which already has a project: '[project-name]'.

   Options:
   1. Continue existing project (/continue)
   2. Create new branch for this project

   What would you like to do?"
```

---

## State Initialization

Once phase plan and branch are confirmed:

```yaml
project:
  name: [inferred from description]
  branch: [current or newly created branch]
  description: [user's description from Q1]
  created_at: [timestamp]

phases:
  discovery:
    enabled: [true/false based on decision tree]
    status: [pending if enabled, skipped if disabled]

  design:
    enabled: [true/false based on decision tree]
    status: [pending if enabled, skipped if disabled]

  implementation:
    enabled: true  # Always true
    status: pending

  review:
    enabled: true  # Always true
    status: pending

  finalize:
    enabled: true  # Always true
    status: pending
```

**First Active Phase**: The first enabled phase is set to `in_progress` and its corresponding `/phase:X` command is invoked.

---

## Decision-Making Rubrics

To prevent over-engineering and ensure consistent phase recommendations, the orchestrator uses scoring rubrics when determining whether phases are warranted.

**Philosophy**: Similar to medical assessment rubrics - clear questions with defined scoring that produces objective recommendations.

---

### Discovery Worthiness Rubric

**When Applied**: When orchestrator needs to determine if discovery phase is warranted (either user asks or orchestrator is uncertain).

**Scoring Questions** (each scored 0-2):

1. **Context Availability**: Does the user have existing context/documentation?
   - 0 = Has comprehensive documentation/notes
   - 1 = Has some notes but gaps exist
   - 2 = No existing context

2. **Problem Clarity**: Is the problem/requirement well-understood?
   - 0 = Crystal clear, no ambiguity
   - 1 = Generally clear but some uncertainty
   - 2 = Unclear, needs investigation

3. **Codebase Familiarity**: Does work require understanding unfamiliar code?
   - 0 = No codebase investigation needed
   - 1 = Some code review needed
   - 2 = Significant codebase exploration required

4. **Research Needs**: Does this require external research (libraries, approaches, best practices)?
   - 0 = No research needed
   - 1 = Light research beneficial
   - 2 = Substantial research required

**Scoring Rubric**:
- **Score 0-2**: Discovery NOT warranted (skip)
- **Score 3-5**: Discovery OPTIONAL (ask user)
- **Score 6-8**: Discovery RECOMMENDED (suggest strongly)

**Example Calculation**:

User wants to "Fix login bug":
- Context: 2 (no investigation done yet)
- Clarity: 1 (know symptom, not root cause)
- Codebase: 1 (need to review auth code)
- Research: 0 (bug fix, no research)
- **Total: 4** → Discovery OPTIONAL (ask user)

User wants to "Implement design from docs/auth-design.md":
- Context: 0 (has comprehensive design)
- Clarity: 0 (design is clear)
- Codebase: 0 (design specifies what to build)
- Research: 0 (design already done)
- **Total: 0** → Discovery NOT warranted (skip)

---

### Design Worthiness Rubric

**When Applied**: When orchestrator needs to determine if design phase is warranted.

**Scoring Questions** (each scored 0-2):

1. **Scope Size**: How large is the implementation?
   - 0 = 1-3 tasks (small)
   - 1 = 4-9 tasks (medium)
   - 2 = 10+ tasks (large)

2. **Architectural Impact**: Does this change system architecture?
   - 0 = No architectural changes
   - 1 = Minor architectural adjustments
   - 2 = Significant architectural changes (new components, patterns, or cross-cutting concerns)

3. **Integration Complexity**: How many integration points exist?
   - 0 = Self-contained, no/minimal integration
   - 1 = Integrates with 1-2 existing components
   - 2 = Integrates with 3+ components or external systems

4. **Design Decisions**: Are there significant technical decisions to document?
   - 0 = No significant decisions (straightforward implementation)
   - 1 = 1-2 decisions worth documenting
   - 2 = Multiple decisions requiring ADRs or design docs

**Scoring Rubric**:
- **Score 0-2**: Design NOT warranted (skip)
- **Score 3-5**: Design OPTIONAL (ask user)
- **Score 6-8**: Design RECOMMENDED (suggest strongly)

**Special Rule for Bug Fixes**: Regardless of score, bug fixes rarely need design phase. If work type is "bug fix", apply -3 penalty to total score (minimum 0).

**Example Calculation**:

User wants to "Add --version flag to CLI":
- Scope: 0 (1 task)
- Architecture: 0 (no changes)
- Integration: 0 (self-contained)
- Decisions: 0 (trivial)
- **Total: 0** → Design NOT warranted (skip)

User wants to "Add notification system with email, SMS, push":
- Scope: 2 (10+ tasks expected)
- Architecture: 2 (new component)
- Integration: 2 (multiple services)
- Decisions: 2 (provider selection, architecture, API design)
- **Total: 8** → Design RECOMMENDED (suggest strongly)

User wants to "Fix complex authentication bug":
- Scope: 1 (4-5 tasks)
- Architecture: 1 (may require auth flow changes)
- Integration: 1 (touches auth system)
- Decisions: 1 (how to fix properly)
- **Subtotal: 4, Bug Fix Penalty: -3**
- **Total: 1** → Design NOT warranted (skip)

---

### One-Off vs Project Rubric

**When Applied**: When orchestrator suspects user-requested "new project" might be better as one-off task.

**Purpose**: Help orchestrator suggest more appropriate workflow without overriding user choice.

**Scoring Questions** (each scored 0 or 1):

1. **Single Action**: Can this be completed in one atomic action?
   - 0 = No, multiple steps required
   - 1 = Yes, single action

2. **No Complexity**: Is this trivial with no decision-making?
   - 0 = No, requires thought/decisions
   - 1 = Yes, completely straightforward

3. **No Tracking Value**: Would tracking this provide no benefit?
   - 0 = No, tracking would be helpful
   - 1 = Yes, tracking is overhead

4. **Immediate Completion**: Can this be done in <5 minutes?
   - 0 = No, will take longer
   - 1 = Yes, very quick

**Scoring Rubric**:
- **Score 0-2**: Project is appropriate (proceed)
- **Score 3-4**: One-off task recommended (suggest to user)

**Orchestrator Behavior When Score = 3-4**:

```
Orchestrator: "This seems like a relatively simple task that could be completed
              as a one-off without creating a full project structure.

              Would you like me to:
              1. Just do it as a one-off task (no project tracking)
              2. Create a project anyway (if you want the structure)

              What do you prefer?"
```

**Important**: Orchestrator SUGGESTS but never overrides user choice. If user says "create project anyway", orchestrator proceeds.

**Example Calculation**:

User wants to "Fix typo in README.md - 'authetication' → 'authentication'":
- Single Action: 1 (one edit)
- No Complexity: 1 (trivial)
- No Tracking: 1 (no value in tracking)
- Immediate: 1 (takes 30 seconds)
- **Total: 4** → One-off recommended (suggest to user)

User wants to "Add comprehensive API documentation":
- Single Action: 0 (multiple files)
- No Complexity: 0 (requires understanding API)
- No Tracking: 0 (tracking helps ensure completeness)
- Immediate: 0 (will take hours)
- **Total: 0** → Project is appropriate (proceed)

---

## Discovery Type Categorization

**Process**: When discovery phase is enabled, `/phase:discovery` is always invoked first. This command's primary purpose is to categorize the discovery work and delegate to the appropriate type-specific command.

**Flow**:
```
1. Orchestrator invokes `/phase:discovery`
2. `/phase:discovery` analyzes user's description and context
3. Categorizes work into one of:
   - bug (bug investigation)
   - feature (feature exploration)
   - docs (documentation gap analysis)
   - refactor (refactoring analysis)
   - general (doesn't fit categories or multiple categories)
4. Invokes appropriate type-specific command:
   - `/phase:discovery:bug`
   - `/phase:discovery:feature`
   - `/phase:discovery:docs`
   - `/phase:discovery:refactor`
   - `/phase:discovery:general` (fallback)
5. Type-specific command provides focused instructions
```

**Why This Approach**:
- Context preservation: Only load relevant instructions, not all 1000+ lines
- Flexibility: Can add new types without changing core flow
- Fallback: `/phase:discovery:general` handles edge cases and uncategorizable work
- Forced categorization: Orchestrator must think about work type

**Categorization Guidelines** (for `/phase:discovery`):

| User Description Contains | Likely Category |
|---------------------------|-----------------|
| "bug", "issue", "broken", "error", "doesn't work" | bug |
| "add", "new feature", "implement", "build" | feature |
| "docs out of date", "documentation gap", "code doesn't match docs" | docs |
| "refactor", "messy", "clean up", "2000 lines", "reorganize" | refactor |
| Multiple types or unclear | general |

**Example**:
```
User: "Fix the login bug"
→ `/phase:discovery` categorizes as "bug"
→ Invokes `/phase:discovery:bug`
→ Bug investigation instructions loaded

User: "Understand how the auth system works and improve the docs"
→ `/phase:discovery` sees both exploration + docs
→ Categorizes as "general" (mixed)
→ Invokes `/phase:discovery:general`
→ General discovery instructions loaded
```

---

## Resolved Design Decisions

### Discovery Type Inference ✅

**Decision**: Always invoke `/phase:discovery` as router. It categorizes and delegates to type-specific commands (`/phase:discovery:bug`, etc.). Fallback to `/phase:discovery:general` when uncategorizable.

**Rationale**: Forces categorization while maintaining flexibility. Preserves context by only loading relevant instructions.

---

### Design Phase Critical Thinking ✅

**Decision**: Use Design Worthiness Rubric (4 questions, 0-2 scoring each). Score 0-2 = skip, 3-5 = optional, 6-8 = recommended. Special rule: bug fixes get -3 penalty.

**Rationale**: Objective scoring prevents over-engineering while allowing flexibility. Medical-style rubric approach provides consistency.

---

### "I'm Not Sure" Responses ✅

**Decision**: Apply appropriate rubric (Discovery Worthiness or Design Worthiness). Calculate score, make recommendation with reasoning.

**Rationale**: Same rubric system used for uncertain users as for orchestrator decision-making. Provides consistent, objective recommendations.

---

### Modifying Phase Plan Mid-Project ✅

**Decision**: Handled via loop-back mechanics (documented in PHASE_SPECIFICATIONS.md). Not part of initial truth table.

**Rationale**: Loop-back is sufficient for adding phases mid-project. Avoids complicating initialization flow.

---

### Question Ordering Flexibility ✅

**Decision**: Orchestrator can skip obvious questions. Always confirm final phase plan before project creation, but can streamline the path to confirmation.

**Rationale**: Maximize efficiency, minimize user friction. Smart orchestrator behavior improves UX.

---

### Multiple Projects on Same Branch ✅

**Decision**: Supported naturally. After finalization deletes `.sow/project/`, user can create new project on same branch. Physical constraint (single `project/` directory) enforces one-at-a-time.

**Rationale**: Current design handles this correctly. No special logic needed.

---

### One-Off Task Threshold ✅

**Decision**: Use One-Off vs Project Rubric (4 binary questions). Score 3-4 = suggest one-off, but never override user. If user chose "new project", they can still accept suggestion or proceed with project.

**Rationale**: Helpful guidance without being prescriptive. Orchestrator suggests better workflow when appropriate but respects user choice.

---

### Context Assessment Depth ✅

**Decision**: Option B - Ask clarifying question ("Are these notes detailed enough to start implementation?") without requiring full review.

**Rationale**: Balances efficiency with understanding. Avoids over-interrogation while getting necessary information.

---

### Re-asking Questions ✅

**Decision**: Always ask follow-up questions for vague/unclear answers. No hard limit, but orchestrator should use judgment.

**Rationale**: Clarity is more important than brevity. Better to ask than make wrong assumptions.

---

### Rubric Question Wording ✅

**Decision**: Keep current rubric questions as-is. Refine based on real-world data showing LLMs consistently guessing wrong due to vagueness.

**Rationale**: Impossible to know if questions are too vague without real-world usage data. Start with current questions and iterate.

---

### Rubric Score Thresholds ✅

**Decision**: Keep current thresholds (0-2 skip, 3-5 optional, 6-8 recommended). Adjust based on real-world usage data.

**Rationale**: Good first attempt using equal ranges. Only data from watching LLMs use rubrics will guide improvements.

---

### Bug Fix Penalty ✅

**Decision**: Keep -3 penalty for bug fixes in Design Worthiness Rubric.

**Rationale**: Works really well and effectively discourages design phase for bug fixes without making it impossible.

---

### Discovery General Fallback ✅

**Decision**: Defer to slash command creation phase (Step 5 in implementation plan). Not pertinent to truth table structure.

**Rationale**: Truth table defines when to invoke `/phase:discovery:general`. Specific instructions are slash command implementation detail.

---

### Rubric Presentation to User ✅

**Decision**: Hide rubric scoring by default. Show only if user asks (e.g., "Why would you suggest that?").

**Rationale**: Avoid noise in normal flow. Provide transparency on demand when user wants to understand reasoning.

---

### Minimum Confirmation Requirements ✅

**Decision**: Only one hard requirement - orchestrator MUST get human confirmation of phase plan before creating project. No other hard requirements.

**Rationale**: When strongly confident, orchestrator can propose phase plan with justifications and ask for confirmation. Flexibility maximizes efficiency.

**Important**: Orchestrator can skip questions and go straight to phase plan confirmation if context is 100% clear. Zero questions is acceptable if final confirmation is obtained.

---

### Rubric Override ✅

**Decision**: User can override ANY orchestrator decision. Orchestrator accepts override without pushback. Exception: Requests that are impossible (nonexistent phases, two projects on same branch).

**Rationale**: User has final authority. For impossible requests, orchestrator explains why it's not possible in sow framework.

**Example**:
```
Orchestrator: "Based on scoring, I don't think we need design phase."
User: "I want to create design docs anyway."
Orchestrator: "Understood. Enabling design phase."
```

---

### Calibration Over Time ✅

**Decision**: Rubrics will be calibrated based on user feedback over time. Scores, thresholds, and questions are version-controlled and can evolve.

**Rationale**: Start with current rubrics, gather user feedback, make data-driven adjustments in future versions.

---

## Next Steps

### Completed ✅

1. ✅ **Resolve Initial Open Questions** - Completed 10 questions
2. ✅ **Add Rubric Framework** - Three rubrics defined (Discovery, Design, One-Off)
3. ✅ **Resolve Rubric-Specific Questions** - Completed 8 additional questions
4. ✅ **Document All Design Decisions** - 17 total resolved decisions documented

### Ready for Next Phase

This truth table is now **complete and ready for implementation**. The next phase of work (Step 3 in NOTES.md implementation plan) is to update the state schema.

**Subsequent Steps**:
- Step 3: Update State Schema (formalize YAML structure, update CUE schemas)
- Step 4: Rewrite Orchestrator (incorporate truth table logic and rubrics)
- Step 5: Update Commands (create `/project:new`, update `/continue`, etc.)
- Step 6: Update Documentation (reflect new phase model)

---

## Notes

- This truth table focuses on **initial project setup only**
- Phase transitions during project execution are covered in PHASE_SPECIFICATIONS.md
- The orchestrator's job is to guide the human to the right configuration, not to dictate it
- When in doubt, the orchestrator should ask rather than assume
