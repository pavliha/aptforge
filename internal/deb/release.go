package deb

import (
	"fmt"
	"strings"
	"time"
)

type ChecksumInfo struct {
	Checksum string
	Size     int64
	Filename string
}

// ReleaseFileContent represents the content for a Release file in an APT repository.
type ReleaseFileContent struct {
	Origin       string
	Label        string
	Archive      string
	Component    string
	Architecture string
	SHA256       []ChecksumInfo
}

// CreateReleaseFileContents generates the content of a Release file
func CreateReleaseFileContents(content ReleaseFileContent) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Origin: %s\n", content.Origin))
	sb.WriteString(fmt.Sprintf("Label: %s\n", content.Label))
	sb.WriteString(fmt.Sprintf("Suite: %s\n", content.Archive))
	sb.WriteString(fmt.Sprintf("Component: %s\n", content.Component))
	sb.WriteString(fmt.Sprintf("Architecture: %s\n", content.Architecture))
	sb.WriteString(fmt.Sprintf("Date: %s\n", generateCurrentDate())) // You can format the date as needed
	sb.WriteString(fmt.Sprintf("MD5Sum:\n"))
	sb.WriteString(fmt.Sprintf("SHA256:\n"))

	// Add SHA256 checksums
	for _, checksum := range content.SHA256 {
		sb.WriteString(fmt.Sprintf(" %s %d %s\n", checksum.Checksum, checksum.Size, checksum.Filename))
	}

	return sb.String()
}

// generateCurrentDate returns the current date in the proper format
func generateCurrentDate() string {
	return time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 MST")
}
