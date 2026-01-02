package endpoints

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/abteilung6/tilmancloud/pkg/api/generated"
	"github.com/abteilung6/tilmancloud/pkg/image"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type ImagesHandler struct {
	ImageLister image.ImageLister
}

func NewImagesHandler(imageLister image.ImageLister) *ImagesHandler {
	return &ImagesHandler{
		ImageLister: imageLister,
	}
}

func (h *ImagesHandler) ListImages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	awsImages, err := h.ImageLister.ListImages(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	images := make([]generated.Image, 0, len(awsImages))
	for _, awsImage := range awsImages {
		image := convertAWSImageToGenerated(awsImage)
		images = append(images, image)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(images)
}

func convertAWSImageToGenerated(awsImage types.Image) generated.Image {
	image := generated.Image{}

	if awsImage.ImageId != nil {
		image.Id = *awsImage.ImageId
	}
	if awsImage.Name != nil {
		image.Name = stringPtrOrNil(*awsImage.Name)
	}
	// State is a value type, not a pointer - always present
	image.State = generated.ImageState(awsImage.State)
	if awsImage.Description != nil {
		image.Description = stringPtrOrNil(*awsImage.Description)
	}
	// Architecture is a value type, not a pointer - always present
	image.Architecture = generated.ImageArchitecture(awsImage.Architecture)
	// VirtualizationType is a value type, not a pointer - always present
	image.VirtualizationType = generated.ImageVirtualizationType(awsImage.VirtualizationType)

	// Extract ImageID and SnapshotID from tags
	for _, tag := range awsImage.Tags {
		if tag.Key != nil && tag.Value != nil {
			if *tag.Key == "ImageID" {
				image.ImageId = stringPtrOrNil(*tag.Value)
			}
			if *tag.Key == "SnapshotID" {
				image.SnapshotId = stringPtrOrNil(*tag.Value)
			}
		}
	}

	// Creation date is always present from AWS
	if awsImage.CreationDate != nil {
		// AWS returns creation date as ISO 8601 string, parse to time.Time
		creationDate, err := time.Parse(time.RFC3339, *awsImage.CreationDate)
		if err != nil {
			// Fallback: try parsing without timezone if RFC3339 fails
			creationDate, err = time.Parse("2006-01-02T15:04:05", *awsImage.CreationDate)
			if err != nil {
				// If parsing fails, use zero time (shouldn't happen with AWS)
				creationDate = time.Time{}
			}
		}
		image.CreationDate = creationDate
	}

	return image
}
