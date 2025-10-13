# User Guide

**Last Updated**: 2025-10-12
**Status**: Comprehensive Architecture Documentation

---

## Table of Contents

- [Getting Started](#getting-started)
  - [Installation](#installation)
  - [First-Time Setup](#first-time-setup)
  - [Verifying Installation](#verifying-installation)
- [Starting a New Project](#starting-a-new-project)
  - [Creating a Feature Branch](#creating-a-feature-branch)
  - [Using /start-project](#using-start-project)
  - [Understanding Initial Planning](#understanding-initial-planning)
- [Working with Projects](#working-with-projects)
  - [Continuing Work](#continuing-work)
  - [Pausing and Resuming](#pausing-and-resuming)
  - [Working Across Sessions](#working-across-sessions)
  - [Switching Between Branches](#switching-between-branches)
- [Understanding Project Status](#understanding-project-status)
  - [Reading state.yaml](#reading-stateyaml)
  - [Interpreting Task States](#interpreting-task-states)
  - [Reviewing Logs](#reviewing-logs)
- [Providing Feedback to Agents](#providing-feedback-to-agents)
  - [When to Provide Feedback](#when-to-provide-feedback)
  - [How Feedback is Stored](#how-feedback-is-stored)
  - [How Agents Use Feedback](#how-agents-use-feedback)
- [Completing and Cleaning Up](#completing-and-cleaning-up)
  - [When to Run /cleanup](#when-to-run-cleanup)
  - [Cleanup Process](#cleanup-process)
  - [Creating Pull Requests](#creating-pull-requests)
- [Working with Sinks](#working-with-sinks)
  - [What Are Sinks?](#what-are-sinks)
  - [Installing Sinks](#installing-sinks)
  - [Updating Sinks](#updating-sinks)
  - [Managing Sinks](#managing-sinks)
- [Working with Linked Repos](#working-with-linked-repos)
  - [Why Link Repositories?](#why-link-repositories)
  - [Adding Linked Repos](#adding-linked-repos)
  - [Syncing Linked Repos](#syncing-linked-repos)
- [One-Off Tasks](#one-off-tasks)
  - [When to Use vs Projects](#when-to-use-vs-projects)
  - [Orchestrator Direct Mode](#orchestrator-direct-mode)
- [Best Practices](#best-practices)
  - [Project Scope](#project-scope)
  - [Branch Management](#branch-management)
  - [Commit Hygiene](#commit-hygiene)
  - [Team Collaboration](#team-collaboration)
- [Troubleshooting](#troubleshooting)
  - [Common Issues](#common-issues)
  - [Recovery Procedures](#recovery-procedures)
  - [Getting Help](#getting-help)
- [Related Documentation](#related-documentation)

---

## Getting Started

### Installation

#### Prerequisites

- **Claude Code** installed and working
- **Git repository** (existing or new)
- **Optional**: CLI tools for enhanced functionality

#### Step 1: Install Plugin

Via marketplace:
```bash
# Add sow marketplace (one-time)
/plugin marketplace add your-org/sow-marketplace

# Install plugin
/plugin install sow@sow-marketplace
```

Via git URL:
```bash
/plugin install https://github.com/your-org/sow
```

#### Step 2: Restart Claude Code

```bash
exit
claude
```

Changes take effect after restart.

#### Step 3: Verify SessionStart Hook

When Claude Code starts, you should see:

```
⚠️  Not a sow repository
💡 Use /init to set up sow
```

This confirms the plugin is installed correctly.

### First-Time Setup

#### Initialize Your Repository

```
/init
```

**What Happens**:
1. Creates `.sow/` directory structure
2. Creates `.sow/knowledge/overview.md` template
3. Creates empty indexes for sinks and repos
4. Adds git ignore rules for `.sow/sinks/` and `.sow/repos/`
5. Commits structure to git
6. Offers optional CLI installation

**Success Output**:
```
✓ Checking prerequisites...
✓ Creating .sow/ structure...
  - .sow/knowledge/ (with overview.md template)
  - .sow/sinks/ (with index.json)
  - .sow/repos/ (with index.json)
  - .sow/.version (tracking version 0.2.0)

✓ Creating .gitignore entries...
  - .sow/sinks/
  - .sow/repos/

✓ Committing structure to git...
  [main abc1234] Initialize sow (v0.2.0)

🚀 Optional: Install sow CLI for enhanced functionality?
   The CLI provides fast operations like logging and sink management.

   Download for your platform:
   - macOS: https://github.com/your-org/sow/releases/download/v0.2.0/sow-macos
   - Linux: https://github.com/your-org/sow/releases/download/v0.2.0/sow-linux
   - Windows: https://github.com/your-org/sow/releases/download/v0.2.0/sow-windows.exe

   After download: mv ~/Downloads/sow-macos ~/.local/bin/sow && chmod +x ~/.local/bin/sow

   Test with: sow --version

[y to download now, n to skip, ? for more info]:
```

#### Optional: Install CLI

The CLI is **optional** but provides faster operations:

```bash
# macOS/Linux
curl -L https://github.com/your-org/sow/releases/download/v0.2.0/sow-macos -o sow
chmod +x sow
mv sow ~/.local/bin/sow

# Verify
sow --version
# sow 0.2.0
```

**CLI Benefits**:
- Fast logging (used by agents)
- Sink management
- Repository management
- Validation commands

### Verifying Installation

After initialization, restart Claude Code:

```bash
exit
claude
```

You should see:

```
📋 You are in a sow-enabled repository
💡 No active project. Use /start-project <name> to begin

✓ Versions aligned (v0.2.0)

📖 Available commands:
   /start-project <name> - Create new project
   /continue - Resume existing project
   /cleanup - Delete project before merge
   /sync - Update sinks and repos
```

---

## Starting a New Project

### Creating a Feature Branch

**Important**: Always create a feature branch before starting a project. `sow` enforces **one project per branch**.

```bash
# Create and switch to feature branch
git checkout -b feat/add-authentication
```

**Branch Naming Conventions**:
- `feat/` - New features
- `fix/` - Bug fixes
- `refactor/` - Code refactoring
- `docs/` - Documentation only

**Why Required?**:
- Project state is committed to feature branch
- Main branch stays clean (no project state)
- CI enforces: no `.sow/project/` on main
- Natural cleanup via branch deletion

### Using /start-project

```
/start-project "Add authentication"
```

**Smart Branch Detection**:

If you're still on main/master:
```
Error: Cannot create project on main/master branch

Would you like to create a new feature branch?
  Branch name suggestion: feat/add-authentication

Options:
1. Create suggested branch (feat/add-authentication)
2. Specify custom branch name
3. Cancel

[1/2/3]:
```

If you already have a project:
```
A project already exists: 'Previous feature'

Options:
1. Continue with existing project (/continue)
2. Create new branch for fresh project
3. Clean up existing project first (/cleanup)

[1/2/3]:
```

### Understanding Initial Planning

After `/start-project`, the orchestrator:

1. **Prompts for Description**:
   ```
   Describe what you want to build:
   ```

   Provide clear requirements:
   ```
   Add JWT-based authentication with user login, token refresh,
   and password hashing. Should integrate with existing User model.
   ```

2. **Assesses Complexity**:
   ```
   🔍 Analyzing requirements...

   Complexity: Moderate (rating: 2/3)

   Estimated scope:
   - Files to modify: 8-12
   - Cross-cutting concerns: Yes (auth middleware)
   - New dependencies: Yes (JWT library, bcrypt)
   ```

3. **Selects Initial Phases**:
   ```
   📋 Proposed plan:

   Phases:
   - design (1 task)
   - implement (4 tasks)

   Note: Starting with minimal structure. Additional phases
   (test, review, document) can be added as needed.

   Approve? [y/n/modify]
   ```

4. **Creates Initial Tasks**:
   ```
   ✓ Created project 'Add authentication'

   Phase: design
   - 010: Design authentication flow [architect]

   Phase: implement
   - 010: Create User model [implementer]
   - 020: Create JWT service [implementer]
   - 030: Add login endpoint [implementer]
   - 040: Add password hashing utility [implementer]

   ✓ Committed to git: [feat/add-auth abc1234] Initialize project 'Add authentication'

   Starting with design phase...
   ```

**Progressive Planning**:
- Don't worry about planning everything upfront
- Orchestrator will request adding phases as needed
- You'll approve phase changes along the way

---

## Working with Projects

### Continuing Work

```
/continue
```

**What Happens**:
1. Reads `.sow/project/state.yaml`
2. Verifies branch matches project
3. Shows current status
4. Identifies next pending task
5. Spawns appropriate worker

**Example Output**:
```
📋 Project: Add authentication (branch: feat/add-auth)

Progress:
  ✓ design (1/1 tasks complete)
  → implement (2/4 tasks complete)

Completed:
  ✓ design/010: Design authentication flow [architect]
  ✓ implement/010: Create User model [implementer]
  ✓ implement/020: Create JWT service [implementer]

Current:
  → implement/030: Add login endpoint [implementer] - IN PROGRESS

Next:
  ⬜ implement/040: Add password hashing utility [implementer]

Resuming task 030...
Spawning implementer worker...
```

### Pausing and Resuming

**Pausing Work**:
- Simply exit Claude Code (Ctrl+D or `exit`)
- All state is saved in `.sow/project/`
- No special command needed

**Resuming Work**:
- Start Claude Code in same repository
- Run `/continue`
- Orchestrator picks up exactly where you left off

**Zero-Context Resumability**:
- No conversation history needed
- Orchestrator reads state and logs
- Workers read task context and feedback
- Can resume even after weeks

### Working Across Sessions

**Scenario**: You work on Monday, pause, resume on Friday

**Monday**:
```bash
# Start work
git checkout feat/add-auth
/continue

# Work happens...
# Task 030 is in progress

# End of day - just exit
exit
```

**Friday**:
```bash
# Resume
git checkout feat/add-auth
claude

# SessionStart shows context
📋 You are in a sow-enabled repository
🚀 Active project: Add authentication (branch: feat/add-auth)
📂 Use /continue to resume work

# Continue where you left off
/continue

# Orchestrator reads state and resumes task 030
```

**What Gets Preserved**:
- Project state (phases, tasks, status)
- Task descriptions and requirements
- Context references (sinks, docs, files)
- Action logs (what's been tried)
- Feedback (your corrections)

**What's NOT Needed**:
- Conversation history
- Previous agent memory
- User re-explanation

### Switching Between Branches

**Key Feature**: Project state switches automatically with branches

```bash
# Work on feature A
git checkout feat/feature-a
/continue
# Works on "Feature A" project

# Switch to feature B
git checkout feat/feature-b
/continue
# Works on "Feature B" project

# Switch back to feature A
git checkout feat/feature-a
/continue
# Back to "Feature A" project - exact same state
```

**How It Works**:
- `.sow/project/` is committed to each feature branch
- Git handles switching project state automatically
- Each branch has its own independent project

**Benefits**:
- Natural branch-based workflow
- No manual project switching
- Clean separation of work
- Team can share branch with project state

---

## Understanding Project Status

### Reading state.yaml

View current project state:

```bash
cat .sow/project/state.yaml
```

**Example**:
```yaml
project:
  name: add-authentication
  branch: feat/add-auth
  created_at: 2025-10-12T14:30:00Z
  updated_at: 2025-10-12T16:45:00Z
  description: Add JWT-based authentication system

  complexity:
    rating: 2
    metrics:
      estimated_files: 8
      cross_cutting: true
      new_dependencies: true

  active_phase: implement

phases:
  - name: design
    status: completed
    created_at: 2025-10-12T14:32:00Z
    completed_at: 2025-10-12T15:20:00Z
    tasks:
      - id: "010"
        name: Design authentication flow
        status: completed
        parallel: false

  - name: implement
    status: in_progress
    created_at: 2025-10-12T15:22:00Z
    tasks:
      - id: "010"
        name: Create User model
        status: completed
        parallel: false

      - id: "020"
        name: Create JWT service
        status: in_progress
        parallel: false

      - id: "030"
        name: Add login endpoint
        status: pending
        parallel: true

      - id: "031"
        name: Add password hashing utility
        status: pending
        parallel: true
```

**Key Fields**:
- `active_phase` - Current phase being worked on
- `phases[].status` - Phase status (pending, in_progress, completed)
- `phases[].tasks[].status` - Task status
- `phases[].tasks[].parallel` - Can run in parallel with other tasks

### Interpreting Task States

| Status | Meaning |
|--------|---------|
| `pending` | Not yet started |
| `in_progress` | Currently being worked on |
| `completed` | Successfully finished |
| `abandoned` | Started but no longer relevant |

**Gap Numbering**:
- Tasks numbered: 010, 020, 030, 040
- Allows insertions: 011, 012, 021, 031
- No renumbering needed when adding tasks

### Reviewing Logs

**Project Log** (orchestrator actions):
```bash
cat .sow/project/log.md
```

**Task Log** (worker actions):
```bash
cat .sow/project/phases/implement/tasks/020/log.md
```

**Example Task Log**:
```markdown
# Task Log

Chronological record of all actions taken during this task.

---

### 2025-10-12 15:30:42

**Agent**: implementer-3
**Action**: started_task
**Result**: success

Started implementing JWT service. Reviewed requirements and acceptance criteria.

---

### 2025-10-12 15:35:18

**Agent**: implementer-3
**Action**: created_file
**Files**:
  - src/auth/jwt.py
**Result**: success

Created initial JWT service file with class structure and method stubs.

---

### 2025-10-12 15:42:03

**Agent**: implementer-3
**Action**: implementation_attempt
**Files**:
  - src/auth/jwt.py
**Result**: error

Attempted to implement token generation. Encountered import error with cryptography library.

---
```

**Use Cases**:
- Understand what's been tried
- Debug issues
- Review progress
- Audit agent actions

---

## Providing Feedback to Agents

### When to Provide Feedback

Provide feedback when:
- Agent made an incorrect assumption
- Implementation doesn't meet requirements
- Code needs specific changes
- Agent got stuck on wrong approach

**Example Scenarios**:
- "The JWT service should use RS256, not HS256"
- "Token expiration should be configurable, not hardcoded"
- "Need to add user role to token payload"
- "This approach won't work with our existing auth middleware"

### How Feedback is Stored

Feedback is stored in chronologically numbered files:

```
.sow/project/phases/implement/tasks/020/feedback/
├── 001.md
├── 002.md
└── 003.md
```

**Process**:
1. You explain issue to orchestrator
2. Orchestrator creates `feedback/00X.md` with your notes
3. Orchestrator increments task iteration counter
4. Orchestrator spawns new worker with feedback context

**Example Feedback File**:
```markdown
# Feedback 001

**Created**: 2025-10-12 16:30:00
**Status**: pending

## Issue

The JWT service is using HS256 (symmetric) algorithm, but we need RS256
(asymmetric) for better security and key management.

## Required Changes

1. Change algorithm from HS256 to RS256
2. Update to use public/private key pair instead of shared secret
3. Add key loading from environment variables
4. Update tests accordingly

## Context

Our infrastructure uses asymmetric keys for all JWT operations.
Keys are managed via Vault and loaded at startup.

Reference: `.sow/sinks/security-guidelines/jwt-best-practices.md`
```

### How Agents Use Feedback

**Worker Startup Process**:
1. Read `state.yaml` (sees iteration=4, meaning 4th attempt)
2. Read `description.md` (original requirements)
3. Read `feedback/` directory (all feedback files in order)
4. Identify pending feedback items
5. Incorporate feedback into work
6. Mark feedback as addressed in `state.yaml`

**Feedback Tracking in state.yaml**:
```yaml
task:
  id: "020"
  iteration: 4
  feedback:
    - id: "001"
      created_at: 2025-10-12T16:30:00Z
      status: addressed
    - id: "002"
      created_at: 2025-10-12T17:15:00Z
      status: pending
```

**Example Conversation**:
```
You: "The JWT tokens are expiring immediately instead of after 1 hour"

Orchestrator: "I'll create feedback for the implementer and have them fix this.
               Creating feedback/002.md and restarting the task..."

Implementer (worker):
- Reads feedback/002.md
- Sees: Token expiration issue
- Checks current implementation
- Fixes the bug (expiration was set to 0 instead of 3600)
- Writes test to prevent regression
- Marks feedback/002 as addressed
- Reports completion

Orchestrator: "Task 020 complete. The token expiration bug has been fixed
               and a test added to prevent regression."
```

---

## Completing and Cleaning Up

### When to Run /cleanup

Run `/cleanup` when:
- All project tasks are complete
- Ready to create pull request
- About to merge to main branch

**Do NOT run cleanup**:
- While work is still in progress
- Before pushing final commits
- If you want to preserve project state

### Cleanup Process

```
/cleanup
```

**Validation**:
```
⚠️ Warning: 1 task is still pending
  → implement/040: Add password hashing utility

Are you sure you want to clean up? [y/n]
```

**If Confirmed**:
```
✓ Deleted .sow/project/
✓ Staged deletion for commit

Next steps:
1. Commit cleanup: git commit -m "chore: cleanup sow project state"
2. Push and create PR
3. CI will verify no .sow/project/ exists before merge

Note: Recommend squash-merge to keep main branch history clean
```

**Git Commands**:
```bash
# Cleanup stages the deletion
git status
# deleted:    .sow/project/

# Commit the cleanup
git commit -m "chore: cleanup sow project state"

# Push to remote
git push origin feat/add-auth
```

### Creating Pull Requests

**After Cleanup**:

1. **Push Branch**:
   ```bash
   git push origin feat/add-auth
   ```

2. **Create PR** (via GitHub UI or CLI):
   ```bash
   gh pr create --title "Add authentication" --body "..."
   ```

3. **CI Validation**:
   - CI should check that `.sow/project/` does not exist
   - If it exists, CI fails (prompts you to run `/cleanup`)

4. **Review and Merge**:
   - Code review proceeds normally
   - **Recommend squash-merge** to keep main branch history clean
   - After merge, delete feature branch

**CI Check Example** (.github/workflows/ci.yml):
```yaml
- name: Ensure no active sow project
  run: |
    if [ -d ".sow/project" ]; then
      echo "Error: .sow/project/ must be removed before merging"
      echo "Run /cleanup or: rm -rf .sow/project"
      exit 1
    fi
```

---

## Working with Sinks

### What Are Sinks?

**Sinks** are collections of markdown files providing focused knowledge:

- **Style guides** (Python conventions, Go idioms)
- **API standards** (REST conventions, GraphQL patterns)
- **Deployment procedures** (Kubernetes configs, CI/CD)
- **Security guidelines** (Auth patterns, encryption)
- **Company policies** (Code review checklist, testing standards)

**Benefits**:
- Agents automatically reference relevant sinks
- Ensures consistency across projects
- Easy to update (just pull latest)
- Shareable across team

**Location**: `.sow/sinks/` (git-ignored)

### Installing Sinks

**Via CLI** (recommended):
```bash
# Install from git repository
sow sinks install https://github.com/your-org/python-style-guide

# Install from git with specific path
sow sinks install https://github.com/your-org/standards style-guides/python

# Install from local path
sow sinks install /path/to/local/sink
```

**Via Slash Command**:
```
/sync
# Then select sinks to install
```

**What Happens**:
1. Clones or copies sink to `.sow/sinks/<name>/`
2. Interrogates sink content (what topics covered)
3. Summarizes sink with "when to use" guidance
4. Updates `.sow/sinks/index.json` with metadata
5. Sink is now available to agents

**Example Sink Structure**:
```
.sow/sinks/
├── index.json                  # Auto-maintained by LLM
├── python-style/
│   ├── formatting.md
│   ├── conventions.md
│   └── testing.md
├── api-conventions/
│   ├── rest-standards.md
│   ├── error-handling.md
│   └── versioning.md
└── deployment-guide/
    ├── kubernetes.md
    └── ci-cd.md
```

**Index Example**:
```json
{
  "sinks": [
    {
      "name": "python-style",
      "path": "python-style",
      "description": "Python code style and conventions",
      "topics": ["formatting", "naming", "testing", "imports"],
      "when_to_use": "When writing or reviewing Python code",
      "version": "v1.2.0",
      "source": "https://github.com/your-org/python-style-guide"
    }
  ]
}
```

### Updating Sinks

**Check for Updates**:
```bash
sow sinks update
```

**What Happens**:
```
🔍 Checking for updates...

Sinks:
  ✓ python-style (v1.2.0) - up to date
  ⬆️ api-conventions (v2.3.0 → v2.4.0) - Added GraphQL conventions
  ⬆️ deployment-guide (v1.0.0 → v1.1.0) - Updated k8s configs

Update all? [y/n/selective]
```

**Selective Update**:
```
Select sinks to update:
  [y] api-conventions
  [n] deployment-guide

Updating...
✓ api-conventions updated to v2.4.0
✓ Regenerated index

1 sink updated, 2 unchanged
```

### Managing Sinks

**List Installed Sinks**:
```bash
sow sinks list
```

**Output**:
```
Installed sinks:

  python-style (v1.2.0)
  - Python code style and conventions
  - Topics: formatting, naming, testing, imports
  - Source: https://github.com/your-org/python-style-guide

  api-conventions (v2.4.0)
  - REST and GraphQL API standards
  - Topics: endpoints, errors, versioning, graphql
  - Source: https://github.com/your-org/api-standards

3 sinks installed
```

**Remove Sink**:
```bash
sow sinks remove python-style
```

**Regenerate Index**:
```bash
# If index gets corrupted or out of sync
sow sinks reindex
```

---

## Working with Linked Repos

### Why Link Repositories?

Link repositories when you need:
- **Cross-repo context** - Reference code from other services
- **Shared libraries** - See implementation examples
- **Multi-repo projects** - Coordinate changes across repos
- **Pseudo-monorepo** - Work with multiple repos as one

**Location**: `.sow/repos/` (git-ignored)

### Adding Linked Repos

**Via CLI**:
```bash
# Clone repository
sow repos add https://github.com/your-org/auth-service

# Symlink local repository
sow repos add /path/to/shared-library --symlink
```

**Via Slash Command**:
```
/sync
# Select repos to add
```

**What Happens**:
1. Clones (or symlinks) repository to `.sow/repos/<name>/`
2. Adds entry to `.sow/repos/index.json`
3. Repository is now available to agents

**Example Structure**:
```
.sow/repos/
├── index.json
├── auth-service/          # Cloned repo
│   ├── src/
│   └── tests/
└── shared-library/        # Symlinked repo
    ├── crypto/
    └── utils/
```

**Index Example**:
```json
{
  "repositories": [
    {
      "name": "auth-service",
      "path": "auth-service",
      "source": "https://github.com/your-org/auth-service",
      "purpose": "Reference auth implementation patterns",
      "type": "clone"
    },
    {
      "name": "shared-library",
      "path": "shared-library",
      "source": "/Users/you/code/shared-library",
      "purpose": "Shared cryptography and utility functions",
      "type": "symlink"
    }
  ]
}
```

### Syncing Linked Repos

**Update All Repos**:
```bash
sow repos sync
```

**What Happens**:
```
🔍 Syncing repositories...

Repositories:
  ✓ auth-service - up to date
  ⬆️ shared-library (abc1234 → def5678) - Added new crypto utilities

Update all? [y/n/selective]
```

**Agent Usage**:

When orchestrator assigns task:
```yaml
# Task state.yaml
references:
  - repos/shared-library/src/crypto/jwt.py
  - repos/auth-service/src/auth/middleware.py
```

Worker reads these files for context and patterns.

---

## One-Off Tasks

### When to Use vs Projects

**Use One-Off** (orchestrator direct):
- Simple changes (fix typo, update config)
- Quick refactoring (rename function)
- Small bug fixes (one-line change)
- Documentation updates

**Use Projects** (structured):
- New features (multiple files)
- Complex refactoring (architectural changes)
- Large bug fixes (requires investigation)
- Multi-step work

**Decision**: You decide, not the orchestrator

### Orchestrator Direct Mode

**Startup Behavior**:

When Claude Code starts with no project:
```
What would you like to work on?

You can describe:
- One-off task (I'll handle it directly)
- New project (structured, multi-task work)

Or use:
- /start-project <name> - Create structured project
```

**One-Off Example**:
```
You: "Fix the typo in README.md line 42 - 'authetication' should be 'authentication'"

Orchestrator:
- Reads README.md
- Makes the edit
- Done

No project structure created.
```

**When Orchestrator Suggests Project**:
```
You: "Add comprehensive authentication with JWT, OAuth, and multi-factor auth"

Orchestrator: "This looks like a complex feature requiring multiple steps.
               Would you like me to create a structured project? [y/n]"

You: y

Orchestrator: "Great! Let's use /start-project to plan this properly."
[Guides you through /start-project flow]
```

**Benefits of Orchestrator Direct Mode**:
- Fast for simple tasks
- No overhead
- Orchestrator still understands sow framework
- Can escalate to project if needed

---

## Best Practices

### Project Scope

**Do**:
- Keep projects short-lived (hours to days, not weeks)
- Scope to what fits in one feature branch
- Delete after merging (cleanup)
- Create new project for follow-up work

**Don't**:
- Create mega-projects spanning weeks
- Try to fit unrelated work in one project
- Keep project state after merge
- Extend completed projects

**Guideline**: If project feels too big, break into multiple feature branches.

### Branch Management

**Do**:
- Always create feature branch before `/start-project`
- Name branches descriptively (`feat/add-auth`, `fix/token-expiry`)
- Keep one project per branch (enforced)
- Clean up before merging

**Don't**:
- Create projects on main/master
- Switch branches mid-task without committing
- Merge with `.sow/project/` still present
- Reuse project state across branches

**Pattern**:
```bash
# Good workflow
git checkout -b feat/my-feature
/start-project "My feature"
# ... work ...
/cleanup
git commit -m "chore: cleanup sow project state"
git push
# Create PR, merge, delete branch

# Bad workflow
# Work on main branch
/start-project "Feature"  # ERROR: Can't create on main
```

### Commit Hygiene

**Do**:
- Commit `.sow/project/` regularly during work
- Push to remote (backup + collaboration)
- Write clear commit messages for feature work
- Run `/cleanup` before final merge

**Don't**:
- Git-ignore `.sow/project/` (breaks branch switching)
- Leave uncommitted project state
- Forget to cleanup before merge
- Include `.sow/project/` in main branch

**Commit Strategy**:
```bash
# During work (feature branch)
git add .sow/project/ src/
git commit -m "feat: implement JWT service"

# Before merge
/cleanup
git commit -m "chore: cleanup sow project state"

# Squash merge recommended
gh pr merge --squash
```

### Team Collaboration

**Do**:
- Commit `.claude/` configuration (shared agents/commands)
- Commit `.sow/knowledge/` (shared docs)
- Commit `.sow/project/` to feature branches (state sharing)
- Document custom sinks in README
- Share sink sources with team

**Don't**:
- Commit `.sow/sinks/` (personal installations)
- Commit `.sow/repos/` (too large)
- Modify `.claude/` structure without team discussion
- Change shared sinks without coordination

**Collaboration Pattern**:
```bash
# Developer A creates feature branch with project
git checkout -b feat/auth
/start-project "Add auth"
# ... some work ...
git push origin feat/auth

# Developer B picks up the work
git fetch
git checkout feat/auth
/continue  # Sees all context, continues work
```

---

## Troubleshooting

### Common Issues

#### Issue: Project won't start on current branch

**Symptoms**:
```
Error: Cannot create project on main/master branch
```

**Solution**:
```bash
# Create feature branch
git checkout -b feat/my-feature

# Then start project
/start-project "My feature"
```

#### Issue: Project already exists

**Symptoms**:
```
A project already exists: 'Old feature'
```

**Solutions**:

Option 1 - Continue existing:
```
/continue
```

Option 2 - Clean up and start new:
```
/cleanup
/start-project "New feature"
```

Option 3 - New branch:
```bash
git checkout -b feat/new-feature
/start-project "New feature"
```

#### Issue: Version mismatch after plugin upgrade

**Symptoms**:
```
⚠️  Version mismatch detected!
   Repository structure: 0.1.0
   Plugin version: 0.2.0
```

**Solution**:
```
/migrate
```

#### Issue: Lost project state after branch switch

**Symptoms**: Project state different than expected after switching branches

**Cause**: `.sow/project/` wasn't committed

**Solution**:
```bash
# Check git status
git status

# If .sow/project/ is untracked
git add .sow/project/
git commit -m "chore: commit project state"

# Now switching branches will work correctly
```

#### Issue: CLI commands not found

**Symptoms**:
```bash
sow: command not found
```

**Solution**:
```bash
# Verify CLI is installed
ls ~/.local/bin/sow

# If not there, install CLI
curl -L https://github.com/your-org/sow/releases/download/v0.2.0/sow-macos -o ~/.local/bin/sow
chmod +x ~/.local/bin/sow

# Ensure ~/.local/bin is in PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Verify
sow --version
```

### Recovery Procedures

#### Corrupted Project State

**Symptoms**: Project state file is malformed or corrupted

**Recovery**:
```bash
# Check if YAML is valid
yq .sow/project/state.yaml

# If corrupted, restore from git
git checkout HEAD~1 .sow/project/state.yaml

# Or start fresh
rm -rf .sow/project
/start-project "Feature name"
```

#### Lost Task Progress

**Symptoms**: Work done but not reflected in state

**Recovery**:
1. Check logs for what was actually done:
   ```bash
   cat .sow/project/phases/*/tasks/*/log.md
   ```

2. Manually update state.yaml if needed:
   ```bash
   # Mark task as completed
   yq -i '.phases[1].tasks[0].status = "completed"' .sow/project/state.yaml
   ```

3. Commit corrected state:
   ```bash
   git add .sow/project/state.yaml
   git commit -m "fix: correct task status"
   ```

#### Broken Index Files

**Symptoms**: Sinks or repos not being found by agents

**Recovery**:
```bash
# Regenerate sink index
sow sinks reindex

# Regenerate repo index
sow repos reindex

# Verify
cat .sow/sinks/index.json
cat .sow/repos/index.json
```

### Getting Help

1. **Check Documentation**: Start with relevant doc sections

2. **Validate Structure**:
   ```bash
   sow validate
   ```

3. **Check Logs**: Review project and task logs for errors

4. **GitHub Issues**: Report bugs with:
   - sow version (`sow --version`)
   - Plugin version (`.claude/.plugin-version`)
   - Error messages
   - Steps to reproduce

5. **Community**: Join discussions, ask questions

---

## Related Documentation

- **[OVERVIEW.md](./OVERVIEW.md)** - System overview and concepts
- **[COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md)** - All available commands
- **[PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md)** - Project lifecycle details
- **[HOOKS_AND_INTEGRATIONS.md](./HOOKS_AND_INTEGRATIONS.md)** - Customization and integrations
- **[DISTRIBUTION.md](./DISTRIBUTION.md)** - Installation and upgrades
- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - CLI command reference
- **[SCHEMAS.md](./SCHEMAS.md)** - File format specifications
