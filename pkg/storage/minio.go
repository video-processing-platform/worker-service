package storage

import (
	"context"
	"github.com/alimarzban99/video-processor-service/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	Client *minio.Client
	Bucket string
}

func NewMinio() (*MinioStorage, error) {
	client, err := minio.New(config.Cfg.Storage.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Cfg.Storage.AccessKey, config.Cfg.Storage.SecretKey, ""),
		Secure: config.Cfg.Storage.UseSSL,
	})

	if err != nil {
		return nil, err
	}

	return &MinioStorage{
		Client: client,
		Bucket: config.Cfg.Storage.Bucket,
	}, nil
}

func (m *MinioStorage) UploadFile(ctx context.Context, objectName, filePath string) error {

	_, err := m.Client.FPutObject(
		ctx,
		m.Bucket,
		objectName,
		filePath,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)

	return err
}

func (m *MinioStorage) DownloadFile(ctx context.Context, objectName, filePath string) error {

	return m.Client.FGetObject(
		ctx,
		m.Bucket,
		objectName,
		filePath,
		minio.GetObjectOptions{},
	)
}

func (m *MinioStorage) DeleteObject(ctx context.Context, objectName string) error {

	return m.Client.RemoveObject(
		ctx,
		m.Bucket,
		objectName,
		minio.RemoveObjectOptions{},
	)
}
