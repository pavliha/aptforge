package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
)

// parseFlags parses command-line arguments and environment variables.
func parseFlags() (filePath, bucket, accessKey, secretKey, endpoint string) {
	flag.StringVar(&filePath, "file", "", "Path to the file to upload")
	flag.StringVar(&bucket, "bucket", "", "Name of the S3 bucket")
	flag.StringVar(&accessKey, "access-key", "", "Access Key")
	flag.StringVar(&secretKey, "secret-key", "", "Secret Access Key")
	flag.StringVar(&endpoint, "endpoint", "", "S3-compatible endpoint (e.g., fra1.digitaloceanspaces.com)")
	flag.Parse()

	// Allow environment variables as fallback
	if accessKey == "" {
		accessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if secretKey == "" {
		secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	// Validate required inputs
	if filePath == "" || bucket == "" || accessKey == "" || secretKey == "" || endpoint == "" {
		log.Fatal("Missing required arguments: file, bucket, access-key, secret-key, endpoint")
	}

	return filePath, bucket, accessKey, secretKey, endpoint
}
