package agent

var baseURL = "http://localhost:8080"

type Client interface {
	HeartBeat() bool
	CheckServer(string) bool
	RegisterServer(string)
	RegisterApp(string)
	Submit(string, interface{})
}

type CanaryClient struct {
	apiKey     string
	serverName string
}

func NewClient(apiKey string, serverName string) *CanaryClient {
	client := &CanaryClient{apiKey: apiKey, serverName: serverName}
	return client
}

func (c *CanaryClient) HeartBeat() bool {
	//TODO
	return true
}

func (c *CanaryClient) CheckServer(name string) bool {
	//TODO
	return true
}

func (c *CanaryClient) RegisterServer(name string) {
	//TODO
}

func (c *CanaryClient) RegisterApp(name string) {
	//TODO
}

func (c *CanaryClient) Submit(app string, data interface{}) {
	//TODO
}
