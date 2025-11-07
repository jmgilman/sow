---

## Discovery Phase Guidance

### Purpose

Gather codebase and design context to inform work unit identification. This ensures work units reference existing code and avoid duplicate work.

**CRITICAL**: DO NOT identify or create work units during Discovery. Work unit identification happens in the Active state after discovery is complete and approved.

### Your Approach

**First, propose your approach to the user:**

Example:
```
To break down this feature effectively, I should first understand the existing codebase.
I could either:
A) Do a quick exploration myself (suitable for familiar areas)
B) Spawn an explorer agent for a thorough investigation (better for complex/unfamiliar areas)

Which approach would work better?
```

**Wait for user to choose**, then proceed.

### Approach Options

**Option A: Orchestrator-led (simple breakdowns)**
- Suitable when: Breaking down small features or familiar code areas
- Process: Create task assigned to self, write discovery doc directly

**Option B: Explorer-led (complex breakdowns)**
- Suitable when: Large features, unfamiliar code, high risk of duplicates
- Process: Create task, spawn explorer agent to investigate codebase

### Discovery Document Contents

Create a discovery artifact that includes:

1. **Existing Code Context**
   - Relevant files, classes, functions that will be extended/modified
   - Third-party libraries already in use
   - Patterns and conventions to follow

2. **Existing Documentation**
   - ADRs that provide architectural decisions
   - Design docs that inform implementation approach
   - Exploratory findings from previous work

3. **Scope Boundaries**
   - What's in scope for this breakdown
   - What already exists and should be reused
   - What's explicitly out of scope

### Workflow

**Step 1: Propose and get approval**
```
Orchestrator: "Should I create a discovery task to explore [specific areas]?"
[Wait for user approval]
```

**Step 2: Create and execute task**
```bash
# Create discovery task
sow task add "Codebase Discovery" --id 001

# Option A: Do it yourself
sow task start 001
# Write project/discovery/analysis.md
sow output add --type discovery --path project/discovery/analysis.md
sow task complete 001

# Option B: Spawn explorer
# (spawn explorer agent with 001 task context)
# Explorer writes discovery doc and completes task
```

**Step 3: Present findings and wait for approval**
```
Orchestrator: "I've documented the existing auth system. Key findings:
- [summary of findings]

The full discovery document is at project/discovery/analysis.md.
Should I proceed to mark this as approved so we can start identifying work units?"
[Wait for user to review and approve]
```

**Step 4: Approve and advance** (only after user confirms)
```bash
# Approve discovery document
sow output approve discovery project/discovery/analysis.md

# Advance to Active state
sow project advance
```
