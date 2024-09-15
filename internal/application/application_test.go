package application

import (
	"bytes"
	"context"
	"github.com/pavliha/aptforge/internal/deb"
	"github.com/pavliha/aptforge/internal/filereader"
	"github.com/pavliha/aptforge/internal/storage"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

// Mock implementations remain the same unless methods have changed
type MockFileReader struct {
	mock.Mock
}

func (m *MockFileReader) Open(name string) (filereader.File, error) {
	args := m.Called(name)
	return args.Get(0).(filereader.File), args.Error(1)
}

type MockFile struct {
	mock.Mock
}

func (m *MockFile) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockFile) Seek(offset int64, whence int) (int64, error) {
	args := m.Called(offset, whence)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFile) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockFile) Stat() (os.FileInfo, error) {
	args := m.Called()
	return args.Get(0).(os.FileInfo), args.Error(1)
}

// Update MockStorage to include DownloadFile
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

func (m *MockStorage) DownloadFile(ctx context.Context, path string, dest *bytes.Buffer) error {
	args := m.Called(ctx, path, dest)
	return args.Error(0)
}

func (m *MockStorage) IsNotFoundError(err error) bool {
	args := m.Called(err)
	return args.Bool(0)
}

type MockDebExtractor struct {
	mock.Mock
}

func (m *MockDebExtractor) ExtractPackageMetadata(file filereader.File) (*deb.PackageMetadata, error) {
	args := m.Called(file)
	return args.Get(0).(*deb.PackageMetadata), args.Error(1)
}

// Test LoadDebFile remains the same
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

// Test CloseFile remains the same
func TestCloseFile(t *testing.T) {
	mockFile := new(MockFile)
	app := applicationImpl{
		logger: log.NewEntry(log.New()),
	}

	mockFile.On("Close").Return(nil)

	app.CloseFile(mockFile)

	mockFile.AssertExpectations(t)
}

// Test ExtractDebMetadata remains the same
func TestExtractDebMetadata(t *testing.T) {
	mockFile := new(MockFile)
	mockMetadata := &deb.PackageMetadata{
		PackageName:  "testpkg",
		Version:      "1.0",
		Architecture: "amd64",
	}

	mockDebExtractor := new(MockDebExtractor)
	mockDebExtractor.On("ExtractPackageMetadata", mockFile).Return(mockMetadata, nil)

	app := applicationImpl{
		logger:    log.NewEntry(log.New()),
		extractor: mockDebExtractor,
	}

	metadata, err := app.ExtractDebMetadata(mockFile)
	assert.NoError(t, err)
	assert.Equal(t, "testpkg", metadata.PackageName)
}

// Test UploadDebFile remains the same
func TestUploadDebFile(t *testing.T) {
	mockStorage := new(MockStorage)
	mockFile := new(MockFile)
	mockMetadata := &deb.PackageMetadata{
		PackageName:  "testpkg",
		Version:      "1.0",
		Architecture: "amd64",
	}

	expectedPath := "pool/main/t/testpkg/testpkg_1.0_amd64.deb"
	mockStorage.On("UploadFile", mock.Anything, expectedPath, mockFile).Return(nil)

	app := applicationImpl{
		storage: mockStorage,
		config: &Config{
			Component: "main",
		},
	}

	err := app.UploadDebFile(context.Background(), mockMetadata, mockFile)
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

// Updated Test downloadPackagesFromStorage
func TestDownloadPackagesFromStorage(t *testing.T) {
	mockStorage := new(MockStorage)
	existingPackagesContent := "existing packages content"

	mockStorage.On("DownloadFile", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*bytes.Buffer)
		dest.WriteString(existingPackagesContent)
	})

	app := applicationImpl{
		storage: mockStorage,
	}

	buffer, err := app.downloadPackagesFromStorage(context.Background(), "packages-path")
	assert.NoError(t, err)
	assert.Equal(t, existingPackagesContent, buffer.String())
	mockStorage.AssertExpectations(t)
}

// Updated Test UpdatePackagesFile
func TestUpdatePackagesFile(t *testing.T) {
	mockStorage := new(MockStorage)
	mockMetadata := &deb.PackageMetadata{
		PackageName:  "testpkg",
		Version:      "1.0",
		Architecture: "amd64",
	}
	existingPackagesContent := "existing packages content"

	// Mock DownloadFile to populate the buffer
	mockStorage.On("DownloadFile", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*bytes.Buffer)
		dest.WriteString(existingPackagesContent)
	})

	// Mock UploadBuffer for Packages and Packages.gz
	mockStorage.On("UploadBuffer", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	app := applicationImpl{
		storage: mockStorage,
		logger:  log.NewEntry(log.New()),
	}

	packagesBuffer, packagesGzBuffer, err := app.UpdatePackagesFile(context.Background(), "packages-path", mockMetadata)
	assert.NoError(t, err)
	assert.Contains(t, packagesBuffer.String(), "testpkg")
	assert.NotNil(t, packagesGzBuffer)
	mockStorage.AssertExpectations(t)
}

// Updated Test UploadReleaseFile
func TestUploadReleaseFile(t *testing.T) {
	mockStorage := new(MockStorage)
	packagesBuffer := bytes.NewBufferString("Packages content")
	packagesGzBuffer := bytes.NewBufferString("Packages.gz content")
	mockStorage.On("UploadBuffer", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	app := applicationImpl{
		logger:  log.NewEntry(log.New()),
		storage: mockStorage,
		config: &Config{
			Archive:      "stable",
			Component:    "main",
			Origin:       "origin",
			Label:        "Test Repo",
			Architecture: "amd64",
		},
	}

	err := app.UploadPackageReleaseFile(context.Background(), "release-file-path", packagesBuffer, packagesGzBuffer)
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}
