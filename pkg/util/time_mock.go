// Code generated by MockGen. DO NOT EDIT.
// Source: time.go

// Package util is a generated GoMock package.
package util

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockTimelineProvider is a mock of TimelineProvider interface
type MockTimelineProvider struct {
	ctrl     *gomock.Controller
	recorder *MockTimelineProviderMockRecorder
}

// MockTimelineProviderMockRecorder is the mock recorder for MockTimelineProvider
type MockTimelineProviderMockRecorder struct {
	mock *MockTimelineProvider
}

// NewMockTimelineProvider creates a new mock instance
func NewMockTimelineProvider(ctrl *gomock.Controller) *MockTimelineProvider {
	mock := &MockTimelineProvider{ctrl: ctrl}
	mock.recorder = &MockTimelineProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTimelineProvider) EXPECT() *MockTimelineProviderMockRecorder {
	return m.recorder
}

// Now mocks base method
func (m *MockTimelineProvider) Now() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Now")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// Now indicates an expected call of Now
func (mr *MockTimelineProviderMockRecorder) Now() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Now", reflect.TypeOf((*MockTimelineProvider)(nil).Now))
}

// Sleep mocks base method
func (m *MockTimelineProvider) Sleep(duration time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Sleep", duration)
}

// Sleep indicates an expected call of Sleep
func (mr *MockTimelineProviderMockRecorder) Sleep(duration interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sleep", reflect.TypeOf((*MockTimelineProvider)(nil).Sleep), duration)
}
