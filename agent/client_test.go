package agent

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stateio/canary-agent/agent/models"
	"github.com/stateio/canary-agent/agent/umwelten"
	"github.com/stateio/testify/suite"
)

type TestJsonRequest map[string]interface{}

//[]map[string]string

type ClientTestSuite struct {
	suite.Suite
	api_key     string
	server_uuid string
	files       models.WatchedFiles
	client      Client
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (t *ClientTestSuite) SetupTest() {
	umwelten.Init("test")
	t.api_key = "my api key"
	t.server_uuid = "server uuid"

	// it needs an ARBITRARY file to watch
	// and the content of the conf file are
	// absolute paths; as a workaround:
	filePath := env.ConfFile
	file := models.NewWatchedFileWithHook(filePath, testCallbackNOP)
	t.files = models.WatchedFiles{file}

	t.client = NewClient(t.api_key, &models.Server{UUID: t.server_uuid})

}

func (t *ClientTestSuite) TestHeartbeat() {

	serverInvoked := false
	ts := testServer(t, "POST", "{\"success\": true}", func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		t.Equal(r.Header.Get("Authorization"), "Token "+t.api_key, "heartbeat api key")

		json_files := rBody["files"].([]interface{})

		// does the json we send look roughly like
		// it's supposed to?
		t.NotNil(json_files)
		monitored_file := json_files[0].(map[string]interface{})

		t.Equal(monitored_file["kind"], "gemfile")
		t.NotNil(monitored_file["path"])
		t.NotEqual(monitored_file["path"], "")
		t.NotNil(monitored_file["updated-at"])
		t.NotEqual(monitored_file["updated-at"], "")
		t.Equal(monitored_file["being-watched"], true)
	})

	// the client uses BaseUrl to set up queries.
	env.BaseUrl = ts.URL

	// actual test execution
	t.client.Heartbeat(t.server_uuid, t.files)

	ts.Close()
	t.files[0].RemoveHook()
	t.True(serverInvoked)
}

func (t *ClientTestSuite) TestSendFile() {
	test_file_path := "/var/foo/whatever"

	serverInvoked := false
	ts := testServer(t, "PUT", "OK", func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		t.Equal(r.Header.Get("Authorization"), "Token "+t.api_key, "heartbeat api key")

		json := rBody

		t.Equal(json["name"], "")
		t.Equal(json["path"], test_file_path)
		t.Equal(json["kind"], "gemfile")
		t.NotEqual(json["contents"], "")

	})

	env.BaseUrl = ts.URL

	contents, _ := t.files[0].Contents()
	t.client.SendFile(test_file_path, contents)

	ts.Close()
	t.True(serverInvoked)
}

func (t *ClientTestSuite) TestCreateServer() {
	server := models.ThisServer("")

	test_uuid := "12345"
	json_response := "{\"uuid\":\"" + test_uuid + "\"}"
	serverInvoked := false

	ts := testServer(t, "POST", json_response, func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		t.Equal(r.Header.Get("Authorization"), "Token "+t.api_key, "heartbeat api key")

		json := rBody

		t.Equal(json["hostname"], server.Hostname)
		t.Equal(json["uname"], server.Uname)
		t.Equal(json["ip"], server.Ip)
		t.Nil(json["uuid"])
	})

	env.BaseUrl = ts.URL
	response_uuid, _ := t.client.CreateServer(server)
	ts.Close()
	t.True(serverInvoked)
	t.Equal(test_uuid, response_uuid)
}

func testCallbackNOP(foo *models.WatchedFile) {
	// NOP
}

//Sends an http.ResponseWriter a string and status
func tsrespond(w http.ResponseWriter, status int, v string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(v))
}

func testServer(assert *ClientTestSuite, method string, respondWithBody string, callback func(*http.Request, TestJsonRequest)) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(r.Method, method, "method")
		assert.Equal(r.Header.Get("Content-Type"), "application/json", "content type")

		body, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()

		var datBody TestJsonRequest
		if err := json.Unmarshal(body, &datBody); err != nil {
			panic(err)
		}

		callback(r, datBody)
		tsrespond(w, 200, respondWithBody)
	}))

	return ts
}

// TODO: handle pathological cases, error handling?
