package ec2

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestCreateInstance_Success(t *testing.T) {
	ctx := context.Background()
	expectedInstanceID := "i-1234567890abcdef0"

	mockClient := &MockEC2Client{
		RunInstancesFunc: func(ctx context.Context, params *awsec2.RunInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.RunInstancesOutput, error) {
			return &awsec2.RunInstancesOutput{
				Instances: []types.Instance{
					{
						InstanceId: aws.String(expectedInstanceID),
						State: &types.InstanceState{
							Name: types.InstanceStateNamePending,
						},
					},
				},
			}, nil
		},
	}

	instanceID, err := CreateInstance(ctx, mockClient)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if instanceID != expectedInstanceID {
		t.Errorf("expected instance ID %s, got %s", expectedInstanceID, instanceID)
	}
}

func TestCreateInstance_RunInstancesError(t *testing.T) {
	ctx := context.Background()
	expectedError := fmt.Errorf("AWS API error: insufficient capacity")

	mockClient := &MockEC2Client{
		RunInstancesFunc: func(ctx context.Context, params *awsec2.RunInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.RunInstancesOutput, error) {
			return nil, expectedError
		},
	}

	instanceID, err := CreateInstance(ctx, mockClient)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if instanceID != "" {
		t.Errorf("expected empty instance ID on error, got %s", instanceID)
	}
	if !strings.Contains(err.Error(), "failed to run instance") {
		t.Errorf("expected error message to contain 'failed to run instance', got %v", err)
	}
}
