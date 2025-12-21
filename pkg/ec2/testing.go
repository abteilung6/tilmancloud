package ec2

import (
	"context"
	"fmt"

	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
)

type MockEC2Client struct {
	RunInstancesFunc       func(ctx context.Context, params *awsec2.RunInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.RunInstancesOutput, error)
	DescribeInstancesFunc  func(ctx context.Context, params *awsec2.DescribeInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeInstancesOutput, error)
	TerminateInstancesFunc func(ctx context.Context, params *awsec2.TerminateInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.TerminateInstancesOutput, error)
}

func (m *MockEC2Client) RunInstances(ctx context.Context, params *awsec2.RunInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.RunInstancesOutput, error) {
	if m.RunInstancesFunc != nil {
		return m.RunInstancesFunc(ctx, params, optFns...)
	}
	return nil, fmt.Errorf("RunInstancesFunc not set")
}

func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *awsec2.DescribeInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeInstancesOutput, error) {
	if m.DescribeInstancesFunc != nil {
		return m.DescribeInstancesFunc(ctx, params, optFns...)
	}
	return nil, fmt.Errorf("DescribeInstancesFunc not set")
}

func (m *MockEC2Client) TerminateInstances(ctx context.Context, params *awsec2.TerminateInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.TerminateInstancesOutput, error) {
	if m.TerminateInstancesFunc != nil {
		return m.TerminateInstancesFunc(ctx, params, optFns...)
	}
	return nil, fmt.Errorf("TerminateInstancesFunc not set")
}
