━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

IMPLEMENTATION PLANNING (Autonomous Mode)

PROJECT: {{.ProjectName}}
DESCRIPTION: {{.ProjectDescription}}

MODE CHANGE: Subservient → Autonomous

You are now in AUTONOMOUS MODE - execute within established boundaries.

AVAILABLE CONTEXT:
{{if .HasDesign}}  ✓ Design Phase: {{.DesignArtifactCount}} artifacts available
{{end}}{{if .HasDiscovery}}  ✓ Discovery Phase: {{.DiscoveryArtifactCount}} artifacts available
{{end}}{{if not (or .HasDesign .HasDiscovery)}}  • Direct implementation (no discovery or design phases)
{{end}}
RESPONSIBILITIES:
  - Create task breakdown independently
  - Break work into discrete, testable units
  - Request human approval when planning is complete
  - Use gap-numbered IDs (010, 020, 030...) to allow insertions

TASK NUMBERING:
  Start at 010, increment by 10 (020, 030, 040...)
  This allows inserting tasks between existing ones if needed (015, 025, etc.)

TASK LIFECYCLE:
  pending → in_progress → needs_review → completed

  Workers mark tasks as "needs_review" when done, NOT "completed"
  You (orchestrator) review and approve/reject in executing phase

TASK DESCRIPTION TEMPLATE:
  Each task should include:
  - What needs to be built/changed
  - Acceptance criteria (how to verify completion)
  - Dependencies on other tasks (if any)
  - Files likely to be modified

NEXT ACTIONS:
  1. Review available artifacts (design docs, discovery notes)
  2. Break work into discrete tasks with clear acceptance criteria
  3. Create tasks: sow agent task add "<name>" --description "..."
  4. When all tasks created, present plan to human
  5. After human confirms: sow agent task approve
  6. Then autonomous execution begins

Reference: PHASES/IMPLEMENTATION.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
