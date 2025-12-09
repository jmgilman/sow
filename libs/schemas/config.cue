package schemas

// Config defines the schema for the sow configuration file at:
// .sow/config.yaml
//
// This allows teams to customize where formal artifacts are stored.
#Config: {
	// Artifact storage locations
	// All paths are relative to repository root
	artifacts?: {
		// Where to store Architecture Decision Records
		// Default: ".sow/knowledge/adrs"
		adrs?: string @go(,optional=nillable)

		// Where to store design documents
		// Default: ".sow/knowledge/design"
		design_docs?: string @go(,optional=nillable)
	} @go(,optional=nillable)
}
