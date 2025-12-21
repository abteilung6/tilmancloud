package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abteilung6/tilmancloud/pkg/api/generated"
	"github.com/abteilung6/tilmancloud/pkg/ec2"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestNodesHandler_CreateNode(t *testing.T) {
	expectedInstanceID := "i-1234567890abcdef0"
	expectedState := types.InstanceStateNamePending
	expectedInstanceType := types.InstanceTypeT2Micro
	expectedPublicIP := "54.123.45.67"
	expectedPrivateIP := "10.0.1.123"

	mockClient := &ec2.MockEC2Client{
		RunInstancesFunc: func(ctx context.Context, params *awsec2.RunInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.RunInstancesOutput, error) {
			return &awsec2.RunInstancesOutput{
				Instances: []types.Instance{
					{
						InstanceId:   aws.String(expectedInstanceID),
						InstanceType: expectedInstanceType,
						State: &types.InstanceState{
							Name: expectedState,
						},
						PublicIpAddress:  aws.String(expectedPublicIP),
						PrivateIpAddress: aws.String(expectedPrivateIP),
					},
				},
			}, nil
		},
	}

	handler := NewNodesHandler(mockClient)

	req := httptest.NewRequest("POST", "/nodes", nil)
	w := httptest.NewRecorder()

	handler.CreateNode(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response generated.Node
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Name != expectedInstanceID {
		t.Errorf("expected instance ID %s, got %s", expectedInstanceID, response.Name)
	}
	if response.State == nil || *response.State != generated.NodeState(expectedState) {
		t.Errorf("expected state %s, got %v", expectedState, response.State)
	}
	if response.InstanceType == nil || *response.InstanceType != string(expectedInstanceType) {
		t.Errorf("expected instance type %s, got %v", expectedInstanceType, response.InstanceType)
	}
	if response.PublicIp == nil || *response.PublicIp != expectedPublicIP {
		t.Errorf("expected public IP %s, got %v", expectedPublicIP, response.PublicIp)
	}
	if response.PrivateIp == nil || *response.PrivateIp != expectedPrivateIP {
		t.Errorf("expected private IP %s, got %v", expectedPrivateIP, response.PrivateIp)
	}
}

func TestNodesHandler_CreateNode_WithNilIPs(t *testing.T) {
	expectedInstanceID := "i-1234567890abcdef0"
	expectedState := types.InstanceStateNamePending
	expectedInstanceType := types.InstanceTypeT2Micro

	mockClient := &ec2.MockEC2Client{
		RunInstancesFunc: func(ctx context.Context, params *awsec2.RunInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.RunInstancesOutput, error) {
			return &awsec2.RunInstancesOutput{
				Instances: []types.Instance{
					{
						InstanceId:   aws.String(expectedInstanceID),
						InstanceType: expectedInstanceType,
						State: &types.InstanceState{
							Name: expectedState,
						},
						// PublicIpAddress and PrivateIpAddress are nil (not yet assigned)
					},
				},
			}, nil
		},
	}

	handler := NewNodesHandler(mockClient)

	req := httptest.NewRequest("POST", "/nodes", nil)
	w := httptest.NewRecorder()

	handler.CreateNode(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response generated.Node
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Name != expectedInstanceID {
		t.Errorf("expected instance ID %s, got %s", expectedInstanceID, response.Name)
	}
	if response.PublicIp != nil {
		t.Errorf("expected public IP to be nil, got %v", response.PublicIp)
	}
	if response.PrivateIp != nil {
		t.Errorf("expected private IP to be nil, got %v", response.PrivateIp)
	}
}

func TestNodesHandler_CreateNode_Error(t *testing.T) {
	mockClient := &ec2.MockEC2Client{
		RunInstancesFunc: func(ctx context.Context, params *awsec2.RunInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.RunInstancesOutput, error) {
			return nil, fmt.Errorf("AWS API error")
		},
	}

	handler := NewNodesHandler(mockClient)

	req := httptest.NewRequest("POST", "/nodes", nil)
	w := httptest.NewRecorder()

	handler.CreateNode(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

