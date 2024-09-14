package deb

import (
	"aptforge/internal/filereader"
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"testing"
	"time"
)

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
	// Return a dummy FileInfo
	return mockFileInfo{}, nil
}

type mockFileInfo struct{}

func (m mockFileInfo) Name() string       { return "mock.deb" }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m mockFileInfo) ModTime() time.Time { return time.Now() }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() interface{}   { return nil }

func createMockDebFile(t *testing.T, controlContent string) filereader.File {
	// Create a buffer to hold the .deb file content
	debBuffer := new(bytes.Buffer)

	// Write the global ar archive header
	_, err := debBuffer.Write([]byte("!<arch>\n"))
	if err != nil {
		t.Fatalf("failed to write ar archive header: %v", err)
	}

	// Helper function to write an ar header and content with proper padding
	writeArEntry := func(name string, content []byte) {
		// Ensure name is exactly 16 bytes, padded with spaces
		nameField := fmt.Sprintf("%-16s", name)
		modTimeField := fmt.Sprintf("%-12d", time.Now().Unix())
		uidField := fmt.Sprintf("%-6d", 0)
		gidField := fmt.Sprintf("%-6d", 0)
		modeField := fmt.Sprintf("%-8o", 0644)
		sizeField := fmt.Sprintf("%-10d", len(content))
		headerTerminator := "`\n"

		header := nameField + modTimeField + uidField + gidField + modeField + sizeField + headerTerminator
		if len(header) != 60 {
			t.Fatalf("ar header is not 60 bytes long, got %d bytes", len(header))
		}

		_, err := debBuffer.Write([]byte(header))
		if err != nil {
			t.Fatalf("failed to write ar header: %v", err)
		}

		_, err = debBuffer.Write(content)
		if err != nil {
			t.Fatalf("failed to write ar content: %v", err)
		}

		// If the file size is odd, add a newline for padding
		if len(content)%2 != 0 {
			err = debBuffer.WriteByte('\n')
			if err != nil {
				t.Fatalf("failed to write padding byte: %v", err)
			}
		}
	}

	// Write the debian-binary file
	debianBinaryContent := []byte("2.0\n")
	writeArEntry("debian-binary", debianBinaryContent)

	// Create control.tar.gz content
	controlTarGzBuffer := new(bytes.Buffer)
	gzipWriter := gzip.NewWriter(controlTarGzBuffer)
	tarWriter := tar.NewWriter(gzipWriter)

	// Write the control file into the tar archive
	controlBytes := []byte(controlContent)
	controlHeader := &tar.Header{
		Name:     "./control",
		Mode:     0644,
		Size:     int64(len(controlBytes)),
		Typeflag: tar.TypeReg,
	}
	err = tarWriter.WriteHeader(controlHeader)
	if err != nil {
		t.Fatalf("failed to write control tar header: %v", err)
	}
	_, err = tarWriter.Write(controlBytes)
	if err != nil {
		t.Fatalf("failed to write control tar content: %v", err)
	}

	// Close tar and gzip writers
	err = tarWriter.Close()
	if err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	// Write control.tar.gz to the ar archive
	controlTarGzBytes := controlTarGzBuffer.Bytes()
	writeArEntry("control.tar.gz", controlTarGzBytes)

	// Return a MockFile that wraps the debBuffer
	return &MockFile{
		Reader: bytes.NewReader(debBuffer.Bytes()),
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

	file := createMockDebFile(t, controlContent)
	extractor := createTestExtractor()

	metadata, err := extractor.ExtractPackageMetadata(file)
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

	file := createMockDebFile(t, controlContent)
	extractor := createTestExtractor()

	_, err := extractor.ExtractPackageMetadata(file)
	if err == nil {
		t.Fatal("expected error due to incomplete metadata, but got no error")
	}

	expectedError := "incomplete control metadata"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error message containing '%s', got '%v'", expectedError, err)
	}
}

func createCorruptDebFile(t *testing.T) filereader.File {
	t.Logf("Create a buffer with invalid content")
	corruptContent := []byte("this is not a valid .deb file")
	return &MockFile{
		Reader: bytes.NewReader(corruptContent),
	}
}

func TestExtractMetadataCorruptFile(t *testing.T) {
	file := createCorruptDebFile(t)
	extractor := createTestExtractor()

	_, err := extractor.ExtractPackageMetadata(file)
	if err == nil {
		t.Fatal("expected error due to corrupted deb file, but got no error")
	}

	expectedError := "failed to read ar archive"
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
