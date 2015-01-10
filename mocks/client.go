package mocks

import "github.com/stretchr/testify/mock"

type Client struct {
	mock.Mock
}

func (m *Client) HeartBeat() error {
	ret := m.Called()
	r0 := ret.Error(0)

	return r0
}

func (m *Client) Submit(_a0 string, _a1 interface{}) error {
	ret := m.Called(_a0, _a1)
	r0 := ret.Error(0)

	return r0
}
