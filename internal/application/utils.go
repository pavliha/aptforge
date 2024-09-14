package application

import (
	"bytes"
	"compress/gzip"
	"github.com/pavliha/aptforge/internal/deb"
)

func compressGzip(data *bytes.Buffer) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err := gzWriter.Write(data.Bytes())
	if err != nil {
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}
	return &buf, nil
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
