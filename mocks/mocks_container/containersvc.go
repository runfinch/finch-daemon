// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/runfinch/finch-daemon/api/handlers/container (interfaces: Service)

// Package mocks_container is a generated GoMock package.
package mocks_container

import (
	context "context"
	io "io"
	reflect "reflect"
	time "time"

	types "github.com/containerd/nerdctl/v2/pkg/api/types"
	gomock "github.com/golang/mock/gomock"
	types0 "github.com/runfinch/finch-daemon/api/types"
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

// Attach mocks base method.
func (m *MockService) Attach(arg0 context.Context, arg1 string, arg2 *types0.AttachOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Attach", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Attach indicates an expected call of Attach.
func (mr *MockServiceMockRecorder) Attach(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Attach", reflect.TypeOf((*MockService)(nil).Attach), arg0, arg1, arg2)
}

// Create mocks base method.
func (m *MockService) Create(arg0 context.Context, arg1 string, arg2 []string, arg3 types.ContainerCreateOptions, arg4 types.NetworkOptions) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockServiceMockRecorder) Create(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockService)(nil).Create), arg0, arg1, arg2, arg3, arg4)
}

// ExecCreate mocks base method.
func (m *MockService) ExecCreate(arg0 context.Context, arg1 string, arg2 types0.ExecConfig) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExecCreate", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExecCreate indicates an expected call of ExecCreate.
func (mr *MockServiceMockRecorder) ExecCreate(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExecCreate", reflect.TypeOf((*MockService)(nil).ExecCreate), arg0, arg1, arg2)
}

// ExtractArchiveInContainer mocks base method.
func (m *MockService) ExtractArchiveInContainer(arg0 context.Context, arg1 *types0.PutArchiveOptions, arg2 io.ReadCloser) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExtractArchiveInContainer", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ExtractArchiveInContainer indicates an expected call of ExtractArchiveInContainer.
func (mr *MockServiceMockRecorder) ExtractArchiveInContainer(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExtractArchiveInContainer", reflect.TypeOf((*MockService)(nil).ExtractArchiveInContainer), arg0, arg1, arg2)
}

// GetPathToFilesInContainer mocks base method.
func (m *MockService) GetPathToFilesInContainer(arg0 context.Context, arg1, arg2 string) (string, func(), error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPathToFilesInContainer", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(func())
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetPathToFilesInContainer indicates an expected call of GetPathToFilesInContainer.
func (mr *MockServiceMockRecorder) GetPathToFilesInContainer(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPathToFilesInContainer", reflect.TypeOf((*MockService)(nil).GetPathToFilesInContainer), arg0, arg1, arg2)
}

// Inspect mocks base method.
func (m *MockService) Inspect(arg0 context.Context, arg1 string, arg2 bool) (*types0.Container, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Inspect", arg0, arg1, arg2)
	ret0, _ := ret[0].(*types0.Container)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Inspect indicates an expected call of Inspect.
func (mr *MockServiceMockRecorder) Inspect(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Inspect", reflect.TypeOf((*MockService)(nil).Inspect), arg0, arg1, arg2)
}

// Kill mocks base method.
func (m *MockService) Kill(arg0 context.Context, arg1 string, arg2 types.ContainerKillOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Kill", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Kill indicates an expected call of Kill.
func (mr *MockServiceMockRecorder) Kill(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kill", reflect.TypeOf((*MockService)(nil).Kill), arg0, arg1, arg2)
}

// List mocks base method.
func (m *MockService) List(arg0 context.Context, arg1 types.ContainerListOptions) ([]types0.ContainerListItem, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0, arg1)
	ret0, _ := ret[0].([]types0.ContainerListItem)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockServiceMockRecorder) List(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockService)(nil).List), arg0, arg1)
}

// Logs mocks base method.
func (m *MockService) Logs(arg0 context.Context, arg1 string, arg2 *types0.LogsOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Logs", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Logs indicates an expected call of Logs.
func (mr *MockServiceMockRecorder) Logs(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Logs", reflect.TypeOf((*MockService)(nil).Logs), arg0, arg1, arg2)
}

// Pause mocks base method.
func (m *MockService) Pause(arg0 context.Context, arg1 string, arg2 types.ContainerPauseOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pause", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Pause indicates an expected call of Pause.
func (mr *MockServiceMockRecorder) Pause(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pause", reflect.TypeOf((*MockService)(nil).Pause), arg0, arg1, arg2)
}

// Remove mocks base method.
func (m *MockService) Remove(arg0 context.Context, arg1 string, arg2, arg3 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockServiceMockRecorder) Remove(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockService)(nil).Remove), arg0, arg1, arg2, arg3)
}

// Rename mocks base method.
func (m *MockService) Rename(arg0 context.Context, arg1, arg2 string, arg3 types.ContainerRenameOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rename", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// Rename indicates an expected call of Rename.
func (mr *MockServiceMockRecorder) Rename(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rename", reflect.TypeOf((*MockService)(nil).Rename), arg0, arg1, arg2, arg3)
}

// Restart mocks base method.
func (m *MockService) Restart(arg0 context.Context, arg1 string, arg2 types.ContainerRestartOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Restart", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Restart indicates an expected call of Restart.
func (mr *MockServiceMockRecorder) Restart(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Restart", reflect.TypeOf((*MockService)(nil).Restart), arg0, arg1, arg2)
}

// Start mocks base method.
func (m *MockService) Start(arg0 context.Context, arg1 string, arg2 types.ContainerStartOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockServiceMockRecorder) Start(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockService)(nil).Start), arg0, arg1, arg2)
}

// Stats mocks base method.
func (m *MockService) Stats(arg0 context.Context, arg1 string) (<-chan *types0.StatsJSON, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stats", arg0, arg1)
	ret0, _ := ret[0].(<-chan *types0.StatsJSON)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stats indicates an expected call of Stats.
func (mr *MockServiceMockRecorder) Stats(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stats", reflect.TypeOf((*MockService)(nil).Stats), arg0, arg1)
}

// Stop mocks base method.
func (m *MockService) Stop(arg0 context.Context, arg1 string, arg2 *time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockServiceMockRecorder) Stop(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockService)(nil).Stop), arg0, arg1, arg2)
}

// Wait mocks base method.
func (m *MockService) Wait(arg0 context.Context, arg1 string, arg2 types.ContainerWaitOptions) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Wait", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Wait indicates an expected call of Wait.
func (mr *MockServiceMockRecorder) Wait(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wait", reflect.TypeOf((*MockService)(nil).Wait), arg0, arg1, arg2)
}

// WriteFilesAsTarArchive mocks base method.
func (m *MockService) WriteFilesAsTarArchive(arg0 string, arg1 io.Writer, arg2 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteFilesAsTarArchive", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteFilesAsTarArchive indicates an expected call of WriteFilesAsTarArchive.
func (mr *MockServiceMockRecorder) WriteFilesAsTarArchive(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteFilesAsTarArchive", reflect.TypeOf((*MockService)(nil).WriteFilesAsTarArchive), arg0, arg1, arg2)
}
