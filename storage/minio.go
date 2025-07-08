package storage

import (
	"crypto/tls"
	"net/http"

	"github.com/tajimyradov/transcoder/models"

	"github.com/minio/minio-go/v7"
	minioCred "github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinio(cfg *models.Minio) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.EndPoint, &minio.Options{
		Creds:  minioCred.NewStaticV4(cfg.AccessKey, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSsl,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return minioClient, nil
}
