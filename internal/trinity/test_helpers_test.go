package trinity

import (
	"context"

	"github.com/nightmare-assault/nightmare-assault/internal/api/client"
)

// createFailingMockProvider creates a mock provider that always fails with the given error
func createFailingMockProvider(name string, err error) *mockProvider {
	return &mockProvider{
		name: name,
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return nil, err
		},
	}
}

// createSuccessMockProvider creates a mock provider that always succeeds
func createSuccessMockProvider(name string) *mockProvider {
	return &mockProvider{
		name: name,
		sendMessageFunc: func(ctx context.Context, messages []client.Message) (*client.Response, error) {
			return &client.Response{
				Content: "test response from " + name,
			}, nil
		},
	}
}
