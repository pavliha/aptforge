package deb

import (
	"fmt"
	"strings"
)

type PackagesContent struct {
	PackageName   string
	Version       string
	Architecture  string
	Maintainer    string
	Description   string
	Section       string
	Priority      string
	InstalledSize string
	Depends       string
	Recommends    string
	Suggests      string
	Conflicts     string
	Provides      string
}

// CreatePackagesFileContents generates a formatted control file section for a .deb package.
func CreatePackagesFileContents(contents *PackagesContent) string {
	var sb strings.Builder

	// Start with required fields
	sb.WriteString(fmt.Sprintf("Package: %s\n", contents.PackageName))
	sb.WriteString(fmt.Sprintf("Version: %s\n", contents.Version))
	sb.WriteString(fmt.Sprintf("Architecture: %s\n", contents.Architecture))
	sb.WriteString(fmt.Sprintf("Maintainer: %s\n", contents.Maintainer))
	sb.WriteString(fmt.Sprintf("Description: %s\n", contents.Description))

	// Add optional fields if present
	if contents.Section != "" {
		sb.WriteString(fmt.Sprintf("Section: %s\n", contents.Section))
	}
	if contents.Priority != "" {
		sb.WriteString(fmt.Sprintf("Priority: %s\n", contents.Priority))
	}
	if contents.InstalledSize != "" {
		sb.WriteString(fmt.Sprintf("Installed-Size: %s\n", contents.InstalledSize))
	}
	if contents.Depends != "" {
		sb.WriteString(fmt.Sprintf("Depends: %s\n", contents.Depends))
	}
	if contents.Recommends != "" {
		sb.WriteString(fmt.Sprintf("Recommends: %s\n", contents.Recommends))
	}
	if contents.Suggests != "" {
		sb.WriteString(fmt.Sprintf("Suggests: %s\n", contents.Suggests))
	}
	if contents.Conflicts != "" {
		sb.WriteString(fmt.Sprintf("Conflicts: %s\n", contents.Conflicts))
	}
	if contents.Provides != "" {
		sb.WriteString(fmt.Sprintf("Provides: %s\n", contents.Provides))
	}

	// Return the full package contents as a string
	return sb.String()
}
