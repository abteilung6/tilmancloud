package image

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Importer struct {
	client *ec2.Client
	region string
}

func NewImporter(ctx context.Context, region string) (*Importer, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Importer{
		client: ec2.NewFromConfig(cfg),
		region: region,
	}, nil
}

func (i *Importer) ImportSnapshot(ctx context.Context, s3Bucket, s3Key, description string) (string, error) {
	slog.Info("Importing snapshot from S3", "bucket", s3Bucket, "key", s3Key, "description", description)

	result, err := i.client.ImportSnapshot(ctx, &ec2.ImportSnapshotInput{
		DiskContainer: &types.SnapshotDiskContainer{
			Format: aws.String("RAW"),
			UserBucket: &types.UserBucket{
				S3Bucket: aws.String(s3Bucket),
				S3Key:    aws.String(s3Key),
			},
		},
		Description: aws.String(description),
	})
	if err != nil {
		return "", fmt.Errorf("failed to initiate snapshot import: %w", err)
	}

	if result.ImportTaskId == nil {
		return "", fmt.Errorf("import task ID is nil")
	}

	taskID := *result.ImportTaskId
	slog.Info("Snapshot import initiated", "task_id", taskID)

	snapshotID, err := i.WaitForImport(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("wait for import failed: %w", err)
	}

	slog.Info("Snapshot import completed", "snapshot_id", snapshotID)
	return snapshotID, nil
}

func (i *Importer) WaitForImport(ctx context.Context, taskID string) (string, error) {
	slog.Info("Waiting for snapshot import to complete", "task_id", taskID)

	importCtx, cancel := context.WithTimeout(ctx, 60*time.Minute)
	defer cancel()

	waiter := ec2.NewSnapshotImportedWaiter(i.client)

	err := waiter.Wait(importCtx, &ec2.DescribeImportSnapshotTasksInput{
		ImportTaskIds: []string{taskID},
	}, 60*time.Minute)
	if err != nil {
		return "", fmt.Errorf("snapshot import timeout or failed: %w", err)
	}

	result, err := i.client.DescribeImportSnapshotTasks(importCtx, &ec2.DescribeImportSnapshotTasksInput{
		ImportTaskIds: []string{taskID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe import task: %w", err)
	}

	if len(result.ImportSnapshotTasks) == 0 {
		return "", fmt.Errorf("import task not found: %s", taskID)
	}

	task := result.ImportSnapshotTasks[0]
	if task.SnapshotTaskDetail == nil {
		return "", fmt.Errorf("snapshot task detail is nil")
	}

	if task.SnapshotTaskDetail.SnapshotId == nil {
		return "", fmt.Errorf("snapshot ID is nil")
	}

	snapshotID := *task.SnapshotTaskDetail.SnapshotId
	return snapshotID, nil
}

