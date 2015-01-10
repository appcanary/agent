package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var baseURL = "http://localhost:8080"

type Client interface {
	HeartBeat() error
	Submit(string, interface{}) error
}

type CanaryClient struct {
	apiKey     string
	serverName string
}

func NewClient(apiKey string, serverName string) *CanaryClient {
	client := &CanaryClient{apiKey: apiKey, serverName: serverName}
	return client
}

func (c *CanaryClient) HeartBeat() error {
	body, err := json.Marshal(map[string]string{"server_name": c.serverName})
	if err != nil {
		return err
	}
	res, err := c.post("/v1/heartbeat", body)
	_ = res
	if err != nil {
		return err
	}
	return nil
}

func (c *CanaryClient) Submit(app string, data interface{}) error {
	//TODO
	return nil
}

func (c *CanaryClient) post(rPath string, body []byte) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", baseURL+rPath, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Canary-Api-Key", c.apiKey)
	return client.Do(req)
}
