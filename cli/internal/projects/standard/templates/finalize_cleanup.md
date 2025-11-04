━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FINALIZE: CLEANUP (Autonomous Mode)

PROJECT: {{.Name}}

MANDATORY: Delete project folder, commit, and push.

This is the final step to clean up project state before merge.
The PR has already been created in the previous phase.

CRITICAL TIMING:
  - PR was created WITH .sow/project/ folder committed
  - This phase immediately follows to delete the folder
  - Final PR will NOT contain .sow/project/ (clean merge)

WORKFLOW:

  1. Verify Current State:
     git status                # Should show clean working tree
     git log -1                # Should show most recent commit

     PR should already be created. You can verify with:
     gh pr view

  2. Delete Project Folder:
     rm -rf .sow/project

     This removes:
     - Project state and metadata
     - All phase artifacts and task logs
     - Context files specific to this project

     Rationale:
     - Project state is ephemeral, only used during development
     - Should not be merged to main branch
     - PR preserves all important information

  3. Commit Deletion:
     git add -A
     git commit -m "chore: remove project state [skip ci]"

     Note: Using [skip ci] to avoid unnecessary CI runs for cleanup commit

  4. Push to Remote:
     git push origin HEAD

     This updates the PR with the cleanup commit.
     The PR now shows all feature work WITHOUT project folder.

  5. Set Completion Flag:
     sow phase set metadata.project_deleted true

     This sets the guard condition for phase completion.

  6. Complete Phase:
     sow advance
     → Transitions to NoProject (terminal state)

PROJECT COMPLETION:
  After this phase completes:
  - Project lifecycle is complete
  - State machine reaches terminal state (NoProject)
  - Branch is ready for human code review and merge
  - PR does not contain .sow/project/ folder
  - Feature branch can be safely merged

AUTONOMY BOUNDARIES:

  Full Autonomy (no approval needed):
    • Deleting .sow/project/ folder
    • Creating cleanup commit
    • Pushing to remote
    • Setting metadata flags
    • Completing phase

  Human Approval Required:
    • None (this is fully autonomous cleanup)

IMPORTANT NOTES:

  - This phase runs immediately after PR creation
  - Do not ask for approval to delete project folder (autonomous)
  - The commit message uses [skip ci] to avoid unnecessary builds
  - All feature work is already in PR, this is just cleanup
  - If push fails, check for force-push protection on branch

NEXT ACTIONS:
  1. Verify PR already created: gh pr view
  2. Delete .sow/project/ folder: rm -rf .sow/project
  3. Commit deletion: git add -A && git commit -m "chore: remove project state [skip ci]"
  4. Push to remote: git push origin HEAD
  5. Set completion flag: sow phase set metadata.project_deleted true
  6. Complete phase: sow advance

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
