package agent

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"net/http"
	"time"

	_ "crypto/sha512"
	//http://bridge.grumpy-troll.org/2014/05/golang-tls-comodo/

	"github.com/cenkalti/backoff"
)

var (
	ErrApi        = errors.New("api error")
	ErrDeprecated = errors.New("api deprecated")
)

type Client interface {
	Heartbeat(string, Watchers) error
	SendFile(string, string, []byte) error
	CreateServer(*Server) (string, error)
	FetchUpgradeablePackages() (map[string]string, error)
}

type CanaryClient struct {
	apiKey string
	server *Server
}

func NewClient(apiKey string, server *Server) *CanaryClient {
	client := &CanaryClient{apiKey: apiKey, server: server}
	return client
}

func (client *CanaryClient) Heartbeat(uuid string, files Watchers) error {

	body, err := json.Marshal(map[string]interface{}{"files": files, "agent-version": CanaryVersion, "distro": client.server.Distro, "release": client.server.Release})

	if err != nil {
		return err
	}

	// TODO SANITIZE UUID input cos this feels abusable
	respBody, err := client.post(ApiHeartbeatPath(uuid), body)

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

func (client *CanaryClient) SendFile(path string, kind string, contents []byte) error {
	// Compute checksum of the file (not base64 encoding)
	crc := crc32.ChecksumIEEE(contents)
	// File needs to be sent base64 encoded
	b64buffer := new(bytes.Buffer)
	b64enc := base64.NewEncoder(base64.StdEncoding, b64buffer)
	b64enc.Write(contents)
	b64enc.Close()

	file_json, err := json.Marshal(map[string]interface{}{
		"name":     "",
		"path":     path,
		"kind":     kind,
		"contents": string(b64buffer.Bytes()),
		"crc":      crc,
	})

	if err != nil {
		return err
	}

	_, err = client.put(ApiServerPath(client.server.UUID), file_json)

	return err

}

func (c *CanaryClient) CreateServer(srv *Server) (string, error) {
	body, err := json.Marshal(*srv)

	if err != nil {
		return "", err
	}

	respBody, err := c.post(ApiServersPath(), body)
	if err != nil {
		return "", err
	}

	var respServer struct {
		UUID string `json:"uuid"`
	}

	json.Unmarshal(respBody, &respServer)
	return respServer.UUID, nil
}

func (client *CanaryClient) FetchUpgradeablePackages() (map[string]string, error) {
	respBody, err := client.get(ApiServerPath(client.server.UUID))

	if err != nil {
		return nil, err
	}

	var package_list map[string]string
	err = json.Unmarshal(respBody, &package_list)

	if err != nil {
		return nil, err
	}

	return package_list, nil
}

func (client *CanaryClient) post(rPath string, body []byte) ([]byte, error) {
	return client.send("POST", rPath, body)
}

func (client *CanaryClient) put(rPath string, body []byte) ([]byte, error) {
	return client.send("PUT", rPath, body)
}

func (client *CanaryClient) get(rPath string) ([]byte, error) {
	return client.send("GET", rPath, []byte{})
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
		log.Debugf("Request: %s %s", method, uri)
		res, err = client.Do(req)
		if err != nil {
			log.Errorf("Error in request %s", err)
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
