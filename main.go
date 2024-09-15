package main

import (
	"context"
	"github.com/pavliha/aptforge/cmd"
	"github.com/pavliha/aptforge/internal/application"
	"github.com/pavliha/aptforge/internal/deb"
	"github.com/pavliha/aptforge/internal/storage"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func main() {
	logger := log.New()
	ctx := context.Background()

	// Execute Cobra command parsing
	config, err := cmd.Execute()
	if err != nil {
		logger.Fatalf("Failed to parse command line arguments: %v", err)
	}

	logger.SetLevel(log.DebugLevel)

	// Fallback to environment variables for an access key and secret key
	if config.AccessKey == "" {
		config.AccessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if config.SecretKey == "" {
		config.SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	app := application.New(logger.WithField("pkg", "application"), &application.Config{
		Storage: &storage.Config{
			Endpoint:  config.Endpoint,
			AccessKey: config.AccessKey,
			SecretKey: config.SecretKey,
			Bucket:    config.Bucket,
			Secure:    config.Secure,
		},
		Component:    config.Component,
		Origin:       config.Origin,
		Label:        config.Label,
		Architecture: config.Architecture,
		Archive:      config.Archive,
	})

	// Load and extract the .deb metadata
	file, err := app.LoadDebFile(config.FilePath)
	if err != nil {
		logger.Fatalf("Failed to load .deb file: %v", err)
	}
	defer app.CloseFile(file)

	packageMetadata, err := app.ExtractDebMetadata(file)
	if err != nil {
		logger.Fatalf("Failed to extract metadata: %v", err)
	}

	// Upload the .deb file
	err = app.UploadDebFile(ctx, packageMetadata, file)
	if err != nil {
		logger.Fatalf("Failed to upload .deb file: %v", err)
	}

	logger.Infof("Updating repository metadata...")
	repoPath := deb.ConstructRepoPath(config.Archive, config.Component, config.Architecture)
	packagesFilePath := filepath.Join(repoPath, "Packages")

	// Update the Packages file and upload both architecture-specific and high-level Release files
	packagesBuffer, packagesGzBuffer, err := app.UpdatePackagesFile(ctx, packagesFilePath, packageMetadata)
	if err != nil {
		logger.Fatalf("Failed to update Packages file: %v", err)
	}

	// Upload an architecture-specific Release file
	err = app.UploadPackageReleaseFile(ctx, filepath.Join(repoPath, "Release"), packagesBuffer, packagesGzBuffer)
	if err != nil {
		logger.Fatalf("Failed to upload architecture-specific Release file: %v", err)
	}

	// Upload suite-level Release file
	err = app.UploadSuiteReleaseFile(ctx, filepath.Join("dists", config.Archive, "Release"), []string{"amd64", "arm64"}, []string{"main", "contrib"})
	if err != nil {
		logger.Fatalf("Failed to upload suite-level Release file: %v", err)
	}

	logger.Infof("File uploaded successfully to %s\n", config.Bucket)
}
