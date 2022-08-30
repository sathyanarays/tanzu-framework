// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vmware-tanzu/tanzu-framework/tkg/avi (interfaces: Client)

// Package avi is a generated GoMock package.
package avi

import (
	reflect "reflect"

	models "github.com/avinetworks/sdk/go/models"
	gomock "github.com/golang/mock/gomock"

	models0 "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// VerifyAccount mocks base method
func (m *MockClient) VerifyAccount(params *models0.AviControllerParams) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyAccount", params)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifyAccount indicates an expected call of VerifyAccount
func (mr *MockClientMockRecorder) VerifyAccount(params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyAccount", reflect.TypeOf((*MockClient)(nil).VerifyAccount), params)
}

// GetClouds mocks base method
func (m *MockClient) GetClouds() ([]*models0.AviCloud, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClouds")
	ret0, _ := ret[0].([]*models0.AviCloud)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClouds indicates an expected call of GetClouds
func (mr *MockClientMockRecorder) GetClouds() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClouds", reflect.TypeOf((*MockClient)(nil).GetClouds))
}

// GetServiceEngineGroups mocks base method
func (m *MockClient) GetServiceEngineGroups() ([]*models0.AviServiceEngineGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServiceEngineGroups")
	ret0, _ := ret[0].([]*models0.AviServiceEngineGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetServiceEngineGroups indicates an expected call of GetServiceEngineGroups
func (mr *MockClientMockRecorder) GetServiceEngineGroups() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServiceEngineGroups", reflect.TypeOf((*MockClient)(nil).GetServiceEngineGroups))
}

// GetVipNetworks mocks base method
func (m *MockClient) GetVipNetworks() ([]*models0.AviVipNetwork, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVipNetworks")
	ret0, _ := ret[0].([]*models0.AviVipNetwork)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVipNetworks indicates an expected call of GetVipNetworks
func (mr *MockClientMockRecorder) GetVipNetworks() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVipNetworks", reflect.TypeOf((*MockClient)(nil).GetVipNetworks))
}

// GetCloudByName mocks base method
func (m *MockClient) GetCloudByName(name string) (*models.Cloud, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCloudByName", name)
	ret0, _ := ret[0].(*models.Cloud)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCloudByName indicates an expected call of GetCloudByName
func (mr *MockClientMockRecorder) GetCloudByName(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCloudByName", reflect.TypeOf((*MockClient)(nil).GetCloudByName), name)
}

// GetServiceEngineGroupByName mocks base method
func (m *MockClient) GetServiceEngineGroupByName(name string) (*models.ServiceEngineGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServiceEngineGroupByName", name)
	ret0, _ := ret[0].(*models.ServiceEngineGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetServiceEngineGroupByName indicates an expected call of GetServiceEngineGroupByName
func (mr *MockClientMockRecorder) GetServiceEngineGroupByName(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServiceEngineGroupByName", reflect.TypeOf((*MockClient)(nil).GetServiceEngineGroupByName), name)
}

// GetVipNetworkByName mocks base method
func (m *MockClient) GetVipNetworkByName(name string) (*models.Network, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVipNetworkByName", name)
	ret0, _ := ret[0].(*models.Network)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVipNetworkByName indicates an expected call of GetVipNetworkByName
func (mr *MockClientMockRecorder) GetVipNetworkByName(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVipNetworkByName", reflect.TypeOf((*MockClient)(nil).GetVipNetworkByName), name)
}
