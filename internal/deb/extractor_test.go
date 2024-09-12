package deb

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"testing"
)

// MockFile is a simple implementation of the File interface to simulate the .deb file
type MockFile struct {
	Reader *bytes.Reader
}

func (m *MockFile) Read(p []byte) (n int, err error) {
	return m.Reader.Read(p)
}

func (m *MockFile) Seek(offset int64, whence int) (int64, error) {
	return m.Reader.Seek(offset, whence)
}

func (m *MockFile) Close() error {
	return nil
}

func (m *MockFile) Stat() (os.FileInfo, error) {
	return nil, nil
}

// Helper function to create a mock .deb file with control.tar.gz content
func createMockDebFile(t *testing.T, controlContent string) *MockFile {
	// Step 1: Create a control file inside a tar archive
	var tarBuffer bytes.Buffer
	tw := tar.NewWriter(&tarBuffer)
	data := []byte(controlContent)
	header := &tar.Header{
		Name: "./control",
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("Failed to close tar writer: %v", err)
	}

	// Step 2: Compress the tar file as gzip
	var gzipBuffer bytes.Buffer
	gw := gzip.NewWriter(&gzipBuffer)
	if _, err := gw.Write(tarBuffer.Bytes()); err != nil {
		t.Fatalf("Failed to write gzip content: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("Failed to close gzip writer: %v", err)
	}

	// Step 3: Create a mock AR archive (using control.tar.gz)
	var arBuffer bytes.Buffer
	arBuffer.WriteString("!<arch>\n") // AR header

	// Correctly format the AR header to be exactly 60 bytes
	arHeader := fmt.Sprintf("%-16s%-12s%-6s%-6s%-10d%-2s",
		"control.tar.gz",        // File name (padded to 16 characters)
		"0",                     // File timestamp (padded to 12 characters)
		"0",                     // Owner ID (padded to 6 characters)
		"0",                     // Group ID (padded to 6 characters)
		len(gzipBuffer.Bytes()), // File size (padded to 10 characters)
		"`\n")                   // End mark (2 characters)

	t.Logf("AR Header: '%s'", arHeader)
	t.Logf("AR Header Length: %d", len(arHeader))

	if len(arHeader) != 60 {
		t.Fatalf("AR header is not the correct length: %d bytes", len(arHeader))
	}

	if _, err := arBuffer.Write([]byte(arHeader)); err != nil {
		t.Fatalf("Failed to write AR header: %v", err)
	}

	if _, err := arBuffer.Write(gzipBuffer.Bytes()); err != nil {
		t.Fatalf("Failed to write AR content: %v", err)
	}

	// Step 4: Return a MockFile
	return &MockFile{
		Reader: bytes.NewReader(arBuffer.Bytes()),
	}
}

func createTestExtractor() *DefaultMetadataExtractor {
	logger := log.New()
	logger.SetLevel(log.DebugLevel)
	entry := log.NewEntry(logger)
	return &DefaultMetadataExtractor{logger: entry}
}

func TestExtractMetadataSuccess(t *testing.T) {
	controlContent := `Package: testpkg
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
Provides: prov1`

	mockFile := createMockDebFile(t, controlContent)
	extractor := createTestExtractor()

	metadata, err := extractor.ExtractPackageMetadata(mockFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if metadata.PackageName != "testpkg" {
		t.Errorf("expected package name 'testpkg', got '%s'", metadata.PackageName)
	}
	if metadata.Version != "1.0" {
		t.Errorf("expected version '1.0', got '%s'", metadata.Version)
	}
	if metadata.Architecture != "amd64" {
		t.Errorf("expected architecture 'amd64', got '%s'", metadata.Architecture)
	}
	if metadata.Maintainer != "John Doe <johndoe@example.com>" {
		t.Errorf("expected maintainer 'John Doe <johndoe@example.com>', got '%s'", metadata.Maintainer)
	}
	if metadata.Description != "Test package" {
		t.Errorf("expected description 'Test package', got '%s'", metadata.Description)
	}
	// Additional field checks can be added here
}

func TestExtractMetadataIncomplete(t *testing.T) {
	controlContent := `Package: testpkg
Version: 1.0`

	mockFile := createMockDebFile(t, controlContent)
	extractor := createTestExtractor()

	_, err := extractor.ExtractPackageMetadata(mockFile)
	if err == nil {
		t.Fatal("expected error due to incomplete metadata, but got no error")
	}

	expectedError := "incomplete control metadata"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error message '%s', got '%v'", expectedError, err)
	}
}

func TestExtractMetadataCorruptFile(t *testing.T) {
	mockFile := createMockDebFile(t, "invalid_tar_content")
	extractor := createTestExtractor()

	_, err := extractor.ExtractPackageMetadata(mockFile)
	if err == nil {
		t.Fatal("expected error due to corrupted tar file, but got no error")
	}

	expectedError := "failed to read tar archive"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error message '%s', got '%v'", expectedError, err)
	}
}

func TestExtractMetadataNoControlTarGz(t *testing.T) {
	mockFile := &MockFile{
		Reader: bytes.NewReader([]byte("invalid_archive")),
	}
	extractor := createTestExtractor()

	_, err := extractor.ExtractPackageMetadata(mockFile)
	if err == nil {
		t.Fatal("expected error due to missing control.tar.gz, but got no error")
	}

	expectedError := "failed to read ar archive"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error message '%s', got '%v'", expectedError, err)
	}
}
