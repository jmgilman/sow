━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FINALIZE: DOCUMENTATION (Autonomous Mode)

PROJECT: {{.ProjectName}}

Check if documentation needs updates based on implementation changes.

RESPONSIBILITIES:
  - Review README, API docs, architectural docs
  - Update documentation to reflect changes
  - Move design artifacts to .sow/knowledge/ if applicable
  - Record all documentation updates

DOCUMENTATION ASSESSMENT:
  Consider updating if:
    • New features added (README examples, API docs)
    • Public APIs changed (API documentation)
    • Architecture decisions made (move ADRs to knowledge)
    • Configuration or setup changed (installation docs)
    • Breaking changes introduced (migration guides)

{{if .HasDocumentationUpdates}}UPDATES RECORDED:
{{range .DocumentationUpdates}}  ✓ {{.}}
{{end}}{{end}}

NEXT ACTIONS:
  1. Review changes from implementation phase
  2. Identify documentation requiring updates
  3. Update documentation files
  4. Record updates in phase metadata: sow agent set documentation_updates "<summary>"
  5. Move design artifacts if applicable
  6. Complete when done: sow agent complete

  When complete:
    (Will auto-transition to checks)

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
