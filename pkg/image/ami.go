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

type AMIRegistrar struct {
	client *ec2.Client
	region string
}

func NewAMIRegistrar(ctx context.Context, region string) (*AMIRegistrar, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AMIRegistrar{
		client: ec2.NewFromConfig(cfg),
		region: region,
	}, nil
}

type AMIFinder interface {
	FindLatestAMI(ctx context.Context) (string, error)
}

type ImageLister interface {
	ListImages(ctx context.Context) ([]types.Image, error)
}

func (r *AMIRegistrar) FindLatestAMI(ctx context.Context) (string, error) {
	slog.Info("Finding latest available AMI")
	result, err := r.client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
		},
		Owners: []string{"self"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to query AMIs: %w", err)
	}

	if len(result.Images) == 0 {
		return "", fmt.Errorf("no available AMIs found")
	}

	var latest *types.Image
	for i := range result.Images {
		img := &result.Images[i]
		if latest == nil {
			latest = img
			continue
		}
		if img.CreationDate != nil && latest.CreationDate != nil {
			if *img.CreationDate > *latest.CreationDate {
				latest = img
			}
		}
	}

	if latest == nil || latest.ImageId == nil {
		return "", fmt.Errorf("no available AMIs found")
	}

	slog.Info("Found latest AMI", "ami_id", *latest.ImageId, "creation_date", latest.CreationDate)
	return *latest.ImageId, nil
}

func (r *AMIRegistrar) FindAMIByImageID(ctx context.Context, imageID string) (string, error) {
	result, err := r.client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:ImageID"),
				Values: []string{imageID},
			},
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
		},
		Owners: []string{"self"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to query AMIs by ImageID: %w", err)
	}

	if len(result.Images) == 0 {
		return "", nil
	}

	var latest *types.Image
	for i := range result.Images {
		img := &result.Images[i]
		if latest == nil {
			latest = img
			continue
		}
		if img.CreationDate != nil && latest.CreationDate != nil {
			if *img.CreationDate > *latest.CreationDate {
				latest = img
			}
		}
	}

	if latest == nil || latest.ImageId == nil {
		return "", nil
	}

	return *latest.ImageId, nil
}

func (r *AMIRegistrar) RegisterAMI(ctx context.Context, snapshotID, imageID, name, description string) (string, error) {
	slog.Info("Checking for existing AMI by ImageID", "image_id", imageID)
	existingID, err := r.FindAMIByImageID(ctx, imageID)
	if err != nil {
		return "", fmt.Errorf("failed to check for existing AMI: %w", err)
	}

	if existingID != "" {
		slog.Info("AMI already exists, reusing", "ami_id", existingID, "image_id", imageID)
		return existingID, nil
	}

	slog.Info("No existing AMI found, registering new AMI", "snapshot_id", snapshotID, "name", name, "image_id", imageID)

	result, err := r.client.RegisterImage(ctx, &ec2.RegisterImageInput{
		Name:               aws.String(name),
		Description:        aws.String(description),
		Architecture:       types.ArchitectureValuesArm64,
		VirtualizationType: aws.String(string(types.VirtualizationTypeHvm)),
		RootDeviceName:     aws.String("/dev/xvda"),
		BlockDeviceMappings: []types.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/xvda"),
				Ebs: &types.EbsBlockDevice{
					SnapshotId:          aws.String(snapshotID),
					DeleteOnTermination: aws.Bool(true),
					VolumeType:          types.VolumeTypeGp3,
				},
			},
		},
		EnaSupport:      aws.Bool(true),
		SriovNetSupport: aws.String("simple"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to register AMI: %w", err)
	}

	if result.ImageId == nil {
		return "", fmt.Errorf("AMI ID is nil")
	}

	amiID := *result.ImageId
	slog.Info("AMI registration initiated", "ami_id", amiID, "image_id", imageID)

	_, err = r.client.CreateTags(ctx, &ec2.CreateTagsInput{
		Resources: []string{amiID},
		Tags: []types.Tag{
			{Key: aws.String("ImageID"), Value: aws.String(imageID)},
			{Key: aws.String("SnapshotID"), Value: aws.String(snapshotID)},
		},
	})
	if err != nil {
		slog.Warn("Failed to tag AMI", "ami_id", amiID, "error", err)
	} else {
		slog.Info("AMI tagged with ImageID", "ami_id", amiID, "image_id", imageID)
	}

	err = r.WaitForAvailable(ctx, amiID)
	if err != nil {
		return "", fmt.Errorf("wait for AMI available failed: %w", err)
	}

	slog.Info("AMI registration completed", "ami_id", amiID)
	return amiID, nil
}

func (r *AMIRegistrar) WaitForAvailable(ctx context.Context, amiID string) error {
	slog.Info("Waiting for AMI to become available", "ami_id", amiID)

	waitCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	waiter := ec2.NewImageAvailableWaiter(r.client)

	err := waiter.Wait(waitCtx, &ec2.DescribeImagesInput{
		ImageIds: []string{amiID},
	}, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("AMI did not become available: %w", err)
	}

	slog.Info("AMI is now available", "ami_id", amiID)
	return nil
}

func (r *AMIRegistrar) ListImages(ctx context.Context) ([]types.Image, error) {
	result, err := r.client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Owners: []string{"self"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query AMIs: %w", err)
	}

	return result.Images, nil
}
