package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/pavliha/aptforge/internal/deb"
	"github.com/pavliha/aptforge/internal/filereader"
	"github.com/pavliha/aptforge/internal/storage"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

type Config struct {
	Storage      *storage.Config
	Archive      string
	Component    string
	Origin       string
	Label        string
	Architecture string
}

type Application interface {
	LoadDebFile(filePath string) (filereader.File, error)
	CloseFile(file filereader.File)
	ExtractDebMetadata(file filereader.File) (*deb.PackageMetadata, error)
	UploadDebFile(ctx context.Context, metadata *deb.PackageMetadata, file filereader.File) error
	UpdatePackagesFile(ctx context.Context, metadata *deb.PackageMetadata) (*bytes.Buffer, *bytes.Buffer, error)
	UploadReleaseFile(ctx context.Context, packagesBuffer *bytes.Buffer, packagesGzBuffer *bytes.Buffer) error
}

type applicationImpl struct {
	logger           *log.Entry
	storage          storage.Storage
	fileReader       filereader.Reader
	extractor        deb.Extractor
	config           *Config
	repoPath         string
	packagesFilePath string
	releaseFilePath  string
}

func New(logger *log.Entry, config *Config) Application {
	repoPath := deb.ConstructRepoPath(config.Archive, config.Component, config.Architecture)
	return &applicationImpl{
		logger:           logger,
		config:           config,
		storage:          storage.Initialize(logger, config.Storage),
		fileReader:       filereader.New(logger.WithField("pkg", "file")),
		extractor:        deb.New(logger.WithField("pkg", "deb")),
		repoPath:         repoPath,
		packagesFilePath: filepath.Join(repoPath, "Packages"),
		releaseFilePath:  filepath.Join(repoPath, "Release"),
	}
}

func (a *applicationImpl) LoadDebFile(filePath string) (filereader.File, error) {
	file, err := a.fileReader.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (a *applicationImpl) CloseFile(file filereader.File) {
	err := file.Close()
	if err != nil {
		a.logger.Errorf("Failed to close file: %v", err)
	}
}

func (a *applicationImpl) ExtractDebMetadata(file filereader.File) (*deb.PackageMetadata, error) {
	metadata, err := a.extractor.ExtractPackageMetadata(file)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	return metadata, nil
}

func (a *applicationImpl) UploadDebFile(ctx context.Context, metadata *deb.PackageMetadata, file filereader.File) error {
	debPath := deb.GeneratePoolPath(a.config.Component, metadata)
	err := a.storage.UploadFile(ctx, debPath, file)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}

func (a *applicationImpl) DownloadPackagesFromStorage(ctx context.Context) (*bytes.Buffer, error) {
	var packagesBuffer bytes.Buffer

	// Download the existing Packages file
	err := a.storage.DownloadFile(ctx, a.packagesFilePath, &packagesBuffer)
	if err != nil {
		// If the error indicates the file does not exist, start with an empty buffer
		if storage.IsNotFoundError(err) {
			a.logger.Info("No existing Packages file found; starting fresh")
			return &packagesBuffer, nil
		}
		return nil, fmt.Errorf("failed to download Packages file: %w", err)
	}

	return &packagesBuffer, nil
}

func (a *applicationImpl) UpdatePackagesFile(ctx context.Context, metadata *deb.PackageMetadata) (*bytes.Buffer, *bytes.Buffer, error) {
	// Download an existing Packages file
	existingPackagesBuffer, err := a.DownloadPackagesFromStorage(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download Packages file: %v", err)
	}

	// Convert package metadata to Packages file format
	newPackageContent := deb.CreatePackagesFileContents(mapMetadataToPackageContents(metadata))

	// Log buffer lengths for debugging
	a.logger.Debugf("Existing Packages buffer length: %d", existingPackagesBuffer.Len())
	a.logger.Debugf("New package content length: %d", len(newPackageContent))

	// Check if the package metadata already exists
	if !strings.Contains(existingPackagesBuffer.String(), newPackageContent) {
		// Append new package content to existing Packages buffer
		if existingPackagesBuffer.Len() > 0 {
			existingPackagesBuffer.WriteString("\n")
		}
		existingPackagesBuffer.WriteString(newPackageContent)
	} else {
		a.logger.Info("Package metadata already exists in Packages file; skipping append")
	}

	// Proceed to upload and compress the Packages file
	err = a.storage.UploadBuffer(ctx, a.packagesFilePath, existingPackagesBuffer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upload Packages file: %w", err)
	}

	// Compress the Packages file into Packages.gz
	packagesGzBuffer, err := compressGzip(existingPackagesBuffer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compress Packages.gz: %v", err)
	}

	// Upload the Packages.gz file
	packagesGzPath := a.packagesFilePath + ".gz"
	err = a.storage.UploadBuffer(ctx, packagesGzPath, packagesGzBuffer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upload Packages.gz file: %v", err)
	}

	return existingPackagesBuffer, packagesGzBuffer, nil
}

func (a *applicationImpl) UploadReleaseFile(ctx context.Context, packagesBuffer, packagesGzBuffer *bytes.Buffer) error {
	// Initialize the SHA256 slice
	var checksums []deb.ChecksumInfo

	// Ensure packagesBuffer is not empty
	if packagesBuffer.Len() == 0 {
		return fmt.Errorf("packages buffer is empty; cannot generate Release file")
	}

	// Compute checksum and size for Packages
	packagesHash := sha256.New()
	_, err := packagesHash.Write(packagesBuffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to compute sha256 for Packages: %v", err)
	}
	packagesSha256Sum := fmt.Sprintf("%x", packagesHash.Sum(nil))
	packagesSize := int64(packagesBuffer.Len())

	a.logger.Debugf("Packages buffer length before checksum: %d", packagesBuffer.Len())
	a.logger.Debugf("Packages.gz buffer length before checksum: %d", packagesGzBuffer.Len())

	checksums = append(checksums, deb.ChecksumInfo{
		Checksum: packagesSha256Sum,
		Size:     packagesSize,
		Filename: "Packages",
	})

	// Compute checksum and size for Packages.gz
	packagesGzHash := sha256.New()
	_, err = packagesGzHash.Write(packagesGzBuffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to compute sha256 for Packages.gz: %v", err)
	}
	packagesGzSha256Sum := fmt.Sprintf("%x", packagesGzHash.Sum(nil))
	packagesGzSize := int64(packagesGzBuffer.Len())

	checksums = append(checksums, deb.ChecksumInfo{
		Checksum: packagesGzSha256Sum,
		Size:     packagesGzSize,
		Filename: "Packages.gz",
	})

	// Construct the Release file content
	releaseContent := deb.CreateReleaseFileContents(deb.ReleaseFileContent{
		Component:    a.config.Component,
		Origin:       a.config.Origin,
		Label:        a.config.Label,
		Architecture: a.config.Architecture,
		SHA256:       checksums,
	})

	// Create a buffer for the Release file
	releaseBuffer := bytes.NewBufferString(releaseContent)

	// Upload the Release file
	err = a.storage.UploadBuffer(ctx, a.releaseFilePath, releaseBuffer)
	if err != nil {
		return fmt.Errorf("failed to upload Release file: %v", err)
	}

	return nil
}
