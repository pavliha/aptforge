package deb

import (
	"aptforge/internal/filereader"
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/blakesmith/ar"
	log "github.com/sirupsen/logrus"
	"io"
	"strings"
)

type Extractor interface {
	ExtractPackageMetadata(file filereader.File) (*PackageMetadata, error)
}

type PackageMetadata struct {
	PackageName   string
	Version       string
	Architecture  string
	Maintainer    string
	Description   string
	Section       string
	Priority      string
	InstalledSize string
	Depends       string
	Recommends    string
	Suggests      string
	Conflicts     string
	Provides      string
}

// DefaultMetadataExtractor is responsible for extracting metadata from .deb files.
type DefaultMetadataExtractor struct {
	logger *log.Entry
}

// New returns a new instance of DefaultMetadataExtractor.
func New(logger *log.Entry) Extractor {
	return &DefaultMetadataExtractor{
		logger: logger,
	}
}

// ExtractPackageMetadata reads metadata from a .deb file and returns it in a DebMetadata struct.
func (d *DefaultMetadataExtractor) ExtractPackageMetadata(file filereader.File) (*PackageMetadata, error) {
	arReader := ar.NewReader(file)
	d.logger.Debug("Starting extraction from .deb file.")

	for {
		header, err := arReader.Next()
		if err != nil {
			d.logger.WithError(err).Error("Failed to read ar archive.")
			return nil, fmt.Errorf("failed to read ar archive: %v", err)
		}

		// Log the name of each file in the .deb archive
		d.logger.Debugf("Found file in .deb archive: %s", header.Name)

		if header.Name == "control.tar.gz" {
			d.logger.Debug("Found control.tar.gz, attempting to read...")

			gzReader, err := gzip.NewReader(arReader)
			if err != nil {
				d.logger.WithError(err).Error("Failed to read control.tar.gz file.")
				return nil, fmt.Errorf("failed to read gzip file: %v", err)
			}
			defer func(gzReader *gzip.Reader) {
				err := gzReader.Close()
				if err != nil {
					d.logger.WithError(err).Error("Failed to close gzip file.")
				}
			}(gzReader)

			// Pass the gzipped tar reader to extract control metadata
			return d.extractControlMetadata(tar.NewReader(gzReader), d.logger)
		}
	}
}

func (d *DefaultMetadataExtractor) extractControlMetadata(tarReader *tar.Reader, logger *log.Entry) (*PackageMetadata, error) {
	logger.Debug("Extracting control metadata from control.tar.gz.")

	for {
		tarHeader, err := tarReader.Next()
		if err != nil {
			logger.WithError(err).Error("Failed to read tar archive.")
			return nil, fmt.Errorf("failed to read tar archive: %v", err)
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
				return nil, fmt.Errorf("failed to read control file: %v", err)
			}

			// Ensure we read the expected number of bytes
			if bytesRead != int(tarHeader.Size) {
				logger.Error("Control file size mismatch.")
				return nil, fmt.Errorf("control file size mismatch: expected %d bytes, got %d bytes", tarHeader.Size, bytesRead)
			}

			// Log the control file content for debugging
			logger.Debug("Control file content:")
			logger.Debug(string(controlData))

			// Parse the control file content
			return d.parseControlFile(string(controlData))
		}
	}
}

// parseControlFile parses the control file text and extracts relevant metadata fields.
func (d *DefaultMetadataExtractor) parseControlFile(controlText string) (*PackageMetadata, error) {
	d.logger.Debug("Parsing control file content.")
	var metadata PackageMetadata

	// Map for better handling of multiple control fields
	controlFields := map[string]*string{
		"Package":        &metadata.PackageName,
		"Version":        &metadata.Version,
		"Architecture":   &metadata.Architecture,
		"Maintainer":     &metadata.Maintainer,
		"Description":    &metadata.Description,
		"Section":        &metadata.Section,
		"Priority":       &metadata.Priority,
		"Installed-Size": &metadata.InstalledSize,
		"Depends":        &metadata.Depends,
		"Recommends":     &metadata.Recommends,
		"Suggests":       &metadata.Suggests,
		"Conflicts":      &metadata.Conflicts,
		"Provides":       &metadata.Provides,
	}

	// Split control text into lines and process each one
	lines := strings.Split(controlText, "\n")
	for _, line := range lines {
		// Skip empty lines or comment lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split line by first occurrence of colon (":")
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			d.logger.Warnf("Invalid control line: %s", line)
			continue
		}

		field := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Try to map the field to the corresponding metadata field
		if target, found := controlFields[field]; found {
			*target = value
			d.logger.Debugf("Parsed %s: %s", field, value)
		} else {
			d.logger.Warnf("Unrecognized control field: %s", field)
		}
	}

	// Ensure that required fields are present
	if metadata.PackageName == "" || metadata.Version == "" || metadata.Architecture == "" {
		d.logger.Error("Incomplete control metadata: missing essential fields.")
		return nil, fmt.Errorf("incomplete control metadata: PackageName, Version, or Architecture is missing")
	}

	d.logger.Debug("Successfully parsed control file.")
	return &metadata, nil
}
