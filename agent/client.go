package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

var baseURL = "http://localhost:8080"

var (
	ErrApi        = errors.New("api error")
	ErrDeprecated = errors.New("api deprecated")
)

type Client interface {
	HeartBeat() error
	Submit(string, interface{}) error
}

type CanaryClient struct {
	apiKey string
	server string
}

func NewClient(apiKey string, server string) *CanaryClient {
	client := &CanaryClient{apiKey: apiKey, server: server}
	return client
}

func (c *CanaryClient) HeartBeat() error {
	body, err := json.Marshal(map[string]string{"server": c.server})
	if err != nil {
		return err
	}
	res, err := c.post("/v1/heartbeat", body)
	_ = res
	if err != nil {
		return err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// check for a false heartbeat response -- indicating deprecated API version
	if string(resBody) == "false" {
		return ErrDeprecated
	}
	return nil
}

func (c *CanaryClient) Submit(app string, deps interface{}) error {
	depsJSON, err := json.Marshal(deps)
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]string{
		"server": c.server,
		"app":    app,
		"deps":   string(depsJSON),
	})
	if err != nil {
		return err
	}
	res, err := c.post("/v1/submit", body)
	_ = res
	if err != nil {
		return err
	}
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
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, ErrApi
	}
	return res, nil
}
