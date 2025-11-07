━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

IMPLEMENTATION: DRAFT PR CREATION (Autonomous Mode)

PROJECT: {{.Name}}
{{if .Description}}DESCRIPTION: {{.Description}}
{{end}}

Create a draft PR to track incremental progress as implementation proceeds.

RESPONSIBILITIES:
  - Create simple initial PR body explaining project intent
  - Create draft PR via gh CLI
  - Store PR URL and number in implementation metadata
  - Advance automatically (no user approval needed)

WORKFLOW:

1. CREATE INITIAL PR BODY

   Create: .sow/project/context/draft_pr_body.md

   Use a simple template explaining intent:

   ```markdown
   # {{.Name}}

   ## Intent

   [Brief 2-3 sentence description of what this PR aims to accomplish]

   ## Status

   🚧 **Draft** - Implementation in progress

   ## Progress

   - [x] Planning phase
   - [ ] Implementation phase
   - [ ] Review phase
   - [ ] Final checks

   ---

   _This PR body will be updated with full details before marking ready for review._

   🤖 Generated with [sow](https://github.com/jmgilman/sow)
   ```

   Extract the intent description from:
   - Project description ({{.Description}})
   - First task descriptions in context/tasks/
   - Any input artifacts (github_issue, feature_request, etc.)

2. GENERATE PR TITLE

   Use conventional commit format for the PR title:

   Format: `<type>(<scope>): <short description>`

   Common types:
   - feat: New feature
   - fix: Bug fix
   - refactor: Code refactoring
   - docs: Documentation changes
   - test: Test additions/changes
   - chore: Maintenance tasks

   Examples:
   - "feat(auth): add JWT authentication system"
   - "fix(api): resolve race condition in request handler"
   - "refactor(db): migrate to connection pooling"

   Derive from project name and description.

3. CREATE DRAFT PR

   Use gh CLI to create draft PR:

   ```bash
   gh pr create \
     --draft \
     --title "feat(scope): description" \
     --body-file .sow/project/context/draft_pr_body.md \
     --json url,number
   ```

   This command:
   - Creates PR in draft state
   - Uses the body from file
   - Returns JSON with URL and number

4. EXTRACT PR METADATA

   Parse the JSON output to get:
   - pr_url: Full GitHub URL
   - pr_number: Numeric PR ID

   Example output:
   ```json
   {
     "url": "https://github.com/owner/repo/pull/123",
     "number": 123
   }
   ```

5. STORE METADATA

   Save PR information to implementation phase metadata:

   ```bash
   sow phase set metadata.pr_url "https://github.com/owner/repo/pull/123" --phase implementation
   sow phase set metadata.pr_number 123 --phase implementation
   sow phase set metadata.draft_pr_created true --phase implementation
   ```

   The draft_pr_created flag enables the guard for advancing to execution.

6. ADVANCE TO EXECUTION

   Once metadata is stored:

   ```bash
   sow advance
   ```

   This transitions to ImplementationExecuting where tasks will be executed
   with commits pushed incrementally.

IMPORTANT NOTES:

  - This step is FULLY AUTONOMOUS - no user approval needed
  - Draft PRs are visible but clearly marked as work-in-progress
  - The PR will be updated with comprehensive body before marking ready
  - PR title uses conventional commit format (will become squash commit message)
  - Early PR creation enables:
    - Incremental progress visibility
    - CI feedback on each pushed commit
    - Clear work-in-progress status

ERROR HANDLING:

  If gh pr create fails:
  - Check if gh CLI is authenticated: gh auth status
  - Check if remote branch exists: git push -u origin HEAD
  - Check if PR already exists for branch: gh pr list --head {{.Branch}}

  If PR already exists:
  - Extract existing PR number and URL
  - Store in metadata
  - Set draft_pr_created flag
  - Advance

NEXT STATE:
  → ImplementationExecuting (task execution with incremental commits)

Reference: PHASES/IMPLEMENTATION.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
