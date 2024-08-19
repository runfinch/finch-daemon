// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/containerd/containerd/cio (interfaces: IO)

// Package mocks_cio is a generated GoMock package.
package mocks_cio

import (
	reflect "reflect"

	cio "github.com/containerd/containerd/cio"
	gomock "github.com/golang/mock/gomock"
)

// MockIO is a mock of IO interface.
type MockIO struct {
	ctrl     *gomock.Controller
	recorder *MockIOMockRecorder
}

// MockIOMockRecorder is the mock recorder for MockIO.
type MockIOMockRecorder struct {
	mock *MockIO
}

// NewMockIO creates a new mock instance.
func NewMockIO(ctrl *gomock.Controller) *MockIO {
	mock := &MockIO{ctrl: ctrl}
	mock.recorder = &MockIOMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIO) EXPECT() *MockIOMockRecorder {
	return m.recorder
}

// Cancel mocks base method.
func (m *MockIO) Cancel() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Cancel")
}

// Cancel indicates an expected call of Cancel.
func (mr *MockIOMockRecorder) Cancel() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Cancel", reflect.TypeOf((*MockIO)(nil).Cancel))
}

// Close mocks base method.
func (m *MockIO) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockIOMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockIO)(nil).Close))
}

// Config mocks base method.
func (m *MockIO) Config() cio.Config {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Config")
	ret0, _ := ret[0].(cio.Config)
	return ret0
}

// Config indicates an expected call of Config.
func (mr *MockIOMockRecorder) Config() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Config", reflect.TypeOf((*MockIO)(nil).Config))
}

// Wait mocks base method.
func (m *MockIO) Wait() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Wait")
}

// Wait indicates an expected call of Wait.
func (mr *MockIOMockRecorder) Wait() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wait", reflect.TypeOf((*MockIO)(nil).Wait))
}
