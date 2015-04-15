package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/stateio/canary-agent/agent/app"
	"github.com/stateio/canary-agent/agent/server"
	"github.com/stateio/canary-agent/agent/umwelten"
)

var env = umwelten.Fetch()

var (
	ErrApi        = errors.New("api error")
	ErrDeprecated = errors.New("api deprecated")
)

type Client interface {
	HeartBeat(string) error
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

func (self *CanaryClient) HeartBeat(uuid string) error {
	// TODO MAKE REAL APPS
	apps := []app.App{app.App{Name: "Foo", MonitoredFiles: "/var/www/foo/Gemfile.lock"}}

	body, err := json.Marshal(map[string][]app.App{"apps": apps})
	if err != nil {
		return err
	}

	respBody, err := self.post(umwelten.API_HEARTBEAT+uuid, body)
	if err != nil {
		return err
	}

	type heartbeatResponse struct {
		Heartbeat time.Time
	}

	var t heartbeatResponse

	err = json.Unmarshal(respBody, &t)
	log.Debug(fmt.Sprintf("Heartbeat: %s", t.Heartbeat))
	if err != nil {
		return err
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

	respBody, err := c.post(umwelten.API_SERVERS, body)
	if err != nil {
		return err
	}

	return json.Unmarshal(respBody, srv)
}

func (c *CanaryClient) post(rPath string, body []byte) ([]byte, error) {
	uri := env.BaseUrl + rPath
	client := &http.Client{}
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Ahem, http://stackoverflow.com/questions/17714494/golang-http-request-results-in-eof-errors-when-making-multiple-requests-successi
	req.Close = true

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Token "+c.apiKey)

	res, err := client.Do(req)
	defer res.Body.Close()

	if err != nil {
		log.Debug("Do err: ", err.Error())
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, errors.New(fmt.Sprintf("API Error: %d %s", res.StatusCode, uri))
	}

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return respBody, err
	}

	return respBody, nil
}
