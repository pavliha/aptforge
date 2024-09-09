package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Set up logrus logger with debug level for testing
	logger := log.New()
	logger.SetLevel(log.DebugLevel)
	entry := log.NewEntry(logger)
	// Parse command-line flags
	filePath, bucket, accessKey, secretKey, endpoint := parseFlags()

	// Load the file
	fileReader := &DefaultFileReader{}
	file, err := fileReader.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Extract metadata
	metadataExtractor := &DefaultMetadataExtractor{
		logger: entry.WithField("service", "extractor"),
	}
	pkgName, pkgVersion, arch, err := metadataExtractor.Extract(file)
	if err != nil {
		log.Fatalf("Failed to extract metadata: %v", err)
	}

	// Generate the S3 key
	s3Key := generateS3Key(pkgName, pkgVersion, arch)

	// Initialize MinIO client
	minioClient, err := initMinioClient(endpoint, accessKey, secretKey)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	// Wrap the MinIO client
	uploader := NewUploader(logger.WithField("service", "uploader"), minioClient)
	// Upload the file
	err = uploader.Upload(context.Background(), bucket, s3Key, file)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	fmt.Printf("File uploaded successfully to %s/%s\n", bucket, s3Key)
}
