package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	_ "crypto/sha512"
	//http://bridge.grumpy-troll.org/2014/05/golang-tls-comodo/
	"github.com/appcanary/agent/Godeps/_workspace/src/github.com/cenkalti/backoff"
	. "github.com/appcanary/agent/agent/models"
	"github.com/appcanary/agent/agent/umwelten"
)

var env = umwelten.Fetch()

var (
	ErrApi        = errors.New("api error")
	ErrDeprecated = errors.New("api deprecated")
)

type Client interface {
	Heartbeat(string, WatchedFiles) error
	SendFile(string, []byte) error
	CreateServer(*Server) (string, error)
}

type CanaryClient struct {
	apiKey string
	server *Server
}

func NewClient(apiKey string, server *Server) *CanaryClient {
	client := &CanaryClient{apiKey: apiKey, server: server}
	return client
}

func (client *CanaryClient) Heartbeat(uuid string, files WatchedFiles) error {

	body, err := json.Marshal(map[string]WatchedFiles{"files": files})

	if err != nil {
		return err
	}

	// TODO SANITIZE UUID input cos this feels abusable
	respBody, err := client.post(umwelten.ApiHeartbeatPath(uuid), body)

	if err != nil {
		return err
	}

	type heartbeatResponse struct {
		Heartbeat time.Time
	}

	var t heartbeatResponse

	// TODO: do something with heartbeat resp
	err = json.Unmarshal(respBody, &t)
	log.Debug(fmt.Sprintf("Heartbeat: %s", t.Heartbeat))
	if err != nil {
		return err
	}

	return nil
}

func (client *CanaryClient) SendFile(path string, contents []byte) error {
	file_json, err := json.Marshal(map[string]string{
		"name":     "",
		"path":     path,
		"kind":     "gemfile",
		"contents": string(contents),
	})

	if err != nil {
		return err
	}

	_, err = client.put(umwelten.ApiServerPath(client.server.UUID), file_json)

	return err

}

func (c *CanaryClient) CreateServer(srv *Server) (string, error) {
	body, err := json.Marshal(*srv)

	if err != nil {
		return "", err
	}

	respBody, err := c.post(umwelten.ApiServersPath(), body)
	if err != nil {
		return "", err
	}

	var respServer struct {
		UUID string `json:"uuid"`
	}

	json.Unmarshal(respBody, &respServer)
	return respServer.UUID, nil
}

func (client *CanaryClient) post(rPath string, body []byte) ([]byte, error) {
	return client.send("POST", rPath, body)
}

func (client *CanaryClient) put(rPath string, body []byte) ([]byte, error) {
	return client.send("PUT", rPath, body)
}

func (c *CanaryClient) send(method string, uri string, body []byte) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Ahem, http://stackoverflow.com/questions/17714494/golang-http-request-results-in-eof-errors-when-making-multiple-requests-successi
	req.Close = true

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Token "+c.apiKey)

	var res *http.Response

	// if the request fails for whatever reason, keep
	// trying to reach the server
	err = backoff.Retry(func() error {
		log.Debug("Request: %s %s", method, uri)
		res, err = client.Do(req)
		if err != nil {
			log.Error("Error in request %s", err)
		}

		return err
	},
		backoff.NewExponentialBackOff())

	if err != nil {
		log.Debug("Do err: ", err.Error())
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, errors.New(fmt.Sprintf("API Error: %d %s", res.StatusCode, uri))
	}

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return respBody, err
	}

	return respBody, nil
}
