package deb

import (
	"path/filepath"
	"testing"
)

func TestGeneratePoolPath(t *testing.T) {
	tests := []struct {
		name         string
		component    string
		metadata     *PackageMetadata
		expectedPath string
	}{
		{
			name:      "Standard package",
			component: "main",
			metadata: &PackageMetadata{
				PackageName:  "myapp",
				Version:      "1.0.0",
				Architecture: "amd64",
			},
			expectedPath: filepath.Join("pool", "main", "m", "myapp", "myapp_1.0.0_amd64.deb"),
		},
		{
			name:      "Package name starts with uppercase",
			component: "contrib",
			metadata: &PackageMetadata{
				PackageName:  "Uppercase",
				Version:      "2.1.3",
				Architecture: "arm64",
			},
			expectedPath: filepath.Join("pool", "contrib", "u", "Uppercase", "Uppercase_2.1.3_arm64.deb"),
		},
		{
			name:      "Package with hyphen",
			component: "non-free",
			metadata: &PackageMetadata{
				PackageName:  "hyphen-package",
				Version:      "0.9-beta",
				Architecture: "i386",
			},
			expectedPath: filepath.Join("pool", "non-free", "h", "hyphen-package", "hyphen-package_0.9-beta_i386.deb"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GeneratePoolPath(tt.component, tt.metadata)
			if result != tt.expectedPath {
				t.Errorf("GeneratePoolPath() = %v, want %v", result, tt.expectedPath)
			}
		})
	}
}

func TestConstructRepoPath(t *testing.T) {
	tests := []struct {
		name         string
		archive      string
		component    string
		architecture string
		expectedPath string
	}{
		{
			name:         "Standard path",
			archive:      "stable",
			component:    "main",
			architecture: "amd64",
			expectedPath: filepath.Join("dists", "stable", "main", "binary-amd64"),
		},
		{
			name:         "Testing archive",
			archive:      "testing",
			component:    "contrib",
			architecture: "arm64",
			expectedPath: filepath.Join("dists", "testing", "contrib", "binary-arm64"),
		},
		{
			name:         "Unstable with non-free",
			archive:      "unstable",
			component:    "non-free",
			architecture: "i386",
			expectedPath: filepath.Join("dists", "unstable", "non-free", "binary-i386"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConstructRepoPath(tt.archive, tt.component, tt.architecture)
			if result != tt.expectedPath {
				t.Errorf("ConstructRepoPath() = %v, want %v", result, tt.expectedPath)
			}
		})
	}
}
