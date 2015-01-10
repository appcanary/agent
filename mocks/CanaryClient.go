package mocks

import "github.com/stretchr/testify/mock"

type CanaryClient struct {
	mock.Mock
}

func (m *CanaryClient) HeartBeat() bool {
	ret := m.Called()

	r0 := ret.Get(0).(bool)

	return r0
}
func (m *CanaryClient) CheckServer(_a0 string) bool {
	ret := m.Called(_a0)

	r0 := ret.Get(0).(bool)

	return r0
}
func (m *CanaryClient) RegisterServer(_a0 string) {
	m.Called(_a0)
}
func (m *CanaryClient) RegisterApp(_a0 string) {
	m.Called(_a0)
}
func (m *CanaryClient) Submit(_a0 string, _a1 interface{}) {
	m.Called(_a0, _a1)
}
