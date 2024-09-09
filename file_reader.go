package main

import (
	"fmt"
	"os"
	"strings"
)

// Interfaces to abstract file operations and MinIO interactions for testability
type FileReader interface {
	Open(name string) (*os.File, error)
}

// Production implementations
type DefaultFileReader struct{}

// Open opens a file for reading.
func (d *DefaultFileReader) Open(name string) (*os.File, error) {
	return os.Open(name)
}

// Generate the S3 key based on APT repository structure
func generateS3Key(pkgName, pkgVersion, arch string) string {
	firstLetter := strings.ToLower(string(pkgName[0]))
	return fmt.Sprintf("pool/main/%s/%s/%s_%s_%s.deb", firstLetter, pkgName, pkgName, pkgVersion, arch)
}
