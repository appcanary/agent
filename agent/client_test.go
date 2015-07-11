package agent

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/appcanary/testify/suite"
)

type TestJsonRequest map[string]interface{}

//[]map[string]string

type ClientTestSuite struct {
	suite.Suite
	api_key     string
	server_uuid string
	files       WatchedFiles
	client      Client
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (t *ClientTestSuite) SetupTest() {
	InitEnv("test")
	t.api_key = "my api key"
	t.server_uuid = "server uuid"

	dpkgPath := DEV_CONF_PATH + "/dpkg/available"
	dpkgFile := NewWatchedFileWithHook(dpkgPath, testCallbackNOP)

	gemfilePath := DEV_CONF_PATH + "/Gemfile.lock"
	gemfile := NewWatchedFileWithHook(gemfilePath, testCallbackNOP)

	t.files = WatchedFiles{dpkgFile, gemfile}

	t.client = NewClient(t.api_key, &Server{UUID: t.server_uuid})

}

func (t *ClientTestSuite) TestHeartbeat() {

	serverInvoked := false
	time.Sleep(TEST_POLL_SLEEP)
	ts := testServer(t, "POST", "{\"success\": true}", func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		t.Equal("Token "+t.api_key, r.Header.Get("Authorization"), "heartbeat api key")

		json_files := rBody["files"].([]interface{})

		// does the json we send look roughly like
		// it's supposed to?
		t.NotNil(json_files)
		t.Equal(2, len(json_files))
		monitored_file := json_files[0].(map[string]interface{})

		t.Equal("ubuntu", monitored_file["kind"])
		t.NotNil(monitored_file["path"])
		t.NotEqual("", monitored_file["path"])
		t.NotNil(monitored_file["updated-at"])
		t.NotEqual("", monitored_file["updated-at"])
		t.Equal(true, monitored_file["being-watched"])

		monitored_file2 := json_files[1].(map[string]interface{})

		t.Equal("gemfile", monitored_file2["kind"])
		t.NotNil(monitored_file2["path"])
		t.NotEqual("", monitored_file2["path"])
		t.NotNil(monitored_file2["updated-at"])
		t.NotEqual("", monitored_file2["updated-at"])
		t.Equal(true, monitored_file2["being-watched"])
	})

	// the client uses BaseUrl to set up queries.
	env.BaseUrl = ts.URL

	// actual test execution
	t.client.Heartbeat(t.server_uuid, t.files)

	ts.Close()
	t.files[0].StopListening()
	t.True(serverInvoked)
}

func (t *ClientTestSuite) TestSendFile() {
	test_file_path := "/var/foo/whatever"

	serverInvoked := false
	ts := testServer(t, "PUT", "OK", func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		t.Equal("Token "+t.api_key, r.Header.Get("Authorization"), "heartbeat api key")

		json := rBody

		t.Equal("", json["name"])
		t.Equal(test_file_path, json["path"])
		t.Equal("gemfile", json["kind"])
		t.NotEqual("", json["contents"])

	})

	env.BaseUrl = ts.URL

	contents, _ := t.files[0].Contents()
	t.client.SendFile(test_file_path, "gemfile", contents)

	ts.Close()
	t.True(serverInvoked)
}

func (t *ClientTestSuite) TestCreateServer() {
	server := NewServer(&ServerConf{})

	test_uuid := "12345"
	json_response := "{\"uuid\":\"" + test_uuid + "\"}"
	serverInvoked := false

	ts := testServer(t, "POST", json_response, func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		t.Equal("Token "+t.api_key, r.Header.Get("Authorization"), "heartbeat api key")

		json := rBody

		t.Equal(server.Hostname, json["hostname"])
		t.Equal(server.Uname, json["uname"])
		t.Equal(server.Ip, json["ip"])
		t.Nil(json["uuid"])
	})

	env.BaseUrl = ts.URL
	response_uuid, _ := t.client.CreateServer(server)
	ts.Close()
	t.True(serverInvoked)
	t.Equal(test_uuid, response_uuid)
}

func testCallbackNOP(foo *WatchedFile) {
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
		assert.Equal(method, r.Method, "method")
		assert.Equal("application/json", r.Header.Get("Content-Type"), "content type")

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
