package mock

import (
	gomock "github.com/golang/mock/gomock"
	pgx "github.com/jackc/pgx"
	reflect "reflect"
)

// DatabaseDriver is a mock of DatabaseDriver interface
type DatabaseDriver struct {
	ctrl     *gomock.Controller
	recorder *DatabaseDriverMockRecorder
}

// DatabaseDriverMockRecorder is the mock recorder for DatabaseDriver
type DatabaseDriverMockRecorder struct {
	mock *DatabaseDriver
}

// NewDatabaseDriver creates a new mock instance
func NewDatabaseDriver(ctrl *gomock.Controller) *DatabaseDriver {
	mock := &DatabaseDriver{ctrl: ctrl}
	mock.recorder = &DatabaseDriverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *DatabaseDriver) EXPECT() *DatabaseDriverMockRecorder {
	return _m.recorder
}

// Query mocks base method
func (_m *DatabaseDriver) Query(_param0 string, _param1 ...interface{}) (*pgx.Rows, error) {
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
func (_mr *DatabaseDriverMockRecorder) Query(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0}, arg1...)
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Query", reflect.TypeOf((*DatabaseDriver)(nil).Query), _s...)
}

// QueryRow mocks base method
func (_m *DatabaseDriver) QueryRow(_param0 string, _param1 ...interface{}) *pgx.Row {
	_s := []interface{}{_param0}
	for _, _x := range _param1 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "QueryRow", _s...)
	ret0, _ := ret[0].(*pgx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow
func (_mr *DatabaseDriverMockRecorder) QueryRow(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0}, arg1...)
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "QueryRow", reflect.TypeOf((*DatabaseDriver)(nil).QueryRow), _s...)
}
