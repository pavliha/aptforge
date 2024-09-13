package main

import (
	"aptforge/cmd"
	"aptforge/internal/application"
	"aptforge/internal/storage"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
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

	// Update the Packages file
	packagesBuffer, err := app.UpdatePackagesFile(ctx, packageMetadata)
	if err != nil {
		logger.Fatalf("Failed to update Packages file: %v", err)
	}

	// Generate and upload the Release file
	err = app.UploadReleaseFile(ctx, packagesBuffer)
	if err != nil {
		logger.Fatalf("Failed to upload Release file: %v", err)
	}

	fmt.Printf("File uploaded successfully to %s\n", config.Bucket)
}
