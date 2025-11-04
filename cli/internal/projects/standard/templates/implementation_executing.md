━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

IMPLEMENTATION EXECUTING (Autonomous Mode)

PROJECT: {{.Name}}

Execute tasks by spawning implementer agents.

{{$impl := phase . "implementation"}}
{{if $impl}}TASK STATUS:
  Total: {{len (index $impl "implementation").Tasks}}
  Completed: {{countTasksByStatus $impl "implementation" "completed"}}
  In Progress: {{countTasksByStatus $impl "implementation" "in_progress"}}
  Pending: {{countTasksByStatus $impl "implementation" "pending"}}

TASKS:
{{range (index $impl "implementation").Tasks}}  [{{.Status}}] {{.Id}} - {{.Name}}
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
  - Review completed tasks before final approval
  - Handle normal execution issues autonomously
  - Request approval only for exceptional situations

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

TASK REVIEW WORKFLOW:

  When a task transitions to "needs_review":

  1. Check task status to see what was done:
     sow task status --id <id>

     Review:
     - Task requirements from description.md
     - Outputs list (modified files)
     - Current iteration number

  2. Review actual changes:
     - Read task log.md for worker's explanation
     - Check git diff for modified files
     - Verify tests were written and pass
     - Validate against requirements

  3. Write review feedback:
     Create: project/phases/implementation/tasks/<id>/feedback/<iteration>.md

     Example for iteration 1: feedback/1.md

     Include in feedback file:
     - Summary of what was reviewed
     - Assessment: pass or fail
     - If fail: Specific issues to address (be detailed and actionable)
     - If pass: Confirmation that work meets requirements

  4. Register feedback and set assessment:

     sow task input add --id <id> --type feedback \
       --path "project/phases/implementation/tasks/<id>/feedback/<iteration>.md"

     sow task input set --id <id> --index <N> metadata.assessment <pass|fail>

     Note: <N> is the index of the feedback just added (0-based, check with `sow task status --id <id>`)

  5. Execute review decision based on assessment:

     APPROVE (assessment = pass):
       sow task set --id <id> status completed
       → Task done, moves to next task

     REQUEST CHANGES (assessment = fail):
       sow task set --id <id> status in_progress
       sow task set --id <id> iteration <N+1>
       → Worker will be restarted and will read feedback/<N>.md

  IMPORTANT: Feedback file numbering
    - Iteration 1 writes no feedback (first attempt)
    - Iteration 2 reads feedback/1.md (review of iteration 1)
    - Iteration 3 reads feedback/2.md (review of iteration 2)
    - Pattern: Worker on iteration N reads feedback/(N-1).md

NEXT ACTIONS:
  - For pending tasks: Spawn implementer agent
    • The implementer will automatically load guidance via sow prompt system
    • Provide task context in spawn message: task ID and location
    • Example: "Execute task 010. Context at .sow/project/phases/implementation/tasks/010/"

  - When task reaches needs_review: Perform review (see workflow above)
  - When all done: sow advance (auto-transition to review)

Reference: PHASES/IMPLEMENTATION.md, AGENTS.md (implementer)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
