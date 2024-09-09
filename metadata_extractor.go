package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/blakesmith/ar"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

type DefaultMetadataExtractor struct {
	logger *log.Entry
}

// Extract reads metadata from a .deb file.
func (d *DefaultMetadataExtractor) Extract(file *os.File) (pkgName, pkgVersion, arch string, err error) {
	arReader := ar.NewReader(file)
	d.logger.Debug("Starting extraction from .deb file.")

	for {
		header, err := arReader.Next()
		if err != nil {
			d.logger.WithError(err).Error("Failed to read ar archive.")
			return "", "", "", fmt.Errorf("failed to read ar archive: %v", err)
		}

		// Log the name of each file in the .deb archive
		d.logger.Debugf("Found file in .deb archive: %s", header.Name)

		if header.Name == "control.tar.gz" {
			d.logger.Debug("Found control.tar.gz, attempting to read...")

			gzReader, err := gzip.NewReader(arReader)
			if err != nil {
				d.logger.WithError(err).Error("Failed to read control.tar.gz file.")
				return "", "", "", fmt.Errorf("failed to read gzip file: %v", err)
			}
			defer gzReader.Close()

			// Pass the gzipped tar reader to extract control metadata
			return extractControlMetadata(tar.NewReader(gzReader), d.logger)
		}
	}
}

func extractControlMetadata(tarReader *tar.Reader, logger *log.Entry) (pkgName, pkgVersion, arch string, err error) {
	logger.Debug("Extracting control metadata from control.tar.gz.")

	for {
		tarHeader, err := tarReader.Next()
		if err != nil {
			logger.WithError(err).Error("Failed to read tar archive.")
			return "", "", "", fmt.Errorf("failed to read tar archive: %v", err)
		}

		// Log the name of each file in control.tar.gz
		logger.Debugf("Found file in control.tar.gz: %s", tarHeader.Name)

		if tarHeader.Name == "./control" {
			// Log the expected file size
			logger.Debugf("Control file expected size: %d bytes", tarHeader.Size)

			controlData := make([]byte, tarHeader.Size)
			bytesRead, err := tarReader.Read(controlData)

			// Log the number of bytes successfully read
			logger.Debugf("Successfully read %d bytes from control file", bytesRead)

			// Handle EOF gracefully if the file has been fully read
			if err != nil && err != io.EOF {
				logger.WithError(err).Error("Failed to read control file.")
				return "", "", "", fmt.Errorf("failed to read control file: %v", err)
			}

			// Ensure we read the expected number of bytes
			if bytesRead != int(tarHeader.Size) {
				logger.Error("Control file size mismatch.")
				return "", "", "", fmt.Errorf("control file size mismatch: expected %d bytes, got %d bytes", tarHeader.Size, bytesRead)
			}

			// Log the control file content for debugging
			logger.Debug("Control file content:")
			logger.Debug(string(controlData))

			// Parse the control file content
			return parseControlFile(string(controlData), logger)
		}
	}
}

func parseControlFile(controlText string, logger *log.Entry) (pkgName, pkgVersion, arch string, err error) {
	logger.Debug("Parsing control file content.")
	lines := strings.Split(controlText, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Package:") {
			pkgName = strings.TrimSpace(strings.Split(line, ":")[1])
			logger.Debugf("Parsed package name: %s", pkgName)
		} else if strings.HasPrefix(line, "Version:") {
			pkgVersion = strings.TrimSpace(strings.Split(line, ":")[1])
			logger.Debugf("Parsed package version: %s", pkgVersion)
		} else if strings.HasPrefix(line, "Architecture:") {
			arch = strings.TrimSpace(strings.Split(line, ":")[1])
			logger.Debugf("Parsed architecture: %s", arch)
		}
	}
	if pkgName == "" || pkgVersion == "" || arch == "" {
		logger.Error("Incomplete control metadata.")
		return "", "", "", fmt.Errorf("incomplete control metadata")
	}
	return pkgName, pkgVersion, arch, nil
}
