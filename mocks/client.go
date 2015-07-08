package mocks

import "github.com/appcanary/testify/mock"

import models "github.com/appcanary/agent/agent/models"

type Client struct {
	mock.Mock
}

func (m *Client) Heartbeat(_a0 string, _a1 models.WatchedFiles) error {
	ret := m.Called()

	r0 := ret.Error(0)

	return r0
}
func (m *Client) SendFile(_a0 string, _a1 string, _a2 []byte) error {
	ret := m.Called()

	r0 := ret.Error(0)

	return r0
}
func (m *Client) CreateServer(_a0 *models.Server) (string, error) {
	return m.Called().String(0), nil
}
