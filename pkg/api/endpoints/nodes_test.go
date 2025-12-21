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

	mockClient := &ec2.MockEC2Client{
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

	if response.Name == nil || *response.Name != expectedInstanceID {
		t.Errorf("expected instance ID %s, got %v", expectedInstanceID, response.Name)
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
