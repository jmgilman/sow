package cmdutil

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
)

// GetArtifactByIndex retrieves an artifact from a collection by its index.
// Returns an error if the index is out of range.
func GetArtifactByIndex(artifacts []state.Artifact, index int) (*state.Artifact, error) {
	if err := IndexInRange(len(artifacts), index); err != nil {
		return nil, err
	}
	return &artifacts[index], nil
}

// UpdateArtifactByIndex updates an artifact in a collection by its index.
// Returns an error if the index is out of range.
func UpdateArtifactByIndex(artifacts *[]state.Artifact, index int, artifact state.Artifact) error {
	if err := IndexInRange(len(*artifacts), index); err != nil {
		return err
	}
	(*artifacts)[index] = artifact
	return nil
}

// SetArtifactField sets a field on an artifact in a collection using a field path.
// Supports both direct fields and metadata paths.
func SetArtifactField(artifacts *[]state.Artifact, index int, fieldPath string, value string) error {
	if err := IndexInRange(len(*artifacts), index); err != nil {
		return err
	}
	return SetField(&(*artifacts)[index], fieldPath, value)
}

// GetArtifactField gets a field value from an artifact in a collection using a field path.
// Supports both direct fields and metadata paths.
func GetArtifactField(artifacts []state.Artifact, index int, fieldPath string) (interface{}, error) {
	if err := IndexInRange(len(artifacts), index); err != nil {
		return nil, err
	}
	return GetField(&artifacts[index], fieldPath)
}

// FormatArtifactList formats a list of artifacts for display.
// Returns a human-readable string representation of all artifacts.
func FormatArtifactList(artifacts []state.Artifact) string {
	if len(artifacts) == 0 {
		return "No artifacts found."
	}

	var sb strings.Builder
	sb.WriteString("Artifacts:\n")

	for i, artifact := range artifacts {
		sb.WriteString(fmt.Sprintf("\n[%d] %s\n", i, FormatArtifact(artifact)))
	}

	return sb.String()
}

// FormatArtifact formats a single artifact for display.
// Returns a human-readable string representation of the artifact.
func FormatArtifact(artifact state.Artifact) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Type: %s\n", artifact.Type))
	sb.WriteString(fmt.Sprintf("  Path: %s\n", artifact.Path))
	sb.WriteString(fmt.Sprintf("  Approved: %t\n", artifact.Approved))

	if len(artifact.Metadata) > 0 {
		sb.WriteString("  Metadata:\n")
		for key, value := range artifact.Metadata {
			sb.WriteString(fmt.Sprintf("    %s: %v\n", key, value))
		}
	}

	return sb.String()
}

// IndexInRange validates that an index is within the valid range [0, length).
// Returns an error if the index is out of range.
func IndexInRange(length int, index int) error {
	if index < 0 || index >= length {
		return fmt.Errorf("index out of range: %d (valid range: 0-%d)", index, length-1)
	}
	return nil
}
