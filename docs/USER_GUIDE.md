# User Guide

**Last Updated**: 2025-10-15
**Purpose**: Day-to-day usage workflows for users

This guide walks you through using `sow` for everyday software development, from installation through project completion.

---

## Table of Contents

- [Getting Started](#getting-started)
- [Starting a New Project](#starting-a-new-project)
- [Working Through Phases](#working-through-phases)
- [Working with External References](#working-with-external-references)
- [Providing Feedback](#providing-feedback)
- [Completing and Finalizing](#completing-and-finalizing)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Related Documentation](#related-documentation)

---

## Getting Started

### Prerequisites

**Claude Code**: Must be installed and working.

**Git Repository**: Existing or new repository (sow works with any git repo).

**sow CLI**: Required for initialization and validation. Download from GitHub releases for your platform.

### Installation Steps

**Step 1: Install Plugin**

Via Claude Code marketplace or git URL:
```
/plugin install https://github.com/your-org/sow
```

**Step 2: Restart Claude Code**

Plugin changes require restart. Exit and restart Claude Code.

**Step 3: Install CLI**

Download and install CLI for your platform:
```bash
# macOS
curl -L https://github.com/your-org/sow/releases/download/v0.2.0/sow-macos -o sow
chmod +x sow
mv sow ~/.local/bin/sow

# Verify
sow version
```

CLI provides: embedded CUE schemas, structure initialization, validation, fast logging, refs management.

**Step 4: Initialize Repository**

In your git repository:
```
/init
```

This creates `.sow/` directory structure with knowledge directory, refs directory with gitignore, version file. Commits structure to git.

**Step 5: Verify**

After restart, you should see:
```
📋 You are in a sow-enabled repository
💡 No active project. Use /project:new to begin

✓ Versions aligned (v0.2.0)
```

---

## Starting a New Project

### Understanding Project Initialization

When you start a new project, orchestrator will ask questions to determine which phases are needed. This truth table decision flow helps avoid over-engineering while ensuring necessary work is done.

**Five Fixed Phases**: Every project has the same five phases: discovery (optional), design (optional), implementation (required), review (required), finalize (required). Orchestrator helps you decide which optional phases to enable.

### Creating Feature Branch

**Critical Rule**: Always create feature branch before starting project. Sow enforces one project per branch.

```bash
# Create and switch to feature branch
git checkout -b feat/add-authentication
```

Branch naming conventions: `feat/` for features, `fix/` for bugs, `refactor/` for refactoring, `docs/` for documentation.

Why required: project state committed to feature branch, main branch stays clean, CI enforces no `.sow/project/` on main, natural cleanup via branch deletion.

### Using /project:new

Start new project:
```
/project:new
```

**Smart Branch Detection**: If on main/master, orchestrator offers to create feature branch. If existing project exists, orchestrator offers to continue existing or create new branch.

### Truth Table Decision Flow

Orchestrator asks questions to infer right phase plan:

**Question 1: What Are You Trying to Accomplish?**

Orchestrator asks: "What would you like to accomplish in this project?"

Provide clear description:
```
Add JWT-based authentication with user login, token refresh,
and password hashing. Should integrate with existing User model.
```

**Question 2: Existing Context Assessment**

Orchestrator asks: "Do you have any existing context, documents, or notes about this?"

Responses guide phase plan:
- "No" or "Not really" → Likely needs discovery/design
- "Yes, I have design docs" → May skip discovery/design
- "There's documentation at X" → Likely skip phases

**Question 3: Discovery Phase Decision** (if limited context)

Orchestrator suggests: "It sounds like we could benefit from discovery to better understand [the problem/requirements]. Would you like to start with discovery phase?"

Responses:
- Yes → Discovery enabled
- No → Discovery disabled
- Tell me more → Orchestrator explains benefits

**Question 4: Design Phase Decision** (after discovery or if some context exists)

Orchestrator asks: "Do you want to create formal design documents before implementation?"

Orchestrator uses Design Worthiness Rubric to make recommendation. Suggests design when: large scope (10+ tasks), new system component, architectural changes, multiple integration points, user uncertain.

Suggests skipping design when: bug fixes, small features (1-5 tasks), minor refactors, discovery notes sufficient.

**Final Confirmation**

Orchestrator presents phase plan:
```
📋 Proposed plan:

Phases enabled:
- Discovery: enabled (research JWT libraries and approach)
- Design: enabled (document architecture decisions)
- Implementation: enabled (always required)
- Review: enabled (always required)
- Finalize: enabled (always required)

Discovery and design will require your approval before continuing.
Implementation, review, and finalize run autonomously.

Ready to proceed? [yes/modify]
```

### Project Creation

After approval, orchestrator creates project structure and transitions to first enabled phase.

Output:
```
✓ Created project 'add-authentication'

Branch: feat/add-auth

Phase structure initialized:
- Discovery: enabled
- Design: enabled
- Implementation: enabled
- Review: enabled
- Finalize: enabled

✓ Committed to git

Starting discovery phase...
```

---

## Working Through Phases

### Discovery Phase (Optional, Human-Led)

**When Enabled**: Research and investigation needed to understand problem or requirements.

**Orchestrator Mode**: Subservient (acts as assistant, you lead).

**Workflow**:

1. **Categorization**: `/phase:discovery` categorizes work as bug, feature, docs, refactor, or general.

2. **Research Work**: Orchestrator helps you research, or spawns researcher agent for focused investigation. Creates research reports in `phases/discovery/research/`.

3. **Conversation and Notes**: Orchestrator takes notes continuously in `phases/discovery/notes.md`. Captures key decisions in `phases/discovery/decisions.md`.

4. **Artifacts Created**: Research reports, notes, decisions. All tracked in project state with approval flags.

5. **Approval Required**: You must approve all artifacts before discovery completes. Orchestrator asks: "Discovery artifacts ready for your review. Approve to continue to design?"

**Example Discovery Workflow**:
```
[Discovery phase active]

Orchestrator: "I've categorized this as feature discovery. Would you like me
to research JWT libraries and authentication approaches?"

You: "Yes, please focus on RS256 implementation and token refresh patterns."

[Orchestrator spawns researcher agent]

Researcher: [Creates research/001-jwt-libraries.md comparing options]

Orchestrator: "Research complete. I've documented findings in
discovery/research/001-jwt-libraries.md. Key recommendation: use PyJWT
with RS256. Ready to approve and move to design?"

You: "Approved."

[Discovery completes, transitions to design]
```

### Design Phase (Optional, Human-Led)

**When Enabled**: Formal architecture decisions and design documentation needed.

**Orchestrator Mode**: Subservient (acts as assistant, you lead).

**Workflow**:

1. **Design Alignment Subphase**: Conversational work refining discovery into high-level architecture decisions. Orchestrator facilitates, you guide. Notes captured in `phases/design/notes.md`.

2. **Formalization**: When ready, orchestrator creates formal design documents. For simple docs, orchestrator handles directly. For complex docs, orchestrator spawns architect agent.

3. **Artifacts Created**: ADRs in `phases/design/adrs/`, design documents in `phases/design/design-docs/`, optional diagrams. All tracked with approval flags.

4. **Approval Required**: You must approve all artifacts before design completes.

5. **Transition**: Once approved, orchestrator moves to implementation automatically.

**Example Design Workflow**:
```
[Design phase active]

Orchestrator: "Based on discovery, we need to decide: symmetric vs asymmetric
JWT signing. What are your thoughts?"

You: "Use RS256 asymmetric signing for better security."

Orchestrator: "Good choice. I'll document this as an ADR. Also need to decide
on token storage approach..."

[Conversation continues, notes captured]

Orchestrator: "Ready to formalize? I'll create ADR for RS256 decision and
design doc for overall auth system architecture."

You: "Yes."

[Orchestrator spawns architect agent]

Architect: [Creates adrs/001-use-jwt-rs256.md and design-docs/auth-system.md]

Orchestrator: "Design documents created. Review phases/design/ directory.
Approve to continue to implementation?"

You: "Approved."

[Design completes, transitions to implementation]
```

### Implementation Phase (Required, AI-Autonomous)

**Always Happens**: Cannot be skipped.

**Orchestrator Mode**: Autonomous (executes independently within boundaries).

**Workflow**:

1. **Task Breakdown**: Orchestrator creates implementation plan from discovery/design artifacts or your description. For very large projects, may spawn planner agent. Presents task breakdown for your approval.

2. **Task Execution**: Orchestrator spawns implementer agents for each task. Maximum autonomy: no approval needed for normal task flow. Implementer uses TDD (test-driven development).

3. **Parallel Execution**: Tasks marked as parallelizable run simultaneously.

4. **Fail-Forward**: If problem reveals new work needed, orchestrator requests approval to add tasks. Original approach may be marked abandoned rather than deleted.

5. **Human Approval Required For**: Adding new tasks (fail-forward), going back to previous phases, implementer stuck needing help.

6. **Automatic Transition**: When all tasks complete, automatically transitions to review.

**Example Implementation Workflow**:
```
[Implementation phase active]

Orchestrator: "Creating task breakdown from design documents..."

Proposed tasks:
- 010: Create User model with validation
- 020: Implement JWT service with RS256
- 030: Create login endpoint
- 040: Add authentication middleware

Approve task breakdown? [yes/modify]

You: yes

[Orchestrator spawns implementer for task 010]

Implementer: [Works on User model, logs actions, completes]

Orchestrator: "Task 010 complete. Moving to task 020..."

[Continues through all tasks]

Orchestrator: "All implementation tasks complete. Transitioning to review..."
```

**Providing Feedback During Implementation**: If implementer makes mistake, tell orchestrator. Orchestrator creates feedback file, increments iteration, spawns implementer with correction context.

Example:
```
You: "The JWT service should use RS256, not HS256."

Orchestrator: "I'll create feedback for the implementer."
[Creates feedback/001.md, increments iteration]

Implementer: [Reads feedback, fixes algorithm, completes]
```

### Review Phase (Required, AI-Autonomous)

**Always Happens**: Cannot be skipped.

**Orchestrator Mode**: Autonomous with mandatory review.

**Workflow**:

1. **Orchestrator Review**: Orchestrator always performs review (mandatory). Reads all requirements and all implementation changes. Compares to validate original intent met. For large/complex changes, may spawn reviewer agent for assistance.

2. **Review Report**: Creates numbered review report (001-review.md, 002-review.md if loop-back). Report includes: requirements summary, changes summary, findings, assessment (pass/fail), next steps.

3. **Human Review**: After orchestrator review, you can review yourself or trust orchestrator's assessment. You have final approval authority.

4. **Loop-Back If Issues Found**: If issues identified (by orchestrator, reviewer agent, or you): orchestrator identifies what needs fixing, adds tasks to implementation, you approve, returns to implementation, executes fixes, automatically returns to review, creates new review report.

5. **Multiple Iterations**: Review → Implementation → Review cycle continues until success.

6. **Approval Required**: You must explicitly approve before proceeding to finalize.

**Example Review Workflow**:
```
[Review phase active]

Orchestrator: "Performing review of implementation..."
[Reads requirements, reads all changes]

Orchestrator: "Review complete. Created phases/review/reports/001-review.md.

Assessment: FAIL
Issue found: Login endpoint missing error handling for invalid credentials.

Recommend adding task to implementation phase to address this. Approve?"

You: "Yes, add the task."

[Orchestrator adds task 050, returns to implementation]

Implementer: [Completes error handling task]

[Automatic transition back to review]

Orchestrator: "Second review complete. Created 002-review.md.

Assessment: PASS
All issues addressed. Implementation meets original intent.

Approve to proceed to finalize?"

You: "Approved."

[Transitions to finalize]
```

### Finalize Phase (Required, AI-Autonomous)

**Always Happens**: Cannot be skipped.

**Orchestrator Mode**: Autonomous for most actions.

**Workflow**:

1. **Documentation Subphase**: Orchestrator determines if documentation updates needed. Proposes changes (README, API docs, changelogs). Decides whether to move design artifacts to repository (ADRs typically moved). Requests your approval before making changes.

2. **Final Checks**: Orchestrator runs: full test suite (must pass), configured linters (must pass), documentation completeness check, git tree cleanliness check.

3. **Project Deletion** (MANDATORY): Orchestrator deletes `.sow/project/` folder. Creates cleanup commit. This is critical gate - cannot complete finalize without deletion.

4. **PR Creation**: Orchestrator creates pull request using GitHub CLI (`gh`) or provides instructions for manual creation. PR includes comprehensive description: summary, changes made, testing, documentation.

5. **Handoff to Human**: Orchestrator hands off. You review PR, merge when ready, delete feature branch.

**Example Finalize Workflow**:
```
[Finalize phase active]

Orchestrator: "Documentation subphase: I recommend updating README.md to
document new auth endpoints. Approve?"

You: "Yes."

[Orchestrator or documenter agent updates README]

Orchestrator: "Running final checks..."
✓ All tests passing (42 tests)
✓ All linters passing
✓ Documentation current
✓ Git tree clean

Orchestrator: "Deleting project folder..."
✓ Deleted .sow/project/
✓ Created cleanup commit

Orchestrator: "Creating pull request..."
✓ PR created: https://github.com/org/repo/pull/42

Work complete! Please review PR and merge when ready.

Recommend squash-merge to keep main branch history clean.
```

---

## Working with External References

### Understanding the Refs System

Refs provide unified system for external git resources: knowledge refs (style guides, API standards, policies), code refs (implementation examples, reference architectures).

Replaces old sinks and repos systems with single consistent mechanism.

**Benefits**: Clone once and reference from multiple repos, automatic staleness detection, AI agents consult refs when making decisions, team-wide sharing via committed index.

### Adding References

**Add Remote Reference**:
```bash
sow refs add https://github.com/acme/style-guides \
  --path python/ \
  --link python-style \
  --type knowledge \
  --tag formatting --tag naming --tag testing \
  --desc "Python coding standards" \
  --summary "Covers PEP 8 formatting, naming conventions, and testing patterns."
```

**Add Local Reference** (not shared with team):
```bash
sow refs add file:///Users/you/local-docs \
  --link local-docs \
  --type knowledge \
  --tag wip \
  --desc "Work-in-progress documentation" \
  --summary "Draft docs before publishing."
```

**Via Orchestrator** (natural language):
```
You: "Add our team's Python style guide from github.com/acme/style-guides"

Orchestrator: "I'll install that as a knowledge reference."
[Examines content, determines tags, adds reference]
```

### Fresh Clone Setup

After cloning repository with existing refs configured:

```bash
# Clone repo
git clone https://github.com/org/repo
cd repo

# Initialize refs
sow refs init
```

This clones repos to cache and creates symlinks per committed index.

When Claude Code starts:
```
Orchestrator: "I notice 3 references configured but not initialized:
- python-style
- api-standards
- security-guide

Should I set them up?"

You: "Yes"

[Orchestrator runs sow refs init]

Orchestrator: "✓ Initialized 3 references. I can now reference these guides
when working on code."
```

### Checking for Updates

```bash
# Check staleness
sow refs status

# Output shows which refs behind
python-style: current
api-standards: behind by 3 commits
```

Orchestrator integration:
```
You: "Check if our references are up to date"

Orchestrator: [Runs sow refs status]
"I checked all references:
- python-style: current
- api-standards: behind by 3 commits

The api-standards ref has updates. Would you like me to update it?"

You: "Yes"

[Orchestrator runs sow refs update api-standards]
```

### How Agents Use Refs

During implementation, orchestrator compiles context for tasks. Includes relevant refs in task state references list.

Example task state:
```yaml
task:
  id: "020"
  references:
    - refs/python-style/conventions.md
    - refs/api-standards/rest-standards.md
    - refs/security-guide/jwt-best-practices.md
```

Implementer reads all referenced files at startup. Ensures code follows team standards.

---

## Providing Feedback

### When to Provide Feedback

Provide feedback when: agent made incorrect assumption, implementation doesn't meet requirements, code needs specific changes, agent got stuck on wrong approach.

### How Feedback Works

**Process**:

1. You identify issue and explain to orchestrator.

2. Orchestrator creates numbered feedback file in task directory.

3. Orchestrator increments task iteration counter.

4. Orchestrator spawns worker with feedback context.

**Example**:
```
You: "The token expiration is hardcoded but should be configurable via
environment variable."

Orchestrator: "I'll create feedback for the implementer."
[Creates feedback/001.md]
[Increments iteration: 2 → 3]
[Spawns implementer-3]

Implementer:
1. Reads state (sees iteration=3)
2. Reads feedback/001.md
3. Makes token expiration configurable
4. Updates tests
5. Marks feedback as addressed
6. Reports completion
```

### Feedback Storage Format

Feedback files numbered chronologically (001.md, 002.md, 003.md) in `phases/implementation/tasks/<id>/feedback/`.

Content includes: issue description, required changes, additional context, references to standards or docs.

### Multiple Feedback Rounds

Single task can have multiple feedback iterations. Each creates new feedback file. Worker addresses all pending feedback before completing task. Iteration counter tracks attempts.

Example progression:
- Iteration 1: Initial attempt, gets basic approach wrong
- Feedback 001 created: "Use RS256 not HS256"
- Iteration 2: Fixes algorithm, discovers missing error handling
- Feedback 002 created: "Add error handling for expired tokens"
- Iteration 3: Adds error handling, completes successfully

---

## Completing and Finalizing

### When Work is Done

All implementation tasks complete → Review phase runs → Review passes → Finalize phase begins.

### Finalize Phase Workflow

**Documentation Updates**: Orchestrator proposes documentation changes. You approve or decline. Updates made to README, API docs, changelogs as needed.

**Final Checks**: Tests must pass (orchestrator runs full suite). Linters must pass (orchestrator runs configured linters). Git tree must be clean (all work committed).

**Project Deletion**: MANDATORY before completion. Orchestrator deletes `.sow/project/` folder. Creates cleanup commit. This prevents project state from being merged to main.

**Pull Request Creation**: Orchestrator uses GitHub CLI to create PR automatically (if available). Otherwise provides instructions for manual PR creation. PR includes comprehensive description with summary, changes, testing notes, documentation updates.

### Merging and Cleanup

After finalize phase:

1. **Review PR**: Use GitHub UI or local review workflow.

2. **Merge**: Orchestrator recommends squash-merge to keep main history clean. After merge, delete feature branch.

3. **CI Enforcement**: Your CI should check that `.sow/project/` does not exist before allowing merge. If present, CI fails with error.

Example CI check:
```yaml
- name: Ensure no active sow project
  run: |
    if [ -d ".sow/project" ]; then
      echo "Error: .sow/project/ must be removed before merging"
      exit 1
    fi
```

### Starting Next Project

After merge, start fresh:

```bash
# Create new feature branch
git checkout main
git pull
git checkout -b feat/next-feature

# Start new project
/project:new
```

Previous project state gone (deleted during finalize). Each project independent.

---

## Best Practices

### Project Scope

**Do**: Keep projects short-lived (hours to days, not weeks). Scope to what fits in one feature branch. Delete after merging. Create new project for follow-up work.

**Don't**: Create mega-projects spanning weeks. Try to fit unrelated work in one project. Keep project state after merge. Extend completed projects.

**Guideline**: If project feels too big, break into multiple feature branches with separate projects.

### Branch Management

**Do**: Always create feature branch before starting project. Name branches descriptively (`feat/add-auth`, `fix/token-bug`). Keep one project per branch (enforced by sow). Run finalize phase before merging.

**Don't**: Create projects on main/master. Switch branches mid-work without committing state. Merge with `.sow/project/` still present. Try to reuse project state across branches.

**Pattern**:
```bash
# Good workflow
git checkout -b feat/my-feature
/project:new
# Work happens through phases
# Finalize deletes project
git push
# Create PR, merge, delete branch

# Bad workflow
# Work on main
/project:new  # ERROR: Can't create on main
```

### Commit Hygiene

**Do**: Commit `.sow/project/` regularly during work (enables resumability and team collaboration). Push to remote for backup. Finalize phase handles cleanup automatically.

**Don't**: Git-ignore `.sow/project/` (breaks branch switching and resumability). Leave uncommitted project state. Manually delete project folder (use finalize phase instead).

**Commit Strategy**:
```bash
# During work (feature branch)
git add .sow/project/ src/
git commit -m "feat: implement JWT service"
git push

# Finalize handles cleanup
# Creates cleanup commit automatically
```

### Team Collaboration

**Do**: Commit `.claude/` configuration (shared agents/commands). Commit `.sow/knowledge/` (shared docs). Commit `.sow/refs/index.json` (ref catalog). Commit `.sow/project/` to feature branches (enables collaboration). Share ref sources with team.

**Don't**: Commit `.sow/refs/` contents (large, gitignored). Modify `.claude/` without team discussion. Change shared refs without coordination.

**Collaboration Pattern**:
```bash
# Developer A creates feature with project
git checkout -b feat/auth
/project:new
# Some work done
git push

# Developer B continues the work
git fetch
git checkout feat/auth
sow refs init  # Initialize refs if needed
/project:continue  # Sees all context, continues seamlessly
```

### Using Optional Phases Wisely

**Skip Discovery When**: You have detailed notes or requirements. Work is straightforward. Implementation path completely clear.

**Skip Design When**: Bug fixes (just implement). Small features (1-5 tasks). Minor refactors. Discovery notes provide sufficient detail.

**Use Discovery When**: Problem or requirements unclear. Need to investigate codebase. Research needed for approach. User uncertain about feasibility.

**Use Design When**: Large scope (10+ tasks). New system component. Architectural changes. Multiple integration points. User uncertain about approach.

**Trust the Rubrics**: Orchestrator uses scoring rubrics to make objective recommendations. Trust the assessment but feel free to override if you have specific needs.

---

## Troubleshooting

### Common Issues

**Issue: Can't start project on main branch**

Error: "Cannot create project on main/master branch"

Solution: Create feature branch first.
```bash
git checkout -b feat/my-feature
/project:new
```

**Issue: Project already exists**

Error: "A project already exists: 'Old Project'"

Solutions:
- Continue existing: `/project:continue`
- Clean up and start new: Finalize old project first, then `/project:new`
- New branch: `git checkout -b feat/new-feature` then `/project:new`

**Issue: Refs not initialized after clone**

Symptom: References configured but not available.

Solution:
```bash
sow refs init
```

Or let orchestrator handle it when prompted.

**Issue: Version mismatch**

Warning: "Repository structure 0.1.0, Plugin version 0.2.0"

Solution: Migration needed (not yet implemented in new design).

**Issue: Lost project state after branch switch**

Symptom: Project state different than expected.

Cause: `.sow/project/` wasn't committed.

Solution:
```bash
git status
git add .sow/project/
git commit -m "chore: commit project state"
```

### Recovery Procedures

**Corrupted Project State**:

Check if YAML valid:
```bash
sow validate
```

Restore from git if corrupted:
```bash
git checkout HEAD~1 .sow/project/state.yaml
```

Or start fresh (if previous state not critical).

**Broken Ref Indexes**:

Regenerate indexes:
```bash
sow refs init
```

**CLI Not Found**:

Verify installation:
```bash
ls ~/.local/bin/sow
```

Ensure PATH includes `~/.local/bin`:
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Getting Help

1. **Check Documentation**: Start with relevant sections in this guide and related docs.

2. **Validate Structure**: Run `sow validate` to check for structural issues.

3. **Review Logs**: Check project and task logs for error details.

4. **GitHub Issues**: Report bugs with sow version, plugin version, error messages, reproduction steps.

5. **Community**: Join discussions, ask questions, share solutions.

---

## Related Documentation

- **[OVERVIEW.md](./OVERVIEW.md)** - System overview and core concepts
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Design philosophy and architecture
- **[PROJECT_LIFECYCLE.md](./PROJECT_LIFECYCLE.md)** - Truth table and 5-phase model details
- **[PHASES/](./PHASES/)** - Individual phase specifications
- **[REFS.md](./REFS.md)** - External references system
- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - Complete CLI command reference
- **[SCHEMAS.md](./SCHEMAS.md)** - File format specifications
- **[AGENTS.md](./AGENTS.md)** - Multi-agent system details
