package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
)

type LocalStorageService struct {
	basePath string
}

func NewLocalStorageService(basePath string) (*LocalStorageService, error) {
	// Create base directory with secure permissions (owner-only access)
	if err := os.MkdirAll(basePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorageService{basePath: basePath}, nil
}

func (s *LocalStorageService) Upload(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	// Create bucket directory with secure permissions (owner-only access)
	bucketPath := filepath.Join(s.basePath, bucketName)
	if err := os.MkdirAll(bucketPath, 0700); err != nil {
		return minio.UploadInfo{}, fmt.Errorf("failed to create bucket directory: %w", err)
	}

	// Create full path for the object
	objectPath := filepath.Join(bucketPath, objectName)

	// Create directory for the object with secure permissions
	objectDir := filepath.Dir(objectPath)
	if err := os.MkdirAll(objectDir, 0700); err != nil {
		return minio.UploadInfo{}, fmt.Errorf("failed to create object directory: %w", err)
	}

	// Create file with secure permissions (owner read/write only)
	file, err := os.OpenFile(objectPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data to file
	written, err := io.Copy(file, reader)
	if err != nil {
		return minio.UploadInfo{}, fmt.Errorf("failed to write file: %w", err)
	}

	return minio.UploadInfo{
		Size: written,
		Key:  objectName,
	}, nil
}

func (s *LocalStorageService) Download(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	objectPath := filepath.Join(s.basePath, bucketName, objectName)
	
	// Check if file exists
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("object not found: %s", objectName)
	}

	// Open and return the file
	file, err := os.Open(objectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

func (s *LocalStorageService) Delete(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	objectPath := filepath.Join(s.basePath, bucketName, objectName)
	
	if err := os.Remove(objectPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
