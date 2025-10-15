# Review Phase

**Last Updated**: 2025-10-15
**Purpose**: Quality validation phase specification

The Review phase validates that implementation changes meet expected outcomes. This is the critical quality gate before finalization.

---

## Table of Contents

- [Overview](#overview)
- [Purpose and Goals](#purpose-and-goals)
- [Entry Criteria](#entry-criteria)
- [Review Principle](#review-principle)
- [Orchestrator Review](#orchestrator-review)
- [Reviewer Agent](#reviewer-agent)
- [Human Review](#human-review)
- [Loop-Back Mechanics](#loop-back-mechanics)
- [Artifacts and Structure](#artifacts-and-structure)
- [Exit Criteria](#exit-criteria)
- [Success Criteria](#success-criteria)
- [Phase Transition](#phase-transition)
- [Related Documentation](#related-documentation)

---

## Overview

**Phase Type**: Required, AI-Autonomous

**Orchestrator Mode**: Autonomous (executes independently)

**Duration**: Variable (minutes to hours, depending on change complexity)

**Output**: Review report with validation of implementation against requirements

---

## Purpose and Goals

Validate that implementation changes meet expected outcome. Provide critical quality gate before finalization. Compare implementation against original requirements to identify gaps or deviations. Not implementation (building phase), not planning (design phase), not cleanup (finalize phase).

---

## Entry Criteria

Review always happens (required phase, cannot be skipped).

**Entry Path**: Always from implementation phase (automatic transition when all tasks complete).

---

## Review Principle

Universal question: "Do the changes made in the implementation phase meet our expected outcome?"

**No Different Modes**: Whether reviewing bug fix, new feature, or refactor, fundamental question remains the same.

**Review Requirements**: Review original requirements (documents provided by user, discovery artifacts if used, design documents if used, original project intent), review ALL implementation changes (every file modified, every commit made, all task logs, test coverage and results), compare and validate (implementation matches requirements, original intent achieved, gaps or deviations identified, quality acceptable).

**Critical Importance**: Easy to get side-tracked during implementation and miss the mark. Review phase is the one chance to validate success before finalizing.

---

## Orchestrator Review

Orchestrator ALWAYS performs review (mandatory, happens automatically).

**Review Process**: Read all original requirements/artifacts → Read all implementation changes → Compare implementation against requirements → Identify gaps, issues, or concerns → Document findings in review report → Present findings to human.

**Why Mandatory**: No harm in doing it, likely net positive. Ensures quality gate always present.

**Orchestrator Options**: Perform complete review itself (simple cases), spawn reviewer agent for assistance (complex cases). Either way, orchestrator presents final report.

---

## Reviewer Agent

Optional specialist for complex review work.

**When to Invoke**: Large number of changes (10+ files modified), complex changes requiring deep analysis, orchestrator uncertain about quality, user requests explicitly.

**When NOT to Invoke**: Small changes (1-5 files), simple straightforward work, orchestrator confident in assessment.

**Review Process**: Review requirements (read discovery artifacts, read design documents, understand original intent, identify acceptance criteria), review implementation (examine every file change, review task logs, check test coverage, analyze code quality), compare and validate (implementation matches requirements, identify gaps or issues, assess quality, provide recommendations), document findings (write review report, identify specific issues, provide recommendations, assess overall success).

**Output**: Review report document (orchestrator presents to human).

---

## Human Review

After orchestrator review, human can choose to perform own review or trust orchestrator's assessment.

**Orchestrator Assistance**: Human asks questions about decisions → Orchestrator explains rationale → Human provides direct feedback → Orchestrator helps navigate changes.

**Human Authority**: Trust orchestrator review and approve, perform own review with orchestrator assistance, override orchestrator assessment (approve despite issues or reject despite clean review).

---

## Loop-Back Mechanics

When issues identified (by orchestrator, reviewer agent, or human):

**Process**: Identify resolution (what needs to change), adjust implementation phase (add tasks to address issues), present changes to human (propose new tasks with rationale), upon approval: return to implementation phase, execute new tasks (implementer agents complete work), return to review (automatic transition back), repeat review process (create new review report with incremented number).

**Multiple Iterations**: Review → Implementation → Review → Implementation continues until success.

**Iteration Tracking**: Review reports numbered sequentially (001-review.md, 002-review.md, 003-review.md). Each iteration represents complete review cycle.

---

## Artifacts and Structure

Artifacts stored in `.sow/project/phases/review/`: log.md (review phase conversation log), reports/ (numbered review reports).

**Review Report Structure**: Date and reviewer identification, original intent summary, changes made summary (files modified, tests added, commits), review findings (requirements met, issues identified, recommendations), overall assessment (pass/fail with reasoning), next steps (approve and finalize, or loop back to implementation with specific tasks).

**Numerical Ordering**: Review reports numbered (001, 002, 003) because multiple iterations possible. If looping back, reports serve as additional context for implementer agents.

---

## Exit Criteria

Review complete when: review report created, report sufficiently proves original intent met, all critical issues addressed (or documented as acceptable), human has reviewed and approved.

**Human Authority**: Human gates whether success achieved. Can approve despite issues (accept trade-offs) or reject despite clean review (wants more changes).

## Success Criteria

Successful when: orchestrator review completed and documented, implementation validated against requirements, gaps and issues identified and addressed, quality meets acceptable threshold, human explicitly approves proceeding to finalize.

## Phase Transition

**To Finalize Phase**: Review report created and approved, human explicitly approved. Orchestrator invokes `/phase:finalize`.

**Loop Back to Implementation**: Issues identified requiring changes, tasks added to implementation phase, human approves loop-back. Orchestrator invokes `/phase:implementation`. After implementation completes, automatic transition back to review.

**Note**: Human approval required before proceeding to finalize. No approval required for automatic transition back to review after loop-back implementation.

---

## Related Documentation

- **[PROJECT_LIFECYCLE.md](../PROJECT_LIFECYCLE.md)** - Review as required phase
- **[AGENTS.md](../AGENTS.md)** - Reviewer agent details
- **[PHASES/IMPLEMENTATION.md](./IMPLEMENTATION.md)** - What happens before review
- **[PHASES/FINALIZE.md](./FINALIZE.md)** - What happens after review
- **[LOGGING_AND_STATE.md](../LOGGING_AND_STATE.md)** - Review artifacts format
