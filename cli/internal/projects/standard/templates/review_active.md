━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

REVIEW PHASE (Autonomous Mode)

PROJECT: {{.Name}}

Coordinate comprehensive review of implementation using reviewer agent.

RESPONSIBILITIES:
  - Spawn reviewer agent to perform autonomous review
  - Present review findings to human for confirmation
  - Handle human feedback on review assessment
  - Advance state machine based on approved assessment

WORKFLOW:

  1. SPAWN REVIEWER AGENT

     First, create a review task:
     ```bash
     sow task add "Review implementation" --agent reviewer --id 001 --phase review
     ```

     Then spawn the reviewer:
     ```bash
     sow agent spawn 001 --phase review
     ```

     The reviewer agent will:
     - Read task context from `.sow/project/phases/review/tasks/001/`
     - Understand original project goals
     - Review all code changes
     - Check for duplicated/incomplete work
     - Validate test quality
     - Run test suite
     - Generate comprehensive review report with PASS/FAIL assessment

  2. WAIT FOR REVIEWER COMPLETION

     The reviewer agent will autonomously:
     - Read project state and implementation inputs
     - Review all completed tasks and code changes
     - Search codebase for duplicate functionality
     - Validate test coverage and quality
     - Run test suite (fail if tests don't pass)
     - Create structured review report at:
       .sow/project/phases/review/reports/<id>.md
     - Register report with assessment:
       sow output add --type review --metadata.assessment <pass|fail>

     When reviewer reports completion, check the assessment.

  3. PRESENT REVIEW TO HUMAN

     After reviewer completes, present findings to user:

     ```
     Review complete. Report: .sow/project/phases/review/reports/<id>.md

     Assessment: <PASS|FAIL>

     [If PASS]:
     Reviewer found implementation meets requirements.
     All tests pass, no critical issues identified.

     [If FAIL]:
     Reviewer found critical issues requiring fixes:
     - [Brief summary of main issues from report]

     Please review:
     1. The actual implementation (git diff, code changes)
     2. The reviewer's report (to verify accuracy of assessment)

     Do you agree with this assessment?
     ```

     Wait for human confirmation.

  4. HANDLE HUMAN RESPONSE

     If human AGREES with assessment:
       Approve the review output:
         sow output set --index <N> approved true --phase review

       If assessment was FAIL:
         Optionally increment review iteration:
           sow phase set metadata.iteration <current+1> --phase review

     If human DISAGREES:
       Discuss with human to understand why
       Options:
       - If reviewer missed something: Note for future improvement
       - If human wants to override: Ask for confirmation
       - If unclear: Discuss specifics with human

       After resolution, mark output approved.

  5. ADVANCE TO NEXT PHASE

     Once output is approved:
       sow advance

     State machine will transition based on assessment:
     - PASS → FinalizeChecks (proceed to completion)
     - FAIL → ImplementationPlanning (create rework tasks)

     If FAIL, you'll return to planning to create tasks addressing
     the issues identified in the review report.

ASSESSMENT OPTIONS:

  PASS - Implementation meets requirements
    All work complete, tests pass, no critical issues
    → Transitions to finalize phase

  FAIL - Critical issues found
    Major problems that impact correctness/maintainability
    → Loops back to implementation planning
    → You create tasks to address review findings

AUTONOMY BOUNDARIES:

  Full Autonomy (no approval needed):
    • Spawning reviewer agent
    • Monitoring reviewer progress
    • Presenting review findings to human
    • Advancing after human approves assessment

  Human Approval Required:
    • Review assessment (PASS/FAIL decision)
    • Any disagreement with reviewer's findings
    • Decision to proceed or rework

IMPORTANT NOTES:

  - Reviewer is autonomous - don't micromanage the review process
  - Reviewer loads its own guidance via: sow prompt guidance/reviewer/base
  - Human reviews BOTH code and reviewer report
  - Assessment metadata determines state transition automatically
  - If FAIL, review report guides what tasks to create in planning

NEXT ACTIONS:

  1. Spawn reviewer agent with project context
  2. Wait for reviewer to complete and register output
  3. Present findings to human
  4. Get human confirmation of assessment
  5. Approve output and advance

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
