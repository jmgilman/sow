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
  1. Verify git status is clean (all work committed)
  2. Run: sow agent project delete
  3. Commit deletion: git add -A && git commit -m "chore: remove project state"
  4. Push: git push
  5. Create pull request:

     Write a comprehensive PR description summarizing:
       - What was implemented
       - Key changes made
       - Testing performed
       - Any important notes for reviewers

     Then create the PR:
       echo "<your-pr-description>" | sow agent project create-pr

     The command will:
       - Auto-generate title from project name
       - Add "Closes #<number>" if issue-linked
       - Add sow footer
       - Create PR via gh CLI
       - Output the PR URL

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

FALLBACK (if gh unavailable):
  - The create-pr command will fail with helpful error
  - Instruct user to create PR manually via GitHub UI
  - Provide your written description
  - Remind them to include "Closes #<number>" for issue-linked projects

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
