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
		}

		expected := `Origin: Debian
Label: Debian
Suite: main
Codename: main
Architectures: amd64
Components: main
SHA256:
 123abc 1024 package1.deb
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
		}

		expected := `Origin: Ubuntu
Label: Ubuntu
Suite: main
Codename: main
Architectures: arm64
Components: main
SHA256:
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
		}

		expected := `Origin: Canonical
Label: Canonical
Suite: contrib
Codename: contrib
Architectures: i386
Components: contrib
SHA256:
 abc123 2048 package2.deb
 def456 4096 package3.deb
`

		result := CreateReleaseFileContents(content)
		if result != expected {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})
}
