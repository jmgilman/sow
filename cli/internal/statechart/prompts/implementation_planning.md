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
  - Use planner agent for complex breakdowns (see guidance below)
  - No approval needed for task creation
  - Gap-numbered IDs (010, 020, 030...)

PLANNING APPROACH:
  Simple (1-5 tasks):   Create breakdown directly
  Medium (6-9 tasks):   Consider using planner agent
  Large (10+ tasks):    Use planner agent (recommended)

  Planner agent provides systematic task breakdown for complex projects.

NEXT ACTIONS:
  1. Review available artifacts (design docs, discovery notes)
  2. Break work into discrete tasks with clear acceptance criteria
  3. Create tasks: sow task add "<name>" [--id <id>]
  4. Once tasks exist, work will begin automatically

Reference: PHASES/IMPLEMENTATION.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
