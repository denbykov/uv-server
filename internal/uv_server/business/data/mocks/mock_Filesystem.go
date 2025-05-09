// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// MockFilesystem is an autogenerated mock type for the Filesystem type
type MockFilesystem struct {
	mock.Mock
}

type MockFilesystem_Expecter struct {
	mock *mock.Mock
}

func (_m *MockFilesystem) EXPECT() *MockFilesystem_Expecter {
	return &MockFilesystem_Expecter{mock: &_m.Mock}
}

// DeleteFile provides a mock function with given fields: path
func (_m *MockFilesystem) DeleteFile(path string) error {
	ret := _m.Called(path)

	if len(ret) == 0 {
		panic("no return value specified for DeleteFile")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockFilesystem_DeleteFile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteFile'
type MockFilesystem_DeleteFile_Call struct {
	*mock.Call
}

// DeleteFile is a helper method to define mock.On call
//   - path string
func (_e *MockFilesystem_Expecter) DeleteFile(path interface{}) *MockFilesystem_DeleteFile_Call {
	return &MockFilesystem_DeleteFile_Call{Call: _e.mock.On("DeleteFile", path)}
}

func (_c *MockFilesystem_DeleteFile_Call) Run(run func(path string)) *MockFilesystem_DeleteFile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockFilesystem_DeleteFile_Call) Return(_a0 error) *MockFilesystem_DeleteFile_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockFilesystem_DeleteFile_Call) RunAndReturn(run func(string) error) *MockFilesystem_DeleteFile_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockFilesystem creates a new instance of MockFilesystem. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockFilesystem(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockFilesystem {
	mock := &MockFilesystem{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
