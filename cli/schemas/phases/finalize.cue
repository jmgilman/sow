package phases

// FinalizePhase represents the finalize phase
#FinalizePhase: {
	#Phase

	// Always enabled
	enabled: true

	// Documentation files updated
	documentation_updates?: [...string] @go(,optional=nillable)

	// Design artifacts moved to knowledge (from→to pairs)
	artifacts_moved?: [...{
		from: string
		to:   string
	}] @go(,optional=nillable)

	// Critical gate: must be true before phase completion
	project_deleted: bool

	// Pull request URL (created during finalize)
	pr_url?: string @go(,optional=nillable)
}
