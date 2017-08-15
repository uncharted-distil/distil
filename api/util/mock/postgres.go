package mock

import (
	gomock "github.com/golang/mock/gomock"
	pgx "github.com/jackc/pgx"
	reflect "reflect"
)

// MockDatabaseDriver is a mock of DatabaseDriver interface
type MockDatabaseDriver struct {
	ctrl     *gomock.Controller
	recorder *MockDatabaseDriverMockRecorder
}

// MockDatabaseDriverMockRecorder is the mock recorder for MockDatabaseDriver
type MockDatabaseDriverMockRecorder struct {
	mock *MockDatabaseDriver
}

// NewMockDatabaseDriver creates a new mock instance
func NewMockDatabaseDriver(ctrl *gomock.Controller) *MockDatabaseDriver {
	mock := &MockDatabaseDriver{ctrl: ctrl}
	mock.recorder = &MockDatabaseDriverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockDatabaseDriver) EXPECT() *MockDatabaseDriverMockRecorder {
	return _m.recorder
}

// Query mocks base method
func (_m *MockDatabaseDriver) Query(_param0 string, _param1 ...interface{}) (*pgx.Rows, error) {
	_s := []interface{}{_param0}
	for _, _x := range _param1 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "Query", _s...)
	ret0, _ := ret[0].(*pgx.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (_mr *MockDatabaseDriverMockRecorder) Query(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0}, arg1...)
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Query", reflect.TypeOf((*MockDatabaseDriver)(nil).Query), _s...)
}
