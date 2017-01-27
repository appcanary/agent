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

func (m *MockClient) SendProcessState(_a0 string, _a1 *processMap) error {
	return m.Called().Error(0)
}

func (m *MockClient) CreateServer(_a0 *Server) (string, error) {
	return m.Called().String(0), nil
}

func (m *MockClient) FetchUpgradeablePackages() (map[string]string, error) {
	ret := m.Called()

	var r0 map[string]string
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(map[string]string)
	}
	r1 := ret.Error(1)

	return r0, r1
}
