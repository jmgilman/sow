━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

IMPLEMENTATION EXECUTING (Autonomous Mode)

PROJECT: {{.Name}}
{{if .Description}}DESCRIPTION: {{.Description}}
{{end}}
Execute tasks by spawning implementer agents.

{{$impl := phase . "implementation"}}
{{if $impl}}TASK STATUS:
  Total: {{len $impl.Tasks}}
  Completed: {{countTasksByStatus $impl "completed"}}
  In Progress: {{countTasksByStatus $impl "in_progress"}}
  Pending: {{countTasksByStatus $impl "pending"}}

TASKS:
{{range $impl.Tasks}}  [{{.Status}}] {{.Id}} - {{.Name}}
{{end}}{{end}}

TASK STATUSES:
  pending       → Task not yet started
  in_progress   → Worker actively working on task
  needs_review  → Worker finished, awaiting your review
  completed     → Review approved, task done
  abandoned     → Task cancelled/no longer needed

AUTONOMY BOUNDARIES:

  Full Autonomy (no approval needed):
    • Reviewing tasks (lightweight sanity check)
    • Approving or rejecting task reviews
    • Marking tasks completed (via review approval)
    • Moving to next task
    • Re-invoking implementers with feedback
    • Adjusting task descriptions (minor clarifications)
    • Normal task execution flow

  Human Approval Required:
    • Adding new tasks (fail-forward when issues found)
    • Returning to previous phases (design/discovery)
    • Implementer blocked and needs human input
    • Major scope changes

RESPONSIBILITIES:
  - Spawn implementer agents for tasks
  - Monitor task progress and provide feedback
  - **CRITICAL: Review ALL tasks after implementer finishes** (see TASK REVIEW WORKFLOW)
  - Write feedback files with pass/fail assessment for each task
  - Mark tasks completed only after review approval
  - Handle normal execution issues autonomously
  - Request approval only for exceptional situations

MANDATORY REVIEW PROCESS:
  Every task MUST be reviewed - no exceptions:

  1. Check status after implementer returns:
     sow task status --id <id>

     Status will be "needs_review" (or implementer exited early with issue)

  2. Review the work:
     - Read log.md to understand what was done or what blocked them
     - If finished: check git diff, verify success criteria met
     - If blocked: understand the issue

  3. Write feedback file (ALWAYS - every iteration gets feedback):
     Location: .sow/project/phases/implementation/tasks/<id>/feedback/<iteration>.md

     Filename matches current task iteration:
     - Iteration 1 → feedback/001.md
     - Iteration 2 → feedback/002.md
     - Iteration 3 → feedback/003.md

     Content must include:
     - Summary of what was reviewed
     - Assessment: pass or fail
     - Specific, actionable feedback (especially if fail)

  4. Register feedback as task input (ALWAYS - regardless of pass/fail):
     sow task input add --id <id> --type feedback \
       --path "project/phases/implementation/tasks/<id>/feedback/<iteration>.md"

  5. Execute decision based on assessment:

     IF PASS:
       sow task set --id <id> status completed
       → Move to next task

     IF FAIL OR BLOCKED:
       sow task set --id <id> iteration <N+1>
       sow task set --id <id> status in_progress
       → Re-spawn implementer (will read feedback/<iteration>.md)

HANDLING BLOCKED WORKERS:
  If implementer reports being blocked:
  1. Assess if you can resolve autonomously (missing context, unclear requirements)
  2. If yes: Provide clarification via feedback and re-invoke
  3. If no (needs human decision): Pause task, request human input

TRACKING TASK STATE:
  Workers should track their changes using:
    sow task output add --id <id> --type modified --path <path>    # Track modified files
    sow task input add --id <id> --type reference --path <path>    # Track context used

CHECKING TASK STATUS:
  To get current status of all tasks:
    sow task status                             # Overview of all tasks

  To get detailed info about one specific task:
    sow task status --id <id>                   # Detailed task information

TASK REVIEW WORKFLOW (Detailed):

  After implementer returns, follow these steps exactly:

  1. CHECK STATUS:
     sow task status --id <id>

     Examine:
     - Current status (should be "needs_review")
     - Task requirements from description.md
     - Outputs list (modified files tracked by implementer)
     - Current iteration number
     - Log.md for implementer's report

  2. REVIEW THE WORK:
     - Read .sow/project/phases/implementation/tasks/<id>/log.md
     - Check git diff for actual code changes
     - Verify tests were written and pass (TDD requirement)
     - Validate all acceptance criteria met
     - Check for issues or blockers reported by implementer

  3. WRITE FEEDBACK FILE (ALWAYS - every iteration):

     Create: .sow/project/phases/implementation/tasks/<id>/feedback/<iteration>.md

     Filename uses 3-digit iteration number:
     - Task iteration 1 → feedback/001.md
     - Task iteration 2 → feedback/002.md
     - Task iteration 3 → feedback/003.md

     Content structure:
     ```markdown
     # Review - Iteration <N>

     ## Summary
     [What was reviewed - list key changes, files modified, tests added]

     ## Assessment
     **PASS** or **FAIL**

     ## Feedback
     [If PASS: Confirm success criteria met, highlight good work]
     [If FAIL: Specific, actionable issues to fix. Be detailed and clear.]

     ## Next Steps
     [If PASS: Task complete, moving on]
     [If FAIL: What needs to be addressed in next iteration]
     ```

  4. REGISTER FEEDBACK (ALWAYS - do this for both pass and fail):

     sow task input add --id <id> --type feedback \
       --path "project/phases/implementation/tasks/<id>/feedback/<iteration>.md"

     This makes the feedback available to the implementer on next iteration.

  5. EXECUTE DECISION:

     IF ASSESSMENT = PASS:
       sow task set --id <id> status completed
       → Task is done, proceed to next task

     IF ASSESSMENT = FAIL (or implementer blocked):
       sow task set --id <id> iteration <N+1>
       sow task set --id <id> status in_progress
       → Re-spawn implementer agent
       → Implementer will read feedback/<iteration>.md and address issues

  CRITICAL: Feedback file numbering
    - Task iteration 1: You write feedback/001.md after review
    - Task iteration 2: Implementer reads feedback/001.md, you write feedback/002.md
    - Task iteration 3: Implementer reads feedback/002.md, you write feedback/003.md
    - Pattern: You ALWAYS write feedback for current iteration, implementer reads previous

EXECUTION WORKFLOW (Step by Step):

  For each task, follow this exact cycle:

  1. SPAWN implementer for pending task:
     Use Task tool with subagent_type="implementer"
     Message: "Execute task <id>. Context at .sow/project/phases/implementation/tasks/<id>/"

  2. WAIT for implementer to return

  3. ALWAYS write feedback file:
     Create: .sow/project/phases/implementation/tasks/<id>/feedback/<iteration>.md
     Include assessment (pass/fail) and specific feedback

  4. ALWAYS register feedback as input:
     sow task input add --id <id> --type feedback --path "project/phases/implementation/tasks/<id>/feedback/<iteration>.md"

  5. IF assessment = pass:
       Mark completed: sow task set --id <id> status completed
       Move to next pending task

     IF assessment = fail or blocked:
       Increment iteration: sow task set --id <id> iteration <N+1>
       Set in_progress: sow task set --id <id> status in_progress
       Re-spawn implementer (go back to step 1 for same task)

  CRITICAL RULES:
  - You MUST write and register feedback for every iteration
  - You can ONLY mark completed AFTER writing and registering passing feedback
  - Never skip the feedback step - implementers need it for next iteration
  - Status flow: pending → in_progress → needs_review → (feedback written) → completed

  When all tasks completed: sow advance (transitions to review phase)

Reference: PHASES/IMPLEMENTATION.md, AGENTS.md (implementer)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
