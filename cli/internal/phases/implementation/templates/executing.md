━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

IMPLEMENTATION EXECUTING (Autonomous Mode)

PROJECT: {{.ProjectName}}

Execute tasks by spawning implementer agents.

TASK STATUS:
  Total: {{.TaskTotal}}
  Completed: {{.TaskCompleted}}
  In Progress: {{.TaskInProgress}}
  Pending: {{.TaskPending}}

{{if .Tasks}}TASKS:
{{range .Tasks}}  [{{.Status}}] {{.Id}} - {{.Name}}
{{end}}{{end}}

AUTONOMY BOUNDARIES:

  Full Autonomy (no approval needed):
    • Marking tasks completed
    • Moving to next task
    • Re-invoking implementers with feedback
    • Adjusting task descriptions
    • Normal task execution flow

  Human Approval Required:
    • Adding new tasks (fail-forward when issues found)
    • Returning to previous phases (design/discovery)
    • Implementer blocked and needs human input

RESPONSIBILITIES:
  - Spawn implementer agents for tasks
  - Monitor task progress and provide feedback
  - Handle normal execution issues autonomously
  - Request approval only for exceptional situations

NEXT ACTIONS:
  - For pending tasks: Spawn implementer agent
  - To update: sow agent task update <id>
  - To mark complete: sow agent task update <id> --status completed
  - When all done: Auto-transition to review

Reference: PHASES/IMPLEMENTATION.md, AGENTS.md (implementer)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
