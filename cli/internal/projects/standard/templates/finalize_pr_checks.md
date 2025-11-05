━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FINALIZE: PR CHECKS (Autonomous Mode)

PROJECT: {{.Name}}

Monitor PR checks, fix failures, and wait for all checks to pass.

RESPONSIBILITIES:
  - Detect if PR has any checks configured
  - Watch checks until they complete
  - View logs and fix any failures autonomously
  - Re-run checks after each fix
  - Proceed when all checks pass (or no checks exist)

WORKFLOW:

  1. Get PR Number:
     The PR was just created. Extract PR number from metadata.

     Check metadata: sow phase get metadata.pr_number

     If pr_number exists in metadata, use it.
     Otherwise, extract from pr_url:
       If pr_url is like "https://github.com/owner/repo/pull/123":
         PR number is 123

  2. Check if PR Has Checks:
     gh pr checks <pr-number> --json name,state

     If output is empty array []:
       → No checks configured for this PR
       → Skip to step 6 (set flag and advance)

     Otherwise:
       → Checks exist, proceed to step 3

  3. Watch Checks Until Complete:
     gh pr checks <pr-number> --watch

     This command blocks until all checks finish.
     Exit codes:
       0 = All checks passed
       1 = Some checks failed
       8 = Checks still pending (shouldn't happen with --watch)

  4. If Checks Failed (exit code 1):

     a. Get detailed check status:
        gh pr checks <pr-number> --json name,state,link,bucket

     b. Identify failed checks (where bucket == "fail"):
        Filter JSON output to find failed check names and run URLs

     c. View logs for each failed check:
        Extract run ID from link field (link format: https://github.com/owner/repo/actions/runs/RUN_ID)

        gh run view <RUN_ID> --log-failed

        This shows logs only for failed steps.

     d. Analyze failures:
        Read error messages, understand what went wrong
        Common failures:
          - Test failures
          - Linting errors
          - Type errors
          - Build failures

     e. Fix the issue:
        Make necessary code changes to fix the failure
        Follow same autonomous approach as FinalizeChecks phase

     f. Commit and push fix:
        git add -A
        git commit -m "fix: address CI failure in [check-name]"
        git push origin HEAD

     g. Return to step 3:
        The push triggers new checks automatically.
        Watch again with: gh pr checks <pr-number> --watch

  5. If All Checks Passed (exit code 0):
     Proceed to step 6

  6. Set Completion Flag:
     sow phase set metadata.pr_checks_passed true

     This sets the guard condition for advancing.

  7. Complete Phase:
     sow advance
     → Transitions to FinalizeCleanup

IMPORTANT NOTES:

  - This phase is fully autonomous (no user approval needed)
  - Treat CI failures like local test failures - fix them automatically
  - The --watch flag makes gh CLI block until checks complete
  - Each fix triggers new checks via push
  - Loop until all checks pass
  - If no checks exist, set flag immediately and advance
  - PR number should be in metadata (set during PR creation)

COMMON SCENARIOS:

  No checks configured:
    → Detect with empty JSON array
    → Set flag immediately
    → Advance (nothing to wait for)

  All checks pass immediately:
    → --watch exits with code 0
    → Set flag
    → Advance

  Some checks fail:
    → View logs
    → Fix issue
    → Push (triggers new checks)
    → Watch again (loop)

AUTONOMY BOUNDARIES:

  Full Autonomy (no approval needed):
    • Monitoring PR checks
    • Viewing failure logs
    • Fixing CI failures
    • Committing and pushing fixes
    • Setting metadata flags
    • Completing phase

  Human Approval Required:
    • None (fully autonomous)

NEXT ACTIONS:
  1. Extract PR number from metadata
  2. Check if PR has checks: gh pr checks <number> --json name
  3. If checks exist: gh pr checks <number> --watch
  4. If failed: view logs, fix, push, repeat step 3
  5. When passed: sow phase set metadata.pr_checks_passed true
  6. Complete: sow advance

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
