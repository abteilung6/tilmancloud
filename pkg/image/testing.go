package image

import "context"

type MockAMIFinder struct {
	FindLatestAMIFunc func(ctx context.Context) (string, error)
}

func (m *MockAMIFinder) FindLatestAMI(ctx context.Context) (string, error) {
	if m.FindLatestAMIFunc != nil {
		return m.FindLatestAMIFunc(ctx)
	}
	return "", nil
}
