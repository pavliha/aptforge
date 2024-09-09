package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
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

func (m *MockFile) Close() error {
	return nil
}

// MockStat simulates the file information (we can return empty data as it's not used)
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
	t.Logf("Writing tar header for control file with size: %d bytes", len(data))
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("Failed to close tar writer: %v", err)
	}
	t.Logf("Successfully wrote tar content: %d bytes", tarBuffer.Len())

	// Step 2: Compress the tar file as gzip
	var gzipBuffer bytes.Buffer
	gw := gzip.NewWriter(&gzipBuffer)
	t.Logf("Compressing tar content of size: %d bytes", tarBuffer.Len())
	if _, err := gw.Write(tarBuffer.Bytes()); err != nil {
		t.Fatalf("Failed to write gzip content: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("Failed to close gzip writer: %v", err)
	}
	t.Logf("Successfully compressed content to gzip: %d bytes", gzipBuffer.Len())

	// Step 3: Validate the gzip content by decompressing it back
	t.Log("Validating gzip content")
	if err := validateGzip(t, gzipBuffer.Bytes()); err != nil {
		t.Fatalf("Validation of gzip content failed: %v", err)
	}

	// Step 4: Create a mock AR archive (using control.tar.gz)
	var arBuffer bytes.Buffer
	arBuffer.WriteString("!<arch>\n") // AR header
	t.Logf("Creating AR archive with control.tar.gz size: %d bytes", gzipBuffer.Len())

	// Correctly format the AR header to be exactly 60 bytes
	arHeader := fmt.Sprintf("%-16s%-12s%-6s%-6s%-10d%-2s",
		"control.tar.gz",        // File name (padded to 16 characters)
		"100644",                // File mode (padded to 12 characters)
		"0",                     // Owner ID (padded to 6 characters)
		"0",                     // Group ID (padded to 6 characters)
		len(gzipBuffer.Bytes()), // File size (padded to 10 characters)
		"`\n")                   // End mark (2 characters)

	if len(arHeader) != 60 { // AR header should be exactly 60 bytes
		t.Fatalf("AR header is not the correct length: %d bytes", len(arHeader))
	}
	if _, err := arBuffer.Write([]byte(arHeader)); err != nil {
		t.Fatalf("Failed to write AR header: %v", err)
	}

	if _, err := arBuffer.Write(gzipBuffer.Bytes()); err != nil {
		t.Fatalf("Failed to write AR content: %v", err)
	}

	// Ensure padding after each file in the AR archive to align to 2-byte boundaries
	if len(gzipBuffer.Bytes())%2 != 0 {
		arBuffer.WriteByte('\n')
		t.Log("Added padding to AR content to align to 2-byte boundary")
	}

	t.Logf("Successfully wrote AR archive content, total size: %d bytes", arBuffer.Len())

	// Step 5: Return a MockFile
	return &MockFile{
		Reader: bytes.NewReader(arBuffer.Bytes()),
	}
}

// Function to validate gzip content by decompressing it back
func validateGzip(t *testing.T, gzipData []byte) error {
	t.Logf("Validating gzip content by decompressing")
	reader := bytes.NewReader(gzipData)
	gr, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	// Read the decompressed content and discard it
	if _, err := io.ReadAll(gr); err != nil {
		return fmt.Errorf("failed to decompress gzip data: %v", err)
	}

	t.Logf("Successfully validated gzip content")
	return nil
}

func createTestExtractor() *DefaultMetadataExtractor {
	// Set up logrus logger with debug level for testing
	logger := log.New()
	logger.SetLevel(log.DebugLevel)
	entry := log.NewEntry(logger)
	return &DefaultMetadataExtractor{logger: entry}
}

// Test the successful extraction of metadata from the .deb file
func TestExtractMetadataSuccess(t *testing.T) {
	// Create a mock .deb file with valid control content
	controlContent := "Package: testpkg\nVersion: 1.0\nArchitecture: amd64\n"
	mockFile := createMockDebFile(t, controlContent)

	// Create the metadata extractor
	extractor := createTestExtractor()

	// Perform the extraction
	pkgName, pkgVersion, arch, err := extractor.Extract(mockFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify the extracted metadata
	if pkgName != "testpkg" {
		t.Errorf("expected package name 'testpkg', got '%s'", pkgName)
	}
	if pkgVersion != "1.0" {
		t.Errorf("expected package version '1.0', got '%s'", pkgVersion)
	}
	if arch != "amd64" {
		t.Errorf("expected architecture 'amd64', got '%s'", arch)
	}
}

// Test handling of missing metadata fields
func TestExtractMetadataIncomplete(t *testing.T) {
	// Set up logrus logger with debug level for testing

	// Create a mock .deb file with incomplete control content (missing Architecture)
	controlContent := "Package: testpkg\nVersion: 1.0\n"
	mockFile := createMockDebFile(t, controlContent)

	// Create the metadata extractor
	extractor := createTestExtractor()

	// Perform the extraction
	_, _, _, err := extractor.Extract(mockFile)
	if err == nil {
		t.Fatal("expected error due to incomplete metadata, but got no error")
	}

	expectedError := "incomplete control metadata"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error message '%s', got '%v'", expectedError, err)
	}
}

// Test handling of corrupted tar file content
func TestExtractMetadataCorruptFile(t *testing.T) {
	// Simulate a corrupt .deb file (invalid tar content inside control.tar.gz)
	mockFile := createMockDebFile(t, "invalid_tar_content")

	// Create the metadata extractor
	extractor := createTestExtractor()

	// Perform the extraction
	_, _, _, err := extractor.Extract(mockFile)
	if err == nil {
		t.Fatal("expected error due to corrupted tar file, but got no error")
	}

	expectedError := "failed to read tar archive"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error message '%s', got '%v'", expectedError, err)
	}
}

// Test handling of no control.tar.gz file in the .deb file
func TestExtractMetadataNoControlTarGz(t *testing.T) {
	// Set up logrus logger with debug level for testing

	// Create a mock .deb file with no control.tar.gz
	mockFile := &MockFile{
		Reader: bytes.NewReader([]byte("invalid_archive")),
	}

	// Create the metadata extractor
	extractor := createTestExtractor()

	// Perform the extraction
	_, _, _, err := extractor.Extract(mockFile)
	if err == nil {
		t.Fatal("expected error due to missing control.tar.gz, but got no error")
	}

	expectedError := "failed to read ar archive"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error message '%s', got '%v'", expectedError, err)
	}
}
func TestExtractMetadataFromRealDeb(t *testing.T) {
	// Load the real .deb file from disk
	filePath := "./test_data/missiya-agent_5.1.13-develop-SNAPSHOT-0e37a13_linux_arm64.deb"
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open .deb file: %v", err)
	}
	defer file.Close()

	// Create the metadata extractor with a logger
	extractor := createTestExtractor()

	// Perform the extraction
	pkgName, pkgVersion, arch, err := extractor.Extract(file)
	if err != nil {
		t.Fatalf("Failed to extract metadata from .deb file: %v", err)
	}

	// Print out the extracted metadata
	t.Logf("Package Name: %s", pkgName)
	t.Logf("Package Version: %s", pkgVersion)
	t.Logf("Architecture: %s", arch)

	// Add your own assertions here based on the expected output
	if pkgName == "" {
		t.Errorf("Expected package name, but got an empty string")
	}
	if pkgVersion == "" {
		t.Errorf("Expected package version, but got an empty string")
	}
	if arch == "" {
		t.Errorf("Expected architecture, but got an empty string")
	}
}
