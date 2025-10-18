━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

IMPLEMENTATION EXECUTING (Autonomous Mode)

Execute tasks by spawning implementer agents.

TASK STATUS:
  Total: {{.TaskTotal}}
  Completed: {{.TaskCompleted}}
  In Progress: {{.TaskInProgress}}
  Pending: {{.TaskPending}}

RESPONSIBILITIES:
  - Spawn implementer agents for tasks
  - Monitor task progress
  - Add new tasks if needed (fail-forward)
  - Update task status

NEXT ACTIONS:
  - For pending tasks: Spawn implementer agent
  - To mark complete: sow task set-status completed <id>
  - When all done: Auto-transition to review

Reference: PHASES/IMPLEMENTATION.md, AGENTS.md (implementer)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
