package image

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type MockAMIFinder struct {
	FindLatestAMIFunc func(ctx context.Context) (string, error)
}

func (m *MockAMIFinder) FindLatestAMI(ctx context.Context) (string, error) {
	if m.FindLatestAMIFunc != nil {
		return m.FindLatestAMIFunc(ctx)
	}
	return "", nil
}

type MockImageLister struct {
	ListImagesFunc func(ctx context.Context) ([]types.Image, error)
}

func (m *MockImageLister) ListImages(ctx context.Context) ([]types.Image, error) {
	if m.ListImagesFunc != nil {
		return m.ListImagesFunc(ctx)
	}
	return []types.Image{}, nil
}
