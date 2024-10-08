// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/runfinch/finch-daemon/pkg/ecc (interfaces: ExecCmdCreator)

// Package mocks_ecc is a generated GoMock package.
package mocks_ecc

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	ecc "github.com/runfinch/finch-daemon/pkg/ecc"
)

// MockExecCmdCreator is a mock of ExecCmdCreator interface.
type MockExecCmdCreator struct {
	ctrl     *gomock.Controller
	recorder *MockExecCmdCreatorMockRecorder
}

// MockExecCmdCreatorMockRecorder is the mock recorder for MockExecCmdCreator.
type MockExecCmdCreatorMockRecorder struct {
	mock *MockExecCmdCreator
}

// NewMockExecCmdCreator creates a new mock instance.
func NewMockExecCmdCreator(ctrl *gomock.Controller) *MockExecCmdCreator {
	mock := &MockExecCmdCreator{ctrl: ctrl}
	mock.recorder = &MockExecCmdCreatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExecCmdCreator) EXPECT() *MockExecCmdCreatorMockRecorder {
	return m.recorder
}

// Command mocks base method.
func (m *MockExecCmdCreator) Command(arg0 string, arg1 ...string) ecc.ExecCmd {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Command", varargs...)
	ret0, _ := ret[0].(ecc.ExecCmd)
	return ret0
}

// Command indicates an expected call of Command.
func (mr *MockExecCmdCreatorMockRecorder) Command(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Command", reflect.TypeOf((*MockExecCmdCreator)(nil).Command), varargs...)
}
