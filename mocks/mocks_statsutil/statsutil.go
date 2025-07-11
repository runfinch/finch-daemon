// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/runfinch/finch-daemon/pkg/statsutil (interfaces: StatsUtil)
//
// Generated by this command:
//
//	mockgen --destination=../../mocks/mocks_statsutil/statsutil.go -package=mocks_statsutil github.com/runfinch/finch-daemon/pkg/statsutil StatsUtil
//

// Package mocks_statsutil is a generated GoMock package.
package mocks_statsutil

import (
	reflect "reflect"

	native "github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
	container "github.com/docker/docker/api/types/container"
	gomock "go.uber.org/mock/gomock"
)

// MockStatsUtil is a mock of StatsUtil interface.
type MockStatsUtil struct {
	ctrl     *gomock.Controller
	recorder *MockStatsUtilMockRecorder
	isgomock struct{}
}

// MockStatsUtilMockRecorder is the mock recorder for MockStatsUtil.
type MockStatsUtilMockRecorder struct {
	mock *MockStatsUtil
}

// NewMockStatsUtil creates a new mock instance.
func NewMockStatsUtil(ctrl *gomock.Controller) *MockStatsUtil {
	mock := &MockStatsUtil{ctrl: ctrl}
	mock.recorder = &MockStatsUtilMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStatsUtil) EXPECT() *MockStatsUtilMockRecorder {
	return m.recorder
}

// CollectNetworkStats mocks base method.
func (m *MockStatsUtil) CollectNetworkStats(pid int, interfaces []native.NetInterface) (map[string]container.NetworkStats, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CollectNetworkStats", pid, interfaces)
	ret0, _ := ret[0].(map[string]container.NetworkStats)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CollectNetworkStats indicates an expected call of CollectNetworkStats.
func (mr *MockStatsUtilMockRecorder) CollectNetworkStats(pid, interfaces any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CollectNetworkStats", reflect.TypeOf((*MockStatsUtil)(nil).CollectNetworkStats), pid, interfaces)
}

// GetNumberOnlineCPUs mocks base method.
func (m *MockStatsUtil) GetNumberOnlineCPUs() (uint32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNumberOnlineCPUs")
	ret0, _ := ret[0].(uint32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNumberOnlineCPUs indicates an expected call of GetNumberOnlineCPUs.
func (mr *MockStatsUtilMockRecorder) GetNumberOnlineCPUs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNumberOnlineCPUs", reflect.TypeOf((*MockStatsUtil)(nil).GetNumberOnlineCPUs))
}

// GetSystemCPUUsage mocks base method.
func (m *MockStatsUtil) GetSystemCPUUsage() (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSystemCPUUsage")
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSystemCPUUsage indicates an expected call of GetSystemCPUUsage.
func (mr *MockStatsUtilMockRecorder) GetSystemCPUUsage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSystemCPUUsage", reflect.TypeOf((*MockStatsUtil)(nil).GetSystemCPUUsage))
}
