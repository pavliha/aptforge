package application

import (
	"aptforge/internal/deb"
	"aptforge/internal/file_reader"
	"aptforge/internal/storage"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
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
	LoadDebFile(filePath string) (file_reader.File, error)
	CloseFile(file file_reader.File)
	ExtractDebMetadata(file file_reader.File) (*deb.PackageMetadata, error)
	UploadDebFile(ctx context.Context, metadata *deb.PackageMetadata, file file_reader.File) error
	DownloadPackagesFromStorage(ctx context.Context) (*bytes.Buffer, error)
	UpdatePackagesFile(ctx context.Context, metadata *deb.PackageMetadata) (*bytes.Buffer, error)
	UploadReleaseFile(ctx context.Context, packagesBuffer *bytes.Buffer) error
}

type applicationImpl struct {
	storage          *storage.Storage
	logger           *log.Entry
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
		repoPath:         repoPath,
		packagesFilePath: filepath.Join(repoPath, "Packages"),
		releaseFilePath:  filepath.Join(repoPath, "Release"),
	}
}

func (a *applicationImpl) LoadDebFile(filePath string) (file_reader.File, error) {
	fileReader := file_reader.New(a.logger.WithField("pkg", "file"))
	file, err := fileReader.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func (a *applicationImpl) CloseFile(file file_reader.File) {
	err := file.Close()
	if err != nil {
		a.logger.Errorf("Failed to close file: %v", err)
	}
}

func (a *applicationImpl) ExtractDebMetadata(file file_reader.File) (*deb.PackageMetadata, error) {
	metadata, err := deb.New(a.logger.WithField("pkg", "deb")).ExtractPackageMetadata(file)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	return metadata, nil
}

func (a *applicationImpl) UploadDebFile(ctx context.Context, metadata *deb.PackageMetadata, file file_reader.File) error {
	debPath := deb.GeneratePoolPath(a.config.Component, metadata)
	err := a.storage.UploadFile(ctx, debPath, file)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}

func (a *applicationImpl) DownloadPackagesFromStorage(ctx context.Context) (*bytes.Buffer, error) {
	packagesFilePath := filepath.Join(a.repoPath, "Packages")

	var packagesBuffer bytes.Buffer
	packagesFile, err := a.storage.Download(ctx, packagesFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			a.logger.Println("No Packages file found, creating a new one.")
			return &packagesBuffer, nil
		}
		return nil, fmt.Errorf("failed to download Packages file: %w", err)
	}
	_, err = io.Copy(&packagesBuffer, packagesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read Packages file: %w", err)
	}
	return &packagesBuffer, nil
}

func (a *applicationImpl) UpdatePackagesFile(ctx context.Context, metadata *deb.PackageMetadata) (*bytes.Buffer, error) {
	packagesFilePath := filepath.Join(a.repoPath, "Packages")
	packagesFile, err := a.DownloadPackagesFromStorage(ctx)
	if err != nil {
		return nil, err
	}
	_, err = packagesFile.WriteString(deb.CreatePackagesFileContents(mapMetadataToPackageContents(metadata)))
	if err != nil {
		return nil, fmt.Errorf("failed to append to Packages file: %w", err)
	}
	err = a.storage.UploadBuffer(ctx, packagesFilePath, packagesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to upload Packages file: %w", err)
	}
	return packagesFile, nil
}

func (a *applicationImpl) UploadReleaseFile(ctx context.Context, packagesBuffer *bytes.Buffer) error {
	releaseFilePath := filepath.Join(a.repoPath, "Release")
	// Compute the SHA256 checksum of the Packages file
	hash := sha256.New()
	if _, err := io.Copy(hash, bytes.NewReader(packagesBuffer.Bytes())); err != nil {
		return fmt.Errorf("failed to compute sha256 for Packages: %v", err)
	}
	sha256Sum := fmt.Sprintf("%x", hash.Sum(nil))

	// Get the size of the Packages file
	packagesSize := int64(packagesBuffer.Len())

	releaseContent := deb.CreateReleaseFileContents(deb.ReleaseFileContent{
		Component:    a.config.Component,
		Origin:       a.config.Origin,
		Label:        a.config.Label,
		Architecture: a.config.Architecture,
		SHA256: []deb.ChecksumInfo{{
			Checksum: sha256Sum,
			Size:     packagesSize,
			Filename: filepath.Base(releaseFilePath),
		}},
		PackagesPath: a.packagesFilePath,
	})

	// Create a buffer for the Release file
	releaseBuffer := bytes.NewBufferString(releaseContent)

	// Upload the Release file
	err := a.storage.UploadBuffer(ctx, releaseFilePath, releaseBuffer)
	if err != nil {
		return fmt.Errorf("failed to upload Release file: %v", err)
	}

	return nil
}

func mapMetadataToPackageContents(metadata *deb.PackageMetadata) *deb.PackagesContent {
	return &deb.PackagesContent{
		PackageName:   metadata.PackageName,
		Version:       metadata.Version,
		Architecture:  metadata.Architecture,
		Maintainer:    metadata.Maintainer,
		Description:   metadata.Description,
		Section:       metadata.Section,
		Priority:      metadata.Priority,
		InstalledSize: metadata.InstalledSize,
		Depends:       metadata.Depends,
		Recommends:    metadata.Recommends,
		Suggests:      metadata.Suggests,
		Conflicts:     metadata.Conflicts,
		Provides:      metadata.Provides,
	}
}
