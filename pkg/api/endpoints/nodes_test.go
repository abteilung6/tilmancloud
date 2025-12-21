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

func TestNodesHandler_ListNodes(t *testing.T) {
	expectedInstanceID1 := "i-1234567890abcdef0"
	expectedInstanceID2 := "i-0987654321fedcba0"
	expectedState1 := types.InstanceStateNameRunning
	expectedState2 := types.InstanceStateNamePending
	expectedInstanceType := types.InstanceTypeT2Micro
	expectedPublicIP1 := "54.123.45.67"
	expectedPrivateIP1 := "10.0.1.123"
	expectedPublicIP2 := "54.123.45.68"
	expectedPrivateIP2 := "10.0.1.124"

	mockClient := &ec2.MockEC2Client{
		DescribeInstancesFunc: func(ctx context.Context, params *awsec2.DescribeInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeInstancesOutput, error) {
			return &awsec2.DescribeInstancesOutput{
				Reservations: []types.Reservation{
					{
						Instances: []types.Instance{
							{
								InstanceId:   aws.String(expectedInstanceID1),
								InstanceType: expectedInstanceType,
								State: &types.InstanceState{
									Name: expectedState1,
								},
								PublicIpAddress:  aws.String(expectedPublicIP1),
								PrivateIpAddress: aws.String(expectedPrivateIP1),
							},
							{
								InstanceId:   aws.String(expectedInstanceID2),
								InstanceType: expectedInstanceType,
								State: &types.InstanceState{
									Name: expectedState2,
								},
								PublicIpAddress:  aws.String(expectedPublicIP2),
								PrivateIpAddress: aws.String(expectedPrivateIP2),
							},
						},
					},
				},
			}, nil
		},
	}

	handler := NewNodesHandler(mockClient)

	req := httptest.NewRequest("GET", "/nodes", nil)
	w := httptest.NewRecorder()

	handler.ListNodes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response []generated.Node
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(response))
	}

	// Check first node
	node1 := response[0]
	if node1.Name != expectedInstanceID1 {
		t.Errorf("expected instance ID %s, got %s", expectedInstanceID1, node1.Name)
	}
	if node1.State == nil || *node1.State != generated.NodeState(expectedState1) {
		t.Errorf("expected state %s, got %v", expectedState1, node1.State)
	}
	if node1.PublicIp == nil || *node1.PublicIp != expectedPublicIP1 {
		t.Errorf("expected public IP %s, got %v", expectedPublicIP1, node1.PublicIp)
	}
	if node1.PrivateIp == nil || *node1.PrivateIp != expectedPrivateIP1 {
		t.Errorf("expected private IP %s, got %v", expectedPrivateIP1, node1.PrivateIp)
	}

	// Check second node
	node2 := response[1]
	if node2.Name != expectedInstanceID2 {
		t.Errorf("expected instance ID %s, got %s", expectedInstanceID2, node2.Name)
	}
	if node2.State == nil || *node2.State != generated.NodeState(expectedState2) {
		t.Errorf("expected state %s, got %v", expectedState2, node2.State)
	}
	if node2.PublicIp == nil || *node2.PublicIp != expectedPublicIP2 {
		t.Errorf("expected public IP %s, got %v", expectedPublicIP2, node2.PublicIp)
	}
	if node2.PrivateIp == nil || *node2.PrivateIp != expectedPrivateIP2 {
		t.Errorf("expected private IP %s, got %v", expectedPrivateIP2, node2.PrivateIp)
	}
}

func TestNodesHandler_ListNodes_Empty(t *testing.T) {
	mockClient := &ec2.MockEC2Client{
		DescribeInstancesFunc: func(ctx context.Context, params *awsec2.DescribeInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeInstancesOutput, error) {
			return &awsec2.DescribeInstancesOutput{
				Reservations: []types.Reservation{},
			}, nil
		},
	}

	handler := NewNodesHandler(mockClient)

	req := httptest.NewRequest("GET", "/nodes", nil)
	w := httptest.NewRecorder()

	handler.ListNodes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response []generated.Node
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 0 {
		t.Errorf("expected empty list, got %d nodes", len(response))
	}
}

func TestNodesHandler_ListNodes_Error(t *testing.T) {
	mockClient := &ec2.MockEC2Client{
		DescribeInstancesFunc: func(ctx context.Context, params *awsec2.DescribeInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeInstancesOutput, error) {
			return nil, fmt.Errorf("AWS API error")
		},
	}

	handler := NewNodesHandler(mockClient)

	req := httptest.NewRequest("GET", "/nodes", nil)
	w := httptest.NewRecorder()

	handler.ListNodes(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
