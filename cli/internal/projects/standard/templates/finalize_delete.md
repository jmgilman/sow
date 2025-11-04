━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FINALIZE: PROJECT DELETION & PR CREATION (Autonomous Mode)

MANDATORY: Delete project folder and create pull request.

This is the final step to clean up project state and prepare for merge.

WORKFLOW:
  1. Verify all work is committed and pushed
  2. Delete the project directory:
     - This removes .sow/project/ from the branch
     - Prevents project files from being merged to main
  3. Commit the deletion
  4. Create pull request via GitHub CLI

NEXT ACTIONS:
  1. Verify all work is committed: git status
  2. Run: sow agent delete
  3. Commit deletion: git add -A && git commit -m "chore: remove project state"
  4. Push to remote: git push -u origin HEAD
  5. Create pull request:

     Write a comprehensive PR description summarizing:
       - What was implemented
       - Key changes made
       - Testing performed
       - Any important notes for reviewers

     Then create the PR (choose one method):
       sow agent create-pr --body "$(cat <<'EOF'
       <your-pr-description>
       EOF
       )"

       OR: echo "<your-pr-description>" | sow agent create-pr

     The command will:
       - Auto-generate title from project name and description
       - Add "Closes #<number>" if issue-linked
       - Add sow footer automatically
       - Create PR via gh CLI
       - Store and output the PR URL

PR DESCRIPTION EXAMPLE:
  ## Summary

  Implemented GitHub issue integration for the sow CLI, enabling projects
  to be created from and linked to GitHub issues.

  ## Changes

  - Added `sow issue` commands for listing, viewing, and checking issues
  - Extended `sow agent project init` with `--issue` flag
  - Integrated `gh issue develop` for branch creation
  - Added automatic issue closing via PR references

  ## Testing

  - All unit tests pass
  - Integration tests verify issue linking workflow
  - Manual testing confirmed end-to-end flow

FALLBACK (if gh CLI unavailable):
  If create-pr fails:
  1. Output your PR description to the user
  2. Inform them to create PR manually at GitHub web UI
  3. Remind them to:
     - Use project name/description as title
     - Include "Closes #<number>" if issue-linked
     - Add footer: "🤖 Generated with sow"

PROJECT COMPLETION:
  After PR is created successfully:
  - Project lifecycle is complete
  - State machine reaches terminal state
  - Branch is ready for human code review and merge

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
