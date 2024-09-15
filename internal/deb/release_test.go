package deb

import (
	"strings"
	"testing"
)

// TestCreatePackageReleaseFileContents tests the architecture-specific release file generation.
func TestCreatePackageReleaseFileContents(t *testing.T) {
	// Test case with typical inputs
	t.Run("typical input", func(t *testing.T) {
		content := ReleaseFileContent{
			Component:    "main",
			Origin:       "Debian",
			Label:        "Debian",
			Archive:      "stable",
			Architecture: "amd64",
			SHA256: []ChecksumInfo{
				{
					Checksum: "123abc",
					Size:     1024,
					Filename: "package1.deb",
				},
			},
		}

		expected := `Origin: Debian
Label: Debian
Suite: stable
Component: main
Architecture: amd64
Date: ` + generateCurrentDate() + `
MD5Sum:
SHA256:
 123abc 1024 package1.deb
`

		result := CreatePackageReleaseFileContents(content)
		if !strings.Contains(result, expected) {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})

	// Test case with no SHA256 entries
	t.Run("no SHA256 entries", func(t *testing.T) {
		content := ReleaseFileContent{
			Component:    "main",
			Origin:       "Ubuntu",
			Label:        "Ubuntu",
			Archive:      "testing",
			Architecture: "arm64",
			SHA256:       []ChecksumInfo{},
		}

		expected := `Origin: Ubuntu
Label: Ubuntu
Suite: testing
Component: main
Architecture: arm64
Date: ` + generateCurrentDate() + `
MD5Sum:
SHA256:
`

		result := CreatePackageReleaseFileContents(content)
		if !strings.Contains(result, expected) {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})

	// Test case with multiple SHA256 entries
	t.Run("multiple SHA256 entries", func(t *testing.T) {
		content := ReleaseFileContent{
			Component:    "contrib",
			Origin:       "Canonical",
			Label:        "Canonical",
			Archive:      "unstable",
			Architecture: "i386",
			SHA256: []ChecksumInfo{
				{
					Checksum: "abc123",
					Size:     2048,
					Filename: "package2.deb",
				},
				{
					Checksum: "def456",
					Size:     4096,
					Filename: "package3.deb",
				},
			},
		}

		expected := `Origin: Canonical
Label: Canonical
Suite: unstable
Component: contrib
Architecture: i386
Date: ` + generateCurrentDate() + `
MD5Sum:
SHA256:
 abc123 2048 package2.deb
 def456 4096 package3.deb
`

		result := CreatePackageReleaseFileContents(content)
		if !strings.Contains(result, expected) {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})
}

// TestCreateSuiteReleaseFileContents tests the high-level suite release file generation.
func TestCreateSuiteReleaseFileContents(t *testing.T) {
	// Test case with typical inputs for suite release
	t.Run("suite typical input", func(t *testing.T) {
		content := ReleaseFileContent{
			Component:    "main",
			Origin:       "Debian",
			Label:        "Debian",
			Archive:      "stable",
			Architecture: "amd64",
			SHA256: []ChecksumInfo{
				{
					Checksum: "123abc",
					Size:     1024,
					Filename: "package1.deb",
				},
			},
		}

		expected := `Origin: Debian
Label: Debian
Suite: stable
Codename: stable
Architectures: amd64
Components: main
Date: ` + generateCurrentDate() + `
SHA256:
 123abc 1024 package1.deb
`

		result := CreateSuiteReleaseFileContents(content)
		if !strings.Contains(result, expected) {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})

	// Test case with no SHA256 entries for suite release
	t.Run("suite no SHA256 entries", func(t *testing.T) {
		content := ReleaseFileContent{
			Component:    "main",
			Origin:       "Ubuntu",
			Label:        "Ubuntu",
			Archive:      "testing",
			Architecture: "arm64",
			SHA256:       []ChecksumInfo{},
		}

		expected := `Origin: Ubuntu
Label: Ubuntu
Suite: testing
Codename: testing
Architectures: arm64
Components: main
Date: ` + generateCurrentDate() + `
SHA256:
`

		result := CreateSuiteReleaseFileContents(content)
		if !strings.Contains(result, expected) {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})

	// Test case with multiple SHA256 entries for suite release
	t.Run("suite multiple SHA256 entries", func(t *testing.T) {
		content := ReleaseFileContent{
			Component:    "contrib",
			Origin:       "Canonical",
			Label:        "Canonical",
			Archive:      "unstable",
			Architecture: "i386",
			SHA256: []ChecksumInfo{
				{
					Checksum: "abc123",
					Size:     2048,
					Filename: "package2.deb",
				},
				{
					Checksum: "def456",
					Size:     4096,
					Filename: "package3.deb",
				},
			},
		}

		expected := `Origin: Canonical
Label: Canonical
Suite: unstable
Codename: unstable
Architectures: i386
Components: contrib
Date: ` + generateCurrentDate() + `
SHA256:
 abc123 2048 package2.deb
 def456 4096 package3.deb
`

		result := CreateSuiteReleaseFileContents(content)
		if !strings.Contains(result, expected) {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})
}
