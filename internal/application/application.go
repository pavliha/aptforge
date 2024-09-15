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
	UpdatePackagesFile(ctx context.Context, packagesPath string, metadata *deb.PackageMetadata) (*bytes.Buffer, *bytes.Buffer, error)
	UploadPackageReleaseFile(ctx context.Context, releasePath string, packagesBuffer, packagesGzBuffer *bytes.Buffer) error
	UploadSuiteReleaseFile(ctx context.Context, suiteReleasePath string, architectures, components []string) error
}

type applicationImpl struct {
	logger     *log.Entry
	storage    storage.Storage
	fileReader filereader.Reader
	extractor  deb.Extractor
	config     *Config
}

func New(logger *log.Entry, config *Config) Application {
	return &applicationImpl{
		logger:     logger,
		config:     config,
		storage:    storage.Initialize(logger, config.Storage),
		fileReader: filereader.New(logger.WithField("pkg", "file")),
		extractor:  deb.New(logger.WithField("pkg", "deb")),
	}
}

func (a *applicationImpl) LoadDebFile(filePath string) (filereader.File, error) {
	if filepath.Ext(filePath) != ".deb" {
		return nil, fmt.Errorf("file is not a .deb file: %s", filePath)
	}

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

func (a *applicationImpl) UpdatePackagesFile(ctx context.Context, packagesPath string, metadata *deb.PackageMetadata) (*bytes.Buffer, *bytes.Buffer, error) {
	// Download an existing Packages file
	existingPackagesBuffer, err := a.downloadPackagesFromStorage(ctx, packagesPath)
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
	err = a.storage.UploadBuffer(ctx, packagesPath, existingPackagesBuffer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upload Packages file: %w", err)
	}

	// Compress the Packages file into Packages.gz
	packagesGzBuffer, err := compressGzip(existingPackagesBuffer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compress Packages.gz: %v", err)
	}

	// Upload the Packages.gz file
	packagesGzPath := packagesPath + ".gz"
	err = a.storage.UploadBuffer(ctx, packagesGzPath, packagesGzBuffer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upload Packages.gz file: %v", err)
	}

	return existingPackagesBuffer, packagesGzBuffer, nil
}

func (a *applicationImpl) UploadPackageReleaseFile(ctx context.Context, releasePath string, packagesBuffer, packagesGzBuffer *bytes.Buffer) error {
	// Initialize the SHA256 slice
	var checksums []deb.ChecksumInfo

	// Compute checksum and size for Packages
	packagesHash := sha256.New()
	_, err := packagesHash.Write(packagesBuffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to compute sha256 for Packages: %v", err)
	}
	packagesSha256Sum := fmt.Sprintf("%x", packagesHash.Sum(nil))
	packagesSize := int64(packagesBuffer.Len())

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
	releaseContent := deb.CreatePackageReleaseFileContents(deb.ReleaseFileContent{
		Origin:       a.config.Origin,
		Label:        a.config.Label,
		Archive:      a.config.Archive,
		Component:    a.config.Component,
		Architecture: a.config.Architecture,
		SHA256:       checksums,
	})

	// Upload the Release file
	err = a.storage.UploadBuffer(ctx, releasePath, bytes.NewBufferString(releaseContent))
	if err != nil {
		return fmt.Errorf("failed to upload Release file: %v", err)
	}

	return nil
}

func (a *applicationImpl) UploadSuiteReleaseFile(ctx context.Context, suiteReleasePath string, architectures, components []string) error {
	// Construct the Release file content for the entire suite
	releaseContent := deb.CreateSuiteReleaseFileContents(deb.ReleaseFileContent{
		Origin:       a.config.Origin,
		Label:        a.config.Label,
		Archive:      a.config.Archive,
		Architecture: strings.Join(architectures, " "),
		Component:    strings.Join(components, " "),
	})

	// Upload the suite-level Release file
	err := a.storage.UploadBuffer(ctx, suiteReleasePath, bytes.NewBufferString(releaseContent))
	if err != nil {
		return fmt.Errorf("failed to upload suite-level Release file: %v", err)
	}

	return nil
}

func (a *applicationImpl) downloadPackagesFromStorage(ctx context.Context, packagesPath string) (*bytes.Buffer, error) {
	var packagesBuffer bytes.Buffer

	// Download the existing Packages file
	err := a.storage.DownloadFile(ctx, packagesPath, &packagesBuffer)
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
