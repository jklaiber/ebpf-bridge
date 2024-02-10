// Code generated by MockGen. DO NOT EDIT.
// Source: bridge.go
//
// Generated by this command:
//
//	mockgen -source=bridge.go -destination=mocks/bridge_mock.go -package=mocks Bridge
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	hostlink "github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	gomock "go.uber.org/mock/gomock"
)

// MockBridge is a mock of Bridge interface.
type MockBridge struct {
	ctrl     *gomock.Controller
	recorder *MockBridgeMockRecorder
}

// MockBridgeMockRecorder is the mock recorder for MockBridge.
type MockBridgeMockRecorder struct {
	mock *MockBridge
}

// NewMockBridge creates a new mock instance.
func NewMockBridge(ctrl *gomock.Controller) *MockBridge {
	mock := &MockBridge{ctrl: ctrl}
	mock.recorder = &MockBridgeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBridge) EXPECT() *MockBridgeMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockBridge) Add() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add")
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockBridgeMockRecorder) Add() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockBridge)(nil).Add))
}

// Interface1 mocks base method.
func (m *MockBridge) Interface1() hostlink.Link {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Interface1")
	ret0, _ := ret[0].(hostlink.Link)
	return ret0
}

// Interface1 indicates an expected call of Interface1.
func (mr *MockBridgeMockRecorder) Interface1() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Interface1", reflect.TypeOf((*MockBridge)(nil).Interface1))
}

// Interface2 mocks base method.
func (m *MockBridge) Interface2() hostlink.Link {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Interface2")
	ret0, _ := ret[0].(hostlink.Link)
	return ret0
}

// Interface2 indicates an expected call of Interface2.
func (mr *MockBridgeMockRecorder) Interface2() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Interface2", reflect.TypeOf((*MockBridge)(nil).Interface2))
}

// MonitorInterface mocks base method.
func (m *MockBridge) MonitorInterface() hostlink.Link {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MonitorInterface")
	ret0, _ := ret[0].(hostlink.Link)
	return ret0
}

// MonitorInterface indicates an expected call of MonitorInterface.
func (mr *MockBridgeMockRecorder) MonitorInterface() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MonitorInterface", reflect.TypeOf((*MockBridge)(nil).MonitorInterface))
}

// Name mocks base method.
func (m *MockBridge) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockBridgeMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockBridge)(nil).Name))
}

// Remove mocks base method.
func (m *MockBridge) Remove() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove")
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockBridgeMockRecorder) Remove() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockBridge)(nil).Remove))
}
