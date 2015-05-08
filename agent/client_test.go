package agent

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stateio/canary-agent/agent/models"
	"github.com/stateio/canary-agent/agent/umwelten"
	"github.com/stretchr/testify/suite"
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

func (self *ClientTestSuite) SetupTest() {
	umwelten.Init("test")
	self.api_key = "my api key"
	self.server_uuid = "server uuid"

	filePath := NewConfFromEnv().Files[0].Path
	file := models.NewWatchedFile(filePath, testCallbackNOP)
	self.files = models.WatchedFiles{file}

	self.client = NewClient(self.api_key, &models.Server{UUID: self.server_uuid})

}

func (self *ClientTestSuite) TestHeartbeat() {

	serverInvoked := false
	ts := testServer(self, "POST", "{\"success\": true}", func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		self.Equal(r.Header.Get("Authorization"), "Token "+self.api_key, "heartbeat api key")

		json_files := rBody["files"].([]interface{})

		// does the json we send look roughly like
		// it's supposed to?
		self.NotNil(json_files)
		monitored_file := json_files[0].(map[string]interface{})

		self.Equal(monitored_file["kind"], "gemfile")
		self.NotNil(monitored_file["path"])
		self.NotEqual(monitored_file["path"], "")
		self.NotNil(monitored_file["updated-at"])
		self.NotEqual(monitored_file["updated-at"], "")
	})

	// the client uses BaseUrl to set up queries.
	env.BaseUrl = ts.URL

	// actual test execution
	self.client.Heartbeat(self.server_uuid, self.files)

	ts.Close()
	self.files[0].RemoveHook()
	self.True(serverInvoked)
}

func (self *ClientTestSuite) TestSendFile() {
	test_file_path := "/var/foo/whatever"

	serverInvoked := false
	ts := testServer(self, "PUT", "OK", func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		self.Equal(r.Header.Get("Authorization"), "Token "+self.api_key, "heartbeat api key")

		json := rBody

		self.Equal(json["name"], "")
		self.Equal(json["path"], test_file_path)
		self.Equal(json["kind"], "gemfile")
		self.NotEqual(json["contents"], "")

	})

	env.BaseUrl = ts.URL

	contents, _ := self.files[0].Contents()
	self.client.SendFile(test_file_path, contents)

	ts.Close()
	self.True(serverInvoked)
}

func (self *ClientTestSuite) TestCreateServer() {
	server := models.ThisServer("")

	test_uuid := "12345"
	json_response := "{\"uuid\":\"" + test_uuid + "\"}"
	serverInvoked := false

	ts := testServer(self, "POST", json_response, func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		self.Equal(r.Header.Get("Authorization"), "Token "+self.api_key, "heartbeat api key")

		json := rBody

		self.Equal(json["hostname"], server.Hostname)
		self.Equal(json["uname"], server.Uname)
		self.Equal(json["ip"], server.Ip)
		self.Nil(json["uuid"])
	})

	env.BaseUrl = ts.URL
	response_uuid, _ := self.client.CreateServer(server)
	ts.Close()
	self.True(serverInvoked)
	self.Equal(test_uuid, response_uuid)
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
