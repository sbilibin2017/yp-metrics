// Code generated by MockGen. DO NOT EDIT.
// Source: /home/sergey/Go/yp-metrics/internal/repositories/metric_get.go

// Package repositories is a generated GoMock package.
package repositories

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/sbilibin2017/yp-metrics/internal/types"
)

// MockGetter is a mock of Getter interface.
type MockGetter struct {
	ctrl     *gomock.Controller
	recorder *MockGetterMockRecorder
}

// MockGetterMockRecorder is the mock recorder for MockGetter.
type MockGetterMockRecorder struct {
	mock *MockGetter
}

// NewMockGetter creates a new mock instance.
func NewMockGetter(ctrl *gomock.Controller) *MockGetter {
	mock := &MockGetter{ctrl: ctrl}
	mock.recorder = &MockGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGetter) EXPECT() *MockGetterMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockGetter) Get(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, id)
	ret0, _ := ret[0].(*types.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockGetterMockRecorder) Get(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockGetter)(nil).Get), ctx, id)
}
