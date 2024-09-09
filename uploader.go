package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

type S3Uploader struct {
	Client MinioUploader
	logger *log.Entry
}

type MinioUploader interface {
	PutObject(ctx context.Context, bucketName string, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (info minio.UploadInfo, err error)
}

func NewUploader(logger *log.Entry, minioClient MinioUploader) *S3Uploader {
	return &S3Uploader{
		Client: minioClient,
		logger: logger,
	}
}

// Upload uploads a file to an S3-compatible bucket.
func (s *S3Uploader) Upload(ctx context.Context, bucket, s3Key string, file *os.File) error {
	// Get file information
	fileInfo, err := file.Stat()
	if err != nil {
		s.logger.WithError(err).Error("Failed to stat file")
		return fmt.Errorf("failed to stat file: %v", err)
	}

	s.logger.Debugf("File size: %d bytes", fileInfo.Size())

	// Reset the file pointer to the start
	if _, err := file.Seek(0, os.SEEK_SET); err != nil {
		s.logger.WithError(err).Error("Failed to reset file pointer")
		return fmt.Errorf("failed to reset file pointer: %v", err)
	}

	// Debug: Log the S3 upload path
	s.logger.Debugf("Uploading file to S3 at path: %s/%s", bucket, s3Key)

	// Perform the file upload
	_, err = s.Client.PutObject(ctx, bucket, s3Key, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		s.logger.WithError(err).Error("Failed to upload file")
		return fmt.Errorf("failed to upload file: %v", err)
	}

	s.logger.Infof("File successfully uploaded to %s/%s", bucket, s3Key)
	return nil
}

// UploadBuffer Uploader method to upload a buffer (for metadata files like Packages or Release)
func (s *S3Uploader) UploadBuffer(ctx context.Context, bucket, s3Key string, buffer *bytes.Buffer) error {
	size := int64(buffer.Len())

	s.logger.Debugf("Uploading buffer to S3 at path: %s/%s", bucket, s3Key)

	// Upload the buffer
	_, err := s.Client.PutObject(ctx, bucket, s3Key, buffer, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		s.logger.WithError(err).Error("Failed to upload buffer")
		return fmt.Errorf("failed to upload buffer: %v", err)
	}

	s.logger.Infof("Buffer successfully uploaded to %s/%s", bucket, s3Key)
	return nil
}
