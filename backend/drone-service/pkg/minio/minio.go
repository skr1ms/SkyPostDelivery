package minio

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/skr1ms/SkyPostDelivery/drone-service/config"
)

type Client struct {
	client     *minio.Client
	bucketName string
}

func New(cfg *config.MinIO) (*Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.RootUser, cfg.RootPassword, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("MinioClient - New - New: %w", err)
	}

	return &Client{
		client:     minioClient,
		bucketName: cfg.BucketRecords,
	}, nil
}

func (c *Client) UploadFrame(ctx context.Context, droneID, deliveryID string, frameData []byte, frameNumber int) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	objectName := fmt.Sprintf("%s/%s/%s_frame_%d.jpg", droneID, deliveryID, timestamp, frameNumber)

	reader := bytes.NewReader(frameData)

	_, err := c.client.PutObject(ctx, c.bucketName, objectName, reader, int64(len(frameData)), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", fmt.Errorf("MinioClient - UploadFrame - PutObject: %w", err)
	}

	return objectName, nil
}
