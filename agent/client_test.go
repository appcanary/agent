package agent

/*
func TestNewClient(t *testing.T) {
	assert := assert.New(t)

	client := NewClient("my api key", "my server")
	assert.Equal(client.apiKey, "my api key", "api key")
	assert.Equal(client.server, "my server", "server name")
}

func TestHeartBeat(t *testing.T) {
	assert := assert.New(t)
	serverInvoked := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverInvoked = true
		assert.Equal(r.Method, "POST", "heartbeat method")
		assert.Equal(r.URL.Path, "/v1/heartbeat", "heartbeat path")
		assert.Equal(r.Header.Get("Content-Type"), "application/json", "heartbeat content type")

		//authentication header
		assert.Equal(r.Header.Get("X-Canary-Api-Key"), "my api key", "heartbeat api key")

		//server name in body
		expectedBody, _ := json.Marshal(map[string]string{"server": "my server"})
		body, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		assert.Equal(body, expectedBody, "heartbeat body")
		respond(w, 200, "{\"success\": true}")
	}))
	defer ts.Close()

	//overwrite the base URL to our testing server
	baseURL = ts.URL

	client := NewClient("my api key", "my server")
	err := client.HeartBeat()
	assert.NoError(err, "heartbeat error")

	assert.True(serverInvoked, "server invoked")
}

func TestHeartBeatDeprecated(t *testing.T) {
	assert := assert.New(t)
	serverInvoked := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverInvoked = true
		respond(w, 200, "{\"success\": false}")
	}))
	defer ts.Close()

	//overwrite the base URL to our testing server
	baseURL = ts.URL

	client := NewClient("my api key", "my server")
	err := client.HeartBeat()
	assert.Equal(err, ErrDeprecated, "false heartbeat response")

	assert.True(serverInvoked, "server invoked")
}

func TestHeartBeatErrorHanding(t *testing.T) {
	assert := assert.New(t)
	serverInvoked := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverInvoked = true
		respond(w, 500, "")
	}))
	defer ts.Close()

	//overwrite the base URL to our testing server
	baseURL = ts.URL

	client := NewClient("my api key", "my server")
	err := client.HeartBeat()
	assert.Equal(err, ErrApi, "error with api serve")

	assert.True(serverInvoked, "server invoked")
}

func TestSubmit(t *testing.T) {
	assert := assert.New(t)
	serverInvoked := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverInvoked = true
		assert.Equal(r.Method, "POST", "submit method")
		assert.Equal(r.URL.Path, "/v1/submit", "submit path")
		assert.Equal(r.Header.Get("Content-Type"), "application/json", "submit content type")

		//authentication header
		assert.Equal(r.Header.Get("X-Canary-Api-Key"), "my api key", "submit api key")

		//server name in body
		expectedBody, _ := json.Marshal(map[string]string{
			"server": "my server",
			"app":    "some app",
			"deps":   "\"foo bar baz\"",
		})
		body, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		assert.Equal(body, expectedBody, "submit body")
		respond(w, 200, "true")
	}))
	defer ts.Close()

	//overwrite the base URL to our testing server
	baseURL = ts.URL

	client := NewClient("my api key", "my server")
	err := client.Submit("some app", "foo bar baz")
	assert.NoError(err)

	assert.True(serverInvoked, "server invoked")
}

//Sends an http.ResponseWriter a string and status
func respond(w http.ResponseWriter, status int, v string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(v))
}
*/
