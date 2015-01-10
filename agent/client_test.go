package agent

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	assert := assert.New(t)

	client := NewClient("my api key", "my server")
	assert.Equal(client.apiKey, "my api key", "api key")
	assert.Equal(client.serverName, "my server", "server name")
}

func TestHeartBeat(t *testing.T) {
	assert := assert.New(t)
	serverInvoked := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverInvoked = true
		assert.Equal(r.Method, "POST")
		assert.Equal(r.URL.Path, "/v1/heartbeat")
		assert.Equal(r.Header.Get("Content-Type"), "application/json")

		//authentication header
		assert.Equal(r.Header.Get("X-Canary-Api-Key"), "my api key")

		//server name in body
		expectedBody, _ := json.Marshal(map[string]string{"server_name": "my server"})
		body, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		assert.Equal(body, expectedBody)
		respond(w, 200, "true")
	}))
	defer ts.Close()

	//overwrite the base URL to our testing server
	baseURL = ts.URL

	client := NewClient("my api key", "my server")
	err := client.HeartBeat()
	assert.NoError(err)

	assert.True(serverInvoked, "server invoked")
}

//Sends an http.ResponseWriter a string and status
func respond(w http.ResponseWriter, status int, v string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(v))
}
