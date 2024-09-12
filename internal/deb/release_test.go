package deb

import (
	"testing"
)

func TestCreateReleaseFileContents(t *testing.T) {
	// Test case with typical inputs
	t.Run("typical input", func(t *testing.T) {
		content := ReleaseFileContent{
			Component:    "main",
			Origin:       "Debian",
			Label:        "Debian",
			Architecture: "amd64",
			SHA256: []ChecksumInfo{
				{
					Checksum: "123abc",
					Size:     1024,
					Filename: "package1.deb",
				},
			},
			PackagesPath: "dists/stable/main/binary-amd64/Packages",
		}

		expected := `Archive: stable
Component: main
Origin: Debian
Label: Debian
Architecture: amd64
SHA256:
 123abc 1024 package1.deb
 dists/stable/main/binary-amd64/Packages
`

		result := CreateReleaseFileContents(content)
		if result != expected {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})

	// Test case with no SHA256 entries
	t.Run("no SHA256 entries", func(t *testing.T) {
		content := ReleaseFileContent{
			Component:    "main",
			Origin:       "Ubuntu",
			Label:        "Ubuntu",
			Architecture: "arm64",
			SHA256:       []ChecksumInfo{},
			PackagesPath: "dists/stable/main/binary-arm64/Packages",
		}

		expected := `Archive: stable
Component: main
Origin: Ubuntu
Label: Ubuntu
Architecture: arm64
SHA256:
 dists/stable/main/binary-arm64/Packages
`

		result := CreateReleaseFileContents(content)
		if result != expected {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})

	// Test case with multiple SHA256 entries
	t.Run("multiple SHA256 entries", func(t *testing.T) {
		content := ReleaseFileContent{
			Component:    "contrib",
			Origin:       "Canonical",
			Label:        "Canonical",
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
			PackagesPath: "dists/stable/contrib/binary-i386/Packages",
		}

		expected := `Archive: stable
Component: contrib
Origin: Canonical
Label: Canonical
Architecture: i386
SHA256:
 abc123 2048 package2.deb
 def456 4096 package3.deb
 dists/stable/contrib/binary-i386/Packages
`

		result := CreateReleaseFileContents(content)
		if result != expected {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})
}
