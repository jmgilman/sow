# Commands and Skills

**Last Updated**: 2025-10-12
**Status**: Comprehensive Architecture Documentation

---

## Table of Contents

- [Overview](#overview)
- [Slash Commands vs Skills](#slash-commands-vs-skills)
- [User Workflow Commands](#user-workflow-commands)
  - [/init](#init)
  - [/start-project](#start-project)
  - [/continue](#continue)
  - [/cleanup](#cleanup)
  - [/migrate](#migrate)
  - [/sync](#sync)
- [Skills System](#skills-system)
  - [What Are Skills?](#what-are-skills)
  - [Skills by Agent](#skills-by-agent)
  - [Architect Skills](#architect-skills)
  - [Implementer Skills](#implementer-skills)
  - [Integration Tester Skills](#integration-tester-skills)
  - [Reviewer Skills](#reviewer-skills)
  - [Documenter Skills](#documenter-skills)
- [Command File Format](#command-file-format)
- [Smart UX Flows](#smart-ux-flows)
- [Creating Custom Commands](#creating-custom-commands)
- [Related Documentation](#related-documentation)

---

## Overview

`sow` uses two types of commands:

1. **User Workflow Commands** - High-level commands users invoke to manage projects
2. **Skills** - Granular capabilities that agents invoke to accomplish specific tasks

Both are implemented as slash commands (markdown files), but serve different purposes and audiences.

---

## Slash Commands vs Skills

### User Workflow Commands

**Location**: `.claude/commands/workflows/`

**Purpose**: User-facing commands for managing the sow lifecycle

**Audience**: Developers using sow

**Examples**:
- `/init` - Bootstrap sow in repository
- `/start-project` - Create new project
- `/continue` - Resume work
- `/cleanup` - Delete project before merge

**Invocation**: Directly by users

### Skills

**Location**: `.claude/commands/skills/`

**Purpose**: Specialized capabilities for agents to accomplish specific tasks

**Audience**: AI agents (orchestrator and workers)

**Examples**:
- `/create-adr` - Create Architecture Decision Record (used by architect)
- `/implement-feature` - Implement feature with TDD (used by implementer)
- `/write-integration-tests` - Write integration tests (used by integration-tester)

**Invocation**: By agents during task execution

**Key Design Decision**: Skills ARE slash commands, not a separate system. This prevents context window bloat while maintaining composability.

---

## User Workflow Commands

### /init

**Purpose**: Bootstrap sow in a repository

**Usage**:
```
/init
```

**Actions**:
1. Check if already initialized (exit if `.sow/` exists)
2. Create `.sow/` structure:
   - `.sow/knowledge/` (with `overview.md` template)
   - `.sow/sinks/` (with `index.json`)
   - `.sow/repos/` (with `index.json`)
   - `.sow/.version` (tracks structure version)
3. Create `.gitignore` entries:
   - `.sow/sinks/`
   - `.sow/repos/`
4. Commit structure to git
5. Offer optional CLI installation

**Success Output**:
```
✓ sow initialized successfully!

Available commands:
  /start-project <name> - Create new project
  /continue - Resume existing project (if any)

Try: /start-project "Add authentication"
```

**Error Cases**:
- Already initialized: "sow already initialized in this repository"
- Not in git repository: "Must be in a git repository to use sow"

---

### /start-project

**Purpose**: Create new project with planning

**Usage**:
```
/start-project <name>
```

**Example**:
```
/start-project "Add authentication"
```

**Validation Checks**:
1. **Not on default branch** - Errors if on `main` or `master`
2. **No existing project** - Errors if `.sow/project/` already exists
3. **Valid git branch** - Must be on a feature branch

**Smart UX Flow**:

#### Case 1: On default branch (main/master)
```
Error: Cannot create project on main/master branch

Would you like to create a new feature branch?
  Branch name suggestion: feat/add-authentication

[y/n] → If yes: git checkout -b feat/add-authentication, then create project
```

#### Case 2: On feature branch, no existing project
```
Creating project 'Add authentication' on branch 'feat/add-auth'...
[Proceed with planning]
```

#### Case 3: On feature branch WITH existing project
```
A project already exists: 'existing-project-name'

Options:
1. Continue with existing project (/continue)
2. Create new branch and start fresh project

[User chooses option]
```

**Planning Workflow**:
1. Prompt user for project description
2. Assess complexity (1-3 scale):
   - Rating: 1=simple, 2=moderate, 3=complex
   - Metrics: estimated files, cross-cutting concerns, new dependencies
3. **Select initial phases** (1-2 phases, minimal start):
   - Simple: just `implement`
   - Moderate: `design` + `implement`
   - Complex: `discovery` + `design` OR `design` + `implement`
   - **Don't select all phases upfront** (add dynamically later)
4. Create initial tasks with gap numbering (010, 020, 030)
5. Assign appropriate agent to each task
6. Create `.sow/project/state.yaml` (includes branch name)
7. Commit to git

**Progressive Planning Philosophy**:
- Start minimal, discover as you go
- Add phases when actually needed, not speculatively
- Request user approval when adding new phases
- Avoid waterfall planning (all phases upfront)

---

### /continue

**Purpose**: Resume existing project

**Usage**:
```
/continue
```

**Prerequisites**:
- `.sow/project/` must exist
- Branch name in state.yaml should match current branch (warning if mismatch)

**Actions**:
1. Read `.sow/project/state.yaml`
2. Verify branch matches (safeguard against branch switching issues)
3. Show current status:
   - Project name
   - Active phase
   - Completed tasks
   - Pending tasks
   - Current task (if any)
4. Identify next pending task
5. Compile context for assigned worker
6. Spawn appropriate worker

**Example Output**:
```
📋 Project: Add authentication (branch: feat/add-auth)

Status:
  Phase: implement (2/4 tasks complete)

  ✓ design/010: Design authentication flow (architect)
  ✓ implement/010: Create User model (implementer)
  → implement/020: Create JWT service (implementer) - IN PROGRESS
  ⬜ implement/030: Add login endpoint (implementer)

Resuming task 020: Create JWT service
Spawning implementer worker...
```

**Error Cases**:
- No project exists: "No active project. Use /start-project <name> to begin"
- Branch mismatch: "Warning: Project was created on branch 'feat/old', but you're on 'feat/new'. Continue anyway? [y/n]"

---

### /cleanup

**Purpose**: Delete project and prepare for merge

**Usage**:
```
/cleanup
```

**Prerequisites**:
- `.sow/project/` must exist

**Validation**:
- Warn if tasks are incomplete (but allow cleanup anyway)
- Show summary of what will be deleted

**Actions**:
1. Check all tasks complete (warn if not, confirm anyway)
2. Delete `.sow/project/` directory
3. Stage deletion (`git rm -rf .sow/project/`)
4. Prompt user to commit cleanup
5. Remind about CI check

**Example Output**:
```
⚠️ Warning: 1 task is still pending
  → implement/030: Add login endpoint

Are you sure you want to clean up? [y/n]

User: y

✓ Deleted .sow/project/
✓ Staged deletion for commit

Next steps:
1. Commit cleanup: git commit -m "chore: cleanup sow project state"
2. Push and create PR
3. CI will verify no .sow/project/ exists before merge

Note: Recommend squash-merge to keep main branch history clean
```

**Why Cleanup is Required**:
- Project state should not be merged to main branch
- CI enforces this by failing if `.sow/project/` exists in PR
- Keeps main branch clean
- Squash merge removes feature branch commits

---

### /migrate

**Purpose**: Migrate sow structure to new version

**Usage**:
```
/migrate
```

**When to Use**:
- After upgrading sow plugin to new version
- SessionStart hook detects version mismatch
- Prompted automatically when mismatch detected

**Actions**:
1. Read versions:
   - Current: `.sow/.version` → `sow_structure_version`
   - Target: `.claude/.plugin-version` → plugin version
2. Build migration chain if skipping versions:
   - If current `0.1.0` and target `0.3.0`:
   - Apply: `0.1.0-to-0.2.0.md`
   - Then: `0.2.0-to-0.3.0.md`
3. Parse migration file(s) from `.claude/migrations/`
4. Execute automated steps:
   - Create directories
   - Convert files
   - Update state files
5. Update `.sow/.version`
6. Commit changes: `chore: migrate sow structure to v<target>`
7. Report completion and changes

**Example Output**:
```
🔍 Analyzing versions...
   Current: 0.1.0
   Target: 0.2.0

📋 Migration: 0.1.0 → 0.2.0

Changes in this migration:
- Add .sow/project/context/ directory for project-specific context
- Convert state.json to state.yaml (YAML format)
- Add iteration field to all task states
- Update hooks.json with SessionStart hook

🚀 Applying migration...

✓ Created .sow/project/context/ (if project exists)
✓ Converted state.json to state.yaml
✓ Added iteration field to 3 tasks
✓ Updated .sow/.version
✓ Committed changes

Migration complete! Your repository is now at v0.2.0

📝 Changelog: https://github.com/your-org/sow/blob/main/CHANGELOG.md#020
```

**Sequential Migration** (skipping versions):
```
🔍 Multiple migrations required: 0.1.0 → 0.3.0

Migration path:
  0.1.0 → 0.2.0 (add context support)
  0.2.0 → 0.3.0 (add parallel task support)

Apply all migrations? [y/n]

User: y

🚀 Applying 0.1.0 → 0.2.0...
✓ Complete

🚀 Applying 0.2.0 → 0.3.0...
✓ Complete

All migrations applied successfully!
Your repository is now at v0.3.0
```

**Rollback** (if issues):
```bash
git revert HEAD  # Revert migration commit
/plugin install sow@sow-marketplace --version 0.1.0  # Reinstall old version
# Restart Claude Code
```

---

### /sync

**Purpose**: Sync installed sinks and linked repositories

**Usage**:
```
/sync
```

**Actions**:
1. Check for sink updates:
   - Poll git remotes for each installed sink
   - Compare current commit with remote HEAD
2. Check for repo updates:
   - Poll git remotes for linked repositories
   - Compare current commit with remote HEAD
3. Show available updates with descriptions
4. Prompt for selective updates
5. Pull updates for selected items
6. Regenerate indexes (`.sow/sinks/index.json`, `.sow/repos/index.json`)

**Example Output**:
```
🔍 Checking for updates...

Sinks:
  ✓ python-style (v1.2.0) - up to date
  ⬆️ api-conventions (v2.3.0 → v2.4.0) - Added GraphQL conventions
  ⬆️ deployment-guide (v1.0.0 → v1.1.0) - Updated k8s configs

Repositories:
  ✓ auth-service - up to date
  ⬆️ shared-library (abc1234 → def5678) - Added new crypto utilities

Update all? [y/n/selective]

User: selective

Select items to update:
  [y] api-conventions
  [n] deployment-guide
  [y] shared-library

Updating...
✓ api-conventions updated to v2.4.0
✓ shared-library updated to def5678
✓ Regenerated indexes

2 items updated, 3 unchanged
```

**Manual Sync Commands** (via CLI):
```bash
# Sync all sinks
sow sinks update

# Sync all repos
sow repos sync

# Sync everything
sow sync
```

---

## Skills System

### What Are Skills?

Skills are **agent-invoked capabilities** implemented as slash commands. They allow agents to perform complex, multi-step operations without bloating the agent's system prompt.

**Design Decision**: Skills ARE slash commands, not a separate abstraction. This prevents context window bloat while maintaining composability.

**Benefits**:
- Agent prompts stay concise (references, not full instructions)
- Skills can be 100+ lines without bloating agent context
- Slash commands already support subdirectories for organization
- Distribution via plugin system (already solved)
- Users can invoke skills directly OR through agents

**Structure**:
```
.claude/commands/skills/
├── architect/
│   ├── create-adr.md
│   └── design-doc.md
├── implementer/
│   ├── implement-feature.md
│   └── fix-bug.md
├── integration-tester/
│   └── write-integration-tests.md
├── reviewer/
│   └── review-code.md
└── documenter/
    └── update-docs.md
```

### Skills by Agent

Skills are organized by the agent that primarily uses them:

| Agent | Skills | Purpose |
|-------|--------|---------|
| **architect** | create-adr, design-doc | Design and architecture decisions |
| **implementer** | implement-feature, fix-bug | Code implementation with TDD |
| **integration-tester** | write-integration-tests | Integration and E2E testing |
| **reviewer** | review-code | Code quality and refactoring |
| **documenter** | update-docs | Documentation maintenance |

---

### Architect Skills

#### /create-adr

**Purpose**: Create an Architecture Decision Record

**Used by**: architect agent during `design` or `discovery` phases

**Parameters**: Decision title and context

**Example Invocation**:
```
/create-adr "Use PostgreSQL for primary database"
```

**Actions**:
1. Determine next ADR number (e.g., `003`)
2. Create file: `.sow/knowledge/adrs/003-use-postgresql.md`
3. Use ADR template:
   - Title
   - Status (proposed, accepted, rejected, deprecated, superseded)
   - Context (situation and forces at play)
   - Decision (what we decided)
   - Consequences (positive and negative impacts)
4. Fill in based on conversation context
5. Commit to git

**Template**:
```markdown
# 3. Use PostgreSQL for Primary Database

**Status**: Accepted

**Date**: 2025-10-12

## Context

[Situation and forces at play...]

## Decision

[What we decided and why...]

## Consequences

### Positive
- ...

### Negative
- ...

## References
- ...
```

#### /design-doc

**Purpose**: Create a design document

**Used by**: architect agent during `design` phase

**Parameters**: Feature or system name

**Example Invocation**:
```
/design-doc "Authentication System"
```

**Actions**:
1. Create file: `.sow/knowledge/architecture/authentication-system.md`
2. Use design doc template:
   - Overview
   - Goals and Non-Goals
   - System Architecture
   - API Design
   - Data Models
   - Security Considerations
   - Performance Considerations
   - Testing Strategy
   - Open Questions
3. Fill in based on project requirements
4. Commit to git

---

### Implementer Skills

#### /implement-feature

**Purpose**: Implement a new feature using Test-Driven Development (TDD)

**Used by**: implementer agent during `implement` phase

**Parameters**: Feature description and requirements

**Example Invocation**:
```
/implement-feature "JWT token generation service"
```

**TDD Workflow**:
1. **Write failing test first**:
   - Create test file
   - Write test cases covering requirements
   - Run tests (should fail)
2. **Implement minimum code to pass**:
   - Create implementation file
   - Write minimal code to satisfy tests
   - Run tests (should pass)
3. **Refactor**:
   - Clean up code
   - Remove duplication
   - Improve design
   - Tests still pass
4. **Log actions** via CLI:
   ```bash
   sow log --file src/auth/jwt.py --action created_file --result success "Implemented JWT service with TDD"
   ```
5. **Update task state**: Mark files modified

**Acceptance Criteria**:
- Unit test coverage >90%
- All tests passing
- Code follows style guide (from sinks)
- Logged all actions

#### /fix-bug

**Purpose**: Fix a bug using test-first approach

**Used by**: implementer agent during `implement` or `review` phases

**Parameters**: Bug description and reproduction steps

**Example Invocation**:
```
/fix-bug "JWT tokens expire immediately instead of after 1 hour"
```

**Bug Fix Workflow**:
1. **Write failing test that reproduces bug**:
   - Create or modify test file
   - Test should fail (confirming bug exists)
2. **Fix the bug**:
   - Modify implementation
   - Test should now pass
3. **Add regression test**:
   - Ensure bug won't reoccur
4. **Log actions**:
   ```bash
   sow log --file src/auth/jwt.py --action modified_file --result success "Fixed token expiration bug"
   ```
5. **Update task state**

---

### Integration Tester Skills

#### /write-integration-tests

**Purpose**: Write integration tests for cross-component interactions or E2E user flows

**Used by**: integration-tester agent during `test` phase

**Parameters**: Components or flows to test

**Example Invocation**:
```
/write-integration-tests "User authentication flow"
```

**Test Types**:

1. **Integration Tests** (multiple components):
   - Database + service interactions
   - API + database interactions
   - Service-to-service communication

2. **End-to-End Tests** (full user flows):
   - User registration → login → access protected resource
   - Complete user journeys
   - UI + API + database

**Workflow**:
1. **Identify test scenarios**:
   - Happy path
   - Error cases
   - Edge cases
2. **Set up test environment**:
   - Test database
   - Mock external services
   - Test fixtures
3. **Write tests**:
   - Use appropriate test framework
   - Clear test names
   - Arrange-Act-Assert pattern
4. **Run tests**:
   - All tests should pass
5. **Log actions**:
   ```bash
   sow log --file tests/integration/test_auth_flow.py --action created_file --result success "Added integration tests for auth flow"
   ```

**Note**: Unit tests are written by implementer during feature development, not by integration-tester.

---

### Reviewer Skills

#### /review-code

**Purpose**: Review code for quality, security, and adherence to standards

**Used by**: reviewer agent during `review` phase

**Parameters**: Files or components to review

**Example Invocation**:
```
/review-code "src/auth/"
```

**Review Checklist**:

1. **Code Quality**:
   - Clear, readable code
   - Proper naming conventions
   - Appropriate abstractions
   - No code duplication
   - Follows SOLID principles

2. **Security**:
   - Input validation
   - Output encoding
   - Authentication/authorization
   - Sensitive data handling
   - Dependency vulnerabilities

3. **Standards Adherence**:
   - Style guide compliance (from sinks)
   - API conventions (from sinks)
   - Architecture patterns
   - Project conventions

4. **Testing**:
   - Adequate test coverage
   - Test quality
   - Edge cases covered

5. **Performance**:
   - No obvious bottlenecks
   - Efficient algorithms
   - Appropriate data structures

**Workflow**:
1. Read files to review
2. Read relevant sinks (style guides, conventions)
3. Analyze code against checklist
4. Create feedback document with findings
5. Suggest refactoring if needed
6. Log review results

**Output**:
- Feedback document with categorized findings
- Priority levels (critical, important, nice-to-have)
- Specific recommendations
- Code examples for improvements

---

### Documenter Skills

#### /update-docs

**Purpose**: Update documentation to reflect code changes

**Used by**: documenter agent during `document` phase

**Parameters**: What changed and what docs to update

**Example Invocation**:
```
/update-docs "Added authentication endpoints to API"
```

**Documentation Types**:

1. **README Updates**:
   - Installation instructions
   - Getting started guide
   - Configuration options
   - Usage examples

2. **Inline Documentation**:
   - Docstrings/JSDoc
   - Function documentation
   - Class documentation
   - Module documentation

3. **API Documentation**:
   - Endpoint descriptions
   - Request/response schemas
   - Authentication requirements
   - Error responses

4. **Architecture Documentation**:
   - System diagrams
   - Component interactions
   - Data flows
   - Design decisions

**Workflow**:
1. Identify what changed (from task context)
2. Read existing documentation
3. Determine what needs updating
4. Update or create documentation
5. Ensure consistency across docs
6. Commit changes
7. Log actions

---

## Command File Format

All commands (workflows and skills) are markdown files with optional YAML frontmatter.

### Basic Format

```markdown
---
description: Short description of what this command does
allowed-tools: Bash(git add:*), Bash(git commit:*)
model: inherit
---

Command instructions go here.

Can reference arguments with $ARGUMENTS or $ARG1, $ARG2, etc.

Can use markdown formatting:
- Lists
- **Bold**
- `code`
- [links](https://example.com)
```

### Frontmatter Options

| Field | Purpose | Example |
|-------|---------|---------|
| `description` | Command description (shown in help) | "Create a git commit" |
| `allowed-tools` | Restrict tools available | "Bash(git add:\*), Bash(git commit:\*)" |
| `argument-hint` | Hint for command arguments | "<message>" |
| `model` | Specific model to use | "sonnet", "opus", "haiku", "inherit" |
| `invoke-model` | When to invoke model | "always", "if-needed" |

### Variable Substitution

- `$ARGUMENTS` - All arguments as single string
- `$ARG1`, `$ARG2`, etc. - Individual arguments
- Can use in instructions

### Example: /create-adr

```markdown
---
description: Create an Architecture Decision Record
allowed-tools: Read, Write, Edit, Grep, Glob, Bash(git add:*), Bash(git commit:*)
argument-hint: "<decision-title>"
model: inherit
---

Create an Architecture Decision Record (ADR) for: $ARGUMENTS

## Steps

1. Determine the next ADR number by checking `.sow/knowledge/adrs/`
2. Create file: `.sow/knowledge/adrs/XXX-<slugified-title>.md`
3. Use the ADR template:
   - Title
   - Status (proposed, accepted, rejected, deprecated, superseded)
   - Date
   - Context (situation and forces at play)
   - Decision (what we decided and why)
   - Consequences (positive and negative impacts)
   - References
4. Fill in based on current conversation context
5. Commit to git: `git add .sow/knowledge/adrs/XXX-*.md && git commit -m "docs: add ADR XXX"`

## ADR Template

```markdown
# X. [Title]

**Status**: [proposed | accepted | rejected | deprecated | superseded]

**Date**: YYYY-MM-DD

## Context

[Describe the context and forces at play...]

## Decision

[What we decided and why...]

## Consequences

### Positive
- ...

### Negative
- ...

## References
- ...
```

Be thorough and capture all relevant context.
```

---

## Smart UX Flows

`sow` commands include intelligent error handling and helpful prompts.

### Branch Protection

When trying to create project on main branch:

```
/start-project "Add feature"

Error: Cannot create project on main/master branch

Would you like to create a new feature branch?
  Branch name suggestion: feat/add-feature

Options:
1. Create suggested branch (feat/add-feature)
2. Specify custom branch name
3. Cancel

[1/2/3]:
```

### Project Conflict Resolution

When project already exists:

```
/start-project "New feature"

A project already exists: 'Add authentication'

Options:
1. Continue with existing project (/continue)
2. Create new branch for fresh project
3. Clean up existing project first (/cleanup)

[1/2/3]:
```

### Migration Prompts

When version mismatch detected:

```
SessionStart:

⚠️  Version mismatch detected!
   Repository structure: 0.1.0
   Plugin version: 0.2.0

💡 Run /migrate to upgrade your repository structure

   Migration path: 0.1.0 → 0.2.0
   Review changes: https://github.com/your-org/sow/blob/main/CHANGELOG.md#020

[Do you want to migrate now? y/n]:
```

### Incomplete Task Warning

When cleaning up with incomplete tasks:

```
/cleanup

⚠️ Warning: 2 tasks are still pending
  → implement/030: Add login endpoint
  → test/010: Write integration tests

Are you sure you want to clean up? [y/n]

If no: Consider completing or abandoning these tasks first
If yes: Project state will be deleted (work is committed to git, but task tracking is lost)
```

---

## Creating Custom Commands

Teams can add custom commands for their specific workflows.

### Step 1: Create Command File

```bash
# Create new workflow command
touch .claude/commands/workflows/deploy-staging.md

# Or create new skill
touch .claude/commands/skills/implementer/create-migration.md
```

### Step 2: Write Command

```markdown
---
description: Deploy to staging environment
allowed-tools: Bash(kubectl:*), Bash(helm:*)
---

Deploy the current branch to staging environment.

## Steps

1. Ensure all tests pass
2. Build Docker image
3. Push to container registry
4. Deploy to staging with helm
5. Run smoke tests
6. Report deployment status

[Detailed instructions...]
```

### Step 3: Use Command

```
# User invokes directly
/deploy-staging

# Or agent references in system prompt
When deployment is needed, use /deploy-staging skill
```

### Best Practices

1. **Clear descriptions** - Help users understand when to use it
2. **Restrict tools** - Limit tools to only what's needed
3. **Document steps** - Make the process transparent
4. **Handle errors** - Provide helpful error messages
5. **Test thoroughly** - Ensure command works in isolation

---

## Related Documentation

- **[AGENTS.md](./AGENTS.md)** - Agent system and how agents use skills
- **[HOOKS_AND_INTEGRATIONS.md](./HOOKS_AND_INTEGRATIONS.md)** - Event automation and integrations
- **[USER_GUIDE.md](./USER_GUIDE.md)** - Day-to-day workflows using commands
- **[PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md)** - Project lifecycle and task management
- **[DISTRIBUTION.md](./DISTRIBUTION.md)** - Plugin packaging and distribution
- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - CLI command reference
- **[../research/CLAUDE_CODE.md](../research/CLAUDE_CODE.md)** - Claude Code slash command documentation
