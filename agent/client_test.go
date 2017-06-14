package agent

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	"github.com/appcanary/agent/conf"
	"github.com/appcanary/testify/suite"
)

type TestJsonRequest map[string]interface{}

type ClientTestSuite struct {
	suite.Suite
	apiKey     string
	serverUUID string
	files      Watchers
	client     Client
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (t *ClientTestSuite) SetupTest() {
	conf.InitEnv("test")
	t.apiKey = "my api key"
	t.serverUUID = "server uuid"

	dpkgPath := conf.DEV_CONF_PATH + "/dpkg/available"
	dpkgFile := NewFileWatcher(dpkgPath, testCallbackNOP)

	gemfilePath := conf.DEV_CONF_PATH + "/Gemfile.lock"
	gemfile := NewFileWatcher(gemfilePath, testCallbackNOP)

	t.files = Watchers{dpkgFile, gemfile}

	t.client = NewClient(t.apiKey, &Server{
		UUID: t.serverUUID,
		Tags: []string{"dogs", "webserver"},
	})
}

func (t *ClientTestSuite) TestHeartbeat() {
	env := conf.FetchEnv()
	serverInvoked := false
	time.Sleep(conf.TEST_POLL_SLEEP)
	ts := testServer(t, "POST", "{\"success\": true}", func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		t.Equal("Token "+t.apiKey, r.Header.Get("Authorization"), "heartbeat api key")

		jsonFiles := rBody["files"].([]interface{})

		// does the json we send look roughly like
		// it's supposed to?
		t.NotNil(jsonFiles)
		t.Equal(2, len(jsonFiles))
		monitoredFile := jsonFiles[0].(map[string]interface{})

		t.Equal("ubuntu", monitoredFile["kind"])
		t.NotNil(monitoredFile["path"])
		t.NotEqual("", monitoredFile["path"])
		t.NotNil(monitoredFile["updated-at"])
		t.NotEqual("", monitoredFile["updated-at"])
		t.Equal(true, monitoredFile["being-watched"])

		monitoredFile2 := jsonFiles[1].(map[string]interface{})

		t.Equal("gemfile", monitoredFile2["kind"])
		t.NotNil(monitoredFile2["path"])
		t.NotEqual("", monitoredFile2["path"])
		t.NotNil(monitoredFile2["updated-at"])
		t.NotEqual("", monitoredFile2["updated-at"])
		t.Equal(true, monitoredFile2["being-watched"])

		if rBody["tags"] == nil {
			t.Fail("tags should not be nil")
		} else {
			jsonTags := rBody["tags"].([]interface{})
			t.Equal("dogs", jsonTags[0].(string))
			t.Equal("webserver", jsonTags[1].(string))
		}
	})

	// the client uses BaseUrl to set up queries.
	env.BaseUrl = ts.URL

	// actual test execution
	t.client.Heartbeat(t.serverUUID, t.files)

	ts.Close()
	t.files[0].Stop()
	t.True(serverInvoked)
}

func (t *ClientTestSuite) TestSendProcessState() {
	env := conf.FetchEnv()

	serverInvoked := false
	ts := testServer(t, "PUT", "OK", func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		t.Equal("Token "+t.apiKey, r.Header.Get("Authorization"), "heartbeat api key")

		// TODO Test what was received
	})

	env.BaseUrl = ts.URL
	script := conf.DEV_CONF_PATH + "/pointless"

	cmd := exec.Command(script)
	err := cmd.Start()
	t.Nil(err)

	defer cmd.Process.Kill()

	done := make(chan bool)

	watcher := NewProcessWatcher("pointless", func(w Watcher) {
		wt := w.(ProcessWatcher)
		jsonBytes := wt.StateJson()
		t.NotNil(jsonBytes)

		var pm map[string]interface{}
		json.Unmarshal(jsonBytes, &pm)

		server := pm["server"]
		t.NotNil(server)

		serverM := server.(map[string]interface{})

		processMap := serverM["system_state"]
		t.NotNil(processMap)

		processMapM := processMap.(map[string]interface{})

		processes := processMapM["processes"]
		t.NotNil(processes)

		processesS := processes.([]interface{})

		var watchedProc map[string]interface{}
		for _, proc := range processesS {
			procM := proc.(map[string]interface{})
			if int(procM["pid"].(float64)) == cmd.Process.Pid {
				watchedProc = procM
				break
			}
		}

		t.NotNil(watchedProc)
		t.Equal(false, watchedProc["outdated"])
		t.NotNil(watchedProc["libraries"])
		t.NotNil(watchedProc["started"])

		// Note this will fail if `dpkg` is unavailable
		if len(watchedProc["libraries"].([]interface{})) == 0 {
			t.Fail("No libraries were found - could be dpkg is not installed?")
		}
		done <- true
	})

	t.NotNil(watcher.(ProcessWatcher))

	// kick things off
	watcher.Start()
	defer watcher.Stop()

	<-done // wait
}

func (t *ClientTestSuite) TestSendFile() {
	env := conf.FetchEnv()
	testFilePath := "/var/foo/whatever"

	serverInvoked := false
	ts := testServer(t, "PUT", "OK", func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		t.Equal("Token "+t.apiKey, r.Header.Get("Authorization"), "heartbeat api key")

		json := rBody

		t.Equal("", json["name"])
		t.Equal(testFilePath, json["path"])
		t.Equal("gemfile", json["kind"])
		t.NotEqual("", json["contents"])

	})

	env.BaseUrl = ts.URL

	contents, _ := t.files[0].(TextWatcher).Contents()
	t.client.SendFile(testFilePath, "gemfile", contents)

	ts.Close()
	t.True(serverInvoked)
}

func (t *ClientTestSuite) TestCreateServer() {
	env := conf.FetchEnv()

	server := NewServer(&conf.Conf{Tags: []string{"dogs", "webserver"}}, &conf.ServerConf{})

	testUUID := "12345"
	jsonResponse := "{\"uuid\":\"" + testUUID + "\"}"
	serverInvoked := false

	ts := testServer(t, "POST", jsonResponse, func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		t.Equal("Token "+t.apiKey, r.Header.Get("Authorization"), "heartbeat api key")

		json := rBody

		t.Equal(server.Hostname, json["hostname"])
		t.Equal(server.Uname, json["uname"])
		t.Equal(server.Ip, json["ip"])
		t.Nil(json["uuid"])

		if json["tags"] == nil {
			t.Fail("tags should not be nil")
		} else {
			tags := json["tags"].([]interface{})
			t.Equal(server.Tags[0], tags[0].(string))
			t.Equal(server.Tags[1], tags[1].(string))
		}
	})

	env.BaseUrl = ts.URL
	responseUUID, _ := t.client.CreateServer(server)
	ts.Close()
	t.True(serverInvoked)
	t.Equal(testUUID, responseUUID)
}

func (t *ClientTestSuite) TestFetchUpgradeablePackages() {
	env := conf.FetchEnv()

	jsonResponse := "{\"libkrb5-3\":\"1.12+dfsg-2ubuntu5.2\",\"isc-dhcp-client\":\"4.2.4-7ubuntu12.4\"}"
	serverInvoked := false
	ts := testServerSansInput(t, "GET", jsonResponse, func(r *http.Request, rBody TestJsonRequest) {
		serverInvoked = true

		t.Equal("Token "+t.apiKey, r.Header.Get("Authorization"), "heartbeat api key")
	})

	env.BaseUrl = ts.URL
	packageList, _ := t.client.FetchUpgradeablePackages()
	ts.Close()

	t.Equal("1.12+dfsg-2ubuntu5.2", packageList["libkrb5-3"])
	t.True(serverInvoked)
}

func testCallbackNOP(foo Watcher) {
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

func testServerSansInput(assert *ClientTestSuite, method string, respondWithBody string, callback func(*http.Request, TestJsonRequest)) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(method, r.Method, "method")
		assert.Equal("application/json", r.Header.Get("Content-Type"), "content type")

		body, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()

		var datBody TestJsonRequest
		if len(body) > 0 {
			if err := json.Unmarshal(body, &datBody); err != nil {
				panic(err)
			}
		}

		callback(r, datBody)
		tsrespond(w, 200, respondWithBody)
	}))

	return ts
}

// TODO: handle pathological cases, error handling?
