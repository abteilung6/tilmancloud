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
	expectedState := types.InstanceStateNamePending
	expectedInstanceType := types.InstanceTypeT4gMicro
	expectedImageID := "ami-1234567890abcdef0"

	mockClient := &MockEC2Client{
		RunInstancesFunc: func(ctx context.Context, params *awsec2.RunInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.RunInstancesOutput, error) {
			if params.ImageId == nil || *params.ImageId != expectedImageID {
				t.Errorf("expected image ID %s, got %v", expectedImageID, params.ImageId)
			}
			if params.InstanceType != expectedInstanceType {
				t.Errorf("expected instance type %s, got %s", expectedInstanceType, params.InstanceType)
			}
			return &awsec2.RunInstancesOutput{
				Instances: []types.Instance{
					{
						InstanceId:   aws.String(expectedInstanceID),
						InstanceType: expectedInstanceType,
						State: &types.InstanceState{
							Name: expectedState,
						},
					},
				},
			}, nil
		},
	}

	config := CreateInstanceConfig{
		ImageID:      expectedImageID,
		InstanceType: expectedInstanceType,
	}
	instanceInfo, err := CreateInstance(ctx, mockClient, config)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if instanceInfo.InstanceID != expectedInstanceID {
		t.Errorf("expected instance ID %s, got %s", expectedInstanceID, instanceInfo.InstanceID)
	}
	if instanceInfo.State != string(expectedState) {
		t.Errorf("expected state %s, got %s", expectedState, instanceInfo.State)
	}
	if instanceInfo.InstanceType != string(expectedInstanceType) {
		t.Errorf("expected instance type %s, got %s", expectedInstanceType, instanceInfo.InstanceType)
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

	config := CreateInstanceConfig{
		ImageID:      "ami-1234567890abcdef0",
		InstanceType: types.InstanceTypeT4gMicro,
	}
	instanceInfo, err := CreateInstance(ctx, mockClient, config)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if instanceInfo.InstanceID != "" {
		t.Errorf("expected empty instance ID on error, got %s", instanceInfo.InstanceID)
	}
	if !strings.Contains(err.Error(), "failed to run instance") {
		t.Errorf("expected error message to contain 'failed to run instance', got %v", err)
	}
}
