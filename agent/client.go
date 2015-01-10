package agent

var baseURL = "http://localhost:8080"

type CanaryClient interface {
	HeartBeat() bool
	CheckServer(string) bool
	RegisterServer(string)
	RegisterApp(string)
	Submit(string, interface{})
}

type Client struct {
	apiKey     string
	serverName string
}

func NewClient(apiKey string, serverName string) *Client {
	client := &Client{apiKey: apiKey, serverName: serverName}
	return client
}

func (c *Client) HeartBeat() bool {
	//TODO
	return true
}

func (c *Client) CheckServer(name string) bool {
	//TODO
	return true
}

func (c *Client) RegisterServer(name string) {
	//TODO
}

func (c *Client) RegisterApp(name string) {
	//TODO
}

func (c *Client) Submit(app string, data interface{}) {
	//TODO
}
