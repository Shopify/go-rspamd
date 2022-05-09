package rspamd

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// client's usage

var _ Client = &mockClient{}

// NewMock creates a mock client, which can be used wherever client is used, to test/
func NewMock() *mockClient {
	return &mockClient{}
}

type mockClient struct {
	mock.Mock
}

func (m *mockClient) Check(ctx context.Context, cr *CheckRequest) (*CheckResponse, error) {
	args := m.Called(ctx, cr)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*CheckResponse), nil
}

func (m *mockClient) LearnSpam(ctx context.Context, lr *LearnRequest) (*LearnResponse, error) {
	args := m.Called(ctx, lr)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*LearnResponse), nil
}

func (m *mockClient) LearnHam(ctx context.Context, lr *LearnRequest) (*LearnResponse, error) {
	args := m.Called(ctx, lr)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*LearnResponse), nil
}

func (m *mockClient) FuzzyAdd(ctx context.Context, fr *FuzzyRequest) (*FuzzyResponse, error) {
	args := m.Called(ctx, fr)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*FuzzyResponse), nil
}

func (m *mockClient) FuzzyDel(ctx context.Context, fr *FuzzyRequest) (*FuzzyResponse, error) {
	args := m.Called(ctx, fr)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*FuzzyResponse), nil
}

func (m *mockClient) Ping(ctx context.Context) (PingResponse, error) {
	args := m.Called(ctx)
	if args.Error(1) != nil {
		return "", args.Error(1)
	}

	return args.Get(0).(PingResponse), nil
}
