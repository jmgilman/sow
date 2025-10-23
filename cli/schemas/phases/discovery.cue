package phases

// DiscoveryPhase represents the discovery phase
#DiscoveryPhase: {
	#Phase

	// Can be disabled
	enabled: bool

	// Discovery type categorization
	discovery_type?: null | "bug" | "feature" | "docs" | "refactor" | "general" @go(,optional=nillable)

	// Discovery artifacts requiring approval
	artifacts: [...#Artifact]
}
