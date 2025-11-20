package minio

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/config"
)

type MinioClient interface {
	UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error
	GetFile(ctx context.Context, objectName string) (io.ReadCloser, error)
	GetFileURL(objectName string) string
	DeleteFile(ctx context.Context, objectName string) error
	EnsureBucket(ctx context.Context, bucketNames ...string) error
}

type Client struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

func New(cfg *config.MinIO) (*Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	client := &Client{
		client:    minioClient,
		bucket:    cfg.BucketQR,
		publicURL: cfg.PublicURL,
	}

	if err := client.EnsureBucket(context.Background(), cfg.BucketQR, cfg.BucketRecords); err != nil {
		return nil, fmt.Errorf("failed to ensure buckets exist: %w", err)
	}

	return client, nil
}

func (c *Client) ensureBucketExists(ctx context.Context, bucketName string) error {
	exists, err := c.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Created MinIO bucket: %s", bucketName)
	}

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetObject",
			"Resource": "arn:aws:s3:::%s/*"
		}]
	}`, bucketName)

	if err := c.client.SetBucketPolicy(ctx, bucketName, policy); err != nil {
		log.Printf("Warning: failed to set bucket policy (bucket may not be public): %v", err)
	}

	return nil
}

func (c *Client) EnsureBucket(ctx context.Context, bucketNames ...string) error {
	for _, bucketName := range bucketNames {
		if err := c.ensureBucketExists(ctx, bucketName); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := c.client.PutObject(ctx, c.bucket, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (c *Client) GetFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	object, err := c.client.GetObject(ctx, c.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio - GetFile: %w", err)
	}
	return object, nil
}

func (c *Client) GetFileURL(objectName string) string {
	return fmt.Sprintf("%s/%s/%s", c.publicURL, c.bucket, objectName)
}

func (c *Client) DeleteFile(ctx context.Context, objectName string) error {
	err := c.client.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
