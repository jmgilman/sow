package phases

// DesignPhase represents the design phase
#DesignPhase: {
	#Phase

	// Can be disabled
	enabled: bool

	// Whether architect agent was used
	architect_used?: null | bool @go(,optional=nillable)

	// Design artifacts requiring approval (ADRs, design docs)
	artifacts: [...#Artifact]
}
