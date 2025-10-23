# Exploration Project Type - Design Document

## Overview

The **exploration project type** is a specialized workflow for research, design, and architectural investigation work. It provides a structured approach to exploring complex problems, evaluating solutions, and decomposing the findings into actionable work items for implementation.

## Purpose

### Why Exploration Projects Exist

Software development often requires upfront research and design work before implementation can begin. This includes:

- **Architecture decisions** - Evaluating multiple approaches for a technical problem
- **Technology research** - Investigating new libraries, frameworks, or tools
- **Design exploration** - Prototyping different UX/UI approaches
- **Problem investigation** - Understanding a complex bug or system behavior
- **Feasibility studies** - Determining if an approach is viable before committing

Previously, this work was handled through optional "discovery" and "design" phases within standard implementation projects. However, this created several problems:

1. **Mixing concerns** - Research work and implementation work have different goals, timelines, and success criteria
2. **Unclear scope** - It was unclear when discovery/design should be used vs. skipped
3. **No clear output** - Discovery/design artifacts lived in the project but weren't actionable
4. **Coupling issues** - Research findings couldn't easily inform multiple related implementation efforts

### The Exploration Project Solution

Exploration projects solve these problems by:

1. **Separating concerns** - Research/design work happens in its own dedicated project lifecycle
2. **Clear deliverable** - Exploration projects explicitly produce GitHub issues as their output
3. **Reusable findings** - A single exploration can inform multiple implementation projects
4. **Appropriate pacing** - Human-led exploration work can take as long as needed without blocking implementation

## Role in the Bigger Picture

### The Two-Project Model

The sow system now operates with two distinct project types:

**Exploration Projects** (this document)
- Human-led research and design work
- Produces knowledge artifacts and GitHub issues
- Flexible timeline, driven by human insight
- Output: Decomposed, actionable work items

**Standard Projects**
- AI-autonomous implementation work
- Consumes GitHub issues as input
- Structured timeline through defined phases
- Output: Working code and pull requests

### Workflow Integration

The typical workflow looks like:

```
User identifies complex problem
    ↓
Create exploration project (if needed)
    ↓
Conduct research, create design artifacts
    ↓
Decompose findings into GitHub issues
    ↓
Close exploration project
    ↓
Create standard project(s) from issues
    ↓
Implement, review, finalize
```

**Key insight:** Not all work requires exploration. Simple features, straightforward bugs, and well-understood tasks can go directly to standard projects. Exploration projects are for situations where the path forward is unclear.

## Phase Structure

Exploration projects follow a **3-phase lifecycle**:

1. **Exploration Phase** (human-led, flexible duration)
2. **Decomposition Phase** (AI-assisted, structured)
3. **Finalize Phase** (AI-autonomous, cleanup)

### Phase 1: Exploration

**Purpose:** Investigate the problem space and create knowledge artifacts

**Mode:** Human-led (subservient mode)
- The AI assists with research, provides suggestions, and helps document findings
- The human drives the direction and makes all key decisions
- No fixed timeline - exploration continues until the human is satisfied

**Activities:**
- Read and understand existing code/systems
- Research external resources (documentation, articles, examples)
- Prototype different approaches
- Create design documents, architecture diagrams, or decision records
- Evaluate tradeoffs between options
- Gather context that will inform implementation

**Artifacts Created:**
- Architecture Decision Records (ADRs)
- Design documents
- Diagrams (architecture, sequence, data flow)
- Research notes
- Prototypes or proof-of-concept code
- Comparison matrices (e.g., comparing libraries or approaches)

**Completion Criteria:**
- The human determines they have sufficient understanding to decompose work
- All key decisions have been documented
- Artifacts are approved by the human

**Special Characteristics:**
- **Iterative** - Can circle back to refine understanding
- **Exploratory** - Dead ends and pivots are expected and acceptable
- **Human judgment** - Success is determined by human satisfaction, not objective metrics
- **Artifact-focused** - The deliverable is knowledge, not code

### Phase 2: Decomposition

**Purpose:** Transform exploration findings into actionable GitHub issues

**Mode:** AI-assisted with human approval
- The AI suggests issue breakdown based on exploration artifacts
- The human reviews, refines, and approves the issues
- Structured workflow with clear deliverables

**Activities:**
- Review all exploration artifacts
- Identify distinct, implementable work items
- Draft GitHub issues with:
  - Clear titles
  - Detailed descriptions referencing exploration findings
  - Acceptance criteria
  - Links to relevant ADRs or design docs
  - Appropriate labels and metadata
- Group related issues (e.g., "these 3 issues implement the auth system")
- Sequence issues if there are dependencies
- Get human approval for each issue or batch of issues

**Deliverable Format:**
Issues are initially drafted as markdown files in the project workspace:
```
.sow/project/phases/decomposition/issues/
├── 001-implement-authentication.md
├── 002-add-user-management.md
└── 003-create-admin-panel.md
```

**Human Approval Loop:**
1. AI drafts issues
2. Human reviews and provides feedback
3. AI refines based on feedback
4. Human approves
5. AI creates GitHub issues via API

**Completion Criteria:**
- All exploration findings have been decomposed into issues
- All issues have been reviewed and approved by human
- Issues have been created in GitHub
- The exploration project state tracks all created issue numbers

**Quality Guidelines:**
- Each issue should be completable in a single standard project
- Issues should reference the exploration project for context
- Issues should be independent where possible (minimal dependencies)
- Issue scope should be appropriate (~1-4 hours of work each)

### Phase 3: Finalize

**Purpose:** Clean up and archive the exploration project

**Mode:** AI-autonomous with human oversight
- The AI handles routine cleanup tasks
- The human confirms completion

**Activities:**
- Verify all artifacts are committed to the repository
- Ensure ADRs are in the correct location (`.sow/knowledge/adrs/`)
- Move any useful reference materials to permanent locations
- Create a summary document linking to all created issues
- Update project metadata with completion information
- Remove temporary workspace files

**Completion Criteria:**
- All valuable artifacts are preserved
- All GitHub issues are created and linked
- Project state is marked complete
- Project directory is deleted (state already committed to branch)

**Special Note:**
Unlike standard projects, exploration projects typically **do not create pull requests**. The exploration work and created issues ARE the deliverable. The project branch with artifacts can be merged directly or kept as reference.

## Design Decisions

### Why Human-Led Exploration?

Research and design work requires:
- **Creative thinking** - Evaluating multiple approaches, considering tradeoffs
- **Context synthesis** - Connecting disparate pieces of information
- **Judgment calls** - Deciding when "good enough" understanding has been reached
- **Domain knowledge** - Understanding business requirements and user needs

These are areas where human insight is essential. The AI assists but doesn't drive.

### Why Separate from Standard Projects?

Exploration and implementation are fundamentally different:

| Aspect | Exploration | Implementation |
|--------|-------------|----------------|
| Goal | Understanding | Working code |
| Timeline | Open-ended | Bounded |
| Success criteria | Human satisfaction | Tests pass, requirements met |
| Mode | Human-led | AI-autonomous |
| Output | Knowledge + issues | Pull request |
| Iteration | Expected, exploratory | Structured, corrective |

Separating these concerns allows each to use the appropriate workflow.

### Why Decomposition as a Separate Phase?

Decomposition is distinct from exploration because:

1. **Different mindset** - Exploration is divergent (explore options), decomposition is convergent (create actionable plan)
2. **Clear gate** - Decomposition shouldn't start until exploration is "done enough"
3. **Different artifacts** - Exploration creates knowledge, decomposition creates work items
4. **AI assistance shifts** - During exploration, AI helps research; during decomposition, AI helps structure

## Usage Patterns

### When to Use Exploration Projects

**Use exploration when:**
- The problem is poorly understood
- Multiple approaches need evaluation
- Architectural decisions need research
- New technology needs investigation
- The scope of work is unclear

**Skip exploration when:**
- The issue already has clear requirements
- The solution approach is straightforward
- Similar work has been done before
- The task is well-scoped and ready for implementation

### Example Scenarios

**Scenario 1: New Authentication System**
1. Create exploration project: "evaluate-auth-approaches"
2. Exploration phase: Research OAuth, JWT, sessions; create ADR; prototype
3. Decomposition phase: Create issues for chosen approach (setup OAuth, add middleware, create user model, add login UI)
4. Finalize and close exploration project
5. Create standard projects from issues

**Scenario 2: Performance Investigation**
1. Create exploration project: "investigate-slow-api"
2. Exploration phase: Profile code, identify bottlenecks, test solutions
3. Decomposition phase: Create issues for each optimization (add caching, optimize query, update index)
4. Finalize and close exploration project
5. Create standard projects from issues

**Scenario 3: Simple Bug Fix**
- No exploration project needed
- Create standard project directly from bug report issue

## Future Considerations

### Potential Enhancements

**Template-based issue generation**
- Predefined issue templates for common patterns
- Automatically populate standard fields

**Artifact validation**
- Check that ADRs follow required format
- Verify all decisions are documented

**Cross-project linking**
- Track which implementation projects came from which exploration
- Show exploration artifacts in related standard projects

**Metrics and insights**
- Track exploration → implementation success rate
- Identify when exploration could have helped a troubled standard project

### Open Questions

- Should exploration projects support sub-phases (e.g., research → prototype → decide)?
- How to handle exploratory work that concludes "don't do this"?
- Can exploration projects create other exploration projects (nested research)?
- Should there be a "maintenance mode" exploration that stays open indefinitely?

## Summary

Exploration projects provide a **dedicated, human-led workflow for research and design work** that produces **actionable GitHub issues** as output. By separating this work from implementation, sow provides appropriate tooling and pacing for both creative investigation and structured development.

The three-phase lifecycle (Exploration → Decomposition → Finalize) ensures that research findings are systematically converted into implementation work, creating a clear bridge between "figuring out what to build" and "building it."
