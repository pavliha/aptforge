package deb

import (
	"path/filepath"
	"strings"
)

// GeneratePoolPath generates the S3 key based on the APT repository structure using `filepath.Join`.
func GeneratePoolPath(component string, metadata *PackageMetadata) string {
	firstLetter := strings.ToLower(string(metadata.PackageName[0]))
	return filepath.Join(
		"pool",
		component,
		firstLetter,
		metadata.PackageName,
		metadata.PackageName+"_"+metadata.Version+"_"+metadata.Architecture+".deb",
	)
}

// ConstructRepoPath Function to construct the APT repository path
func ConstructRepoPath(archive, component, architecture string) string {
	// Construct the path dynamically using the provided archive, component, and architecture
	return filepath.Join("dists", archive, component, "binary-"+architecture)
}
