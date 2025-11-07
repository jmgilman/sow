━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FINALIZE: PR READY (Requires Approval)

PROJECT: {{.Name}}

Update PR body with comprehensive summary and mark ready for review.

RESPONSIBILITIES:
  - Create comprehensive PR body document
  - Present PR body for user approval
  - Update existing draft PR with new body
  - Mark PR as ready for review

WORKFLOW:

  1. Gather PR Information:
     Review project history to understand complete change set:
     - git log origin/{{.Branch}}..HEAD          # Review all commits
     - git diff origin/{{.Branch}}...HEAD        # Review all changes
     - Read .sow/project/context/*                # Review project decisions
     - Read .sow/project/phases/*/tasks/*/log.md  # Review task work logs

  2. Create Comprehensive PR Body Document:
     Create: .sow/project/phases/finalize/pr_body.md

     Include:
     - Clear title summarizing the change (conventional commit format)
     - High-level summary (2-3 paragraphs)
     - Detailed changes (organized by task or component)
     - Testing performed
     - Any breaking changes or migration notes
     - Related issues/tickets if applicable

     Structure:
     ```markdown
     # [Title in conventional commit format]

     ## Summary
     [High-level overview of what this PR accomplishes]

     ## Changes
     [Detailed breakdown of changes, organized logically]

     ## Testing
     [What tests were run, what was verified]

     ## Notes
     [Any additional context, breaking changes, migration steps]

     🤖 Generated with [sow](https://github.com/jmgilman/sow)
     ```

  3. Register PR Body as Output:
     sow output add --type pr_body --path "project/phases/finalize/pr_body.md"

  4. Wait for Verbal Approval:
     Present the PR body to the user:
     "I've created an updated PR body document. Please review:
      .sow/project/phases/finalize/pr_body.md

      Do you approve this PR description?"

     If user approves:
       → Run: sow output set --index 0 approved true
       → Proceed to step 5

     If user requests changes:
       → Update pr_body.md based on feedback
       → Return to step 4 (ask for approval again)

  5. Update PR Description:
     After approval, update the existing draft PR:

     Extract PR number from implementation metadata:
     ```bash
     PR_NUMBER=$(sow phase get metadata.pr_number --phase implementation)
     ```

     Update PR body:
     ```bash
     gh pr edit $PR_NUMBER --body-file .sow/project/phases/finalize/pr_body.md
     ```

  6. Mark PR Ready for Review:
     Remove draft status:

     ```bash
     gh pr ready $PR_NUMBER
     ```

     This marks the PR as ready for review, triggering notifications to reviewers.

  7. Complete Phase:
     sow advance
     → Auto-transitions to PR checks phase

AUTONOMY BOUNDARIES:

  Full Autonomy (no approval needed):
    • Reading project history and context
    • Creating PR body document
    • Registering output artifact
    • Running gh pr edit command (after approval)
    • Running gh pr ready command (after approval)

  Human Approval Required:
    • PR body content (must present and get verbal approval)

IMPORTANT NOTES:

  - PR was created as draft in ImplementationDraftPRCreation state
  - PR number is stored in implementation phase metadata
  - This step updates the existing PR (doesn't create new one)
  - Marking PR ready triggers GitHub notifications
  - Draft → Ready transition is a signal that work is complete
  - If gh CLI not authenticated, inform user to run: gh auth login

NEXT ACTIONS:
  1. Gather project context and commit history
  2. Create comprehensive PR body document
  3. Register as output and request approval
  4. After approval: update PR body via gh pr edit
  5. Mark PR ready: gh pr ready <number>
  6. When complete: sow advance

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
