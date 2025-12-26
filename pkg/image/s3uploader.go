package image

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type S3Uploader struct {
	client *s3.Client
	bucket string
}

func NewS3Uploader(ctx context.Context, bucket, region string) (*S3Uploader, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &S3Uploader{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}, nil
}

func (u *S3Uploader) Upload(ctx context.Context, filePath, key string) error {
	if exists, _ := u.Exists(ctx, key); exists {
		slog.Info("File already exists in S3, skipping upload", "bucket", u.bucket, "key", key)
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	slog.Info("Uploading file to S3", "bucket", u.bucket, "key", key, "size", fileInfo.Size())

	uploader := manager.NewUploader(u.client)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("upload to S3 failed: %w", err)
	}

	slog.Info("File uploaded to S3 successfully", "bucket", u.bucket, "key", key)
	return nil
}

func (u *S3Uploader) Exists(ctx context.Context, key string) (bool, error) {
	_, err := u.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NotFound" {
				return false, nil
			}
		}
		return false, fmt.Errorf("check if object exists: %w", err)
	}
	return true, nil
}

func (u *S3Uploader) GetS3URL(key string) string {
	return fmt.Sprintf("s3://%s/%s", u.bucket, key)
}

func GenerateS3Key(filename string) string {
	return "images/" + filepath.Base(filename)
}
