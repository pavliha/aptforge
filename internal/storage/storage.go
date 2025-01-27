package storage

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/pavliha/aptforge/internal/filereader"
	log "github.com/sirupsen/logrus"
	"io"
)

import "errors"

var ErrNotFound = errors.New("object not found")

type Storage interface {
	UploadFile(ctx context.Context, s3Key string, file filereader.File) error
	UploadBuffer(ctx context.Context, s3Key string, buffer *bytes.Buffer) error
	Download(ctx context.Context, s3Key string) (Object, error)
	DownloadFile(ctx context.Context, s3Key string, dest *bytes.Buffer) error
}

type storageImpl struct {
	client MinioClient
	logger *log.Entry
	bucket string
}

type MinioClient interface {
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (info minio.UploadInfo, err error)
	GetObject(ctx context.Context, bucketName string, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
}

type Object interface {
	Read(p []byte) (n int, err error)
}

func New(logger *log.Entry, minioClient MinioClient, bucket string) Storage {
	return &storageImpl{
		client: minioClient,
		logger: logger,
		bucket: bucket,
	}
}

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Secure    bool
}

func Initialize(logger *log.Entry, config *Config) Storage {
	minioClient, err := InitMinioClient(config.Endpoint, config.AccessKey, config.SecretKey, config.Secure)
	if err != nil {
		logger.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	return New(logger, minioClient, config.Bucket)
}

// UploadFile uploads a file to an S3-compatible bucket.
func (s *storageImpl) UploadFile(ctx context.Context, s3Key string, file filereader.File) error {
	// Get file information
	fileInfo, err := file.Stat()
	if err != nil {
		s.logger.WithError(err).Error("Failed to stat file")
		return fmt.Errorf("failed to stat file: %v", err)
	}

	s.logger.Debugf("File size: %d bytes", fileInfo.Size())

	// Reset the file pointer to the start
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		s.logger.WithError(err).Error("Failed to reset file pointer")
		return fmt.Errorf("failed to reset file pointer: %v", err)
	}

	// Debug: Log the S3 upload path
	s.logger.Debugf("Uploading file to S3 at path: %s/%s", s.bucket, s3Key)

	// Perform the file upload
	_, err = s.client.PutObject(ctx, s.bucket, s3Key, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		s.logger.WithError(err).Error("Failed to upload file")
		return fmt.Errorf("failed to upload file: %v", err)
	}

	s.logger.Infof("File successfully uploaded to %s/%s", s.bucket, s3Key)
	return nil
}

// UploadBuffer Uploader method to upload a buffer (for metadata files like Packages or Release)
func (s *storageImpl) UploadBuffer(ctx context.Context, s3Key string, buffer *bytes.Buffer) error {
	size := int64(buffer.Len())

	s.logger.Debugf("Uploading buffer to S3 at path: %s/%s", s.bucket, s3Key)

	// Create a new reader from the buffer's bytes
	reader := bytes.NewReader(buffer.Bytes())

	// Upload the buffer without consuming it
	_, err := s.client.PutObject(ctx, s.bucket, s3Key, reader, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		s.logger.WithError(err).Error("Failed to upload buffer")
		return fmt.Errorf("failed to upload buffer: %v", err)
	}

	s.logger.Infof("Buffer successfully uploaded to %s/%s", s.bucket, s3Key)
	return nil
}

// Download downloads a file from an S3-compatible bucket and saves it to the provided file path.
func (s *storageImpl) Download(ctx context.Context, s3Key string) (Object, error) {
	s.logger.Debugf("Downloading file from S3 at path: %s/%s", s.bucket, s3Key)

	// Get the object from S3
	object, err := s.client.GetObject(ctx, s.bucket, s3Key, minio.GetObjectOptions{})
	if err != nil {
		s.logger.WithError(err).Error("Failed to download file from S3")
		return nil, fmt.Errorf("failed to download file from S3: %v", err)
	}

	return object, nil
}

func (s *storageImpl) DownloadFile(ctx context.Context, s3Key string, dest *bytes.Buffer) error {
	s.logger.Debugf("Downloading file from S3 at path: %s/%s", s.bucket, s3Key)

	// Get the object from S3
	object, err := s.client.GetObject(ctx, s.bucket, s3Key, minio.GetObjectOptions{})
	if err != nil {
		s.logger.WithError(err).Error("Failed to get object from S3")
		return fmt.Errorf("failed to get object from S3: %v", err)
	}
	defer func(object *minio.Object) {
		err := object.Close()
		if err != nil {
			s.logger.WithError(err).Error("Failed to close object")
		}
	}(object)

	// Stat the object to check if it exists
	_, err = object.Stat()
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			s.logger.Info("Object does not exist in S3")
			return ErrNotFound
		}
		return fmt.Errorf("failed to stat object: %v", err)
	}

	// Copy the content to the destination buffer
	_, err = io.Copy(dest, object)
	if err != nil {
		return fmt.Errorf("failed to read downloaded file: %v", err)
	}

	return nil
}

func IsNotFoundError(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}
	errResp := minio.ToErrorResponse(err)
	return errResp.StatusCode == 404 || errResp.Code == "NoSuchKey"
}
