━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FINALIZE: DOCUMENTATION (Autonomous Mode)

PROJECT: {{.ProjectName}}

Check if documentation needs updates based on implementation changes.

RESPONSIBILITIES:
  - Review README, API docs, architectural docs
  - Update documentation to reflect changes
  - Record all documentation updates
  - Minimal viable documentation is acceptable

DOCUMENTATION ASSESSMENT:
  At minimum, update if:
    • User-facing changes (README, examples)
    • Public API changes (API docs, function signatures)
    • Breaking changes (migration notes)

  Optionally update:
    • Architecture decisions (ADRs)
    • Configuration changes (setup docs)
    • Internal documentation

{{if .HasDocumentationUpdates}}UPDATES RECORDED:
{{range .DocumentationUpdates}}  ✓ {{.}}
{{end}}{{end}}

NEXT ACTIONS:
  1. Review changes from implementation phase (check git log)
  2. Identify documentation requiring updates
  3. Update documentation files as needed
  4. Record updates (optional): sow agent set documentation_updates "<summary>"
  5. Complete phase: sow agent complete

  COMPLETION CRITERIA:
    All user-facing changes are documented at minimum level
    (This phase can be completed quickly if no doc updates needed)

  When complete:
    → Auto-transitions to checks phase

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
