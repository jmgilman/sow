package schemas

// RefManifest defines the schema for .sow-ref.yaml manifest files.
//
// These manifests are required in all OCI-distributed refs and define
// metadata including title, link name, content description, classifications,
// tags, and optional provenance/packaging/hints sections.
#RefManifest: {
	// schema_version is the semantic version of this manifest format.
	// Format: MAJOR.MINOR.PATCH
	schema_version: string & =~"^[0-9]+\\.[0-9]+\\.[0-9]+$"

	// ref contains core identification for this reference.
	ref: #RefIdentification

	// content describes what this reference contains.
	content: #RefContent

	// provenance contains optional authorship and source information.
	provenance?: #RefProvenance

	// packaging contains optional publishing configuration.
	packaging?: #RefPackaging

	// hints provides optional LLM usage suggestions.
	hints?: #RefHints

	// metadata is a freeform map for organization-specific data.
	metadata?: {...}
}

// RefIdentification contains core identification fields for a reference.
#RefIdentification: {
	// title is the human-readable name for this reference.
	title: string & !=""

	// link is the symlink name in kebab-case format.
	// Must start and end with alphanumeric, can contain hyphens in the middle.
	// Examples: "go-standards", "api-patterns", "my-ref"
	link: string & =~"^[a-z0-9][a-z0-9-]*[a-z0-9]$"
}

// RefContent describes the content of a reference.
#RefContent: {
	// description is a non-empty string describing the reference.
	description: string & !=""

	// summary is an optional longer description.
	summary?: string

	// classifications is an array of at least one classification.
	classifications: [#RefClassification, ...#RefClassification]

	// tags is an array of at least one tag.
	tags: [string, ...string]
}

// RefClassification categorizes a reference by type.
#RefClassification: {
	// type is one of the predefined classification types.
	type: #ClassificationType

	// description is an optional string describing the classification.
	description?: string
}

// ClassificationType defines the allowed classification values.
#ClassificationType: "tutorial" | "api-reference" | "guidelines" |
	"architecture" | "runbook" | "specification" | "reference" |
	"code-examples" | "code-templates" | "code-library" | "uncategorized"

// RefProvenance contains optional authorship and source information.
#RefProvenance: {
	// authors is an optional array of author names.
	authors?: [...string]

	// created is an optional RFC 3339 timestamp string.
	created?: string

	// updated is an optional RFC 3339 timestamp string.
	updated?: string

	// source is an optional source URL (e.g., git URL).
	source?: string

	// license is an optional license identifier (e.g., "MIT", "Apache-2.0").
	license?: string
}

// RefPackaging contains optional publishing configuration.
#RefPackaging: {
	// exclude is an optional array of glob patterns to exclude when packaging.
	exclude?: [...string]
}

// RefHints provides optional LLM usage suggestions.
#RefHints: {
	// suggested_queries is an optional array of example queries.
	suggested_queries?: [...string]

	// primary_files is an optional array of important file paths.
	primary_files?: [...string]
}
