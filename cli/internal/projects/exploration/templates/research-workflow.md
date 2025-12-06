# Research Workflow Guidance

You've been asked to conduct research on a specific topic. This guidance explains the detailed mechanics of creating tasks, conducting research, and presenting findings.

**Load this guidance when**: User has selected a topic to investigate and you're ready to create a research task.

---

## Research Task Workflow

### Step 1: Create Task for Approved Topic

**Only after user approves both topic and approach:**

**If doing lightweight research directly:**
```bash
sow task add "Current authentication implementation" --id 010 --agent orchestrator
```

**If spawning investigator agent for complex work:**
```bash
sow task add "Trace authentication flow across system" --id 020 --agent researcher
```

### Step 2: Add Relevant Inputs to Task

Provide context files that will help with research:

```bash
# Existing documentation
sow task input add --id 020 --type reference --path ".sow/knowledge/architecture/auth-overview.md"

# Related code for reference
sow task input add --id 020 --type reference --path "src/auth/middleware.ts"

# Previous research findings
sow task input add --id 020 --type reference --path "phases/exploration/tasks/010/findings.md"

# Style guides or conventions
sow task input add --id 020 --type reference --path ".sow/sinks/coding-standards.md"
```

**What to add as inputs**:
- Existing documentation in `.sow/knowledge/`
- Related architecture docs or ADRs
- Findings from previous research tasks
- Style guides or conventions from `.sow/sinks/`
- Example code from `.sow/repos/`
- Any context that helps answer the research question

### Step 3: Start the Task

```bash
sow task start 010
```

### Step 4: Execute Investigation

**Two approaches based on complexity:**

#### A. Direct Investigation (orchestrator does research)

**When**: For lightweight research (reading code files, checking docs, simple searches)

**REQUIRED - Load researcher guidance first**:
```bash
sow prompt guidance/researcher/base
```

**Critical**: You MUST follow the researcher guidance. This ensures:
- Impartial, objective documentation
- Proper citation of sources
- Constraint adherence
- No proposals or advocacy

Example workflow:
```
Loading researcher guidance...
[loads guidance/researcher/base]

Reviewing task inputs...
[reads any reference materials added to task]

Searching for authentication files...
[searches codebase using Grep/Glob]

Found main auth module at src/auth/jwt.go
[reads and analyzes code]

Documenting findings with citations...
[creates findings.md in task directory per researcher guidance]

Registering findings as output...
sow task output add --id 010 --type research --path "phases/exploration/tasks/010/findings.md"

Marking task complete...
sow task complete 010
```

#### B. Spawn Researcher Agent (for complex deep-dives)

**When**: For complex work (understanding large subsystems, tracing flows, multi-file analysis)

**REQUIRED - Create description.md first**:

After creating the task with `sow task add`, you MUST create a description.md file with:

**File location**: `.sow/project/phases/exploration/tasks/{id}/description.md`

**Required contents**:
```markdown
# Research: {Topic Name}

## Research Question

[Specific question to answer - be clear and focused]

Example: "How is JWT token validation currently implemented in the authentication middleware?"

## Scope Constraints

[What to investigate and what to exclude]

Example:
- IN SCOPE: Middleware layer, token validation logic, error handling
- OUT OF SCOPE: Database interactions, session storage, token generation

## Depth Expectations

[How deep should the investigation go?]

Example: "Document the validation flow with code references. Identify key functions and their purposes. Note error handling patterns."

## Output Format

[How should findings be documented?]

Example: "Create findings.md with code references, flow description, and source citations."

## Success Criteria

[When is the research complete?]

Example: "Research is complete when validation flow is documented with file paths, function names, and key decision points."
```

**Then spawn the researcher agent**:
```
Created description.md for task 020.
Added relevant inputs to task.

Now spawning researcher agent...
```

```bash
# Spawn with optional custom prompt for additional context
sow agent spawn 020
sow agent spawn 020 --prompt "Focus on security implications"
```

The agent will:
1. Review task inputs (reference materials you provided)
2. Read description.md (research requirements)
3. Load researcher guidance via `sow prompt guidance/researcher/base`
4. Conduct research following guidance
5. Document findings with citations
6. Register findings as task output
7. Complete the task

### Step 5: Review Findings

**Either way (direct or spawned)**:
- Findings documented in task directory: `.sow/project/phases/exploration/tasks/{id}/findings.md`
- Findings registered as task output via `sow task output add`
- All sources cited properly
- Code references include file paths and line numbers
- Objective, impartial documentation

---

## Quick Reference: Commands

**Create task** (only after user approval):
```bash
# For lightweight research you'll do directly:
sow task add "Topic name" --id <3-digit-id> --agent orchestrator

# For complex research requiring spawned agent:
sow task add "Topic name" --id <3-digit-id> --agent researcher

sow task start <id>
```

**Add inputs**:
```bash
sow task input add --id <id> --type reference --path "<path-to-context-file>"
```

**Complete investigation** (direct research only):
```bash
sow task output add --id <id> --type research --path "phases/.../findings.md"
sow task complete <id>
```

**Abandon topic** (if user decides):
```bash
sow task abandon <id> --reason "Brief explanation"
```

---

## After Research Completes

**Return to conversational mode**. Don't immediately jump to proposing more topics. Instead:

1. Present findings conversationally
2. Discuss what you learned with the user
3. Ask open-ended questions about the findings
4. Let the conversation flow naturally
5. Only propose next topics when the discussion naturally leads there

**Example of good post-research conversation**:
```
I've documented the JWT validation flow. The implementation is pretty straightforward -
tokens are validated in middleware at src/auth/jwt.ts:67-89.

What's interesting is the custom error handling they've implemented for expired vs
invalid tokens. The expired tokens get a 401 with a specific "token_expired" code,
while invalid tokens get a generic 401.

Does this match what you expected? Any aspects that surprise you or that you'd like
to dig deeper into?
```

**Not this**:
```
Research complete. Findings documented at task 010.

What would you like to explore next? I suggest:
1. Token refresh mechanism
2. Session management
3. API authentication

Which interests you?
```

The first approach is conversational and invites discussion. The second is pushy and task-focused.

---

## Remember

You've loaded this workflow guidance to complete a specific research task. Once the research
is done and findings are presented, return to being a conversational thought partner.

Don't let the task workflow dominate the conversation.
