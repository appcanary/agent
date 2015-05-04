package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/stateio/canary-agent/agent/models"
	"github.com/stateio/canary-agent/agent/umwelten"
)

var env = umwelten.Fetch()

var (
	ErrApi        = errors.New("api error")
	ErrDeprecated = errors.New("api deprecated")
)

type Client interface {
	HeartBeat(string, WatchedFiles) error
	SendFile(string, []byte) error
	CreateServer(*Server) error
}

type CanaryClient struct {
	apiKey string
	server *Server
}

func NewClient(apiKey string, server *Server) *CanaryClient {
	client := &CanaryClient{apiKey: apiKey, server: server}
	return client
}

func (self *CanaryClient) HeartBeat(uuid string, files WatchedFiles) error {

	body, err := json.Marshal(map[string]WatchedFiles{"files": files})

	if err != nil {
		return err
	}

	// TODO SANITIZE UUID input cos this feels abusable
	respBody, err := self.post(umwelten.API_HEARTBEAT+"/"+uuid, body)

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

func (self *CanaryClient) SendFile(path string, contents []byte) error {
	file_json, err := json.Marshal(map[string]string{
		"name":     "",
		"path":     path,
		"kind":     "gemfile",
		"contents": string(contents),
	})

	if err != nil {
		return err
	}

	_, err = self.put(umwelten.API_SERVERS+"/"+self.server.UUID, file_json)

	return err

}

func (c *CanaryClient) CreateServer(srv *Server) error {
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

func (self *CanaryClient) post(rPath string, body []byte) ([]byte, error) {
	return self.send("POST", rPath, body)
}

func (self *CanaryClient) put(rPath string, body []byte) ([]byte, error) {
	return self.send("PUT", rPath, body)
}

func (c *CanaryClient) send(method string, rPath string, body []byte) ([]byte, error) {
	uri := env.BaseUrl + rPath
	client := &http.Client{}
	req, err := http.NewRequest(method, uri, bytes.NewBuffer(body))
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
