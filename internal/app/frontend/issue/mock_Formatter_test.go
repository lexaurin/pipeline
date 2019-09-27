// Code generated by mockery v1.0.0. DO NOT EDIT.

package issue

import mock "github.com/stretchr/testify/mock"

// MockFormatter is an autogenerated mock type for the Formatter type
type MockFormatter struct {
	mock.Mock
}

// FormatIssue provides a mock function with given fields: data
func (_m *MockFormatter) FormatIssue(data NewIssueData) (string, error) {
	ret := _m.Called(data)

	var r0 string
	if rf, ok := ret.Get(0).(func(NewIssueData) string); ok {
		r0 = rf(data)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(NewIssueData) error); ok {
		r1 = rf(data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}