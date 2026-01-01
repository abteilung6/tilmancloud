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

func (i *Importer) ImportSnapshot(ctx context.Context, s3Bucket, s3Key, description, imageID string) (string, error) {
	slog.Info("Checking for existing snapshot by ImageID", "image_id", imageID)
	existingID, err := i.FindSnapshotByImageID(ctx, imageID)
	if err != nil {
		return "", fmt.Errorf("failed to check for existing snapshot: %w", err)
	}

	if existingID != "" {
		slog.Info("Snapshot already exists, reusing", "snapshot_id", existingID, "image_id", imageID)
		return existingID, nil
	}

	slog.Info("No existing snapshot found, importing from S3", "bucket", s3Bucket, "key", s3Key, "image_id", imageID)

	result, err := i.client.ImportSnapshot(ctx, &ec2.ImportSnapshotInput{
		ClientToken: aws.String(imageID),
		DiskContainer: &types.SnapshotDiskContainer{
			Format: aws.String("RAW"),
			UserBucket: &types.UserBucket{
				S3Bucket: aws.String(s3Bucket),
				S3Key:    aws.String(s3Key),
			},
		},
		Description: aws.String(description),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeImportSnapshotTask,
				Tags: []types.Tag{
					{Key: aws.String("source_object_name"), Value: aws.String(s3Key)},
					{Key: aws.String("ImageID"), Value: aws.String(imageID)},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to initiate snapshot import: %w", err)
	}

	if result.ImportTaskId == nil {
		return "", fmt.Errorf("import task ID is nil")
	}

	taskID := *result.ImportTaskId
	slog.Info("Snapshot import initiated", "task_id", taskID, "image_id", imageID)

	snapshotID, err := i.WaitForImport(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("wait for import failed: %w", err)
	}

	_, err = i.client.CreateTags(ctx, &ec2.CreateTagsInput{
		Resources: []string{snapshotID},
		Tags: []types.Tag{
			{Key: aws.String("ImageID"), Value: aws.String(imageID)},
			{Key: aws.String("source_object_name"), Value: aws.String(s3Key)},
			{Key: aws.String("S3Bucket"), Value: aws.String(s3Bucket)},
			{Key: aws.String("S3Key"), Value: aws.String(s3Key)},
		},
	})
	if err != nil {
		slog.Warn("Failed to tag snapshot", "snapshot_id", snapshotID, "error", err)
	} else {
		slog.Info("Snapshot tagged with ImageID and S3 source", "snapshot_id", snapshotID, "image_id", imageID)
	}

	slog.Info("Snapshot import completed", "snapshot_id", snapshotID)
	return snapshotID, nil
}

func (i *Importer) FindSnapshotByImageID(ctx context.Context, imageID string) (string, error) {
	result, err := i.client.DescribeSnapshots(ctx, &ec2.DescribeSnapshotsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:ImageID"),
				Values: []string{imageID},
			},
			{
				Name:   aws.String("status"),
				Values: []string{"completed"},
			},
		},
		OwnerIds: []string{"self"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to query snapshots by ImageID: %w", err)
	}

	if len(result.Snapshots) == 0 {
		return "", nil
	}

	var latest *types.Snapshot
	for i := range result.Snapshots {
		snap := &result.Snapshots[i]
		if latest == nil || (snap.StartTime != nil && latest.StartTime != nil && snap.StartTime.After(*latest.StartTime)) {
			latest = snap
		}
	}

	if latest == nil || latest.SnapshotId == nil {
		return "", nil
	}

	return *latest.SnapshotId, nil
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
