// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/containerd/nerdctl/v2/pkg/containerutil (interfaces: NetworkOptionsManager)

// Package mocks_container is a generated GoMock package.
package mocks_container

import (
	context "context"
	reflect "reflect"

	client "github.com/containerd/containerd/v2/client"
	oci "github.com/containerd/containerd/v2/pkg/oci"
	types "github.com/containerd/nerdctl/v2/pkg/api/types"
	gomock "github.com/golang/mock/gomock"
)

// MockNetworkOptionsManager is a mock of NetworkOptionsManager interface.
type MockNetworkOptionsManager struct {
	ctrl     *gomock.Controller
	recorder *MockNetworkOptionsManagerMockRecorder
}

// MockNetworkOptionsManagerMockRecorder is the mock recorder for MockNetworkOptionsManager.
type MockNetworkOptionsManagerMockRecorder struct {
	mock *MockNetworkOptionsManager
}

// NewMockNetworkOptionsManager creates a new mock instance.
func NewMockNetworkOptionsManager(ctrl *gomock.Controller) *MockNetworkOptionsManager {
	mock := &MockNetworkOptionsManager{ctrl: ctrl}
	mock.recorder = &MockNetworkOptionsManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNetworkOptionsManager) EXPECT() *MockNetworkOptionsManagerMockRecorder {
	return m.recorder
}

// CleanupNetworking mocks base method.
func (m *MockNetworkOptionsManager) CleanupNetworking(arg0 context.Context, arg1 client.Container) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CleanupNetworking", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CleanupNetworking indicates an expected call of CleanupNetworking.
func (mr *MockNetworkOptionsManagerMockRecorder) CleanupNetworking(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CleanupNetworking", reflect.TypeOf((*MockNetworkOptionsManager)(nil).CleanupNetworking), arg0, arg1)
}

// ContainerNetworkingOpts mocks base method.
func (m *MockNetworkOptionsManager) ContainerNetworkingOpts(arg0 context.Context, arg1 string) ([]oci.SpecOpts, []client.NewContainerOpts, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContainerNetworkingOpts", arg0, arg1)
	ret0, _ := ret[0].([]oci.SpecOpts)
	ret1, _ := ret[1].([]client.NewContainerOpts)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ContainerNetworkingOpts indicates an expected call of ContainerNetworkingOpts.
func (mr *MockNetworkOptionsManagerMockRecorder) ContainerNetworkingOpts(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerNetworkingOpts", reflect.TypeOf((*MockNetworkOptionsManager)(nil).ContainerNetworkingOpts), arg0, arg1)
}

// InternalNetworkingOptionLabels mocks base method.
func (m *MockNetworkOptionsManager) InternalNetworkingOptionLabels(arg0 context.Context) (types.NetworkOptions, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InternalNetworkingOptionLabels", arg0)
	ret0, _ := ret[0].(types.NetworkOptions)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InternalNetworkingOptionLabels indicates an expected call of InternalNetworkingOptionLabels.
func (mr *MockNetworkOptionsManagerMockRecorder) InternalNetworkingOptionLabels(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InternalNetworkingOptionLabels", reflect.TypeOf((*MockNetworkOptionsManager)(nil).InternalNetworkingOptionLabels), arg0)
}

// NetworkOptions mocks base method.
func (m *MockNetworkOptionsManager) NetworkOptions() types.NetworkOptions {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NetworkOptions")
	ret0, _ := ret[0].(types.NetworkOptions)
	return ret0
}

// NetworkOptions indicates an expected call of NetworkOptions.
func (mr *MockNetworkOptionsManagerMockRecorder) NetworkOptions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NetworkOptions", reflect.TypeOf((*MockNetworkOptionsManager)(nil).NetworkOptions))
}

// SetupNetworking mocks base method.
func (m *MockNetworkOptionsManager) SetupNetworking(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetupNetworking", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetupNetworking indicates an expected call of SetupNetworking.
func (mr *MockNetworkOptionsManagerMockRecorder) SetupNetworking(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetupNetworking", reflect.TypeOf((*MockNetworkOptionsManager)(nil).SetupNetworking), arg0, arg1)
}

// VerifyNetworkOptions mocks base method.
func (m *MockNetworkOptionsManager) VerifyNetworkOptions(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyNetworkOptions", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// VerifyNetworkOptions indicates an expected call of VerifyNetworkOptions.
func (mr *MockNetworkOptionsManagerMockRecorder) VerifyNetworkOptions(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyNetworkOptions", reflect.TypeOf((*MockNetworkOptionsManager)(nil).VerifyNetworkOptions), arg0)
}
