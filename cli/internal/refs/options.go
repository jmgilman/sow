package refs

import "encoding/json"

// RefOption configures ref operations.
type RefOption func(*refConfig)

// refConfig holds configuration for ref operations.
type refConfig struct {
	id          string
	semantic    string
	link        string
	tags        []string
	description string
	branch      string // git-specific
	path        string // git-specific
	local       bool
}

// RefListOption configures ref listing operations.
type RefListOption func(*refListConfig)

// refListConfig holds configuration for listing refs.
type refListConfig struct {
	typeFilter     string
	semanticFilter string
	tagsFilter     []string
	localOnly      bool
	committedOnly  bool
}

// WithRefID sets an explicit ref ID (auto-generated if not specified).
func WithRefID(id string) RefOption {
	return func(c *refConfig) {
		c.id = id
	}
}

// WithRefLink sets the workspace symlink name (required).
func WithRefLink(link string) RefOption {
	return func(c *refConfig) {
		c.link = link
	}
}

// WithRefSemantic sets the semantic type (knowledge or code).
func WithRefSemantic(semantic string) RefOption {
	return func(c *refConfig) {
		c.semantic = semantic
	}
}

// WithRefTags sets topic tags for categorization.
func WithRefTags(tags ...string) RefOption {
	return func(c *refConfig) {
		c.tags = tags
	}
}

// WithRefDescription sets the ref description.
func WithRefDescription(desc string) RefOption {
	return func(c *refConfig) {
		c.description = desc
	}
}

// WithRefBranch sets the git branch (only valid for git refs).
func WithRefBranch(branch string) RefOption {
	return func(c *refConfig) {
		c.branch = branch
	}
}

// WithRefPath sets the subpath within repository (only valid for git refs).
func WithRefPath(path string) RefOption {
	return func(c *refConfig) {
		c.path = path
	}
}

// WithRefLocal marks the ref as local-only (not shared with team).
func WithRefLocal(local bool) RefOption {
	return func(c *refConfig) {
		c.local = local
	}
}

// WithRefTypeFilter filters by structural type (git, file).
func WithRefTypeFilter(typeName string) RefListOption {
	return func(c *refListConfig) {
		c.typeFilter = typeName
	}
}

// WithRefSemanticFilter filters by semantic type (knowledge, code).
func WithRefSemanticFilter(semantic string) RefListOption {
	return func(c *refListConfig) {
		c.semanticFilter = semantic
	}
}

// WithRefTagsFilter filters by tags (all tags must match).
func WithRefTagsFilter(tags ...string) RefListOption {
	return func(c *refListConfig) {
		c.tagsFilter = tags
	}
}

// WithRefLocalOnly shows only local refs.
func WithRefLocalOnly() RefListOption {
	return func(c *refListConfig) {
		c.localOnly = true
		c.committedOnly = false
	}
}

// WithRefCommittedOnly shows only committed refs.
func WithRefCommittedOnly() RefListOption {
	return func(c *refListConfig) {
		c.committedOnly = true
		c.localOnly = false
	}
}

// Helper functions for JSON marshaling

// marshalJSON marshals a value to JSON with indentation.
func marshalJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// unmarshalJSON unmarshals JSON data into a value.
func unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
