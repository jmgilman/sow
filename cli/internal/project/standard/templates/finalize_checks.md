━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FINALIZE: CHECKS (Autonomous Mode)

PROJECT: {{.ProjectName}}

Run final validation checks before completion.

RESPONSIBILITIES:
  - Run all available quality checks
  - Ensure tests pass
  - Verify linters pass
  - Confirm build succeeds
  - Fix any failures autonomously

STANDARD CHECKS:
  Run these if available in the project:
    □ Tests:   go test ./...  (or npm test, pytest, etc.)
    □ Linter:  golangci-lint run  (or eslint, flake8, etc.)
    □ Build:   go build  (or npm run build, make, etc.)
    □ Format:  go fmt  (or prettier, black, etc.)

APPROACH:
  1. Detect available tooling (Makefile, package.json, go.mod, etc.)
  2. Run appropriate checks for the project type
  3. Fix any failures autonomously
  4. Re-run ALL checks after each fix to prevent regressions
  5. Repeat until all checks pass

  All checks must pass before proceeding.

  If project has no tests/linters/build:
    Complete phase immediately - no checks to run

NEXT ACTIONS:
  1. Identify available checks in the project
  2. Run all checks sequentially
  3. Address any failures autonomously
  4. Re-run all checks after each fix
  5. When all checks pass: sow agent complete

  COMPLETION CRITERIA:
    All available checks pass successfully
    OR no checks available in project

  When complete:
    → Auto-transitions to project deletion phase

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
