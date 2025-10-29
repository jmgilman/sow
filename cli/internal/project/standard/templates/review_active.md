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
    → Human approval required before proceeding to finalize
    → Command: sow agent artifact add <path> --metadata type=review --metadata assessment=pass

  FAIL - Issues need addressing
    → Creates detailed review report with specific issues
    → Human approval required before loop-back to implementation
    → Command: sow agent artifact add <path> --metadata type=review --metadata assessment=fail
    → Then: sow agent set iteration <n+1>

NEXT ACTIONS:
  1. Review implementation artifacts (code, tests, task logs)
  2. Validate against original requirements
  3. Create comprehensive review report
  4. Add report with assessment:
     sow agent artifact add project/phases/review/reports/<id>.md \
       --metadata type=review \
       --metadata assessment=<pass|fail>
  5. Request human approval of your review
  6. Human approves: sow agent artifact approve project/phases/review/reports/<id>.md
  7. When all review artifacts approved: sow agent complete
  8. Transition occurs based on assessment (pass → finalize, fail → implementation)

Reference: PHASES/REVIEW.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
