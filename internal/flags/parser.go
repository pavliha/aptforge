package flags

import (
	"flag"
	log "github.com/sirupsen/logrus"
)

// Config holds the values parsed from command-line flags and environment variables.
type Config struct {
	FilePath     string
	Bucket       string
	AccessKey    string
	SecretKey    string
	Endpoint     string
	Component    string
	Origin       string
	Label        string
	Architecture string
	Archive      string
}

// Parse parses command-line arguments and environment variables.
func Parse() *Config {
	// Define flags for file upload and Release file metadata
	var filePath, bucket, accessKey, secretKey, endpoint string
	var component, origin, label, architecture, archive string

	// File upload flags
	flag.StringVar(&filePath, "file", "", "Path to the file to upload")
	flag.StringVar(&bucket, "bucket", "", "Name of the S3 bucket")
	flag.StringVar(&accessKey, "access-key", "", "Access Key")
	flag.StringVar(&secretKey, "secret-key", "", "Secret Access Key")
	flag.StringVar(&endpoint, "endpoint", "", "S3-compatible endpoint (e.g., fra1.digitaloceanspaces.com)")

	// Release file metadata flags
	flag.StringVar(&component, "component", "main", "Component of the APT repository (e.g., main, contrib, non-free)")
	flag.StringVar(&origin, "origin", "Custom Repository", "Origin of the APT repository")
	flag.StringVar(&label, "label", "Custom Repo", "Label for the APT repository")
	flag.StringVar(&architecture, "arch", "amd64", "Target architecture for the repository (e.g., amd64, arm64)")
	flag.StringVar(&archive, "archive", "stable", "Archive type of the APT repository (e.g., stable, testing, unstable)")

	flag.Parse()

	// Validate required inputs
	if filePath == "" || bucket == "" || accessKey == "" || secretKey == "" || endpoint == "" {
		log.Fatal("Missing required arguments: file, bucket, access-key, secret-key, endpoint")
	}

	// Return the configuration in a struct
	return &Config{
		FilePath:     filePath,
		Bucket:       bucket,
		AccessKey:    accessKey,
		SecretKey:    secretKey,
		Endpoint:     endpoint,
		Component:    component,
		Origin:       origin,
		Label:        label,
		Architecture: architecture,
		Archive:      archive,
	}
}
