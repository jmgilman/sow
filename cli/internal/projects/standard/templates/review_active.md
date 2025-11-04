━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

REVIEW PHASE (Autonomous Mode)

PROJECT: {{.Name}}

Perform mandatory review of implementation.
RESPONSIBILITIES:
  - Review all completed work thoroughly
  - Validate against original requirements
  - Create detailed review report with specific findings
  - Decide assessment: pass or fail
  - Present findings to human for confirmation

ASSESSMENT OPTIONS:

  PASS - Implementation meets requirements
    All work is complete and meets quality standards
    → Will transition to finalize phase after human approval

  FAIL - Issues need addressing
    Specific problems found that require fixes
    → Will loop back to implementation planning after human approval
    → Orchestrator will create new tasks to address review issues

WORKFLOW:

  1. REVIEW THE IMPLEMENTATION
     - Examine all completed tasks and their changes
     - Review code quality, tests, and documentation
     - Check git log and task logs for context
     - Validate against original requirements

  2. CREATE REVIEW REPORT
     Write a comprehensive report documenting:
     - Summary of what was implemented
     - Specific findings (good and problematic)
     - Assessment decision (pass or fail)
     - If FAIL: Detailed list of issues to address

     Save to: project/phases/review/reports/<id>.md
     (Use sequential IDs: 001.md, 002.md, etc.)

  3. ADD ARTIFACT WITH ASSESSMENT
     sow agent artifact add project/phases/review/reports/<id>.md \
       --metadata type=review \
       --metadata assessment=<pass|fail>

  4. PRESENT TO HUMAN FOR REVIEW
     The human will review BOTH:
     - The actual implementation (code changes)
     - Your review report (to verify accuracy)

     Wait for human confirmation that your review is accurate.

  5. AFTER HUMAN CONFIRMS
     Approve the artifact:
       sow agent artifact approve project/phases/review/reports/<id>.md

     If assessment was FAIL, increment iteration for next review cycle:
       sow agent set iteration <current+1>

  6. COMPLETE THE REVIEW PHASE
     sow agent complete

     The state machine will automatically transition based on assessment:
     - PASS → FinalizeDocumentation (proceed to merge preparation)
     - FAIL → ImplementationPlanning (create tasks to fix issues)

Reference: PHASES/REVIEW.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
