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

// ReleaseFileContent represents the content for a ReleaseFileContent file in an APT repository.
type ReleaseFileContent struct {
	Component    string
	Origin       string
	Label        string
	Architecture string
	SHA256       []ChecksumInfo
	PackagesPath string // New field for dynamic path
}

// CreateReleaseFileContents generates a formatted ReleaseFileContent file for an APT repository.
func CreateReleaseFileContents(content ReleaseFileContent) string {
	var sb strings.Builder

	// Write the main fields of the Release file
	sb.WriteString("Archive: stable\n")
	sb.WriteString("Component: ")
	sb.WriteString(content.Component)
	sb.WriteString("\n")
	sb.WriteString("Origin: ")
	sb.WriteString(content.Origin)
	sb.WriteString("\n")
	sb.WriteString("Label: ")
	sb.WriteString(content.Label)
	sb.WriteString("\n")
	sb.WriteString("Architecture: ")
	sb.WriteString(content.Architecture)
	sb.WriteString("\n")

	// Write the SHA256 section
	sb.WriteString("SHA256:\n")
	for _, checksum := range content.SHA256 {
		sb.WriteString(fmt.Sprintf(" %s %d %s\n", checksum.Checksum, checksum.Size, checksum.Filename))
	}

	sb.WriteString(" ")
	sb.WriteString(content.PackagesPath)
	sb.WriteString("\n")

	// Return the full contents as a string
	return sb.String()
}
