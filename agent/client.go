package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/stateio/canary-agent/agent/server"
	"github.com/stateio/canary-agent/agent/umwelten"
)

var env = umwelten.Fetch()

var (
	ErrApi        = errors.New("api error")
	ErrDeprecated = errors.New("api deprecated")
)

type Client interface {
	HeartBeat() error
	Submit(string, interface{}) error
	CreateServer(*server.Server) error
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
	res, err := c.post(umwelten.API_HEARTBEAT, body)
	_ = res
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	type heartbeatResponse struct {
		Success bool
	}

	var t heartbeatResponse

	err = json.Unmarshal(b, &t)
	if err != nil {
		return err
	}
	// check for a false heartbeat response -- indicating deprecated API version
	if t.Success == false {
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

func (c *CanaryClient) CreateServer(srv *server.Server) error {
	body, err := json.Marshal(*srv)

	if err != nil {
		return err
	}

	res, err := c.post(umwelten.API_SERVERS, body)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, srv)
}

func (c *CanaryClient) post(rPath string, body []byte) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", env.BaseUrl+rPath, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Token "+c.apiKey)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, ErrApi
	}
	return res, nil
}
