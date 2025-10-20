package services

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageService interface {
	Upload(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	Download(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error)
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

func (s *MinIOService) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	return s.client.ListBuckets(ctx)
}

func (s *MinIOService) EnsureBucket(ctx context.Context, bucketName string) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MinIOService) Upload(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return s.client.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
}

func (s *MinIOService) Download(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *MinIOService) Delete(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	return s.client.RemoveObject(ctx, bucketName, objectName, opts)
}
