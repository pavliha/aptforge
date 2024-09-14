package deb

import (
	"fmt"
	"strings"
)

type ChecksumInfo struct {
	Checksum string
	Size     int64
	Filename string
}

// ReleaseFileContent represents the content for a Release file in an APT repository.
type ReleaseFileContent struct {
	Component    string
	Origin       string
	Label        string
	Architecture string
	SHA256       []ChecksumInfo
}

// CreateReleaseFileContents generates a formatted Release file for an APT repository.
func CreateReleaseFileContents(content ReleaseFileContent) string {
	var sb strings.Builder

	// Write the main fields of the Release file
	sb.WriteString("Origin: ")
	sb.WriteString(content.Origin)
	sb.WriteString("\nLabel: ")
	sb.WriteString(content.Label)
	sb.WriteString("\nSuite: ")
	sb.WriteString(content.Component)
	sb.WriteString("\nCodename: ")
	sb.WriteString(content.Component)
	sb.WriteString("\nArchitectures: ")
	sb.WriteString(content.Architecture)
	sb.WriteString("\nComponents: ")
	sb.WriteString(content.Component)
	sb.WriteString("\n")

	// Write the SHA256 section
	sb.WriteString("SHA256:\n")
	for _, checksum := range content.SHA256 {
		sb.WriteString(fmt.Sprintf(" %s %d %s\n", checksum.Checksum, checksum.Size, checksum.Filename))
	}

	return sb.String()
}
