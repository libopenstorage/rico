// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/libopenstorage/rico/pkg/storageprovider (interfaces: Interface)

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	storageprovider "github.com/libopenstorage/rico/pkg/storageprovider"
	reflect "reflect"
)

// MockInterface is a mock of Interface interface
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// DeviceAdd mocks base method
func (m *MockInterface) DeviceAdd(arg0 *storageprovider.StorageNode, arg1 *storageprovider.Device) error {
	ret := m.ctrl.Call(m, "DeviceAdd", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeviceAdd indicates an expected call of DeviceAdd
func (mr *MockInterfaceMockRecorder) DeviceAdd(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeviceAdd", reflect.TypeOf((*MockInterface)(nil).DeviceAdd), arg0, arg1)
}

// DeviceRemove mocks base method
func (m *MockInterface) DeviceRemove(arg0 *storageprovider.StorageNode, arg1 *storageprovider.Device) error {
	ret := m.ctrl.Call(m, "DeviceRemove", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeviceRemove indicates an expected call of DeviceRemove
func (mr *MockInterfaceMockRecorder) DeviceRemove(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeviceRemove", reflect.TypeOf((*MockInterface)(nil).DeviceRemove), arg0, arg1)
}

// Event mocks base method
func (m *MockInterface) Event() {
	m.ctrl.Call(m, "Event")
}

// Event indicates an expected call of Event
func (mr *MockInterfaceMockRecorder) Event() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Event", reflect.TypeOf((*MockInterface)(nil).Event))
}

// GetTopology mocks base method
func (m *MockInterface) GetTopology() (*storageprovider.Topology, error) {
	ret := m.ctrl.Call(m, "GetTopology")
	ret0, _ := ret[0].(*storageprovider.Topology)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTopology indicates an expected call of GetTopology
func (mr *MockInterfaceMockRecorder) GetTopology() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTopology", reflect.TypeOf((*MockInterface)(nil).GetTopology))
}

// Utilization mocks base method
func (m *MockInterface) Utilization() (int, error) {
	ret := m.ctrl.Call(m, "Utilization")
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Utilization indicates an expected call of Utilization
func (mr *MockInterfaceMockRecorder) Utilization() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Utilization", reflect.TypeOf((*MockInterface)(nil).Utilization))
}