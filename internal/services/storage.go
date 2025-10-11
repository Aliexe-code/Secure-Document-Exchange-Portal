package services

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageService interface {
	Upload(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	Download(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	Delete(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
}

type MinIOService struct {
	client *minio.Client
}

func NewMinIOService(endpoint, accessKey, secretKey string, useSSL bool) (*MinIOService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOService{client: client}, nil
}

func (s *MinIOService) Upload(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return s.client.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
}

func (s *MinIOService) Download(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return s.client.GetObject(ctx, bucketName, objectName, opts)
}

func (s *MinIOService) Delete(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	return s.client.RemoveObject(ctx, bucketName, objectName, opts)
}
