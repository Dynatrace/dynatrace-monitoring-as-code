// Code generated by MockGen. DO NOT EDIT.
// Source: config.go

// Package config is a generated GoMock package.
package config

import (
	api "github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/api"
	environment "github.com/dynatrace-oss/dynatrace-monitoring-as-code/pkg/environment"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockConfig is a mock of Config interface
type MockConfig struct {
	ctrl     *gomock.Controller
	recorder *MockConfigMockRecorder
}

// MockConfigMockRecorder is the mock recorder for MockConfig
type MockConfigMockRecorder struct {
	mock *MockConfig
}

// NewMockConfig creates a new mock instance
func NewMockConfig(ctrl *gomock.Controller) *MockConfig {
	mock := &MockConfig{ctrl: ctrl}
	mock.recorder = &MockConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConfig) EXPECT() *MockConfigMockRecorder {
	return m.recorder
}

// GetConfigForEnvironment mocks base method
func (m *MockConfig) GetConfigForEnvironment(environment environment.Environment, dict map[string]api.DynatraceEntity) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfigForEnvironment", environment, dict)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetConfigForEnvironment indicates an expected call of GetConfigForEnvironment
func (mr *MockConfigMockRecorder) GetConfigForEnvironment(environment, dict interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfigForEnvironment", reflect.TypeOf((*MockConfig)(nil).GetConfigForEnvironment), environment, dict)
}

// IsSkipDeployment mocks base method
func (m *MockConfig) IsSkipDeployment(environment environment.Environment) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsSkipDeployment", environment)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsSkipDeployment indicates an expected call of IsSkipDeployment
func (mr *MockConfigMockRecorder) IsSkipDeployment(environment interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsSkipDeployment", reflect.TypeOf((*MockConfig)(nil).IsSkipDeployment), environment)
}

// GetApi mocks base method
func (m *MockConfig) GetApi() api.Api {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApi")
	ret0, _ := ret[0].(api.Api)
	return ret0
}

// GetApi indicates an expected call of GetApi
func (mr *MockConfigMockRecorder) GetApi() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApi", reflect.TypeOf((*MockConfig)(nil).GetApi))
}

// GetObjectNameForEnvironment mocks base method
func (m *MockConfig) GetObjectNameForEnvironment(environment environment.Environment, dict map[string]api.DynatraceEntity) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetObjectNameForEnvironment", environment, dict)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetObjectNameForEnvironment indicates an expected call of GetObjectNameForEnvironment
func (mr *MockConfigMockRecorder) GetObjectNameForEnvironment(environment, dict interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetObjectNameForEnvironment", reflect.TypeOf((*MockConfig)(nil).GetObjectNameForEnvironment), environment, dict)
}

// HasDependencyOn mocks base method
func (m *MockConfig) HasDependencyOn(config Config) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasDependencyOn", config)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasDependencyOn indicates an expected call of HasDependencyOn
func (mr *MockConfigMockRecorder) HasDependencyOn(config interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasDependencyOn", reflect.TypeOf((*MockConfig)(nil).HasDependencyOn), config)
}

// GetFilePath mocks base method
func (m *MockConfig) GetFilePath() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFilePath")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetFilePath indicates an expected call of GetFilePath
func (mr *MockConfigMockRecorder) GetFilePath() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFilePath", reflect.TypeOf((*MockConfig)(nil).GetFilePath))
}

// GetFullQualifiedId mocks base method
func (m *MockConfig) GetFullQualifiedId() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFullQualifiedId")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetFullQualifiedId indicates an expected call of GetFullQualifiedId
func (mr *MockConfigMockRecorder) GetFullQualifiedId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFullQualifiedId", reflect.TypeOf((*MockConfig)(nil).GetFullQualifiedId))
}

// GetType mocks base method
func (m *MockConfig) GetType() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetType")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetType indicates an expected call of GetType
func (mr *MockConfigMockRecorder) GetType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetType", reflect.TypeOf((*MockConfig)(nil).GetType))
}

// GetMeIdsOfEnvironment mocks base method
func (m *MockConfig) GetMeIdsOfEnvironment(environment environment.Environment) map[string]map[string]string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMeIdsOfEnvironment", environment)
	ret0, _ := ret[0].(map[string]map[string]string)
	return ret0
}

// GetMeIdsOfEnvironment indicates an expected call of GetMeIdsOfEnvironment
func (mr *MockConfigMockRecorder) GetMeIdsOfEnvironment(environment interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMeIdsOfEnvironment", reflect.TypeOf((*MockConfig)(nil).GetMeIdsOfEnvironment), environment)
}

// GetId mocks base method
func (m *MockConfig) GetId() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetId")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetId indicates an expected call of GetId
func (mr *MockConfigMockRecorder) GetId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetId", reflect.TypeOf((*MockConfig)(nil).GetId))
}

// GetProject mocks base method
func (m *MockConfig) GetProject() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProject")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetProject indicates an expected call of GetProject
func (mr *MockConfigMockRecorder) GetProject() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProject", reflect.TypeOf((*MockConfig)(nil).GetProject))
}

// GetProperties mocks base method
func (m *MockConfig) GetProperties() map[string]map[string]string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProperties")
	ret0, _ := ret[0].(map[string]map[string]string)
	return ret0
}

// GetProperties indicates an expected call of GetProperties
func (mr *MockConfigMockRecorder) GetProperties() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProperties", reflect.TypeOf((*MockConfig)(nil).GetProperties))
}

// GetRequiredByConfigIdList mocks base method
func (m *MockConfig) GetRequiredByConfigIdList() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRequiredByConfigIdList")
	ret0, _ := ret[0].([]string)
	return ret0
}

// GetRequiredByConfigIdList indicates an expected call of GetRequiredByConfigIdList
func (mr *MockConfigMockRecorder) GetRequiredByConfigIdList() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRequiredByConfigIdList", reflect.TypeOf((*MockConfig)(nil).GetRequiredByConfigIdList))
}

// addToRequiredByConfigIdList mocks base method
func (m *MockConfig) addToRequiredByConfigIdList(config string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "addToRequiredByConfigIdList", config)
}

// addToRequiredByConfigIdList indicates an expected call of addToRequiredByConfigIdList
func (mr *MockConfigMockRecorder) addToRequiredByConfigIdList(config interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "addToRequiredByConfigIdList", reflect.TypeOf((*MockConfig)(nil).addToRequiredByConfigIdList), config)
}

// MockConfigFactory is a mock of ConfigFactory interface
type MockConfigFactory struct {
	ctrl     *gomock.Controller
	recorder *MockConfigFactoryMockRecorder
}

// MockConfigFactoryMockRecorder is the mock recorder for MockConfigFactory
type MockConfigFactoryMockRecorder struct {
	mock *MockConfigFactory
}

// NewMockConfigFactory creates a new mock instance
func NewMockConfigFactory(ctrl *gomock.Controller) *MockConfigFactory {
	mock := &MockConfigFactory{ctrl: ctrl}
	mock.recorder = &MockConfigFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConfigFactory) EXPECT() *MockConfigFactoryMockRecorder {
	return m.recorder
}

// NewConfig mocks base method
func (m *MockConfigFactory) NewConfig(id, project, fileName string, properties map[string]map[string]string, api api.Api) (Config, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewConfig", id, project, fileName, properties, api)
	ret0, _ := ret[0].(Config)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewConfig indicates an expected call of NewConfig
func (mr *MockConfigFactoryMockRecorder) NewConfig(id, project, fileName, properties, api interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewConfig", reflect.TypeOf((*MockConfigFactory)(nil).NewConfig), id, project, fileName, properties, api)
}
