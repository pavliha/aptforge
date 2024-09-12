package deb

import (
	"strings"
	"testing"
)

func TestCreatePackagesFileContents(t *testing.T) {
	// Test case with all fields present
	t.Run("all fields present", func(t *testing.T) {
		contents := &PackagesContent{
			PackageName:   "testpkg",
			Version:       "1.0",
			Architecture:  "amd64",
			Maintainer:    "John Doe <johndoe@example.com>",
			Description:   "Test package",
			Section:       "utils",
			Priority:      "optional",
			InstalledSize: "2048",
			Depends:       "dep1, dep2",
			Recommends:    "rec1",
			Suggests:      "sug1",
			Conflicts:     "conf1",
			Provides:      "prov1",
		}

		expected := `Package: testpkg
Version: 1.0
Architecture: amd64
Maintainer: John Doe <johndoe@example.com>
Description: Test package
Section: utils
Priority: optional
Installed-Size: 2048
Depends: dep1, dep2
Recommends: rec1
Suggests: sug1
Conflicts: conf1
Provides: prov1
`

		result := CreatePackagesFileContents(contents)
		if result != expected {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})

	// Test case with only required fields present
	t.Run("only required fields", func(t *testing.T) {
		contents := &PackagesContent{
			PackageName:  "testpkg",
			Version:      "1.0",
			Architecture: "amd64",
			Maintainer:   "John Doe <johndoe@example.com>",
			Description:  "Test package",
		}

		expected := `Package: testpkg
Version: 1.0
Architecture: amd64
Maintainer: John Doe <johndoe@example.com>
Description: Test package
`

		result := CreatePackagesFileContents(contents)
		if result != expected {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})

	// Test case with some optional fields present
	t.Run("some optional fields", func(t *testing.T) {
		contents := &PackagesContent{
			PackageName:   "testpkg",
			Version:       "1.0",
			Architecture:  "amd64",
			Maintainer:    "John Doe <johndoe@example.com>",
			Description:   "Test package",
			Section:       "utils",
			InstalledSize: "2048",
		}

		expected := `Package: testpkg
Version: 1.0
Architecture: amd64
Maintainer: John Doe <johndoe@example.com>
Description: Test package
Section: utils
Installed-Size: 2048
`

		result := CreatePackagesFileContents(contents)
		if result != expected {
			t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
		}
	})

	// Test case with no optional fields
	t.Run("no optional fields", func(t *testing.T) {
		contents := &PackagesContent{
			PackageName:  "testpkg",
			Version:      "1.0",
			Architecture: "amd64",
			Maintainer:   "John Doe <johndoe@example.com>",
			Description:  "Test package",
		}

		result := CreatePackagesFileContents(contents)
		if strings.Contains(result, "Section") || strings.Contains(result, "Priority") ||
			strings.Contains(result, "Installed-Size") || strings.Contains(result, "Depends") ||
			strings.Contains(result, "Recommends") || strings.Contains(result, "Suggests") ||
			strings.Contains(result, "Conflicts") || strings.Contains(result, "Provides") {
			t.Errorf("expected no optional fields, but found optional fields in output:\n%s", result)
		}
	})
}
