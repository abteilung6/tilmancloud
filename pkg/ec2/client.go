package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
)

type EC2Client interface {
	RunInstances(ctx context.Context, params *awsec2.RunInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.RunInstancesOutput, error)
	DescribeInstances(ctx context.Context, params *awsec2.DescribeInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeInstancesOutput, error)
	TerminateInstances(ctx context.Context, params *awsec2.TerminateInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.TerminateInstancesOutput, error)
	DescribeImages(ctx context.Context, params *awsec2.DescribeImagesInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeImagesOutput, error)
}

func NewClient(ctx context.Context, region string) (EC2Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := awsec2.NewFromConfig(cfg)
	return client, nil
}
