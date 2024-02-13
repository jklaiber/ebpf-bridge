// Code generated by MockGen. DO NOT EDIT.
// Source: printer.go
//
// Generated by this command:
//
//	mockgen -source=printer.go -destination=mocks/printer_mock.go -package=mocks Printer
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	api "github.com/jklaiber/ebpf-bridge/pkg/api"
	gomock "go.uber.org/mock/gomock"
)

// MockPrinter is a mock of Printer interface.
type MockPrinter struct {
	ctrl     *gomock.Controller
	recorder *MockPrinterMockRecorder
}

// MockPrinterMockRecorder is the mock recorder for MockPrinter.
type MockPrinterMockRecorder struct {
	mock *MockPrinter
}

// NewMockPrinter creates a new mock instance.
func NewMockPrinter(ctrl *gomock.Controller) *MockPrinter {
	mock := &MockPrinter{ctrl: ctrl}
	mock.recorder = &MockPrinterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPrinter) EXPECT() *MockPrinterMockRecorder {
	return m.recorder
}

// PrintBridgeDescriptions mocks base method.
func (m *MockPrinter) PrintBridgeDescriptions(bridges []*api.BridgeDescription) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PrintBridgeDescriptions", bridges)
	ret0, _ := ret[0].(string)
	return ret0
}

// PrintBridgeDescriptions indicates an expected call of PrintBridgeDescriptions.
func (mr *MockPrinterMockRecorder) PrintBridgeDescriptions(bridges any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PrintBridgeDescriptions", reflect.TypeOf((*MockPrinter)(nil).PrintBridgeDescriptions), bridges)
}
