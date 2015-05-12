package mocks

import "github.com/stateio/testify/mock"

type File struct {
	mock.Mock
}

func (m *File) GetPath() string {
	args := m.Mock.Called()
	return args.String(0)
}

func (m *File) Parse() interface{} {
	args := m.Mock.Called()
	return args.Get(0)
}
