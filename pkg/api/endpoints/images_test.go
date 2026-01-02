package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/abteilung6/tilmancloud/pkg/api/generated"
	"github.com/abteilung6/tilmancloud/pkg/image"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestImagesHandler_ListImages(t *testing.T) {
	expectedAMIID1 := "ami-1234567890abcdef0"
	expectedAMIID2 := "ami-0987654321fedcba0"
	expectedName1 := "fedora-43-aarch64-base-76f2ddd3bac7da2b"
	expectedName2 := "fedora-43-aarch64-base-87e3eee4cbd8eb3c"
	expectedState1 := types.ImageStateAvailable
	expectedState2 := types.ImageStatePending
	expectedImageID1 := "fedora-43-aarch64-76f2ddd3bac7da2b"
	expectedImageID2 := "fedora-43-aarch64-87e3eee4cbd8eb3c"
	expectedSnapshotID1 := "snap-abcdef1234567890"
	expectedSnapshotID2 := "snap-fedcba0987654321"
	expectedDescription := "Fedora 43 aarch64 base image"
	expectedArchitecture := types.ArchitectureValuesArm64
	expectedVirtualizationType := types.VirtualizationTypeHvm
	expectedCreationDate1 := "2024-01-15T10:30:00Z"
	expectedCreationDate2 := "2024-01-16T11:45:00Z"

	mockImageLister := &image.MockImageLister{
		ListImagesFunc: func(ctx context.Context) ([]types.Image, error) {
			return []types.Image{
				{
					ImageId:            aws.String(expectedAMIID1),
					Name:               aws.String(expectedName1),
					State:              expectedState1,
					Description:        aws.String(expectedDescription),
					Architecture:       expectedArchitecture,
					VirtualizationType: expectedVirtualizationType,
					CreationDate:       aws.String(expectedCreationDate1),
					Tags: []types.Tag{
						{Key: aws.String("ImageID"), Value: aws.String(expectedImageID1)},
						{Key: aws.String("SnapshotID"), Value: aws.String(expectedSnapshotID1)},
					},
				},
				{
					ImageId:            aws.String(expectedAMIID2),
					Name:               aws.String(expectedName2),
					State:              expectedState2,
					Description:        aws.String(expectedDescription),
					Architecture:       expectedArchitecture,
					VirtualizationType: expectedVirtualizationType,
					CreationDate:       aws.String(expectedCreationDate2),
					Tags: []types.Tag{
						{Key: aws.String("ImageID"), Value: aws.String(expectedImageID2)},
						{Key: aws.String("SnapshotID"), Value: aws.String(expectedSnapshotID2)},
					},
				},
			}, nil
		},
	}

	handler := NewImagesHandler(mockImageLister)

	req := httptest.NewRequest("GET", "/images", nil)
	w := httptest.NewRecorder()

	handler.ListImages(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response []generated.Image
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Fatalf("expected 2 images, got %d", len(response))
	}

	// Check first image
	image1 := response[0]
	if image1.Id != expectedAMIID1 {
		t.Errorf("expected AMI ID %s, got %s", expectedAMIID1, image1.Id)
	}
	if image1.Name == nil || *image1.Name != expectedName1 {
		t.Errorf("expected name %s, got %v", expectedName1, image1.Name)
	}
	if image1.State != generated.ImageState(expectedState1) {
		t.Errorf("expected state %s, got %s", expectedState1, image1.State)
	}
	if image1.ImageId == nil || *image1.ImageId != expectedImageID1 {
		t.Errorf("expected imageId %s, got %v", expectedImageID1, image1.ImageId)
	}
	if image1.SnapshotId == nil || *image1.SnapshotId != expectedSnapshotID1 {
		t.Errorf("expected snapshotId %s, got %v", expectedSnapshotID1, image1.SnapshotId)
	}
	if image1.Description == nil || *image1.Description != expectedDescription {
		t.Errorf("expected description %s, got %v", expectedDescription, image1.Description)
	}
	if image1.Architecture != generated.ImageArchitecture(expectedArchitecture) {
		t.Errorf("expected architecture %s, got %s", expectedArchitecture, image1.Architecture)
	}
	if image1.VirtualizationType != generated.ImageVirtualizationType(expectedVirtualizationType) {
		t.Errorf("expected virtualizationType %s, got %s", expectedVirtualizationType, image1.VirtualizationType)
	}
	// Parse expected creation date for comparison
	expectedDate1, _ := time.Parse(time.RFC3339, expectedCreationDate1)
	if !image1.CreationDate.Equal(expectedDate1) {
		t.Errorf("expected creationDate %s, got %s", expectedCreationDate1, image1.CreationDate.Format(time.RFC3339))
	}

	// Check second image
	image2 := response[1]
	if image2.Id != expectedAMIID2 {
		t.Errorf("expected AMI ID %s, got %s", expectedAMIID2, image2.Id)
	}
	if image2.State != generated.ImageState(expectedState2) {
		t.Errorf("expected state %s, got %s", expectedState2, image2.State)
	}
}

func TestImagesHandler_ListImages_Empty(t *testing.T) {
	mockImageLister := &image.MockImageLister{
		ListImagesFunc: func(ctx context.Context) ([]types.Image, error) {
			return []types.Image{}, nil
		},
	}

	handler := NewImagesHandler(mockImageLister)

	req := httptest.NewRequest("GET", "/images", nil)
	w := httptest.NewRecorder()

	handler.ListImages(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response []generated.Image
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 0 {
		t.Errorf("expected empty list, got %d images", len(response))
	}
}

func TestImagesHandler_ListImages_Error(t *testing.T) {
	mockImageLister := &image.MockImageLister{
		ListImagesFunc: func(ctx context.Context) ([]types.Image, error) {
			return nil, fmt.Errorf("AWS API error")
		},
	}

	handler := NewImagesHandler(mockImageLister)

	req := httptest.NewRequest("GET", "/images", nil)
	w := httptest.NewRecorder()

	handler.ListImages(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestImagesHandler_ListImages_MissingTags(t *testing.T) {
	expectedAMIID := "ami-1234567890abcdef0"
	expectedName := "fedora-43-aarch64-base-76f2ddd3bac7da2b"
	expectedState := types.ImageStateAvailable

	mockImageLister := &image.MockImageLister{
		ListImagesFunc: func(ctx context.Context) ([]types.Image, error) {
			return []types.Image{
				{
					ImageId: aws.String(expectedAMIID),
					Name:    aws.String(expectedName),
					State:   expectedState,
					Tags:    []types.Tag{}, // No tags
				},
			}, nil
		},
	}

	handler := NewImagesHandler(mockImageLister)

	req := httptest.NewRequest("GET", "/images", nil)
	w := httptest.NewRecorder()

	handler.ListImages(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response []generated.Image
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Fatalf("expected 1 image, got %d", len(response))
	}

	image := response[0]
	if image.Id != expectedAMIID {
		t.Errorf("expected AMI ID %s, got %s", expectedAMIID, image.Id)
	}
	if image.ImageId != nil {
		t.Errorf("expected imageId to be nil when tag is missing, got %v", image.ImageId)
	}
	if image.SnapshotId != nil {
		t.Errorf("expected snapshotId to be nil when tag is missing, got %v", image.SnapshotId)
	}
}
