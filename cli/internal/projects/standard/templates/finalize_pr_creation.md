━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FINALIZE: PR CREATION (Requires Approval)

PROJECT: {{.Name}}

Create pull request with comprehensive summary and documentation.

RESPONSIBILITIES:
  - Verify git working directory is clean
  - Ensure local branch is synced with remote
  - Create comprehensive PR body document
  - Present PR body for user approval
  - Create PR via gh CLI after approval

WORKFLOW:

  1. Verify Repository State:
     git status                # Must be clean
     git fetch origin
     git status                # Check if ahead/behind remote

     If behind remote:
       → Issue: Local branch not synced
       → Action: Inform user to sync before continuing

     If ahead but not pushed:
       → Action: Push local commits
       → Command: git push origin HEAD

     If working directory not clean:
       → Issue: Uncommitted changes exist
       → Action: Review what's uncommitted, create commit if needed

  2. Gather PR Information:
     Review project history to understand complete change set:
     - git log origin/{{.Branch}}..HEAD          # Review all commits
     - git diff origin/{{.Branch}}...HEAD        # Review all changes
     - Read .sow/project/context/*                # Review project decisions
     - Read .sow/project/phases/*/tasks/*/log.md  # Review task work logs

  3. Create PR Body Document:
     Create: .sow/project/phases/finalize/pr_body.md

     Include:
     - Clear title summarizing the change
     - High-level summary (2-3 paragraphs)
     - Detailed changes (organized by task or component)
     - Testing performed
     - Any breaking changes or migration notes
     - Related issues/tickets if applicable

     Structure:
     ```markdown
     # [Title]

     ## Summary
     [High-level overview of what this PR accomplishes]

     ## Changes
     [Detailed breakdown of changes, organized logically]

     ## Testing
     [What tests were run, what was verified]

     ## Notes
     [Any additional context, breaking changes, migration steps]
     ```

  4. Register PR Body as Output:
     sow output add --type pr_body --path "project/phases/finalize/pr_body.md"

  5. Wait for Verbal Approval:
     Present the PR body to the user:
     "I've created a PR body document. Please review:
      .sow/project/phases/finalize/pr_body.md

      Do you approve this PR description?"

     If user approves:
       → Run: sow output set --index 0 approved true
       → Proceed to step 6

     If user requests changes:
       → Update pr_body.md based on feedback
       → Return to step 5 (ask for approval again)

  6. Create Pull Request:
     After approval command succeeds, create PR using gh CLI:

     gh pr create \
       --title "$(head -1 .sow/project/phases/finalize/pr_body.md | sed 's/^# //')" \
       --body-file .sow/project/phases/finalize/pr_body.md \
       --base {{.Branch}}

     Save PR URL for user reference.

  7. Complete Phase:
     sow advance
     → Auto-transitions to cleanup phase

AUTONOMY BOUNDARIES:

  Full Autonomy (no approval needed):
    • Verifying git status
    • Pushing local commits to remote
    • Reading project history and context
    • Creating PR body document
    • Registering output artifact
    • Running gh pr create command (after approval)

  Human Approval Required:
    • PR body content (must present and get verbal approval)
    • Any git operations if working tree has unexpected state

IMPORTANT NOTES:

  - PR body document lives in .sow/project/ which will be deleted in next phase
  - This is expected: PR is created WITH project folder, then immediately cleaned up
  - Do not push the cleanup commit until next phase
  - If gh CLI not authenticated, inform user to run: gh auth login

NEXT ACTIONS:
  1. Verify repository state (clean, synced)
  2. Create comprehensive PR body document
  3. Register as output and request approval
  4. After approval: create PR via gh CLI
  5. When PR created: sow advance

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
