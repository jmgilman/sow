━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

REVIEW PHASE (Autonomous Mode) - Iteration {{.ReviewIteration}}

PROJECT: {{.ProjectName}}

Perform mandatory review of implementation.
{{if .HasPreviousReview}}
PREVIOUS ITERATION:
  Assessment: {{.PreviousAssessment}}
  This is a re-review after addressing issues.
{{end}}
RESPONSIBILITIES:
  - Review all completed work
  - Validate against requirements
  - Create review report with specific findings
  - Decide: pass or fail

ASSESSMENT OPTIONS:

  PASS - Implementation meets requirements
    → Proceeds to finalization phase
    → Command: sow project review add-report <path> --assessment pass

  FAIL - Issues need addressing
    → Creates detailed review report with specific issues
    → Loops back to implementation planning
    → Human approval required for loop-back
    → Command: sow project review add-report <path> --assessment fail
    → Then: sow project review increment

NEXT ACTIONS:
  1. Review implementation artifacts (code, tests, task logs)
  2. Validate against original requirements
  3. Create comprehensive review report
  4. Add report with assessment: sow project review add-report <path> --assessment <pass|fail>

Reference: PHASES/REVIEW.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
