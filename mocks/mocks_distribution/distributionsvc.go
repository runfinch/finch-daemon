// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/runfinch/finch-daemon/api/handlers/distribution (interfaces: Service)

// Package mocks_distribution is a generated GoMock package.
package mocks_distribution

import (
	context "context"
	reflect "reflect"

	types "github.com/docker/cli/cli/config/types"
	registry "github.com/docker/docker/api/types/registry"
	gomock "github.com/golang/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// Inspect mocks base method.
func (m *MockService) Inspect(arg0 context.Context, arg1 string, arg2 *types.AuthConfig) (*registry.DistributionInspect, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Inspect", arg0, arg1, arg2)
	ret0, _ := ret[0].(*registry.DistributionInspect)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Inspect indicates an expected call of Inspect.
func (mr *MockServiceMockRecorder) Inspect(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Inspect", reflect.TypeOf((*MockService)(nil).Inspect), arg0, arg1, arg2)
}
