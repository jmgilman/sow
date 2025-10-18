━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

REVIEW PHASE (Autonomous Mode) - Iteration {{.ReviewIteration}}

Perform mandatory review of implementation.

RESPONSIBILITIES:
  - Review all completed work
  - Validate against requirements
  - Create review report
  - Decide: pass or fail

NEXT ACTIONS:
  1. Review implementation artifacts
  2. Create review report
  3. Add report: sow project review add-report <path> --assessment <pass|fail>

  If FAIL:
    - sow project review increment (loops back to implementation)

  If PASS:
    - sow project phase complete review (→ finalize)

Reference: PHASES/REVIEW.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
