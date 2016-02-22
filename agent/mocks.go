package agent

import (
	"github.com/appcanary/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Heartbeat(_a0 string, _a1 Watchers) error {
	ret := m.Called()

	r0 := ret.Error(0)

	return r0
}
func (m *MockClient) SendFile(_a0 string, _a1 string, _a2 []byte) error {
	ret := m.Called()

	r0 := ret.Error(0)

	return r0
}
func (m *MockClient) CreateServer(_a0 *Server) (string, error) {
	return m.Called().String(0), nil
}
