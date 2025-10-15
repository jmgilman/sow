# Finalize Phase

**Last Updated**: 2025-10-15
**Purpose**: Cleanup and PR creation phase specification

The Finalize phase updates documentation, performs final checks, deletes project files, and creates a pull request for merge.

---

## Table of Contents

- [Overview](#overview)
- [Purpose and Goals](#purpose-and-goals)
- [Entry Criteria](#entry-criteria)
- [Documentation Subphase](#documentation-subphase)
- [Final Checks](#final-checks)
- [Project Deletion](#project-deletion)
- [PR Creation](#pr-creation)
- [Artifacts and Structure](#artifacts-and-structure)
- [Exit Criteria](#exit-criteria)
- [Success Criteria](#success-criteria)
- [Phase Transition](#phase-transition)
- [Related Documentation](#related-documentation)

---

## Overview

**Phase Type**: Required, AI-Autonomous

**Orchestrator Mode**: Autonomous (executes independently)

**Duration**: Variable (minutes to hour, depending on documentation needs)

**Output**: Updated documentation, clean git tree, pull request created

---

## Purpose and Goals

Prepare work for merge through documentation updates, final validation, and PR creation. Update repository documentation to reflect changes. Move design artifacts to permanent locations. Ensure all checks pass. Delete project files (mandatory). Create pull request with comprehensive description. Not implementation (building phase), not validation (review phase), not planning (design phase).

---

## Entry Criteria

Finalize always happens (required phase, cannot be skipped).

**Entry Path**: Always from review phase (human-approved).

---

## Documentation Subphase

Initial subphase where orchestrator determines documentation updates needed.

**Process**: Identify key changes (review implementation, summarize major changes), review existing documentation (READMEs, changelogs, developer guides, code documentation, API documentation, architecture documentation), determine updates needed (outdated documentation, new documentation needed, gaps to fill), answer design artifact question (should design documents be moved to repository, ADRs typically move to repository ADR folder, architecture docs may move to docs/, implementation-specific notes stay in project and deleted), propose changes to human (list proposed documentation updates, explain rationale, request approval), make approved changes (update/create documentation, move design artifacts if approved, commit changes).

**When to Skip**: Change sufficiently small (bug fix with no user-facing impact), no existing documentation needs updating, orchestrator determines no documentation changes warranted.

**Human Approval**: Required before making documentation changes.

---

## Final Checks

After documentation subphase complete, orchestrator verifies readiness.

**Verification Steps**: Tests passing (run full test suite, all tests must pass, if failures loop back to implementation), linters passing (run configured linters, all checks must pass, if failures fix or loop back to implementation), documentation complete (all approved documentation updates made, commits pushed), git tree clean (no uncommitted changes, all work committed and pushed, ready for PR).

---

## Project Deletion

Mandatory step before PR creation: project folder MUST be deleted.

**Central Rule**: Projects are per branch and never present on non-feature branches (main/master).

**Why Critical**: PR should never include `.sow/project/` files, CI should reject PRs with project files present, keeps main branch clean, project state is branch-specific not repository-wide.

**Deletion Process**: Verify everything committed (all implementation work, all documentation updates, git tree clean) → Delete project folder → Create cleanup commit → Push cleanup commit.

**After Deletion**: Project state gone from branch. Cannot resume project after this point. This is intentional (work complete).

---

## PR Creation

Default method uses GitHub CLI (`gh` command).

**Prerequisites**: `gh` CLI installed and authenticated, repository on GitHub, branch pushed to remote.

**With GitHub CLI**: Check for `gh` CLI availability → Create PR automatically with descriptive title and comprehensive description following best practices (summary, changes made, implementation details, testing, documentation, related issues).

**Without GitHub CLI**: Instruct human to create PR manually with generated title and description.

**Best Practices**: Clear descriptive title, comprehensive description, link related issues, include testing information, note documentation changes, use conventional commit format for title if applicable.

**Orchestrator Does NOT**: Merge PR (always human responsibility), delete branch (human decides when), request reviewers (human knows who to ask).

---

## Artifacts and Structure

After finalization: `.sow/project/` deleted, no project artifacts remain.

**What Persists in Repository**: Code changes (in git history), documentation updates (committed), moved design artifacts (ADRs, design docs in repository locations).

**What's Deleted**: All `.sow/project/` contents (discovery artifacts, design phase notes unless moved, implementation task logs, review reports).

**What's In PR**: Code changes, documentation updates, moved design artifacts if any. NO `.sow/project/` files (deleted in cleanup commit).

---

## Exit Criteria

Finalize complete when: documentation updates proposed and approved (if needed), all documentation changes made and committed, tests passing, linters passing, git tree clean (all work committed), project folder deleted (cleanup commit made), PR created (via `gh` or manually by human).

## Success Criteria

Successful when: all documentation current and accurate, final checks pass, project files deleted from branch, PR created with comprehensive description, work ready for human review and merge, orchestrator has handed off to human.

## Phase Transition

**Final Phase**: Finalize is last phase (no automatic transition).

**Project Complete**: After finalization, project complete and cannot be resumed.

**Next Steps**: Human merges PR and deletes feature branch.

---

## Related Documentation

- **[PROJECT_LIFECYCLE.md](../PROJECT_LIFECYCLE.md)** - Finalize as required phase and completion workflow
- **[PHASES/REVIEW.md](./REVIEW.md)** - What happens before finalize
- **[FILE_STRUCTURE.md](../FILE_STRUCTURE.md)** - Project deletion and git versioning
- **[ARCHITECTURE.md](../ARCHITECTURE.md)** - Single project constraint and cleanup rationale
