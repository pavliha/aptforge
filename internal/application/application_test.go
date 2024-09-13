package application

import (
	"aptforge/internal/deb"
	"aptforge/internal/filereader"
	"aptforge/internal/storage"
	"bytes"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

type MockFileReader struct {
	mock.Mock
}

// Read method implementation
func (m *MockFileReader) Open(name string) (filereader.File, error) {
	args := m.Called(name)
	return args.Get(0).(filereader.File), args.Error(1)
}

type MockFile struct {
	mock.Mock
}

// Read method implementation
func (m *MockFile) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

// Seek method implementation
func (m *MockFile) Seek(offset int64, whence int) (int64, error) {
	args := m.Called(offset, whence)
	return args.Get(0).(int64), args.Error(1)
}

// Close method implementation
func (m *MockFile) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Stat method implementation
func (m *MockFile) Stat() (os.FileInfo, error) {
	args := m.Called()
	return args.Get(0).(os.FileInfo), args.Error(1)
}

// MockStorage is a mock implementation of storage.Storage
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) UploadFile(ctx context.Context, path string, file filereader.File) error {
	args := m.Called(ctx, path, file)
	return args.Error(0)
}

func (m *MockStorage) UploadBuffer(ctx context.Context, path string, buffer *bytes.Buffer) error {
	args := m.Called(ctx, path, buffer)
	return args.Error(0)
}

func (m *MockStorage) Download(ctx context.Context, path string) (storage.Object, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(storage.Object), args.Error(1)
}

type MockDebExtractor struct {
	mock.Mock
}

func (m *MockDebExtractor) ExtractPackageMetadata(file filereader.File) (*deb.PackageMetadata, error) {
	args := m.Called(file)
	return args.Get(0).(*deb.PackageMetadata), args.Error(1)
}

// Test LoadDebFile
func TestLoadDebFile(t *testing.T) {
	mockFileReader := new(MockFileReader)
	mockFile := new(MockFile)
	app := applicationImpl{
		logger:     log.NewEntry(log.New()),
		fileReader: mockFileReader,
	}

	mockFileReader.On("Open", "test_linux_arm64.deb").Return(mockFile, nil)

	file, err := app.LoadDebFile("test_linux_arm64.deb")

	assert.NoError(t, err)
	assert.NotNil(t, file)
	mockFileReader.AssertExpectations(t)
}

// Test CloseFile
func TestCloseFile(t *testing.T) {
	mockFile := new(MockFile)
	app := applicationImpl{
		logger: log.NewEntry(log.New()),
	}

	mockFile.On("Close").Return(nil)

	app.CloseFile(mockFile)

	mockFile.AssertExpectations(t)
}

// Test ExtractDebMetadata
func TestExtractDebMetadata(t *testing.T) {
	mockFile := new(MockFile)
	mockMetadata := &deb.PackageMetadata{
		PackageName:  "testpkg",
		Version:      "1.0",
		Architecture: "amd64",
	}

	mockDebExtractor := new(MockDebExtractor) // assuming DefaultMetadataExtractor exists in deb package
	mockDebExtractor.On("ExtractPackageMetadata", mockFile).Return(mockMetadata, nil)

	app := applicationImpl{
		logger:    log.NewEntry(log.New()),
		extractor: mockDebExtractor,
	}

	metadata, err := app.ExtractDebMetadata(mockFile)
	assert.NoError(t, err)
	assert.Equal(t, "testpkg", metadata.PackageName)
}

// Test UploadDebFile
func TestUploadDebFile(t *testing.T) {
	mockStorage := new(MockStorage)
	mockDebExtractor := new(MockDebExtractor)
	mockFile := new(MockFile)
	mockMetadata := &deb.PackageMetadata{
		PackageName:  "testpkg",
		Version:      "1.0",
		Architecture: "amd64",
	}
	mockStorage.On("UploadFile", mock.Anything, "pool/main/t/testpkg/testpkg_1.0_amd64.deb", mock.Anything).Return(nil)

	app := applicationImpl{
		logger:     nil,
		storage:    mockStorage,
		fileReader: nil,
		extractor:  mockDebExtractor,
		config: &Config{
			Storage:      nil,
			Archive:      "stable",
			Component:    "main",
			Origin:       "origin",
			Label:        "Test Repo",
			Architecture: "amd64",
		},
		repoPath:         "repo-path",
		packagesFilePath: "packages-path",
		releaseFilePath:  "release-file-path",
	}

	err := app.UploadDebFile(context.Background(), mockMetadata, mockFile)
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

// Test DownloadPackagesFromStorage
func TestDownloadPackagesFromStorage(t *testing.T) {
	mockStorage := new(MockStorage)
	packagesBuffer := bytes.NewBufferString("existing packages content")
	mockStorage.On("Download", mock.Anything, mock.Anything).Return(packagesBuffer, nil)

	app := applicationImpl{
		storage: mockStorage,
	}

	buffer, err := app.DownloadPackagesFromStorage(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "existing packages content", buffer.String())
	mockStorage.AssertExpectations(t)
}

// Test UpdatePackagesFile
func TestUpdatePackagesFile(t *testing.T) {
	mockStorage := new(MockStorage)
	mockMetadata := &deb.PackageMetadata{
		PackageName:  "testpkg",
		Version:      "1.0",
		Architecture: "amd64",
	}
	packagesBuffer := bytes.NewBufferString("existing packages content")
	mockStorage.On("Download", mock.Anything, mock.Anything).Return(packagesBuffer, nil)
	mockStorage.On("UploadBuffer", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	app := applicationImpl{
		storage: mockStorage,
	}

	buffer, err := app.UpdatePackagesFile(context.Background(), mockMetadata)
	assert.NoError(t, err)
	assert.Contains(t, buffer.String(), "testpkg")
	mockStorage.AssertExpectations(t)
}

// Test UploadReleaseFile
func TestUploadReleaseFile(t *testing.T) {
	mockStorage := new(MockStorage)
	packagesBuffer := bytes.NewBufferString("Packages content")
	mockStorage.On("UploadBuffer", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	app := applicationImpl{
		logger:     nil,
		storage:    mockStorage,
		fileReader: nil,
		extractor:  nil,
		config: &Config{
			Storage:      nil,
			Archive:      "stable",
			Component:    "main",
			Origin:       "origin",
			Label:        "Test Repo",
			Architecture: "amd64",
		},
		repoPath:         "repo-path",
		packagesFilePath: "packages-path",
		releaseFilePath:  "release-file-path",
	}

	err := app.UploadReleaseFile(context.Background(), packagesBuffer)
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}
