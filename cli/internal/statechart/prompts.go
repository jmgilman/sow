package statechart

import "fmt"

// PromptContext contains all information needed to generate contextual prompts.
type PromptContext struct {
	State        State
	ProjectState *ProjectState
}

// GeneratePrompt generates a contextual prompt for the current state.
func GeneratePrompt(ctx PromptContext) string {
	switch ctx.State {
	case NoProject:
		return promptNoProject()
	case DiscoveryDecision:
		return promptDiscoveryDecision()
	case DiscoveryActive:
		return promptDiscoveryActive(ctx.ProjectState)
	case DesignDecision:
		return promptDesignDecision()
	case DesignActive:
		return promptDesignActive(ctx.ProjectState)
	case ImplementationPlanning:
		return promptImplementationPlanning()
	case ImplementationExecuting:
		return promptImplementationExecuting(ctx.ProjectState)
	case ReviewActive:
		return promptReviewActive(ctx.ProjectState)
	case FinalizeDocumentation:
		return promptFinalizeDocumentation()
	case FinalizeChecks:
		return promptFinalizeChecks()
	case FinalizeDelete:
		return promptFinalizeDelete()
	default:
		return fmt.Sprintf("Unknown state: %s", ctx.State)
	}
}

func promptNoProject() string {
	return `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

NO ACTIVE PROJECT

No active project found in this repository.

NEXT ACTION:
  Run: sow project init <name> --description "<description>"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`
}

func promptDiscoveryDecision() string {
	return `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DISCOVERY PHASE DECISION

Determine if discovery phase is warranted for this work.

APPROACH:
  1. Consider the Discovery Worthiness Rubric:
     - Context availability (0-2)
     - Problem clarity (0-2)
     - Codebase familiarity (0-2)
     - Research needs (0-2)

  2. Make decision:
     Score 0-2: Not warranted (skip)
     Score 3-5: Optional (ask user)
     Score 6-8: Recommended (suggest strongly)

NEXT ACTION:
  If discovery needed:
    sow project phase enable discovery --type <bug|feature|docs|refactor|general>

  If discovery not needed:
    (Will auto-transition to design decision)

Reference: PROJECT_LIFECYCLE.md (Discovery Rubric)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`
}

func promptDiscoveryActive(state *ProjectState) string {
	artifactCount := len(state.Phases.Discovery.Artifacts)
	approvedCount := 0
	for _, a := range state.Phases.Discovery.Artifacts {
		if a.Approved {
			approvedCount++
		}
	}

	return fmt.Sprintf(`━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DISCOVERY PHASE (Subservient Mode)

You are operating in SUBSERVIENT MODE - act as assistant to the human.

RESPONSIBILITIES:
  - Facilitate research and investigation
  - Create research artifacts
  - Never make unilateral decisions
  - Request approval for all artifacts

CURRENT STATUS:
  Artifacts: %d total, %d approved

NEXT ACTIONS:
  1. Create research artifacts as needed
  2. Register artifacts: sow project artifact add <path> --phase discovery
  3. Request human approval for each artifact
  4. When all approved: sow project phase complete discovery

Reference: PHASES/DISCOVERY.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`, artifactCount, approvedCount)
}

func promptDesignDecision() string {
	return `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DESIGN PHASE DECISION

Determine if design phase is warranted for this work.

APPROACH:
  1. Consider the Design Worthiness Rubric:
     - Scope size (0-2)
     - Architectural impact (0-2)
     - Integration complexity (0-2)
     - Design decisions (0-2)
     - Bug fix penalty: -3 (minimum 0)

  2. Make decision:
     Score 0-2: Not warranted (skip)
     Score 3-5: Optional (ask user)
     Score 6-8: Recommended (suggest strongly)

NEXT ACTION:
  If design needed:
    sow project phase enable design

  If design not needed:
    (Will auto-transition to implementation)

Reference: PROJECT_LIFECYCLE.md (Design Rubric)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`
}

func promptDesignActive(state *ProjectState) string {
	artifactCount := len(state.Phases.Design.Artifacts)
	approvedCount := 0
	for _, a := range state.Phases.Design.Artifacts {
		if a.Approved {
			approvedCount++
		}
	}

	return fmt.Sprintf(`━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DESIGN PHASE (Subservient Mode)

You are operating in SUBSERVIENT MODE - act as assistant to the human.

RESPONSIBILITIES:
  - Facilitate design alignment through conversation
  - Create design artifacts (ADRs, design docs)
  - Never make unilateral decisions
  - Request approval for all artifacts

CURRENT STATUS:
  Artifacts: %d total, %d approved

NEXT ACTIONS:
  1. Create design artifacts as needed
  2. Register artifacts: sow project artifact add <path> --phase design
  3. Request human approval for each artifact
  4. When all approved: sow project phase complete design

Reference: PHASES/DESIGN.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`, artifactCount, approvedCount)
}

func promptImplementationPlanning() string {
	return `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

IMPLEMENTATION PLANNING (Autonomous Mode)

You are now in AUTONOMOUS MODE - execute within established boundaries.

MODE CHANGE: Subservient → Autonomous

RESPONSIBILITIES:
  - Create task breakdown independently
  - Use planner agent for complex breakdowns (10+ tasks)
  - No approval needed for task creation
  - Gap-numbered IDs (010, 020, 030...)

NEXT ACTIONS:
  1. Review design artifacts (if any)
  2. Break work into discrete tasks
  3. Create tasks: sow task init "<name>" [--id <id>]
  4. Once tasks exist, work will begin automatically

Reference: PHASES/IMPLEMENTATION.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`
}

func promptImplementationExecuting(state *ProjectState) string {
	tasks := state.Phases.Implementation.Tasks
	completed := 0
	inProgress := 0
	pending := 0

	for _, t := range tasks {
		switch t.Status {
		case "completed":
			completed++
		case "in_progress":
			inProgress++
		case "pending":
			pending++
		}
	}

	return fmt.Sprintf(`━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

IMPLEMENTATION EXECUTING (Autonomous Mode)

Execute tasks by spawning implementer agents.

TASK STATUS:
  Total: %d
  Completed: %d
  In Progress: %d
  Pending: %d

RESPONSIBILITIES:
  - Spawn implementer agents for tasks
  - Monitor task progress
  - Add new tasks if needed (fail-forward)
  - Update task status

NEXT ACTIONS:
  - For pending tasks: Spawn implementer agent
  - To mark complete: sow task set-status completed <id>
  - When all done: Auto-transition to review

Reference: PHASES/IMPLEMENTATION.md, AGENTS.md (implementer)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`, len(tasks), completed, inProgress, pending)
}

func promptReviewActive(state *ProjectState) string {
	iteration := state.Phases.Review.Iteration
	if iteration == 0 {
		iteration = 1 // Default to 1 if not set
	}

	return fmt.Sprintf(`━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

REVIEW PHASE (Autonomous Mode) - Iteration %d

Perform mandatory review of implementation.

RESPONSIBILITIES:
  - Review all completed work
  - Validate against requirements
  - Create review report
  - Decide: pass or fail

NEXT ACTIONS:
  1. Review implementation artifacts
  2. Create review report
  3. Add report: sow project review add-report <path> --assessment <pass|fail>

  If FAIL:
    - sow project review increment (loops back to implementation)

  If PASS:
    - sow project phase complete review (→ finalize)

Reference: PHASES/REVIEW.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`, iteration)
}

func promptFinalizeDocumentation() string {
	return `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FINALIZE: DOCUMENTATION (Autonomous Mode)

Check if documentation needs updates.

RESPONSIBILITIES:
  - Review README, API docs, etc.
  - Update documentation as needed
  - Record updates

NEXT ACTIONS:
  If documentation updates needed:
    1. Update documentation files
    2. Register: sow project finalize add-document <path>

  When done (or no updates needed):
    (Will auto-transition to checks)

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`
}

func promptFinalizeChecks() string {
	return `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FINALIZE: CHECKS (Autonomous Mode)

Run final validation checks.

RESPONSIBILITIES:
  - Run tests
  - Run linters
  - Run build
  - Ensure all pass

NEXT ACTIONS:
  If checks needed:
    1. Run: tests, linters, build
    2. Fix any failures

  When done (or no checks needed):
    (Will auto-transition to project deletion)

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`
}

func promptFinalizeDelete() string {
	return `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FINALIZE: PROJECT DELETION (Autonomous Mode)

MANDATORY: Delete project folder before completion.

This is a required step to clean up project state from the branch.

NEXT ACTIONS:
  1. Ensure all work is committed
  2. Run: sow project delete
  3. This will:
     - Set project_deleted flag
     - Remove .sow/project/ directory
  4. Complete: sow project phase complete finalize
  5. Create PR

Reference: PHASES/FINALIZE.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`
}
