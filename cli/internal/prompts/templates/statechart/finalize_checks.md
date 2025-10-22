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
  4. Re-run checks to confirm success

  All checks must pass before proceeding.

NEXT ACTIONS:
  1. Identify available checks in the project
  2. Run all checks sequentially
  3. Address any failures
  4. Confirm all checks pass

  When all checks pass:
    (Will auto-transition to project deletion)

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
